package controllers

import (
	"errors"
	"net/http"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotifyRuleController 通知规则控制器
type NotifyRuleController struct {
	DB *gorm.DB
}

func NewNotifyRuleController(db *gorm.DB) *NotifyRuleController {
	return &NotifyRuleController{DB: db}
}

func (c *NotifyRuleController) List(ctx *gin.Context) {
	var list []models.NotifyRule
	if err := c.DB.Order("created_at desc").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (c *NotifyRuleController) Create(ctx *gin.Context) {
	var r models.NotifyRule
	if err := ctx.ShouldBindJSON(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if r.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "规则名称不能为空"})
		return
	}
	if err := c.DB.Create(&r).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": r})
}

func (c *NotifyRuleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var r models.NotifyRule
	if err := c.DB.First(&r, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "规则不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	var body models.NotifyRule
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	updates := map[string]interface{}{
		"name":           body.Name,
		"remark":         body.Remark,
		"media_id":       body.MediaID,
		"template_id":    body.TemplateID,
		"trigger_types":  body.TriggerTypes,
		"severity_filter": body.SeverityFilter,
		"is_enabled":     body.IsEnabled,
	}
	if body.Name == "" {
		delete(updates, "name")
	}
	if err := c.DB.Model(&r).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.DB.First(&r, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": r})
}

func (c *NotifyRuleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.DB.Delete(&models.NotifyRule{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0})
}

func (c *NotifyRuleController) Toggle(ctx *gin.Context) {
	id := ctx.Param("id")
	var r models.NotifyRule
	if err := c.DB.First(&r, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "规则不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	newStatus := 0
	if r.IsEnabled == 0 {
		newStatus = 1
	}
	if err := c.DB.Model(&r).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败"})
		return
	}
	c.DB.First(&r, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": r})
}
