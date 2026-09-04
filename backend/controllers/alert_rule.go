package controllers

import (
	"errors"
	"net/http"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AlertRuleController 告警规则控制器（系统自建，不再代理 Nightingale）
type AlertRuleController struct {
	DB *gorm.DB
}

// NewAlertRuleController 创建控制器
func NewAlertRuleController(db *gorm.DB) *AlertRuleController {
	return &AlertRuleController{DB: db}
}

// List 获取所有告警规则列表
func (c *AlertRuleController) List(ctx *gin.Context) {
	var list []models.AlertRule
	if err := c.DB.Order("created_at desc").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": list,
	})
}

// Get 根据 ID 获取单个告警规则
func (c *AlertRuleController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}

// validateRule 校验规则必填字段
func validateRule(rule *models.AlertRule) error {
	if rule.Name == "" {
		return errors.New("规则名称不能为空")
	}
	if rule.PromQL == "" {
		return errors.New("PromQL 不能为空")
	}
	if rule.Severity < 1 || rule.Severity > 3 {
		return errors.New("告警级别必须是 1(P1紧急)、2(P2警告) 或 3(P3提醒)")
	}
	if rule.Duration < 0 {
		return errors.New("持续时间不能为负数")
	}
	return nil
}

// Create 创建告警规则
func (c *AlertRuleController) Create(ctx *gin.Context) {
	var rule models.AlertRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := validateRule(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	if err := c.DB.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}

// Update 更新告警规则
func (c *AlertRuleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	var updateData models.AlertRule
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if err := validateRule(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	updates := map[string]interface{}{
		"name":       updateData.Name,
		"prom_ql":    updateData.PromQL,
		"duration":   updateData.Duration,
		"severity":   updateData.Severity,
		"is_enabled": updateData.IsEnabled,
	}

	if err := c.DB.Model(&rule).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}

// Delete 删除告警规则（软删除）
func (c *AlertRuleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	if err := c.DB.Delete(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ToggleEnabled 切换启用状态
func (c *AlertRuleController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")

	var rule models.AlertRule
	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "告警规则不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	newStatus := 0
	if rule.IsEnabled == 0 {
		newStatus = 1
	}

	if err := c.DB.Model(&rule).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "操作失败",
			"error":   err.Error(),
		})
		return
	}

	if err := c.DB.First(&rule, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}
