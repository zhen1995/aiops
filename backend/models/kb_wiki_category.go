package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KBWikiCategory Wiki 分类：树形，parent_id 为空表示顶级分类
type KBWikiCategory struct {
	ID        string         `gorm:"primaryKey;size:64;comment:id" json:"id"`
	ParentID  string         `gorm:"size:64;index;comment:父分类id" json:"parent_id"`
	Name      string         `gorm:"size:128;not null;uniqueIndex;comment:分类名称" json:"name"`
	Sort      int            `gorm:"default:0;comment:排序（同级从小到大）" json:"sort"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (KBWikiCategory) TableName() string {
	return "kb_wiki_category"
}

// BeforeCreate 自动生成 UUID
func (c *KBWikiCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
