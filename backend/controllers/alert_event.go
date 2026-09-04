package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"aiops/internal/nightingale"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// loadEnabledConfig 查询当前启用的告警引擎配置（告警事件查询依赖）
func loadEnabledConfig(db *gorm.DB) (*models.AlertEngineConfig, error) {
	var cfg models.AlertEngineConfig
	if err := db.Where("is_enabled = ?", 1).Order("is_default DESC, created_at ASC").First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未找到启用的告警引擎配置，请先配置")
		}
		return nil, err
	}
	return &cfg, nil
}

type AlertEventController struct {
	DB *gorm.DB
}

func NewAlertEventController(db *gorm.DB) *AlertEventController {
	return &AlertEventController{DB: db}
}

func (c *AlertEventController) List(ctx *gin.Context) {
	cfg, err := loadEnabledConfig(c.DB)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	cli := nightingale.NewClient(cfg)

	hours, err := parseInt64(ctx.DefaultQuery("hours", "24"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "hours 参数必须是整数"})
		return
	}
	page, err := parseInt(ctx.DefaultQuery("page", "1"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "page 参数必须是整数"})
		return
	}
	limit, err := parseInt(ctx.DefaultQuery("limit", "20"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "limit 参数必须是整数"})
		return
	}

	req := nightingale.EventListRequest{
		Hours: hours,
		Page:  page,
		Limit: limit,
		Query: ctx.Query("query"),
	}
	if s := ctx.Query("severity"); s != "" {
		sev, err := parseInt(s)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "severity 参数必须是整数"})
			return
		}
		req.Severity = sev
	}

	scope := ctx.DefaultQuery("scope", "active")
	if scope != "active" && scope != "history" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "scope 参数必须是 active 或 history"})
		return
	}
	var result *nightingale.EventList
	if scope == "history" {
		result, err = cli.ListHistoryEvents(ctx, req)
	} else {
		result, err = cli.ListActiveEvents(ctx, req)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警事件失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
