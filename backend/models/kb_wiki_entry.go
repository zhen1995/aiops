package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Wiki 词条状态
const (
	WikiEntryStatusDraft     = "draft"     // 草稿（待审核）
	WikiEntryStatusPublished = "published" // 已发布（可被检索与对话引用）
	WikiEntryStatusArchived  = "archived"  // 已归档（不可检索）
)

// Wiki 词条来源类型
const (
	WikiEntrySourceDocument  = "document"  // 由文档提炼
	WikiEntrySourceAlert     = "alert_event" // 由告警/RCA 沉淀
	WikiEntrySourceInspect   = "inspection"  // 由巡检报告沉淀
	WikiEntrySourceManual    = "manual"      // 手工创建
)

// KBWikiEntry Wiki 词条：结构化知识一等公民（LLM Wiki）
// 原始文档（kb_document/kb_chunk）降级为来源素材，词条快照式存储、人审发布后进入混合检索
type KBWikiEntry struct {
	ID          string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Title       string         `gorm:"size:200;not null;comment:词条标题" json:"title"`
	CategoryID  string         `gorm:"size:64;index;comment:分类id（空为未分类）" json:"category_id"`
	Content     string         `gorm:"type:text;comment:正文 Markdown" json:"content"`
	Summary     string         `gorm:"size:500;comment:一句话摘要（列表展示）" json:"summary"`
	Tags        string         `gorm:"size:300;comment:标签 JSON 数组" json:"tags"`
	Status      string         `gorm:"size:16;not null;default:draft;index;comment:状态 draft/published/archived" json:"status"`
	SourceType  string         `gorm:"size:16;comment:来源 document/alert_event/inspection/manual" json:"source_type"`
	SourceRef   string         `gorm:"size:64;index;comment:来源引用（文档/事件/报告 ID）" json:"source_ref"`
	SourceName  string         `gorm:"size:200;comment:来源展示名（快照式，源删除不丢失）" json:"source_name"`
	CreatedBy   string         `gorm:"size:64;comment:创建人" json:"created_by"`
	UpdatedBy   string         `gorm:"size:64;comment:更新人" json:"updated_by"`
	PublishedAt *time.Time     `gorm:"comment:发布时间" json:"published_at"`
	CreatedAt   time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (KBWikiEntry) TableName() string {
	return "kb_wiki_entry"
}

// BeforeCreate 自动生成 UUID
func (e *KBWikiEntry) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}
