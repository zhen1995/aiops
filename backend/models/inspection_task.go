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
