package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LogClusterStat 聚类时间分桶统计（1 小时桶，保留 7 天）：趋势图与趋势打标数据源
type LogClusterStat struct {
	ID          string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	ClusterID   string    `gorm:"size:64;uniqueIndex:uk_log_cluster_stat;index;comment:聚类ID" json:"cluster_id"`
	ServiceID   string    `gorm:"size:64;index;comment:关联服务ID" json:"service_id"`
	BucketStart time.Time `gorm:"uniqueIndex:uk_log_cluster_stat;index;comment:分桶起始（整点对齐，1小时）" json:"bucket_start"`
	TotalCount  int64     `gorm:"comment:该桶总量" json:"total_count"`
	ErrorCount  int64     `gorm:"comment:该桶ERROR量" json:"error_count"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (LogClusterStat) TableName() string {
	return "log_cluster_stats"
}

// BeforeCreate 自动生成 UUID
func (s *LogClusterStat) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
