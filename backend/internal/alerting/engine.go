// Package alerting 告警规则评估引擎：按每条规则的执行频率探测 PromQL，
// 持续命中达到持续时间后自动生成告警事件，持续未命中后自动生成告警恢复事件。
// 规则持续时间为 0 时表示命中/未命中一次即触发告警/恢复。
package alerting

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"aiops/internal/datasource"
	"aiops/models"

	"gorm.io/gorm"
)

// defaultEvalInterval 规则未配置执行频率时的默认值
const defaultEvalInterval = 30

// ruleState 每条规则的内存评估状态
type ruleState struct {
	alertingSince    time.Time // 本次持续命中的起始时间
	notAlertingSince time.Time // 本次持续未命用的起始时间（用于恢复判断）
	eventID          string    // 当前未恢复的告警事件 ID，空表示当前无未恢复事件
}

// Engine 告警规则评估引擎
type Engine struct {
	db   *gorm.DB
	stop chan struct{}

	mu       sync.Mutex
	states   map[string]*ruleState  // ruleID → 评估状态
	stopChan map[string]chan struct{} // ruleID → 该规则评估循环的停止通道
}

// NewEngine 创建评估引擎
func NewEngine(db *gorm.DB) *Engine {
	return &Engine{
		db:       db,
		stop:     make(chan struct{}),
		states:   make(map[string]*ruleState),
		stopChan: make(map[string]chan struct{}),
	}
}

// Start 启动引擎，加载所有启用的告警规则并开始评估
func (e *Engine) Start() error {
	var rules []models.AlertRule
	if err := e.db.Where("is_enabled = ?", 1).Find(&rules).Error; err != nil {
		return fmt.Errorf("加载告警规则失败: %w", err)
	}
	for _, r := range rules {
		e.startRule(r)
	}
	fmt.Printf("[告警引擎] 已启动，共加载 %d 条告警规则\n", len(rules))
	return nil
}

// Stop 停止所有评估循环
func (e *Engine) Stop() {
	close(e.stop)
	e.mu.Lock()
	for id, ch := range e.stopChan {
		close(ch)
		delete(e.stopChan, id)
	}
	e.mu.Unlock()
}

// Add 新增规则评估（仅启用状态）
func (e *Engine) Add(rule models.AlertRule) {
	if rule.IsEnabled != 1 {
		return
	}
	e.startRule(rule)
}

// Update 规则更新后重启其评估循环
func (e *Engine) Update(rule models.AlertRule) {
	e.stopRule(rule.ID)
	if rule.IsEnabled == 1 {
		e.startRule(rule)
	}
}

// Remove 移除规则评估循环，并把该规则未恢复的事件置为已恢复
func (e *Engine) Remove(ruleID string) {
	e.stopRule(ruleID)
	e.mu.Lock()
	delete(e.states, ruleID)
	e.mu.Unlock()
	now := time.Now()
	if err := e.db.Model(&models.AlertEvent{}).
		Where("rule_id = ? AND status = ?", ruleID, models.AlertEventStatusFiring).
		Updates(map[string]interface{}{"status": models.AlertEventStatusResolved, "recovered_at": &now}).Error; err != nil {
		fmt.Printf("[告警引擎] 关闭规则 %s 的未恢复事件失败: %v\n", ruleID, err)
	}
}

// startRule 为单条规则启动评估循环
func (e *Engine) startRule(rule models.AlertRule) {
	e.mu.Lock()
	if _, ok := e.stopChan[rule.ID]; ok {
		e.mu.Unlock()
		return
	}
	stopCh := make(chan struct{})
	e.stopChan[rule.ID] = stopCh
	e.mu.Unlock()

	interval := rule.EvalInterval
	if interval <= 0 {
		interval = defaultEvalInterval
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-e.stop:
				return
			case <-ticker.C:
				e.eval(rule)
			}
		}
	}()
}

// stopRule 停止单条规则的评估循环
func (e *Engine) stopRule(ruleID string) {
	e.mu.Lock()
	if ch, ok := e.stopChan[ruleID]; ok {
		close(ch)
		delete(e.stopChan, ruleID)
	}
	e.mu.Unlock()
}

// eval 评估单条规则的一次探测
func (e *Engine) eval(rule models.AlertRule) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ds, err := e.findPrometheus(ctx)
	if err != nil {
		fmt.Printf("[告警引擎] 规则 %s 评估失败: %v\n", rule.Name, err)
		return
	}

	cli := datasource.NewPrometheusClient(ds)
	res, err := cli.QueryInstant(ctx, rule.PromQL)
	if err != nil {
		fmt.Printf("[告警引擎] 规则 %s 探测失败: %v\n", rule.Name, err)
		return
	}

	now := time.Now()
	e.mu.Lock()
	defer e.mu.Unlock()

	state := e.states[rule.ID]
	if state == nil {
		state = &ruleState{}
		e.states[rule.ID] = state
	}

	if res.Hit {
		// 命中：若有未恢复事件，重置恢复计时；否则累计触发计时
		state.notAlertingSince = time.Time{}
		if state.eventID != "" {
			return
		}
		if rule.Duration > 0 {
			// 持续时间 > 0：需持续命中达到持续时间后才触发
			if state.alertingSince.IsZero() {
				state.alertingSince = now
				return
			}
			if now.Sub(state.alertingSince) < time.Duration(rule.Duration)*time.Second {
				return
			}
		}
		// 持续时间 = 0：命中一次即触发
		eventID, err := e.fireLocked(rule, res, now)
		if err != nil {
			fmt.Printf("[告警引擎] 创建告警事件失败(规则 %s): %v\n", rule.Name, err)
			return
		}
		state.eventID = eventID
		state.alertingSince = time.Time{}
		return
	}

	// 未命中：重置触发计时；若有未恢复事件，累计恢复计时
	state.alertingSince = time.Time{}
	if state.eventID == "" {
		return
	}
	if rule.Duration > 0 {
		// 持续时间 > 0：需持续未命中达到持续时间后才恢复
		if state.notAlertingSince.IsZero() {
			state.notAlertingSince = now
			return
		}
		if now.Sub(state.notAlertingSince) < time.Duration(rule.Duration)*time.Second {
			return
		}
	}
	// 持续时间 = 0：未命中一次即恢复
	if err := e.recoverLocked(rule, now); err != nil {
		fmt.Printf("[告警引擎] 创建恢复事件失败(规则 %s): %v\n", rule.Name, err)
		return
	}
	state.eventID = ""
	state.notAlertingSince = time.Time{}
}

// fireLocked 创建告警事件，返回事件 ID（调用方需已持有 e.mu）
func (e *Engine) fireLocked(rule models.AlertRule, res *datasource.InstantQueryResult, now time.Time) (string, error) {
	event := models.AlertEvent{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Type:        models.AlertEventTypeAlert,
		Status:      models.AlertEventStatusFiring,
		TriggerTime: now,
	}
	if len(res.Samples) > 0 {
		s := res.Samples[0]
		event.TargetIdent = targetIdent(s.Labels)
		event.Tags = serializeTags(s.Labels)
		event.TriggerValue = fmt.Sprintf("%g", s.Value)
	}
	if err := e.db.Create(&event).Error; err != nil {
		return "", err
	}
	fmt.Printf("[告警引擎] 规则 %s 触发告警事件: %s\n", rule.Name, event.ID)
	return event.ID, nil
}

// recoverLocked 把未恢复事件置为已恢复，并创建告警恢复事件（调用方需已持有 e.mu）
func (e *Engine) recoverLocked(rule models.AlertRule, now time.Time) error {
	var firing models.AlertEvent
	err := e.db.Where("rule_id = ? AND type = ? AND status = ?", rule.ID, models.AlertEventTypeAlert, models.AlertEventStatusFiring).
		Order("trigger_time DESC").First(&firing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if err := e.db.Model(&firing).Updates(map[string]interface{}{
		"status":       models.AlertEventStatusResolved,
		"recovered_at": &now,
	}).Error; err != nil {
		return err
	}

	recovery := models.AlertEvent{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Type:        models.AlertEventTypeRecovery,
		Status:      models.AlertEventStatusResolved,
		TargetIdent: firing.TargetIdent,
		Tags:        firing.Tags,
		TriggerTime: now,
	}
	if err := e.db.Create(&recovery).Error; err != nil {
		return err
	}
	fmt.Printf("[告警引擎] 规则 %s 触发恢复事件: %s\n", rule.Name, recovery.ID)
	return nil
}

// findPrometheus 查找第一个启用的 Prometheus 数据源
func (e *Engine) findPrometheus(ctx context.Context) (*models.Datasource, error) {
	manager := datasource.NewManager(e.db)
	list, err := manager.LoadEnabled(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Type == datasource.TypePrometheus {
			return &list[i], nil
		}
	}
	return nil, fmt.Errorf("未找到启用的 Prometheus 数据源")
}

// targetIdent 提取告警对象：优先 instance，其次第一个标签值
func targetIdent(labels map[string]string) string {
	if v := labels["instance"]; v != "" {
		return v
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 0 {
		return labels[keys[0]]
	}
	return ""
}

// serializeTags 将标签序列化为 k=v 逗号分隔
func serializeTags(labels map[string]string) string {
	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}
