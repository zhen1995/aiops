package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemConfigKeyQdrantURL system_configs 中 Qdrant 向量库地址的键名
const SystemConfigKeyQdrantURL = "qdrant_url"

// DefaultQdrantURL system_configs 无记录时的回退默认值
const DefaultQdrantURL = "http://localhost:6333"

// SystemConfig 系统配置（key-value 存储）
type SystemConfig struct {
	ID        string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	Key       string    `gorm:"size:128;not null;uniqueIndex;comment:配置键" json:"key"`
	Value     string    `gorm:"size:512;not null;comment:配置值" json:"value"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// BeforeCreate 自动生成 UUID
func (c *SystemConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// SeedSystemConfig 种子系统配置：qdrant_url 无记录时插入配置默认值（幂等）
func SeedSystemConfig(db *gorm.DB, qdrantURL string) error {
	if qdrantURL == "" {
		qdrantURL = DefaultQdrantURL
	}
	var count int64
	if err := db.Model(&SystemConfig{}).Where("`key` = ?", SystemConfigKeyQdrantURL).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&SystemConfig{Key: SystemConfigKeyQdrantURL, Value: qdrantURL}).Error
}
