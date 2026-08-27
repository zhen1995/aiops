package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatSession 对话会话
type ChatSession struct {
	ID        string         `gorm:"primaryKey;size:36;comment:会话ID" json:"id"`
	UserID    string         `gorm:"size:36;not null;index;comment:用户ID" json:"user_id"`
	Title     string         `gorm:"size:100;not null;default:'新会话';comment:会话标题" json:"title"`
	Status    string         `gorm:"size:20;not null;default:'active';comment:状态 active/archived" json:"status"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间（软删除）" json:"deleted_at"`
}

func (ChatSession) TableName() string {
	return "chat_session"
}

func (s *ChatSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID               string         `gorm:"primaryKey;size:36;comment:消息ID" json:"id"`
	SessionID        string         `gorm:"size:36;not null;index;comment:会话ID" json:"session_id"`
	Role             string         `gorm:"size:20;not null;comment:角色 system/user/assistant" json:"role"`
	Content          string         `gorm:"type:longtext;not null;comment:消息内容" json:"content"`
	PromptTokens     int            `gorm:"default:0;comment:输入token数" json:"prompt_tokens"`
	CompletionTokens int            `gorm:"default:0;comment:输出token数" json:"completion_tokens"`
	TotalTokens      int            `gorm:"default:0;comment:总token数" json:"total_tokens"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"comment:删除时间（软删除）" json:"deleted_at"`
}

func (ChatMessage) TableName() string {
	return "chat_message"
}

func (m *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
