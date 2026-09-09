package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 服务注册：将业务服务与其日志/指标/剖析数据源位置关联，供 Agent 查询时精确定位
type Service struct {
	ID               string         `gorm:"primaryKey;size:40;comment:id" json:"id"`
	Name             string         `gorm:"size:50;comment:服务名称" json:"name"`
	Code             string         `gorm:"size:50;uniqueIndex;comment:服务编码 小写字母数字中划线" json:"code"`
	ESDatasourceID   string         `gorm:"size:40;comment:ElasticSearch数据源ID" json:"es_datasource_id"`
	ESIndexPatterns  string         `gorm:"type:text;comment:ES索引模式 JSON数组字符串" json:"es_index_patterns"`
	PromDatasourceID string         `gorm:"size:40;comment:Prometheus数据源ID" json:"prom_datasource_id"`
	PromLabels       string         `gorm:"type:text;comment:Prometheus标签选择器 JSON对象字符串" json:"prom_labels"`
	PyroscopeApp     string         `gorm:"size:100;comment:Pyroscope应用名" json:"pyroscope_app"`
	Owner            string         `gorm:"size:50;comment:负责人" json:"owner"`
	Description      string         `gorm:"size:200;comment:描述" json:"description"`
	Status           int            `gorm:"comment:状态 1-启用 0-停用" json:"status"`
	Verified         int            `gorm:"comment:探测是否通过 1-是 0-否" json:"verified"`
	VerifyMessage    string         `gorm:"type:text;comment:探测结果信息" json:"verify_message"`
	LastVerifiedAt   *time.Time     `gorm:"comment:最近探测时间" json:"last_verified_at"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"comment:删除时间" json:"deleted_at"`
}

func (Service) TableName() string {
	return "services"
}

// BeforeCreate 自动生成 UUID 并设置默认值
func (s *Service) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == 0 {
		s.Status = 1
	}
	return nil
}
