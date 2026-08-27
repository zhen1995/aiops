package chat

import (
	"errors"

	"aiops/models"

	"gorm.io/gorm"
)

// DefaultConfig 从 llm_config 读取默认且启用的大模型配置
func DefaultConfig(db *gorm.DB) (*models.LLMConfig, error) {
	var cfg models.LLMConfig
	err := db.Where("is_default = ? AND is_enabled = ?", 1, 1).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("未找到默认且启用的大模型配置")
		}
		return nil, err
	}
	return &cfg, nil
}
