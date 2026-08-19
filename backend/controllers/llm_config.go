package controllers

import (
	"errors"
	"net/http"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LLMConfigController 大模型配置控制器
type LLMConfigController struct {
	DB *gorm.DB
}

// NewLLMConfigController 创建控制器
func NewLLMConfigController(db *gorm.DB) *LLMConfigController {
	return &LLMConfigController{DB: db}
}

// List 获取所有 LLM 配置列表
func (c *LLMConfigController) List(ctx *gin.Context) {
	var configs []models.LLMConfig
	if err := c.DB.Find(&configs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": configs,
	})
}

// Get 根据 ID 获取单个 LLM 配置
func (c *LLMConfigController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	var config models.LLMConfig
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "配置不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
	})
}

// Create 创建 LLM 配置
func (c *LLMConfigController) Create(ctx *gin.Context) {
	var config models.LLMConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error": err.Error(),
		})
		return
	}

	// 开启事务
	tx := c.DB.Begin()

	// 如果设置为默认，需要先取消其他默认项
	if config.IsDefault == 1 {
		if err := tx.Model(&models.LLMConfig{}).Where("is_default = ?", 1).Update("is_default", 0).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"message": "创建失败",
			})
			return
		}
	}

	if err := tx.Create(&config).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "创建失败",
			"error": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "提交失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
	})
}

// Update 更新 LLM 配置
func (c *LLMConfigController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var config models.LLMConfig
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "配置不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	var updateData models.LLMConfig
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "参数错误",
			"error": err.Error(),
		})
		return
	}

	// 开启事务
	tx := c.DB.Begin()

	// 如果设置为默认，需要先取消其他默认项
	if updateData.IsDefault == 1 {
		if err := tx.Model(&models.LLMConfig{}).Where("id != ? AND is_default = ?", id, 1).Update("is_default", 0).Error; err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"message": "更新失败",
			})
			return
		}
	}

	// 更新字段
	updates := map[string]interface{}{
		"name":              updateData.Name,
		"description":       updateData.Description,
		"supplier_category": updateData.SupplierCategory,
		"model":             updateData.Model,
		"base_url":          updateData.BaseURL,
		"api_key":           updateData.APIKey,
		"is_default":        updateData.IsDefault,
		"is_enabled":        updateData.IsEnabled,
	}

	if err := tx.Model(&config).Updates(updates).Error; err != nil {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "更新失败",
			"error": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "提交失败",
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
	})
}

// Delete 删除 LLM 配置（软删除）
func (c *LLMConfigController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	var config models.LLMConfig
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "配置不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	if err := c.DB.Delete(&config).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "删除失败",
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "删除成功",
	})
}

// ToggleEnabled 切换启用状态
func (c *LLMConfigController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")

	var config models.LLMConfig
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"code": 404,
				"message": "配置不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	// 切换状态
	newStatus := 0
	if config.IsEnabled == 0 {
		newStatus = 1
	}

	if err := c.DB.Model(&config).Update("is_enabled", newStatus).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "操作失败",
			"error": err.Error(),
		})
		return
	}

	// 重新查询返回最新数据
	if err := c.DB.First(&config, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"message": "查询失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
	})
}

// SupplierCategoryList 获取供应商类型列表
func (c *LLMConfigController) SupplierCategoryList(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": models.SupplierCategoryOptions,
	})
}