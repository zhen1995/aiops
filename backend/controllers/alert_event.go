package controllers

import (
	"net/http"
	"strconv"

	"aiops/internal/nightingale"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	req := nightingale.EventListRequest{
		Hours: parseInt64(ctx.DefaultQuery("hours", "24")),
		Page:  parseInt(ctx.DefaultQuery("page", "1")),
		Limit: parseInt(ctx.DefaultQuery("limit", "20")),
		Query: ctx.Query("query"),
	}
	if s := ctx.Query("severity"); s != "" {
		req.Severity = parseInt(s)
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

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
