package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"aiops/internal/shield"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ShieldRuleController 告警屏蔽规则控制器
type ShieldRuleController struct {
	DB     *gorm.DB
	Shield *shield.Service
}

// NewShieldRuleController 创建控制器
func NewShieldRuleController(db *gorm.DB, shieldSvc *shield.Service) *ShieldRuleController {
	return &ShieldRuleController{DB: db, Shield: shieldSvc}
}

// List 屏蔽规则列表
// GET /api/shield-rules?q=&include_expired=
func (c *ShieldRuleController) List(ctx *gin.Context) {
	db := c.DB.Model(&models.AlertShieldRule{})
	if q := strings.TrimSpace(ctx.Query("q")); q != "" {
		like := "%" + q + "%"
		db = db.Where("name LIKE ? OR remark LIKE ?", like, like)
	}
	if ctx.Query("include_expired") != "1" {
		db = db.Where("end_at >= ?", time.Now().Add(-24*time.Hour))
	}
	var rules []models.AlertShieldRule
	if err := db.Order("created_at DESC").Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询屏蔽规则失败", "error": err.Error()})
		return
	}
	if rules == nil {
		rules = []models.AlertShieldRule{}
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rules})
}

type shieldRulePayload struct {
	Name      string `json:"name"`
	Tags      string `json:"tags"` // JSON 数组字符串，如 ["instance=wms_prd","type=应用"]
	StartAt   string `json:"start_at"`
	EndAt     string `json:"end_at"`
	IsEnabled int    `json:"is_enabled"`
	Remark    string `json:"remark"`
}

// validateAndParse 校验并解析屏蔽规则入参
func (p *shieldRulePayload) validateAndParse() (startAt, endAt time.Time, err error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return startAt, endAt, errors.New("规则名称不能为空")
	}
	p.Tags = strings.TrimSpace(p.Tags)
	if p.Tags == "" || p.Tags == "[]" {
		return startAt, endAt, errors.New("事件标签不能为空")
	}
	var arr []string
	if err := json.Unmarshal([]byte(p.Tags), &arr); err != nil || len(arr) == 0 {
		return startAt, endAt, errors.New("事件标签格式错误（应为 [\"k=v\",...] 的 JSON 数组）")
	}
	for _, t := range arr {
		if !strings.Contains(t, "=") {
			return startAt, endAt, errors.New("事件标签格式错误，每项应为 k=v 形式")
		}
	}
	startAt, err = time.ParseInLocation("2006-01-02T15:04:05", p.StartAt, time.Local)
	if err != nil {
		return startAt, endAt, errors.New("屏蔽开始时间格式错误")
	}
	endAt, err = time.ParseInLocation("2006-01-02T15:04:05", p.EndAt, time.Local)
	if err != nil {
		return startAt, endAt, errors.New("屏蔽结束时间格式错误")
	}
	if !endAt.After(startAt) {
		return startAt, endAt, errors.New("屏蔽结束时间必须晚于开始时间")
	}
	return startAt, endAt, nil
}

// Create 新建屏蔽规则
func (c *ShieldRuleController) Create(ctx *gin.Context) {
	var payload shieldRulePayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	startAt, endAt, err := payload.validateAndParse()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if payload.IsEnabled != 0 {
		payload.IsEnabled = 1
	}
	rule := models.AlertShieldRule{
		Name:      payload.Name,
		Tags:      payload.Tags,
		StartAt:   startAt,
		EndAt:     endAt,
		IsEnabled: payload.IsEnabled,
		Remark:    strings.TrimSpace(payload.Remark),
		CreatedBy: currentOperator(ctx),
	}
	if err := c.DB.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建屏蔽规则失败", "error": err.Error()})
		return
	}
	c.Shield.Reload()
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rule})
}

// Update 更新屏蔽规则
func (c *ShieldRuleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var rule models.AlertShieldRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "屏蔽规则不存在"})
		return
	}
	var payload shieldRulePayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	startAt, endAt, err := payload.validateAndParse()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if payload.IsEnabled != 0 {
		payload.IsEnabled = 1
	}
	updates := map[string]interface{}{
		"name":       payload.Name,
		"tags":       payload.Tags,
		"start_at":   startAt,
		"end_at":     endAt,
		"is_enabled": payload.IsEnabled,
		"remark":     strings.TrimSpace(payload.Remark),
	}
	if err := c.DB.Model(&rule).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新屏蔽规则失败", "error": err.Error()})
		return
	}
	c.Shield.Reload()
	c.DB.First(&rule, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rule})
}

// Toggle 启用/停用屏蔽规则
func (c *ShieldRuleController) Toggle(ctx *gin.Context) {
	id := ctx.Param("id")
	var rule models.AlertShieldRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "屏蔽规则不存在"})
		return
	}
	enabled := 0
	if rule.IsEnabled != 1 {
		enabled = 1
	}
	if err := c.DB.Model(&rule).Update("is_enabled", enabled).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "切换状态失败", "error": err.Error()})
		return
	}
	c.Shield.Reload()
	c.DB.First(&rule, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rule})
}

// Delete 删除屏蔽规则（软删除）
func (c *ShieldRuleController) Delete(ctx *gin.Context) {
	var rule models.AlertShieldRule
	if err := c.DB.First(&rule, "id = ?", ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "屏蔽规则不存在"})
		return
	}
	if err := c.DB.Delete(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除屏蔽规则失败", "error": err.Error()})
		return
	}
	c.Shield.Reload()
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
