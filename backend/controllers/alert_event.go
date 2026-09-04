package controllers

import (
	"net/http"
	"strconv"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AlertEventController 告警事件控制器（查询系统按告警规则自动生成的本地事件）
type AlertEventController struct {
	DB *gorm.DB
}

func NewAlertEventController(db *gorm.DB) *AlertEventController {
	return &AlertEventController{DB: db}
}

// List 查询告警事件
// scope=active 返回未恢复（告警中）事件；scope=history 返回已解决/已恢复事件（含告警恢复事件）
func (c *AlertEventController) List(ctx *gin.Context) {
	hours, err := parseInt64(ctx.DefaultQuery("hours", "24"))
	if err != nil || hours <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "hours 参数必须是正整数"})
		return
	}
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

	scope := ctx.DefaultQuery("scope", "active")
	if scope != "active" && scope != "history" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "scope 参数必须是 active 或 history"})
		return
	}

	db := c.DB.Model(&models.AlertEvent{})
	if scope == "active" {
		db = db.Where("status = ?", models.AlertEventStatusFiring)
	} else {
		db = db.Where("status = ?", models.AlertEventStatusResolved)
	}
	db = db.Where("trigger_time >= ?", time.Now().Add(-time.Duration(hours)*time.Hour))

	if s := ctx.Query("severity"); s != "" {
		sev, err := parseInt(s)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "severity 参数必须是整数"})
			return
		}
		db = db.Where("severity = ?", sev)
	}
	if q := ctx.Query("query"); q != "" {
		like := "%" + q + "%"
		db = db.Where("rule_name LIKE ? OR target_ident LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}

	var list []models.AlertEvent
	if err := db.Order("trigger_time DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list, "total": total}})
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
