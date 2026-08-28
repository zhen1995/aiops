package datasource

import (
	"context"
	"errors"
	"fmt"

	"aiops/models"

	"gorm.io/gorm"
)

// Manager 负责从数据库加载数据源配置
type Manager struct {
	db *gorm.DB
}

// NewManager 创建数据源管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// LoadEnabled 加载所有启用的数据源
func (m *Manager) LoadEnabled(ctx context.Context) ([]models.Datasource, error) {
	var list []models.Datasource
	if err := m.db.WithContext(ctx).
		Where("is_enabled = ?", 1).
		Order("type ASC, name ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindEnabled 按名称查找启用的数据源
func (m *Manager) FindEnabled(ctx context.Context, name string) (*models.Datasource, error) {
	var ds models.Datasource
	err := m.db.WithContext(ctx).
		Where("name = ? AND is_enabled = ?", name, 1).
		First(&ds).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("数据源不存在或未启用: %s", name)
		}
		return nil, err
	}
	return &ds, nil
}
