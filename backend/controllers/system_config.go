package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SystemConfigController 系统配置控制器：管理知识库 Qdrant 向量库地址等全局配置
type SystemConfigController struct {
	DB *gorm.DB
}

// NewSystemConfigController 创建控制器
func NewSystemConfigController(db *gorm.DB) *SystemConfigController {
	return &SystemConfigController{DB: db}
}

// systemConfigPayload 请求/响应体：qdrant_url 为空字符串表示未传
type systemConfigPayload struct {
	QdrantURL string `json:"qdrant_url"`
}

// getQdrantURL 从 system_configs 读取 qdrant_url，无记录或读错误时回退默认值
func (c *SystemConfigController) getQdrantURL() string {
	var cfg models.SystemConfig
	if err := c.DB.Where("`key` = ?", models.SystemConfigKeyQdrantURL).First(&cfg).Error; err != nil {
		return models.DefaultQdrantURL
	}
	return cfg.Value
}

// Get 获取系统配置
// GET /api/system-config → {"qdrant_url": "..."}
func (c *SystemConfigController) Get(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"qdrant_url": c.getQdrantURL()}})
}

// Update 更新系统配置（upsert）
// PUT /api/system-config，body {"qdrant_url": "..."}
func (c *SystemConfigController) Update(ctx *gin.Context) {
	var payload systemConfigPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	url := strings.TrimSpace(payload.QdrantURL)
	if err := validateQdrantURL(url); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	var cfg models.SystemConfig
	err := c.DB.Where("`key` = ?", models.SystemConfigKeyQdrantURL).First(&cfg).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		cfg = models.SystemConfig{Key: models.SystemConfigKeyQdrantURL, Value: url}
		if err := c.DB.Create(&cfg).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存系统配置失败", "error": err.Error()})
			return
		}
	case err != nil:
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询系统配置失败", "error": err.Error()})
		return
	default:
		if err := c.DB.Model(&cfg).Update("value", url).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存系统配置失败", "error": err.Error()})
			return
		}
		cfg.Value = url
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"qdrant_url": cfg.Value}})
}

// TestQdrant 测试 Qdrant 向量库连通性（Go 直接 GET {url}/collections，3s 超时）
// POST /api/system-config/qdrant/test，body 可带 {"qdrant_url": "..."}，为空时用当前配置值
func (c *SystemConfigController) TestQdrant(ctx *gin.Context) {
	url := ""
	var payload systemConfigPayload
	if err := ctx.ShouldBindJSON(&payload); err == nil {
		url = strings.TrimSpace(payload.QdrantURL)
	}
	if url == "" {
		url = c.getQdrantURL()
	} else if err := validateQdrantURL(url); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(strings.TrimRight(url, "/") + "/collections")
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"ok": false, "message": "连接失败: " + err.Error()}})
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"ok": false, "message": "连接失败: Qdrant 返回状态码 " + resp.Status}})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"ok": true, "message": "连接成功"}})
}

// validateQdrantURL 校验 Qdrant 地址：非空且以 http:// 或 https:// 开头
func validateQdrantURL(url string) error {
	if url == "" {
		return errors.New("Qdrant 地址不能为空")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("Qdrant 地址必须以 http:// 或 https:// 开头")
	}
	return nil
}
