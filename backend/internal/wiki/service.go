// Package wiki 知识库 LLM Wiki：词条检索与 LLM 提炼生成（供 REST 控制器与对话 Agent 工具共用）
package wiki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"aiops/internal/chat"
	"aiops/models"

	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// maxExtractChars 文档生成词条时输入 LLM 的正文上限
const maxExtractChars = 12000

// Service Wiki 词条服务
type Service struct {
	DB *gorm.DB
}

// NewService 创建服务
func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// EntryHit 词条检索命中（附加分类名便于展示）
type EntryHit struct {
	models.KBWikiEntry
	CategoryName string `json:"category_name"`
}

// SearchEntries 关键词检索词条：title/summary/tags/content 模糊匹配，published 优先、draft 次之
func (s *Service) SearchEntries(q string, limit int, includeDraft bool) ([]EntryHit, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return []EntryHit{}, nil
	}
	like := "%" + q + "%"
	db := s.DB.Model(&models.KBWikiEntry{}).
		Where("title LIKE ? OR summary LIKE ? OR tags LIKE ? OR content LIKE ?", like, like, like, like).
		Where("status <> ?", models.WikiEntryStatusArchived)
	if !includeDraft {
		db = db.Where("status = ?", models.WikiEntryStatusPublished)
	}
	var hits []EntryHit
	if err := db.
		Order("CASE status WHEN 'published' THEN 0 ELSE 1 END").
		Order("updated_at DESC").
		Limit(limit).
		Find(&hits).Error; err != nil {
		return nil, err
	}
	// 补充分类名
	if len(hits) > 0 {
		names := map[string]string{}
		var cats []models.KBWikiCategory
		if err := s.DB.Find(&cats).Error; err == nil {
			for _, c := range cats {
				names[c.ID] = c.Name
			}
		}
		for i := range hits {
			hits[i].CategoryName = names[hits[i].CategoryID]
		}
	}
	if hits == nil {
		hits = []EntryHit{}
	}
	return hits, nil
}

// textExts 支持直接按文本提取内容的文档扩展名
var textExts = map[string]bool{
	"txt": true, "md": true, "log": true, "csv": true, "json": true,
	"xml": true, "yaml": true, "yml": true, "conf": true, "ini": true, "sh": true,
}

// draftCandidate LLM 输出的候选词条
type draftCandidate struct {
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags"`
	CategoryHint string   `json:"category_hint"`
}

// GenerateFromDocument 用 LLM 从知识库文档提炼 Wiki 词条草稿（status=draft），返回创建的词条
func (s *Service) GenerateFromDocument(ctx context.Context, docID, operator string) ([]models.KBWikiEntry, error) {
	var doc models.KBDocument
	if err := s.DB.First(&doc, "id = ?", docID).Error; err != nil {
		return nil, errors.New("文档不存在")
	}
	if doc.Status != "indexed" {
		return nil, errors.New("文档尚未索引完成，暂不能生成词条")
	}
	if !textExts[strings.ToLower(doc.Type)] {
		return nil, fmt.Errorf("暂不支持从 .%s 文档生成词条，请转换为文本类文件（txt/md/log 等）", doc.Type)
	}
	data, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return nil, errors.New("读取文档文件失败")
	}
	text := string(data)
	if len(text) > maxExtractChars {
		text = text[:maxExtractChars]
	}

	cfg, err := chat.DefaultConfig(s.DB, models.LLMModelTypeChat)
	if err != nil {
		return nil, fmt.Errorf("加载对话模型配置失败: %w", err)
	}
	cm, err := chat.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	system := `你是运维知识库工程师。根据用户提供的文档内容提炼运维知识词条。
要求：
1. 每个词条聚焦一个知识点（概念、操作步骤、故障处理、配置说明等）；
2. content 用 Markdown 编写，必须基于文档内容，禁止编造文档中没有的信息；
3. summary 为一句话摘要（50 字以内）；
4. tags 为 3~6 个检索关键词；
5. 严格只输出 JSON，格式：{"entries":[{"title":"...","summary":"...","content":"...","tags":["..."],"category_hint":"..."}]}，不要输出其他文字。`
	user := fmt.Sprintf("文档名称：%s\n\n文档内容：\n%s", doc.Name, text)

	messages := []*schema.Message{
		schema.SystemMessage(system),
		schema.UserMessage(user),
	}
	genCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	resp, err := cm.Generate(genCtx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 生成失败: %w", err)
	}

	raw := extractJSON(resp.Content)
	var parsed struct {
		Entries []draftCandidate `json:"entries"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || len(parsed.Entries) == 0 {
		return nil, errors.New("LLM 输出解析失败，未获得候选词条")
	}

	// 分类提示匹配：按名称模糊匹配已有分类，匹配不到则留空（未分类）
	categoryIDs := map[string]string{}
	var cats []models.KBWikiCategory
	if err := s.DB.Find(&cats).Error; err == nil {
		for _, c := range cats {
			categoryIDs[c.Name] = c.ID
		}
	}

	created := make([]models.KBWikiEntry, 0, len(parsed.Entries))
	for _, c := range parsed.Entries {
		title := strings.TrimSpace(c.Title)
		content := strings.TrimSpace(c.Content)
		if title == "" || content == "" {
			continue
		}
		tags, _ := json.Marshal(c.Tags)
		entry := models.KBWikiEntry{
			Title:      title,
			Content:    content,
			Summary:    strings.TrimSpace(c.Summary),
			Tags:       string(tags),
			Status:     models.WikiEntryStatusDraft,
			SourceType: models.WikiEntrySourceDocument,
			SourceRef:  doc.ID,
			SourceName: doc.Name,
			CreatedBy:  operator,
			UpdatedBy:  operator,
		}
		if id, ok := categoryIDs[c.CategoryHint]; ok {
			entry.CategoryID = id
		}
		if err := s.DB.Create(&entry).Error; err != nil {
			return created, fmt.Errorf("词条落库失败: %w", err)
		}
		created = append(created, entry)
	}
	if len(created) == 0 {
		return nil, errors.New("LLM 未产出有效词条（标题/正文为空）")
	}
	return created, nil
}

// extractJSON 从 LLM 输出中提取第一个 JSON 对象（容忍 ```json 围栏与前后说明文字）
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "json")
		if j := strings.Index(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	return strings.TrimSpace(s)
}
