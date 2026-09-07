package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LLMConfig 大模型配置
type LLMConfig struct {
	ID               string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name             string         `gorm:"size:30;comment:LLM 名称" json:"name"`
	Description      string         `gorm:"size:100;comment:描述" json:"description"`
	SupplierCategory string         `gorm:"size:10;comment:供应商类型 openai/claude/gemini等" json:"supplier_category"`
	ModelType        string         `gorm:"size:10;not null;default:chat;comment:模型类型 chat-对话 embedding-向量化" json:"model_type"`
	Model            string         `gorm:"size:20;comment:模型名称" json:"model"`
	BaseURL          string         `gorm:"size:50;comment:api url" json:"base_url"`
	APIKey           string         `gorm:"size:100;comment:密钥" json:"api_key"`
	IsDefault        int            `gorm:"not null;default:0;comment:0 否，1是" json:"is_default"`
	IsEnabled        int            `gorm:"default:1;comment:是否启用 0-否 1-是" json:"is_enabled"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (LLMConfig) TableName() string {
	return "llm_config"
}

// BeforeCreate 自动生成 UUID
func (l *LLMConfig) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

// 模型类型常量
const (
	LLMModelTypeChat      = "chat"
	LLMModelTypeEmbedding = "embedding"
)

// LLMModelTypeOptions 模型类型选项
var LLMModelTypeOptions = []string{LLMModelTypeChat, LLMModelTypeEmbedding}

// IsValidLLMModelType 校验模型类型是否合法
func IsValidLLMModelType(t string) bool {
	for _, o := range LLMModelTypeOptions {
		if o == t {
			return true
		}
	}
	return false
}

// SupplierCategoryOptions 供应商类型选项
var SupplierCategoryOptions = []string{
	"openai",
	"claude",
	"gemini",
	"qwen",
	"deepseek",
	"kimi",
	"other",
}
