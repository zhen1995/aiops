package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"aiops/internal/n9e"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// N9eController 夜莺告警相关
type N9eController struct {
	db *gorm.DB
}

func NewN9eController(db *gorm.DB) *N9eController {
	return &N9eController{db: db}
}

// ---------- 引擎配置 ----------

func (c *N9eController) ListConfig(ctx *gin.Context) {
	var list []models.N9eConfig
	if err := c.db.Order("created_at desc").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	for i := range list {
		list[i].Token = maskToken(list[i].Token)
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (c *N9eController) SaveConfig(ctx *gin.Context) {
	var cfg models.N9eConfig
	if err := ctx.ShouldBindJSON(&cfg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if cfg.Address == "" || cfg.Token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址和 Token 必填"})
		return
	}

	if cfg.IsEnabled == 1 {
		c.db.Model(&models.N9eConfig{}).Update("is_enabled", 0)
	}

	if cfg.ID == "" {
		if err := c.db.Create(&cfg).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存失败: " + err.Error()})
			return
		}
	} else {
		updates := map[string]interface{}{
			"name":       cfg.Name,
			"address":    cfg.Address,
			"is_enabled": cfg.IsEnabled,
			"updated_at": time.Now(),
		}
		if cfg.Token != "" {
			updates["token"] = cfg.Token
		}
		if err := c.db.Model(&models.N9eConfig{}).Where("id = ?", cfg.ID).Updates(updates).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
			return
		}
	}
	cfg.Token = maskToken(cfg.Token)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (c *N9eController) DeleteConfig(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.db.Delete(&models.N9eConfig{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0})
}

func (c *N9eController) TestConfig(ctx *gin.Context) {
	var cfg models.N9eConfig
	if err := ctx.ShouldBindJSON(&cfg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if cfg.Address == "" || cfg.Token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "地址和 Token 必填"})
		return
	}
	nc := n9e.NewClient(cfg.Address, cfg.Token)
	body, code, err := nc.GetAlertRules(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": gin.H{"ok": false, "message": err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"ok":     code >= 200 && code < 300,
		"status": code,
		"raw":    truncateStr(string(body), 500),
	}})
}

// ---------- 代理到夜莺 ----------

func (c *N9eController) GetBusiGroups(ctx *gin.Context) {
	nc, err := n9e.NewClientFromDB(c.db)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": err.Error()})
		return
	}
	body, code, err := nc.GetBusiGroups(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": "夜莺请求失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "http_code": code, "data": parseN9eBody(body)})
}

func (c *N9eController) GetAlertRules(ctx *gin.Context) {
	nc, err := n9e.NewClientFromDB(c.db)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": err.Error()})
		return
	}
	body, code, err := nc.GetAlertRules(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": "夜莺请求失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "http_code": code, "data": parseN9eBody(body)})
}

func (c *N9eController) GetCurEvents(ctx *gin.Context) {
	nc, err := n9e.NewClientFromDB(c.db)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": err.Error()})
		return
	}
	q := ctx.Request.URL.Query()
	p := 1
	if v := q.Get("p"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			p = n
		}
	}
	limit := 30
	if v := q.Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			limit = n
		}
	}
	myGroups := true
	if v := q.Get("my_groups"); v != "" && v != "true" {
		myGroups = false
	}
	body, code, err := nc.GetCurEvents(ctx.Request.Context(), n9e.CurEventsParams{
		P: p, Limit: limit, MyGroups: myGroups,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": "夜莺请求失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "http_code": code, "data": parseN9eBody(body)})
}

func (c *N9eController) GetHisEvents(ctx *gin.Context) {
	nc, err := n9e.NewClientFromDB(c.db)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": err.Error()})
		return
	}
	q := ctx.Request.URL.Query()
	p := 1
	if v := q.Get("p"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			p = n
		}
	}
	limit := 30
	if v := q.Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			limit = n
		}
	}
	var stime, etime int64
	if v := q.Get("stime"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			stime = n
		}
	}
	if v := q.Get("etime"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			etime = n
		}
	}
	body, code, err := nc.GetHisEvents(ctx.Request.Context(), n9e.HisEventsParams{
		P: p, Limit: limit, Stime: stime, Etime: etime,
		RuleIDs: q.Get("rule_ids"), BusiGroupIDs: q.Get("busi_group_ids"),
		Cate: q.Get("cate"), Tags: q.Get("tags"),
	})
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "data": nil, "message": "夜莺请求失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "http_code": code, "data": parseN9eBody(body)})
}

// ---------- helpers ----------

func maskToken(t string) string {
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "****" + t[len(t)-4:]
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func parseN9eBody(body []byte) interface{} {
	if body == nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return string(body)
	}
	return v
}

// suppress unused import if any
var _ = fmt.Sprintf
