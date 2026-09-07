package knowledge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiops/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 用内存 SQLite 代替 MySQL（结构一致，仅用于单测）
// 每个测试使用独立的内存库名，避免 cache=shared 串数据
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.KBDocument{}, &models.KBChunk{}, &models.LLMConfig{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSaveUpload(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()
	svc := NewService(db, NewClient("http://unused"), dir)

	doc, err := svc.SaveUpload(context.Background(), "手册.md", 12, strings.NewReader("# 内容内容内容"), "root")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Status != "pending" || doc.Type != "md" {
		t.Errorf("doc = %+v", doc)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); err != nil {
		t.Errorf("文件未落盘: %v", err)
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	dir := t.TempDir()

	var deletedVec bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletedVec = true
		}
		w.Write([]byte(`{"document_id":"x","chunks":1}`))
	}))
	defer srv.Close()

	svc := NewService(db, NewClient(srv.URL), dir)
	doc, _ := svc.SaveUpload(context.Background(), "a.md", 2, strings.NewReader("hi"), "root")
	svc.IndexDocument(doc.ID) // 假服务返回 chunks，走完流程

	if err := svc.Delete(context.Background(), doc.ID); err != nil {
		t.Fatal(err)
	}
	if !deletedVec {
		t.Error("未调用向量删除")
	}
	var count int64
	db.Model(&models.KBDocument{}).Count(&count)
	if count != 0 {
		t.Errorf("文档未删除, count=%d", count)
	}
	if _, err := os.Stat(filepath.Join(dir, doc.ID+".md")); !os.IsNotExist(err) {
		t.Error("文件未删除")
	}
}
