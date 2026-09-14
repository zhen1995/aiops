package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 日志聚类趋势常量
const (
	LogClusterTrendSpike  = "spike"  // 激增
	LogClusterTrendRising = "rising" // 上升
	LogClusterTrendFlat   = "flat"   // 平稳
)

// LogCluster 日志聚类表：一期与模板一一对应（Drain 等价类），预留语义合并后多模板成簇
type LogCluster struct {
	ID          string    `gorm:"primaryKey;size:64;comment:id" json:"id"`
	ServiceID   string    `gorm:"size:64;uniqueIndex:uk_log_cluster_code;index;comment:关联服务ID" json:"service_id"`
	TemplateID  string    `gorm:"size:64;comment:主模板ID" json:"template_id"`
	Code        string    `gorm:"size:16;uniqueIndex:uk_log_cluster_code;comment:聚类短码 如 C-018" json:"code"`
	Pattern     string    `gorm:"type:text;comment:展示用模式（=模板）" json:"pattern"`
	Level       string    `gorm:"size:16;comment:级别 error/warn/info" json:"level"`
	Count       int64     `gorm:"comment:当前窗口累计条数" json:"count"`
	Trend       string    `gorm:"size:16;index;comment:趋势 spike/rising/flat" json:"trend"`
	FirstSeenAt time.Time `gorm:"comment:首次出现" json:"first_seen_at"`
	LastSeenAt  time.Time `gorm:"comment:最近出现" json:"last_seen_at"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (LogCluster) TableName() string {
	return "log_clusters"
}

// BeforeCreate 自动生成 UUID
func (c *LogCluster) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
