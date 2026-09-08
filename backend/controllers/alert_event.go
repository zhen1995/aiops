package controllers

import (
	"net/http"
	"strconv"
	"time"

	"aiops/internal/alerting"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// alertEngine 告警评估引擎实例（由 main.go 注入，用于事件删除后联动清理序列状态）
var alertEngine *alerting.Engine

// SetAlertingEngine 注入告警评估引擎实例
func SetAlertingEngine(e *alerting.Engine) {
	alertEngine = e
}

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

// Delete 删除单条告警事件；若删除的是未恢复的告警事件，同步清理引擎中对应序列状态
func (c *AlertEventController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var ev models.AlertEvent
	if err := c.DB.First(&ev, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "告警事件不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}

	if err := c.DB.Delete(&models.AlertEvent{}, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除告警事件失败", "error": err.Error()})
		return
	}

	// 删除未恢复事件后通知引擎清理序列状态，条件仍满足时可按持续时间重新触发
	if ev.Type == models.AlertEventTypeAlert && ev.Status == models.AlertEventStatusFiring && alertEngine != nil {
		alertEngine.ClearSeries(ev.RuleID, ev.Tags)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id}})
}

// BatchDelete 批量删除告警事件
// 请求体: {"ids": ["uuid1", "uuid2", ...]}
func (c *AlertEventController) BatchDelete(ctx *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(body.IDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请至少选择一条告警事件"})
		return
	}
	if len(body.IDs) > 500 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "单次最多删除 500 条"})
		return
	}

	// 查出所有待删事件，用来清理引擎序列状态
	var evs []models.AlertEvent
	if err := c.DB.Where("id IN ?", body.IDs).Find(&evs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}

	// 批量删除
	if err := c.DB.Delete(&models.AlertEvent{}, "id IN ?", body.IDs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "批量删除失败", "error": err.Error()})
		return
	}

	// 逐条清理引擎序列状态
	if alertEngine != nil {
		for _, ev := range evs {
			if ev.Type == models.AlertEventTypeAlert && ev.Status == models.AlertEventStatusFiring {
				alertEngine.ClearSeries(ev.RuleID, ev.Tags)
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"deleted": len(evs)}})
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
