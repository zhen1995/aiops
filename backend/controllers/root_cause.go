package controllers

import (
	"net/http"

	"aiops/internal/rca"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RootCauseController 根因分析控制器（针对告警事件触发 AI 根因分析并查询结果）
type RootCauseController struct {
	DB *gorm.DB
}

func NewRootCauseController(db *gorm.DB) *RootCauseController {
	return &RootCauseController{DB: db}
}

// Trigger 对指定告警事件触发根因分析
// POST /api/alert-events/:id/root-cause
func (c *RootCauseController) Trigger(ctx *gin.Context) {
	eventID := ctx.Param("id")
	if eventID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少告警事件ID"})
		return
	}

	// 校验告警事件存在
	var event models.AlertEvent
	if err := c.DB.Where("id = ?", eventID).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "告警事件不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}

	// 异步启动根因分析，返回分析任务ID
	analysisID, err := rca.StartAnalysis(c.DB, eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "启动根因分析失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": analysisID, "status": models.RootCauseStatusRunning}})
}

// List 查询根因分析列表（可按告警事件过滤，trigger_time DESC 分页）
// GET /api/root-cause-analyses?alert_event_id=&page=&limit=
func (c *RootCauseController) List(ctx *gin.Context) {
	page, err := parseInt(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "page 参数必须是正整数"})
		return
	}
	limit, err := parseInt(ctx.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "limit 参数必须是正整数"})
		return
	}

	db := c.DB.Model(&models.RootCauseAnalysis{})
	if eventID := ctx.Query("alert_event_id"); eventID != "" {
		db = db.Where("alert_event_id = ?", eventID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询根因分析列表失败", "error": err.Error()})
		return
	}

	var list []models.RootCauseAnalysis
	if err := db.Order("trigger_time DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询根因分析列表失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list, "total": total}})
}

// Detail 查询根因分析详情
// GET /api/root-cause-analyses/:id
func (c *RootCauseController) Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少分析ID"})
		return
	}

	var analysis models.RootCauseAnalysis
	if err := c.DB.Where("id = ?", id).First(&analysis).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "根因分析不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询根因分析详情失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": analysis})
}
