package controllers

// 业务分组父子层级接口测试：创建子分组、仅改名时 parent_id 保持不变、防成环校验、
// 删除分组时子分组上移一级且规则 group_id 置空。使用内存 sqlite，不依赖外部 MySQL。

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupVerifyRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.BusinessGroup{}, &models.AlertRule{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	c := NewBusinessGroupController(db)
	r := gin.New()
	r.POST("/business-groups", c.Create)
	r.PUT("/business-groups/:id", c.Update)
	r.DELETE("/business-groups/:id", c.Delete)
	r.GET("/business-groups", c.List)
	return r, db
}

func doReq(t *testing.T, r *gin.Engine, method, path string, body any) (int, map[string]any) {
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func TestBusinessGroupTreeVerify(t *testing.T) {
	r, db := setupVerifyRouter(t)

	// 1. 创建顶级分组
	code, resp := doReq(t, r, "POST", "/business-groups", map[string]any{"name": "验证-父"})
	if code != http.StatusOK {
		t.Fatalf("create parent: code=%d resp=%v", code, resp)
	}
	parentID := resp["data"].(map[string]any)["id"].(string)

	// 2. 创建子分组
	code, resp = doReq(t, r, "POST", "/business-groups", map[string]any{"name": "验证-子", "parent_id": parentID})
	if code != http.StatusOK {
		t.Fatalf("create child: code=%d resp=%v", code, resp)
	}
	child := resp["data"].(map[string]any)
	childID := child["id"].(string)
	if child["parent_id"] != parentID {
		t.Fatalf("child parent_id=%v, want %v", child["parent_id"], parentID)
	}

	// 3. 仅传 name 更新（不传 parent_id），parent_id 应保持不变
	code, resp = doReq(t, r, "PUT", "/business-groups/"+childID, map[string]any{"name": "验证-子2"})
	if code != http.StatusOK {
		t.Fatalf("rename child: code=%d resp=%v", code, resp)
	}
	if got := resp["data"].(map[string]any)["parent_id"]; got != parentID {
		t.Fatalf("after rename parent_id=%v, want unchanged %v", got, parentID)
	}

	// 4. 防成环：子分组父级指向自己 / 指向自己的后代（先建孙子分组）
	code, resp = doReq(t, r, "POST", "/business-groups", map[string]any{"name": "验证-孙", "parent_id": childID})
	if code != http.StatusOK {
		t.Fatalf("create grandchild: %v", resp)
	}
	grandID := resp["data"].(map[string]any)["id"].(string)
	code, _ = doReq(t, r, "PUT", "/business-groups/"+childID, map[string]any{"name": "验证-子2", "parent_id": childID})
	if code != http.StatusBadRequest {
		t.Fatalf("self parent: code=%d, want 400", code)
	}
	code, _ = doReq(t, r, "PUT", "/business-groups/"+parentID, map[string]any{"name": "验证-父", "parent_id": grandID})
	if code != http.StatusBadRequest {
		t.Fatalf("descendant parent: code=%d, want 400", code)
	}

	// 5. 父分组下挂一条规则，删除子分组：规则 group_id 置空、孙分组上移到父分组
	rule := models.AlertRule{ID: "rule-1", Name: "r", GroupID: childID, Severity: 2, EvalInterval: 30, Duration: 0, IsEnabled: 1}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("create rule: %v", err)
	}
	code, _ = doReq(t, r, "DELETE", "/business-groups/"+childID, nil)
	if code != http.StatusOK {
		t.Fatalf("delete child: code=%d", code)
	}
	var g models.BusinessGroup
	if err := db.First(&g, "id = ?", grandID).Error; err != nil {
		t.Fatalf("grandchild gone: %v", err)
	}
	if g.ParentID != parentID {
		t.Fatalf("grandchild parent=%q, want %q (moved up)", g.ParentID, parentID)
	}
	var r2 models.AlertRule
	db.First(&r2, "id = ?", "rule-1")
	if r2.GroupID != "" {
		t.Fatalf("rule group_id=%q, want empty", r2.GroupID)
	}

	// 6. 列表包含 parent_id 与 rule_count
	code, resp = doReq(t, r, "GET", "/business-groups", nil)
	if code != http.StatusOK {
		t.Fatalf("list: code=%d", code)
	}
	list := resp["data"].([]any)
	if len(list) != 2 { // 父 + 孙（子已软删）
		t.Fatalf("list len=%d, want 2", len(list))
	}
	found := false
	for _, item := range list {
		m := item.(map[string]any)
		if m["id"] == grandID {
			found = true
			if m["parent_id"] != parentID {
				t.Fatalf("list grandchild parent_id=%v", m["parent_id"])
			}
		}
	}
	if !found {
		t.Fatal("grandchild not in list")
	}

	t.Log("全部业务分组树校验通过")
}
