package controllers

import (
	"net/http"
	"time"

	"aiops/internal/denoise"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DenoiseController 告警降噪控制器
type DenoiseController struct {
	DB      *gorm.DB
	denoise *denoise.DenoiseService
}

// NewDenoiseController 创建控制器
func NewDenoiseController(db *gorm.DB, svc *denoise.DenoiseService) *DenoiseController {
	return &DenoiseController{DB: db, denoise: svc}
}

// policyView 降噪策略出参
type policyView struct {
	Strategy    string `json:"strategy"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Config        string `json:"config"`
	SuppressedToday int64 `json:"suppressed_today"` // 今日拦截数
}

// ListPolicies 返回两条降噪策略列表（含今日拦截数）
// GET /api/denoise/policies
func (c *DenoiseController) ListPolicies(ctx *gin.Context) {
	stats, err := c.denoise.TodayStats(time.Now())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询降噪统计失败", "error": err.Error()})
		return
	}

	strategies := []string{
		models.DenoiseStrategyWindowAggregation,
		models.DenoiseStrategyTopologySuppression,
	}
	list := make([]policyView, 0, len(strategies))
	for _, strategy := range strategies {
		var p models.DenoisePolicy
		if err := c.DB.Where("strategy = ?", strategy).First(&p).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询降噪策略失败", "error": err.Error()})
			return
		}
		list = append(list, policyView{
			Strategy:        p.Strategy,
			Name:            p.Name,
			Description:     p.Description,
			Enabled:         p.Enabled,
			Config:          p.Config,
			SuppressedToday: stats.Suppressed[strategy],
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"policies": list}})
}

// updatePolicyForm 更新策略入参（仅允许修改 enabled，config 不接受前端修改）
type updatePolicyForm struct {
	Enabled *bool `json:"enabled"`
}

// UpdatePolicy 更新指定策略的启用状态
// PUT /api/denoise/policies/:strategy
func (c *DenoiseController) UpdatePolicy(ctx *gin.Context) {
	strategy := ctx.Param("strategy")
	if strategy != models.DenoiseStrategyWindowAggregation && strategy != models.DenoiseStrategyTopologySuppression {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未知的降噪策略类型"})
		return
	}

	var form updatePolicyForm
	if err := ctx.ShouldBindJSON(&form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if form.Enabled == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "enabled 字段不能为空"})
		return
	}

	var p models.DenoisePolicy
	if err := c.DB.Where("strategy = ?", strategy).First(&p).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询降噪策略失败", "error": err.Error()})
		return
	}
	if err := c.DB.Model(&p).Update("enabled", *form.Enabled).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新降噪策略失败", "error": err.Error()})
		return
	}
	p.Enabled = *form.Enabled
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": policyView{
		Strategy:    p.Strategy,
		Name:        p.Name,
		Description: p.Description,
		Enabled:     p.Enabled,
		Config:      p.Config,
	}})
}

// Stats 今日降噪 KPI 与漏斗
// GET /api/denoise/stats
func (c *DenoiseController) Stats(ctx *gin.Context) {
	stats, err := c.denoise.TodayStats(time.Now())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询降噪统计失败", "error": err.Error()})
		return
	}

	windowSuppressed := stats.Suppressed[models.DenoiseStrategyWindowAggregation]
	topoSuppressed := stats.Suppressed[models.DenoiseStrategyTopologySuppression]

	rawTotal := stats.AlertEvents + windowSuppressed + topoSuppressed
	afterWindow := rawTotal - windowSuppressed
	afterTopology := afterWindow - topoSuppressed
	suppressedTotal := windowSuppressed + topoSuppressed

	compressionRate := 0.0
	if rawTotal > 0 {
		compressionRate = float64(suppressedTotal) * 100 / float64(rawTotal)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"raw_total":        rawTotal,
		"effective":        stats.AlertEvents,
		"compression_rate": compressionRate,
		"suppressed_total": suppressedTotal,
		"funnel": []gin.H{
			{"name": "原始告警", "value": rawTotal},
			{"name": "窗口聚合后", "value": afterWindow},
			{"name": "拓扑抑制后", "value": afterTopology},
			{"name": "有效通知", "value": stats.AlertEvents},
		},
		"details": gin.H{
			"window_aggregation_suppressed":   windowSuppressed,
			"topology_suppression_suppressed": topoSuppressed,
			"alert_events":                    stats.AlertEvents,
		},
	}})
}
