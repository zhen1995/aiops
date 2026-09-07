package controllers

import (
	"errors"
	"net/http"

	"aiops/internal/knowledge"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 上传限制：整个 multipart 请求 100MB，单个文件 50MB
const (
	maxUploadTotal = 100 << 20
	maxFileSize    = 50 << 20
)

// KnowledgeController 运维知识库控制器
type KnowledgeController struct {
	DB      *gorm.DB
	Service *knowledge.Service
}

func NewKnowledgeController(db *gorm.DB, svc *knowledge.Service) *KnowledgeController {
	return &KnowledgeController{DB: db, Service: svc}
}

// List 文档列表（软删过滤、按创建时间倒序）
func (c *KnowledgeController) List(ctx *gin.Context) {
	docs := make([]models.KBDocument, 0)
	if err := c.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": docs})
}

// Upload 多文件上传，逐个落库并异步索引；逐文件收集成功/失败结果
func (c *KnowledgeController) Upload(ctx *gin.Context) {
	// 限制整个 multipart 请求体大小，超限由 MaxBytesReader 截断并报错
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxUploadTotal)
	form, err := ctx.MultipartForm()
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"code": 413, "message": "上传内容超过 100MB 限制"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请使用 multipart/form-data 上传"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未接收到文件"})
		return
	}
	uploader := "unknown"
	if info, ok := ctx.Get("username"); ok {
		if name, ok2 := info.(string); ok2 {
			uploader = name
		}
	}
	type failedItem struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	}
	succeeded := make([]*models.KBDocument, 0, len(files))
	failed := make([]failedItem, 0)
	for _, fh := range files {
		if fh.Size > maxFileSize {
			failed = append(failed, failedItem{Name: fh.Filename, Error: "文件超过 50MB 限制"})
			continue
		}
		src, err := fh.Open()
		if err != nil {
			failed = append(failed, failedItem{Name: fh.Filename, Error: "打开文件失败: " + err.Error()})
			continue
		}
		doc, err := c.Service.SaveUpload(ctx, fh.Filename, fh.Size, src, uploader)
		src.Close()
		if err != nil {
			failed = append(failed, failedItem{Name: fh.Filename, Error: err.Error()})
			continue
		}
		c.Service.IndexDocument(doc.ID)
		succeeded = append(succeeded, doc)
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"succeeded": succeeded, "failed": failed}})
}

// Delete 删除文档
func (c *KnowledgeController) Delete(ctx *gin.Context) {
	if err := c.Service.Delete(ctx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// Reindex 重新索引
func (c *KnowledgeController) Reindex(ctx *gin.Context) {
	if err := c.Service.Reindex(ctx, ctx.Param("id")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "重新索引失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "已触发重新索引"})
}

// Retrieve 语义检索
func (c *KnowledgeController) Retrieve(ctx *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
		TopK  int    `json:"top_k"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "query 不能为空"})
		return
	}
	if req.TopK <= 0 || req.TopK > 50 {
		req.TopK = 5
	}
	hits, err := c.Service.Retrieve(ctx, req.Query, req.TopK)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "检索失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": hits})
}
