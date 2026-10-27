package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiops/internal/knowledge"
	"aiops/internal/wiki"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// KnowledgeWikiController 知识库 LLM Wiki 控制器：分类树、词条 CRUD/发布流、混合检索、LLM 生成
type KnowledgeWikiController struct {
	DB   *gorm.DB
	Wiki *wiki.Service
	KB   *knowledge.Service
}

// NewKnowledgeWikiController 创建控制器
func NewKnowledgeWikiController(db *gorm.DB, wikiSvc *wiki.Service, kbSvc *knowledge.Service) *KnowledgeWikiController {
	return &KnowledgeWikiController{DB: db, Wiki: wikiSvc, KB: kbSvc}
}

// ---------- 分类 ----------

// ListCategories 分类列表（扁平数组，前端组树），附每个分类下的词条数
func (c *KnowledgeWikiController) ListCategories(ctx *gin.Context) {
	var cats []models.KBWikiCategory
	if err := c.DB.Order("sort asc, created_at asc").Find(&cats).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询分类失败", "error": err.Error()})
		return
	}
	items := make([]gin.H, 0, len(cats))
	for _, cat := range cats {
		var count int64
		c.DB.Model(&models.KBWikiEntry{}).Where("category_id = ?", cat.ID).Count(&count)
		items = append(items, gin.H{"id": cat.ID, "parent_id": cat.ParentID, "name": cat.Name, "sort": cat.Sort, "entry_count": count})
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// CreateCategory 新建分类（可选父分类）
func (c *KnowledgeWikiController) CreateCategory(ctx *gin.Context) {
	var payload struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
		Sort     int    `json:"sort"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称不能为空"})
		return
	}
	var dup int64
	c.DB.Model(&models.KBWikiCategory{}).Where("name = ?", payload.Name).Count(&dup)
	if dup > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称已存在"})
		return
	}
	cat := models.KBWikiCategory{Name: payload.Name, ParentID: strings.TrimSpace(payload.ParentID), Sort: payload.Sort}
	if err := c.DB.Create(&cat).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建分类失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cat})
}

// UpdateCategory 重命名/调整父分类/排序
func (c *KnowledgeWikiController) UpdateCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	var cat models.KBWikiCategory
	if err := c.DB.First(&cat, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}
	var payload struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"` // 传 null/缺省表示不修改；传 "" 表示移到顶级
		Sort     *int    `json:"sort"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if name := strings.TrimSpace(payload.Name); name != "" && name != cat.Name {
		var dup int64
		c.DB.Model(&models.KBWikiCategory{}).Where("name = ? AND id <> ?", name, id).Count(&dup)
		if dup > 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称已存在"})
			return
		}
		cat.Name = name
	}
	if payload.ParentID != nil {
		pid := strings.TrimSpace(*payload.ParentID)
		if pid == id {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "父分类不能是自身"})
			return
		}
		cat.ParentID = pid
	}
	if payload.Sort != nil {
		cat.Sort = *payload.Sort
	}
	if err := c.DB.Save(&cat).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新分类失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": cat})
}

// DeleteCategory 删除分类：子分类上移一级，词条改为未分类
func (c *KnowledgeWikiController) DeleteCategory(ctx *gin.Context) {
	id := ctx.Param("id")
	var cat models.KBWikiCategory
	if err := c.DB.First(&cat, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}
	if err := c.DB.Model(&models.KBWikiCategory{}).Where("parent_id = ?", id).Update("parent_id", cat.ParentID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "迁移子分类失败", "error": err.Error()})
		return
	}
	if err := c.DB.Model(&models.KBWikiEntry{}).Where("category_id = ?", id).Update("category_id", "").Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解除词条关联失败", "error": err.Error()})
		return
	}
	if err := c.DB.Delete(&cat).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除分类失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ---------- 词条 ----------

// ListEntries 词条列表：category_id（空=全部）、status（空=全部，review=draft）、q 关键词、分页
func (c *KnowledgeWikiController) ListEntries(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	db := c.DB.Model(&models.KBWikiEntry{})
	if cid := ctx.Query("category_id"); cid != "" {
		db = db.Where("category_id = ?", cid)
	}
	switch status := ctx.Query("status"); status {
	case "review":
		db = db.Where("status = ?", models.WikiEntryStatusDraft)
	case models.WikiEntryStatusPublished, models.WikiEntryStatusArchived:
		db = db.Where("status = ?", status)
	}
	if q := strings.TrimSpace(ctx.Query("q")); q != "" {
		like := "%" + q + "%"
		db = db.Where("title LIKE ? OR summary LIKE ? OR tags LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询词条失败", "error": err.Error()})
		return
	}
	var entries []models.KBWikiEntry
	if err := db.Order("updated_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&entries).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询词条失败", "error": err.Error()})
		return
	}
	if entries == nil {
		entries = []models.KBWikiEntry{}
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": entries, "total": total}})
}

// GetEntry 词条详情
func (c *KnowledgeWikiController) GetEntry(ctx *gin.Context) {
	var entry models.KBWikiEntry
	if err := c.DB.First(&entry, "id = ?", ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "词条不存在"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": entry})
}

type wikiEntryPayload struct {
	Title      string `json:"title"`
	CategoryID string `json:"category_id"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	Tags       string `json:"tags"` // JSON 数组字符串
}

func (p *wikiEntryPayload) validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("词条标题不能为空")
	}
	if strings.TrimSpace(p.Content) == "" {
		return errors.New("词条正文不能为空")
	}
	return nil
}

// currentOperator 取当前登录用户名（认证中间件注入）
func currentOperator(ctx *gin.Context) string {
	if name, ok := ctx.Get("username"); ok {
		if s, ok := name.(string); ok && s != "" {
			return s
		}
	}
	return "unknown"
}

// CreateEntry 新建词条（默认草稿；status 显式传 published 时直接发布）
func (c *KnowledgeWikiController) CreateEntry(ctx *gin.Context) {
	var payload wikiEntryPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if err := payload.validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	status := models.WikiEntryStatusDraft
	if ctx.Query("publish") == "1" {
		status = models.WikiEntryStatusPublished
	}
	now := time.Now()
	entry := models.KBWikiEntry{
		Title:       strings.TrimSpace(payload.Title),
		CategoryID:  strings.TrimSpace(payload.CategoryID),
		Content:     payload.Content,
		Summary:     strings.TrimSpace(payload.Summary),
		Tags:        strings.TrimSpace(payload.Tags),
		Status:      status,
		SourceType:  models.WikiEntrySourceManual,
		CreatedBy:   currentOperator(ctx),
		UpdatedBy:   currentOperator(ctx),
		PublishedAt: nil,
	}
	if status == models.WikiEntryStatusPublished {
		entry.PublishedAt = &now
	}
	if err := c.DB.Create(&entry).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建词条失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": entry})
}

// UpdateEntry 更新词条（标题/分类/正文/摘要/标签）；published 词条更新后回到 draft 需人工确认？保持状态不变
func (c *KnowledgeWikiController) UpdateEntry(ctx *gin.Context) {
	id := ctx.Param("id")
	var entry models.KBWikiEntry
	if err := c.DB.First(&entry, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "词条不存在"})
		return
	}
	var payload wikiEntryPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误", "error": err.Error()})
		return
	}
	if err := payload.validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	updates := map[string]interface{}{
		"title":      strings.TrimSpace(payload.Title),
		"category_id": strings.TrimSpace(payload.CategoryID),
		"content":    payload.Content,
		"summary":    strings.TrimSpace(payload.Summary),
		"tags":       strings.TrimSpace(payload.Tags),
		"updated_by": currentOperator(ctx),
	}
	if err := c.DB.Model(&entry).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新词条失败", "error": err.Error()})
		return
	}
	c.DB.First(&entry, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": entry})
}

// DeleteEntry 删除词条（软删除）
func (c *KnowledgeWikiController) DeleteEntry(ctx *gin.Context) {
	var entry models.KBWikiEntry
	if err := c.DB.First(&entry, "id = ?", ctx.Param("id")).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "词条不存在"})
		return
	}
	if err := c.DB.Delete(&entry).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除词条失败", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// setEntryStatus 词条状态流转：draft→published（发布）、任意→draft（退回）、任意→archived（归档）
func (c *KnowledgeWikiController) setEntryStatus(ctx *gin.Context, target string) {
	id := ctx.Param("id")
	var entry models.KBWikiEntry
	if err := c.DB.First(&entry, "id = ?", id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "词条不存在"})
		return
	}
	updates := map[string]interface{}{
		"status":     target,
		"updated_by": currentOperator(ctx),
	}
	if target == models.WikiEntryStatusPublished {
		now := time.Now()
		updates["published_at"] = &now
	}
	if err := c.DB.Model(&entry).Updates(updates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "状态变更失败", "error": err.Error()})
		return
	}
	c.DB.First(&entry, "id = ?", id)
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": entry})
}

// PublishEntry 发布词条
func (c *KnowledgeWikiController) PublishEntry(ctx *gin.Context) {
	c.setEntryStatus(ctx, models.WikiEntryStatusPublished)
}

// WithdrawEntry 退回草稿
func (c *KnowledgeWikiController) WithdrawEntry(ctx *gin.Context) {
	c.setEntryStatus(ctx, models.WikiEntryStatusDraft)
}

// ArchiveEntry 归档词条
func (c *KnowledgeWikiController) ArchiveEntry(ctx *gin.Context) {
	c.setEntryStatus(ctx, models.WikiEntryStatusArchived)
}

// GenerateFromDocument 从文档生成词条草稿（LLM）
// POST /api/knowledge-wiki/entries/generate-from-document {"document_id": "..."}
func (c *KnowledgeWikiController) GenerateFromDocument(ctx *gin.Context) {
	var payload struct {
		DocumentID string `json:"document_id"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil || strings.TrimSpace(payload.DocumentID) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误：document_id 不能为空"})
		return
	}
	entries, err := c.Wiki.GenerateFromDocument(ctx, strings.TrimSpace(payload.DocumentID), currentOperator(ctx))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entries": entries, "count": len(entries)}})
}

// Search 混合检索：词条关键词检索 + 文档向量检索
// GET /api/knowledge-wiki/search?q=&limit= → {"entries": [...], "chunks": [...]}
func (c *KnowledgeWikiController) Search(ctx *gin.Context) {
	q := strings.TrimSpace(ctx.Query("q"))
	if q == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "q 不能为空"})
		return
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	if limit < 1 || limit > 10 {
		limit = 5
	}
	entries, err := c.Wiki.SearchEntries(q, limit, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "词条检索失败", "error": err.Error()})
		return
	}
	chunks := []interface{}{}
	if c.KB != nil {
		hits, err := c.KB.Retrieve(ctx, q, limit)
		if err != nil {
			// 向量检索失败不阻断词条结果
			chunks = []interface{}{}
		} else {
			for _, h := range hits {
				chunks = append(chunks, h)
			}
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entries": entries, "chunks": chunks}})
}
