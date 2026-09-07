package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Datasource 数据源
type Datasource struct {
	ID        string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name      string         `gorm:"size:30;comment:名称" json:"name"`
	Type      string         `gorm:"size:20;comment:类型 Prometheus/ElasticSearch/Pyroscope" json:"type"`
	URL       string         `gorm:"size:100;not null;comment:http地址" json:"url"`
	Timeout   int            `gorm:"comment:超时(ms)" json:"timeout"`
	Username  string         `gorm:"size:100;comment:用户名" json:"username"`
	Password  string         `gorm:"size:100;comment:密码" json:"password"`
	IsSkipSSL int            `gorm:"comment:是否跳过SSL验证 0-否 1-是" json:"is_skip_ssl"`
	Remark    string         `gorm:"size:100;comment:备注" json:"remark"`
	IsEnabled int            `gorm:"comment:状态 0-不启用 1-启用" json:"is_enabled"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (Datasource) TableName() string {
	return "datasources"
}

// BeforeCreate 自动生成 UUID 并设置默认值
func (d *Datasource) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	if d.Timeout == 0 {
		d.Timeout = 5000
	}
	if d.IsEnabled == 0 {
		d.IsEnabled = 1
	}
	return nil
}

// DatasourceTypeOptions 数据源类型选项
var DatasourceTypeOptions = []string{
	"Prometheus",
	"ElasticSearch",
	"Pyroscope",
}
