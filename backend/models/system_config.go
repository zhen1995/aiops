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

// SystemConfigKeyFrontendBaseURL system_configs 中前端访问地址的键名
// （巡检报告通知中的"完整报告"链接、告警通知模板 $.domain 变量均使用此地址）
const SystemConfigKeyFrontendBaseURL = "frontend_base_url"

// DefaultFrontendBaseURL system_configs 无记录时的回退默认值
const DefaultFrontendBaseURL = "http://localhost:5173"

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

// SeedSystemConfig 种子系统配置：qdrant_url / frontend_base_url 无记录时插入配置
// 默认值（幂等）。frontendBaseURL 为空时取 DefaultFrontendBaseURL。
func SeedSystemConfig(db *gorm.DB, qdrantURL, frontendBaseURL string) error {
	if qdrantURL == "" {
		qdrantURL = DefaultQdrantURL
	}
	if frontendBaseURL == "" {
		frontendBaseURL = DefaultFrontendBaseURL
	}
	for key, value := range map[string]string{
		SystemConfigKeyQdrantURL:       qdrantURL,
		SystemConfigKeyFrontendBaseURL: frontendBaseURL,
	} {
		var count int64
		if err := db.Model(&SystemConfig{}).Where("`key` = ?", key).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&SystemConfig{Key: key, Value: value}).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetSystemConfigValue 读取指定 key 的配置值，无记录或读错误时回退 fallback
func GetSystemConfigValue(db *gorm.DB, key, fallback string) string {
	var cfg SystemConfig
	if err := db.Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return fallback
	}
	return cfg.Value
}
