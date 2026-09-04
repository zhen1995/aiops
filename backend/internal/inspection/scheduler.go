package inspection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aiops/models"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler 管理所有巡检定时任务的生命周期
type Scheduler struct {
	cron *cron.Cron
	db   *gorm.DB

	// frontendBaseURL 前端访问地址，用于报告通知中的"完整报告"链接
	frontendBaseURL string

	mu     sync.Mutex
	entries map[uint]cron.EntryID // taskID → cron.EntryID
}

// NewScheduler 创建 Scheduler 实例
func NewScheduler(db *gorm.DB, frontendBaseURL string) *Scheduler {
	return &Scheduler{
		cron:            cron.New(cron.WithSeconds()), // 支持秒级 cron（可选），也兼容 5 段标准 cron
		db:              db,
		frontendBaseURL: frontendBaseURL,
		entries:         make(map[uint]cron.EntryID),
	}
}

// Start 启动 scheduler，并从 DB 加载所有启用的任务注册进去
func (s *Scheduler) Start() error {
	var tasks []models.InspectionTask
	if err := s.db.Where("enabled = ?", true).Find(&tasks).Error; err != nil {
		return fmt.Errorf("加载巡检任务失败: %w", err)
	}

	for _, t := range tasks {
		s.registerTask(t)
	}

	s.cron.Start()
	fmt.Printf("[巡检调度器] 已启动，共加载 %d 个巡检任务\n", len(tasks))
	return nil
}

// Stop 优雅停止 scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
	}
}

// Add 新增任务到调度器
func (s *Scheduler) Add(task models.InspectionTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[task.ID]; exists {
		return fmt.Errorf("任务 %d 已在调度器中", task.ID)
	}

	s.registerTask(task)
	return nil
}

// Update 更新任务（先移除再重新添加）
func (s *Scheduler) Update(task models.InspectionTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[task.ID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, task.ID)
	}

	if task.Enabled {
		s.registerTask(task)
	}

	return nil
}

// Remove 从调度器移除任务
func (s *Scheduler) Remove(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[taskID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, taskID)
	}
}

// ParseCron 解析 cron 表达式，返回下次执行时间序列（用于前端预览）
func ParseCron(expr string, count int) ([]time.Time, error) {
	// 先尝试带秒的 6 段表达式，再回退标准 5 段
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(expr)
	if err != nil {
		parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sched, err = parser.Parse(expr)
		if err != nil {
			return nil, err
		}
	}

	if count <= 0 {
		count = 5
	}
	now := time.Now()
	times := make([]time.Time, 0, count)
	t := now
	for i := 0; i < count; i++ {
		t = sched.Next(t)
		times = append(times, t)
	}
	return times, nil
}

func (s *Scheduler) registerTask(task models.InspectionTask) {
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		fmt.Printf("[巡检调度器] 开始执行任务: %s (id=%d)\n", task.Name, task.ID)

		// 更新 last_run_at
		now := time.Now()
		_ = s.db.Model(&task).Update("last_run_at", now).Error

		// 估算下次执行时间并更新 next_run_at
		if nextTimes, err2 := ParseCron(task.CronExpr, 1); err2 == nil && len(nextTimes) > 0 {
			_ = s.db.Model(&task).Update("next_run_at", nextTimes[0]).Error
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if _, err2 := Executor(ctx, s.db, task, s.frontendBaseURL); err2 != nil {
			fmt.Printf("[巡检调度器] 任务 %s 执行失败: %v\n", task.Name, err2)
		} else {
			fmt.Printf("[巡检调度器] 任务 %s 执行完成\n", task.Name)
		}
	})

	if err != nil {
		fmt.Printf("[巡检调度器] 注册任务 %s(cron=%s) 失败: %v\n", task.Name, task.CronExpr, err)
		return
	}

	s.entries[task.ID] = entryID

	// 同步 next_run_at 到 DB
	if nextTimes, err2 := ParseCron(task.CronExpr, 1); err2 == nil && len(nextTimes) > 0 {
		_ = s.db.Model(&task).Update("next_run_at", nextTimes[0]).Error
	}
	fmt.Printf("[巡检调度器] 已注册任务: %s [%s]\n", task.Name, task.CronExpr)
}

// TriggerNow 立即执行一次指定任务（不经过 cron 调度）
func (s *Scheduler) TriggerNow(ctx context.Context, taskID uint) (models.InspectionReport, error) {
	var task models.InspectionTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return models.InspectionReport{}, err
	}
	return Executor(ctx, s.db, task, s.frontendBaseURL)
}
