package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KBChunk 知识库分块文本表（向量与相似度在 Qdrant）
type KBChunk struct {
	ID         string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	DocumentID string         `gorm:"column:document_id;size:40;not null;index;comment:文档id" json:"document_id"`
	ChunkIndex int            `gorm:"column:chunk_index;comment:块序号" json:"chunk_index"`
	Content    string         `gorm:"type:text;comment:块文本" json:"content"`
	CreatedAt  *time.Time     `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (KBChunk) TableName() string {
	return "kb_chunk"
}

// BeforeCreate 创建前自动生成 ID 并设置时间戳
func (c *KBChunk) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now()
	c.CreatedAt = &now
	return nil
}
