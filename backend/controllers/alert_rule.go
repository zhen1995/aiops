package controllers

import (
	"errors"
	"net/http"

	"aiops/internal/nightingale"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlertRuleController struct {
	DB *gorm.DB
}

func NewAlertRuleController(db *gorm.DB) *AlertRuleController {
	return &AlertRuleController{DB: db}
}

func (c *AlertRuleController) List(ctx *gin.Context) {
	cfg, err := loadEnabledConfig(c.DB)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	cli := nightingale.NewClient(cfg)
	gids := ctx.Query("gids")
	if gids == "" {
		gids = cfg.Gids
	}
	rules, err := cli.ListRules(ctx, gids)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询告警规则失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": rules})
}

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
