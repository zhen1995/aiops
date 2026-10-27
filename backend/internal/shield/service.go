// Package shield 告警屏蔽：内存缓存屏蔽规则，引擎建事件前按事件标签匹配拦截
package shield

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"aiops/models"

	"gorm.io/gorm"
)

// shieldStrategy 通知记录中标识屏蔽拦截的策略名
const shieldStrategy = "shield"

// Service 屏蔽规则服务：内存缓存 + CRUD 后热加载
type Service struct {
	db    *gorm.DB
	mu    sync.RWMutex
	rules []models.AlertShieldRule
}

// NewService 创建服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Load 从数据库加载启用中的屏蔽规则到内存（启动时调用）
func (s *Service) Load() error {
	var rules []models.AlertShieldRule
	if err := s.db.Where("is_enabled = 1").Find(&rules).Error; err != nil {
		return err
	}
	s.mu.Lock()
	s.rules = rules
	s.mu.Unlock()
	return nil
}

// Reload 重载屏蔽规则（CRUD 后调用）
func (s *Service) Reload() {
	if err := s.Load(); err != nil {
		// 加载失败保留旧缓存，避免误放行
		return
	}
}

// Strategy 返回通知记录中的屏蔽策略标识
func (s *Service) Strategy() string { return shieldStrategy }

// Match 判断给定事件标签与时间是否命中屏蔽规则：
// 存在启用中的规则，当前时间在 [start_at, end_at] 内，且事件标签包含规则的全部标签
func (s *Service) Match(tags string, now time.Time) (bool, string) {
	eventTags := tagSet(tags)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rule := range s.rules {
		if now.Before(rule.StartAt) || now.After(rule.EndAt) {
			continue
		}
		if allMatch(rule.Tags, eventTags) {
			return true, rule.Name
		}
	}
	return false, ""
}

// tagSet 把 "k=v,k=v" 形式的事件标签转成集合
func tagSet(tags string) map[string]bool {
	set := map[string]bool{}
	for _, t := range strings.Split(tags, ",") {
		if t = strings.TrimSpace(t); t != "" {
			set[t] = true
		}
	}
	return set
}

// allMatch 规则标签（JSON 数组）是否全部包含在事件标签集合中
func allMatch(ruleTags string, eventTags map[string]bool) bool {
	var arr []string
	if err := json.Unmarshal([]byte(ruleTags), &arr); err != nil {
		return false
	}
	if len(arr) == 0 {
		return false
	}
	for _, t := range arr {
		if !eventTags[strings.TrimSpace(t)] {
			return false
		}
	}
	return true
}
