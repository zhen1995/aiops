package controllers

import (
	"net/http"
	"strings"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotifyRecordController 通知记录控制器
type NotifyRecordController struct {
	DB *gorm.DB
}

// NewNotifyRecordController 创建控制器
func NewNotifyRecordController(db *gorm.DB) *NotifyRecordController {
	return &NotifyRecordController{DB: db}
}

// List 分页查询通知记录
// GET /api/notify-records?status=&event_type=&strategy=&keyword=&start=&end=&page=&page_size=
func (c *NotifyRecordController) List(ctx *gin.Context) {
	db := c.DB.Model(&models.NotifyRecord{})

	if status := strings.TrimSpace(ctx.Query("status")); status != "" {
		db = db.Where("status = ?", status)
	}
	if eventType := strings.TrimSpace(ctx.Query("event_type")); eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	// 降噪策略过滤（用于查询被指定策略拦截的通知记录）
	if strategy := strings.TrimSpace(ctx.Query("strategy")); strategy != "" {
		db = db.Where("strategy = ?", strategy)
	}
	if kw := strings.TrimSpace(ctx.Query("keyword")); kw != "" {
		// 参数化 LIKE 并转义 %/_，防止通配符注入
		like := "%" + escapeLikeKeyword(kw) + "%"
		esc := " ESCAPE '!'"
		db = db.Where("(rule_name LIKE ?"+esc+" OR target_ident LIKE ?"+esc+")", like, like)
	}
	if start := strings.TrimSpace(ctx.Query("start")); start != "" {
		if t, ok := parseNotifyRecordTime(start); ok {
			db = db.Where("created_at >= ?", t)
		}
	}
	if end := strings.TrimSpace(ctx.Query("end")); end != "" {
		if t, ok := parseNotifyRecordTime(end); ok {
			db = db.Where("created_at <= ?", t)
		}
	}

	page, err := parseInt(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := parseInt(ctx.DefaultQuery("page_size", "20"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询通知记录总数失败", "error": err.Error()})
		return
	}

	list := make([]models.NotifyRecord, 0)
	if err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询通知记录失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"total": total, "list": list}})
}

// parseNotifyRecordTime 解析时间筛选参数：兼容 RFC3339 与 datetime-local（无时区/秒，按本地时区）
func parseNotifyRecordTime(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", s, time.Local); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// Detail 查询单条通知记录详情
// GET /api/notify-records/:id
func (c *NotifyRecordController) Detail(ctx *gin.Context) {
	var rec models.NotifyRecord
	if err := c.DB.Where("id = ?", ctx.Param("id")).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知记录不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询通知记录失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rec})
}
