package models

import (
	"time"
)

// LogAnalysisCursor 日志分析增量游标：每服务记录最近分析到的日志时间
type LogAnalysisCursor struct {
	ServiceID      string    `gorm:"primaryKey;size:64;comment:服务ID" json:"service_id"`
	LastAnalyzedAt time.Time `gorm:"comment:最近分析到的日志时间（本轮 end）" json:"last_analyzed_at"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (LogAnalysisCursor) TableName() string {
	return "log_analysis_cursors"
}
