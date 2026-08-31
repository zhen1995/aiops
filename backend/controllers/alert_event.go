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
