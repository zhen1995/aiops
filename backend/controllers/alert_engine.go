package controllers

import (
	"errors"
	"net/http"

	"aiops/internal/nightingale"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlertEngineController struct {
	DB *gorm.DB
}

func NewAlertEngineController(db *gorm.DB) *AlertEngineController {
	return &AlertEngineController{DB: db}
}

func (c *AlertEngineController) List(ctx *gin.Context) {
	var list []models.AlertEngineConfig
	if err := c.DB.Order("created_at DESC").Find(&list).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (c *AlertEngineController) Get(ctx *gin.Context) {
	var cfg models.AlertEngineConfig
	if err := c.DB.First(&cfg, "id = ?", ctx.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (c *AlertEngineController) Create(ctx *gin.Context) {
	var cfg models.AlertEngineConfig
	if err := ctx.ShouldBindJSON(&cfg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if err := c.DB.Create(&cfg).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (c *AlertEngineController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var cfg models.AlertEngineConfig
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	var req models.AlertEngineConfig
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	updates := map[string]interface{}{
		"name":       req.Name,
		"base_url":   req.BaseURL,
		"token":      req.Token,
		"gids":       req.Gids,
		"remark":     req.Remark,
		"is_enabled": req.IsEnabled,
		"is_default": req.IsDefault,
	}
	if err := c.DB.Model(&cfg).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败", "error": err.Error()})
		return
	}
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (c *AlertEngineController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	var cfg models.AlertEngineConfig
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	if err := c.DB.Delete(&cfg).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

func (c *AlertEngineController) ToggleEnabled(ctx *gin.Context) {
	id := ctx.Param("id")
	var cfg models.AlertEngineConfig
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	next := 0
	if cfg.IsEnabled == 0 {
		next = 1
	}
	if err := c.DB.Model(&cfg).Update("is_enabled", next).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败", "error": err.Error()})
		return
	}
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (c *AlertEngineController) Test(ctx *gin.Context) {
	id := ctx.Param("id")
	var cfg models.AlertEngineConfig
	if err := c.DB.First(&cfg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	cli := nightingale.NewClient(&cfg)
	if err := cli.TestConnection(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "message": "连接失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "连接成功"})
}
