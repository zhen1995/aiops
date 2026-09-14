package models

import (
	"time"

	"gorm.io/gorm"
)

// InspectionTask 巡检任务表
type InspectionTask struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"size:128;not null" json:"name"`        // 任务名称
	CronExpr   string `gorm:"size:64;not null" json:"cron_expr"`    // cron 表达式
	Prompt     string `gorm:"type:text;not null" json:"prompt"`     // 任务提示词
	// Mode 执行模式：single=单 Agent 巡检（默认）；multi=多 Agent 多维度巡检
	Mode string `gorm:"size:16;default:single;comment:执行模式 single-单Agent multi-多维度" json:"mode"`
	// Dimensions 多维度模式下的巡检维度 JSON：[{"name":"容量趋势","prompt":"..."}]，1~8 项
	Dimensions string `gorm:"type:text;comment:巡检维度JSON(multi模式)" json:"dimensions"`
	NotifyMediaIDs string `gorm:"type:text;comment:通知媒介ID列表JSON" json:"notify_media_ids"` // 报告生成后推送的媒介
	Enabled    bool   `gorm:"default:true" json:"enabled"`          // 是否启用
	LastRunAt  *time.Time `json:"last_run_at"`                    // 上次执行时间
	NextRunAt  *time.Time `json:"next_run_at"`                    // 预计下次执行时间
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InspectionTask) TableName() string {
	return "inspection_tasks"
}
