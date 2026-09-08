package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// N9eConfig 夜莺引擎配置（单例，只允许一条 enabled 记录）
type N9eConfig struct {
	ID          string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name        string         `gorm:"size:64;not null;default:夜莺" json:"name"`
	Address     string         `gorm:"size:255;not null;comment:夜莺服务地址，如 http://10.2.209.145:17000" json:"address"`
	Token       string         `gorm:"size:128;not null;comment:X-User-Token" json:"token"`
	IsEnabled   int            `gorm:"default:1;comment:是否启用" json:"is_enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"`
}

func (N9eConfig) TableName() string { return "n9e_config" }

func (c *N9eConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// GetEnabledN9eConfig 获取当前启用的夜莺配置
func GetEnabledN9eConfig(db *gorm.DB) (*N9eConfig, error) {
	var cfg N9eConfig
	if err := db.Where("is_enabled = 1").First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BaseURL 返回带斜杠的基础地址
func (c *N9eConfig) BaseURL() string {
	addr := strings.TrimRight(c.Address, "/")
	return addr
}
