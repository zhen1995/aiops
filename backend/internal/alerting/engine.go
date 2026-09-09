// Package alerting 告警规则评估引擎：按每条规则的执行频率探测 PromQL，
// 每条满足条件的时序序列独立计数：持续命中达到持续时间后自动生成告警事件，
// 持续未命中达到持续时间后自动生成告警恢复事件；持续时间为 0 时表示命中/未命中一次即触发。
package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"aiops/internal/datasource"
	"aiops/internal/denoise"
	"aiops/internal/notify"
	"aiops/models"

	"gorm.io/gorm"
)

// defaultEvalInterval 规则未配置执行频率时的默认值
const defaultEvalInterval = 30

// seriesState 单个时序序列（标签组合）的评估状态
type seriesState struct {
	alertingSince    time.Time // 本次持续命中的起始时间
	notAlertingSince time.Time // 本次持续未命用的起始时间（用于恢复判断）
	eventID          string    // 当前未恢复的告警事件 ID，空表示当前无未恢复事件
}

// ruleState 每条规则的内存评估状态（key 为序列标签指纹）
type ruleState struct {
	series map[string]*seriesState
}

// Engine 告警规则评估引擎
type Engine struct {
	db       *gorm.DB
	domain   string // 站点地址（用于模板 $.domain 变量）
	denoise  *denoise.DenoiseService // 可选降噪依赖（nil 时跳过降噪）
	stop     chan struct{}

	mu       sync.Mutex
	states   map[string]*ruleState  // ruleID → 评估状态
	stopChan map[string]chan struct{} // ruleID → 该规则评估循环的停止通道
}

// NewEngine 创建评估引擎
func NewEngine(db *gorm.DB, domain string) *Engine {
	return &Engine{
		db:       db,
		domain:   domain,
		stop:     make(chan struct{}),
		states:   make(map[string]*ruleState),
		stopChan: make(map[string]chan struct{}),
	}
}

// SetDenoise 注入降噪服务（由 main.go 在引擎启动后调用；denoise 包不反向依赖本包）
func (e *Engine) SetDenoise(svc *denoise.DenoiseService) {
	e.denoise = svc
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

// ClearSeries 清理某条序列的评估状态（如对应告警事件被人工删除后调用），
// 之后若条件仍满足，引擎会按持续时间重新触发告警
func (e *Engine) ClearSeries(ruleID, tags string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if state, ok := e.states[ruleID]; ok {
		delete(state.series, tags)
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
		state = &ruleState{series: make(map[string]*seriesState)}
		e.states[rule.ID] = state
	}

	// 本次探测命中的序列：标签指纹 → 样本
	hits := make(map[string]datasource.InstantSample, len(res.Samples))
	for _, s := range res.Samples {
		hits[serializeTags(s.Labels)] = s
	}

	// 命中序列：若有未恢复事件，重置恢复计时；否则累计触发计时
	for key, sample := range hits {
		st := state.series[key]
		if st == nil {
			st = &seriesState{}
			state.series[key] = st
		}
		st.notAlertingSince = time.Time{}
		if st.eventID != "" {
			continue
		}
		if rule.Duration > 0 {
			// 持续时间 > 0：需持续命中达到持续时间后才触发
			if st.alertingSince.IsZero() {
				st.alertingSince = now
				continue
			}
			if now.Sub(st.alertingSince) < time.Duration(rule.Duration)*time.Second {
				continue
			}
		}
		// 优先复用同规则同标签仍未恢复的事件（如引擎重启后内存状态丢失），避免重复告警
		eventID := e.findFiringEventLocked(rule.ID, key)
		if eventID == "" {
			var err error
			eventID, err = e.fireLocked(rule, sample, now)
			if err != nil {
				fmt.Printf("[告警引擎] 创建告警事件失败(规则 %s): %v\n", rule.Name, err)
				continue
			}
		}
		st.eventID = eventID
		st.alertingSince = time.Time{}
	}

	// 重复通知检查：对所有 firing 中的序列，判断是否到了重复间隔、是否超最大次数
	for _, st := range state.series {
		if st.eventID == "" {
			continue
		}
		go e.maybeResend(st.eventID, rule)
	}

	// 已知但未命中的序列：重置触发计时；若有未恢复事件，累计恢复计时
	for key, st := range state.series {
		if _, ok := hits[key]; ok {
			continue
		}
		st.alertingSince = time.Time{}
		if st.eventID == "" {
			continue
		}
		if rule.Duration > 0 {
			// 持续时间 > 0：需持续未命中达到持续时间后才恢复
			if st.notAlertingSince.IsZero() {
				st.notAlertingSince = now
				continue
			}
			if now.Sub(st.notAlertingSince) < time.Duration(rule.Duration)*time.Second {
				continue
			}
		}
		// 持续时间 = 0：未命中一次即恢复
		if err := e.recoverLocked(rule, st.eventID, now); err != nil {
			fmt.Printf("[告警引擎] 创建恢复事件失败(规则 %s): %v\n", rule.Name, err)
			continue
		}
		st.eventID = ""
		st.notAlertingSince = time.Time{}
	}
}

// findFiringEventLocked 查找同规则同标签仍未恢复的告警事件 ID，无则返回空串（调用方需已持有 e.mu）
func (e *Engine) findFiringEventLocked(ruleID, tags string) string {
	var ev models.AlertEvent
	err := e.db.Where("rule_id = ? AND type = ? AND status = ? AND tags = ?", ruleID, models.AlertEventTypeAlert, models.AlertEventStatusFiring, tags).
		Order("trigger_time DESC").First(&ev).Error
	if err != nil {
		return ""
	}
	return ev.ID
}

// fireLocked 创建告警事件，返回事件 ID（调用方需已持有 e.mu）
func (e *Engine) fireLocked(rule models.AlertRule, sample datasource.InstantSample, now time.Time) (string, error) {
	targetIdent := targetIdent(sample.Labels)
	tags := serializeTags(sample.Labels)

	// 降噪：时间窗口聚合 → 拓扑抑制（均为独立 db 查询，不会与 e.mu 死锁）
	if e.denoise != nil {
		if e.denoise.SuppressWindow(rule.ID, tags, now) {
			e.denoise.Record(models.DenoiseStrategyWindowAggregation, rule.ID, rule.Name, rule.Severity, targetIdent, tags)
			fmt.Printf("[告警降噪] 规则 %s (%s) 命中时间窗口聚合，已拦截\n", rule.Name, targetIdent)
			return "", nil
		}
		if e.denoise.SuppressTopology(tags) {
			e.denoise.Record(models.DenoiseStrategyTopologySuppression, rule.ID, rule.Name, rule.Severity, targetIdent, tags)
			fmt.Printf("[告警降噪] 规则 %s (%s) 命中拓扑抑制，已拦截\n", rule.Name, targetIdent)
			return "", nil
		}
	}

	event := models.AlertEvent{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Severity:    rule.Severity,
		Type:        models.AlertEventTypeAlert,
		Status:      models.AlertEventStatusFiring,
		TargetIdent: targetIdent,
		Tags:        tags,
		TriggerValue: fmt.Sprintf("%g", sample.Value),
		TriggerTime: now,
	}
	if err := e.db.Create(&event).Error; err != nil {
		return "", err
	}
	fmt.Printf("[告警引擎] 规则 %s 触发告警事件: %s (%s)\n", rule.Name, event.ID, event.TargetIdent)

	// 异步发送通知（在 goroutine 里，不持有 e.mu 锁）
	if rule.NotifyRuleID != "" {
		evtCopy := event
		go e.dispatch(evtCopy, rule)
	}
	return event.ID, nil
}

// recoverLocked 把指定未恢复事件置为已恢复，并创建告警恢复事件（调用方需已持有 e.mu）
func (e *Engine) recoverLocked(rule models.AlertRule, eventID string, now time.Time) error {
	var firing models.AlertEvent
	err := e.db.Where("id = ? AND type = ? AND status = ?", eventID, models.AlertEventTypeAlert, models.AlertEventStatusFiring).
		First(&firing).Error
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
	fmt.Printf("[告警引擎] 规则 %s 触发恢复事件: %s (%s)\n", rule.Name, recovery.ID, recovery.TargetIdent)

	// 异步发送通知
	if rule.NotifyRuleID != "" {
		evtCopy := recovery
		go e.dispatch(evtCopy, rule)
	}
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

// maybeResend 判断告警事件是否需要发送重复通知，需要则调用 dispatch 发送。
// 必须在 goroutine 中调用（不持有 e.mu 锁）
func (e *Engine) maybeResend(eventID string, rule models.AlertRule) {
	// 重复间隔为 0 表示不重复，仅首次 fire 时发送
	if rule.RepeatIntervalMinutes <= 0 {
		return
	}

	var ev models.AlertEvent
	if err := e.db.Where("id = ?", eventID).First(&ev).Error; err != nil {
		return
	}
	if ev.Status != models.AlertEventStatusFiring {
		return // 已恢复，不重复
	}

	// 最大发送次数限制（0 = 不限制）
	if rule.MaxSendCount > 0 && ev.NotifyCount >= rule.MaxSendCount {
		return
	}

	// 第一次发送（首次 fire 已发过）不在这里处理，由 fireLocked 里的 dispatch 完成
	// 这里只处理后续重复发送
	if ev.NotifyCount == 0 {
		return
	}

	interval := time.Duration(rule.RepeatIntervalMinutes) * time.Minute
	if ev.LastNotifiedAt != nil && time.Since(*ev.LastNotifiedAt) < interval {
		return // 还没到重复间隔时间
	}

	// 到了重复间隔，发送
	fmt.Printf("[通知] 规则 %s 事件 %s 触发重复通知（已发送 %d 次）\n", rule.Name, eventID, ev.NotifyCount)
	e.dispatch(ev, rule)
}

// dispatch 完整通知链路：NotifyRule → Media + Template → 渲染 → 发送
// 必须在 goroutine 里调用（不持有 e.mu 锁）
func (e *Engine) dispatch(event models.AlertEvent, rule models.AlertRule) {
	logPrefix := fmt.Sprintf("[通知] 规则 %s 事件 %s", rule.Name, event.ID)

	// 1. 加载通知规则
	var nr models.NotifyRule
	if err := e.db.Where("id = ? AND is_enabled = ?", rule.NotifyRuleID, 1).First(&nr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("%s 通知规则 %s 不存在或已停用，跳过\n", logPrefix, rule.NotifyRuleID)
			return
		}
		fmt.Printf("%s 加载通知规则失败: %v\n", logPrefix, err)
		return
	}

	// 2. 检查 TriggerTypes 匹配
	eventType := "firing"
	if event.Type == models.AlertEventTypeRecovery {
		eventType = "recovered"
	}
	if !triggerTypeMatches(nr.TriggerTypes, eventType) {
		fmt.Printf("%s 通知规则触发类型=%s，事件类型=%s，跳过\n", logPrefix, nr.TriggerTypes, eventType)
		return
	}

	// 3. 检查 SeverityFilter
	if nr.SeverityFilter != "" {
		var filter []int
		if err := json.Unmarshal([]byte(nr.SeverityFilter), &filter); err == nil {
			matched := false
			for _, s := range filter {
				if s == event.Severity {
					matched = true
					break
				}
			}
			if !matched {
				fmt.Printf("%s 事件级别 %d 不在过滤列表 %v 中，跳过\n", logPrefix, event.Severity, filter)
				return
			}
		}
	}

	// 4. 加载通知媒介
	var media models.NotifyMedia
	if err := e.db.Where("id = ? AND is_enabled = ?", nr.MediaID, 1).First(&media).Error; err != nil {
		fmt.Printf("%s 加载媒介 %s 失败: %v\n", logPrefix, nr.MediaID, err)
		return
	}

	// 5. 加载消息模板（优先匹配事件类型，否则用通用模板）
	var tpl models.NotifyTemplate
	tplQuery := e.db.Where("id = ?", nr.TemplateID)
	if err := tplQuery.First(&tpl).Error; err != nil {
		fmt.Printf("%s 加载模板 %s 失败: %v\n", logPrefix, nr.TemplateID, err)
		return
	}
	// 模板停用时也跳过
	if tpl.IsEnabled != 1 {
		fmt.Printf("%s 模板 %s 已停用，跳过\n", logPrefix, tpl.ID)
		return
	}
	// 模板 Type 匹配检查：firing 模板只给 firing 事件，recovered 模板只给 recovered，all 通用
	if tpl.Type != "all" && tpl.Type != eventType {
		fmt.Printf("%s 模板类型=%s 事件类型=%s，跳过\n", logPrefix, tpl.Type, eventType)
		return
	}

	// 6. 渲染模板
	evt := buildAlertEvent(event, rule)
	content, err := notify.RenderTemplate(tpl.Content, evt, e.domain)
	if err != nil {
		fmt.Printf("%s 模板渲染失败: %v\n", logPrefix, err)
		return
	}

	// 7. 确定 msgtype：模板用 markdown 语法 + 钉钉媒介 → markdown；其他 → text
	msgType := "text"
	if media.Type == notify.TypeDingtalk {
		msgType = "markdown"
	}

	// 8. 发送
	if err := notify.SendTyped(media.Type, media.Config, content, msgType); err != nil {
		fmt.Printf("%s 发送失败（媒介=%s 模板=%s 媒介名=%s）: %v\n", logPrefix, media.Type, tpl.Name, media.Name, err)
		return
	}
	fmt.Printf("%s ✅ 发送成功（媒介=%s 模板=%s 媒介名=%s msgtype=%s）\n", logPrefix, media.Type, tpl.Name, media.Name, msgType)

	// 9. 更新事件的通知次数和时间（频控）
	now := time.Now()
	if err := e.db.Model(&models.AlertEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{
		"notify_count":      gorm.Expr("notify_count + 1"),
		"last_notified_at":  now,
	}).Error; err != nil {
		fmt.Printf("%s 更新通知频控状态失败: %v\n", logPrefix, err)
	}
}

// triggerTypeMatches 判断通知规则的触发类型是否匹配事件类型
// nrTrigger: "all" / "firing" / "recovered"
// eventType: "firing" / "recovered"
func triggerTypeMatches(nrTrigger, eventType string) bool {
	switch nrTrigger {
	case "", "all":
		return true
	case "firing":
		return eventType == "firing"
	case "recovered":
		return eventType == "recovered"
	}
	return true
}

// parseSeverityFilter 解析 SeverityFilter JSON 数组
func parseSeverityFilter(raw string) []int {
	if raw == "" {
		return nil
	}
	var out []int
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

// buildAlertEvent 把 models.AlertEvent + AlertRule 组装成 notify.AlertEvent（模板引擎上下文）
func buildAlertEvent(ev models.AlertEvent, rule models.AlertRule) *notify.AlertEvent {
	tags := parseSerializedTags(ev.Tags)
	sevLabel := "P3-提醒"
	switch ev.Severity {
	case 1:
		sevLabel = "P1-紧急"
	case 2:
		sevLabel = "P2-警告"
	case 3:
		sevLabel = "P3-提醒"
	}

	isRecovered := ev.Type == models.AlertEventTypeRecovery
	triggerTime := ev.TriggerTime
	lastEval := ev.TriggerTime
	firstTrigger := ev.TriggerTime
	if ev.Type == models.AlertEventTypeAlert {
		// firing 事件：LastEval 就是 TriggerTime；FirstTrigger 也是 TriggerTime（简化版，后续可以加 first_trigger 字段）
		lastEval = ev.TriggerTime
		firstTrigger = ev.TriggerTime
	} else if ev.RecoveredAt != nil {
		// recovery 事件：用 RecoveredAt
		triggerTime = *ev.RecoveredAt
		lastEval = *ev.RecoveredAt
	}

	// 计算持续时长
	durationSec := int64(0)
	if ev.RecoveredAt != nil {
		durationSec = int64(ev.RecoveredAt.Sub(ev.TriggerTime).Seconds())
	} else {
		durationSec = int64(time.Since(ev.TriggerTime).Seconds())
	}

	id := ev.ID
	if id == "" {
		id = rule.ID + ":" + ev.TargetIdent
	}

	return &notify.AlertEvent{
		Id:             id,
		RuleID:         0,
		RuleName:       rule.Name,
		RuleNote:       "",
		Cluster:        "",
		BusiGroupID:    0,
		BusiGroupName:  "",
		Severity:       ev.Severity,
		SeverityLabel:  sevLabel,
		TriggerValue:   ev.TriggerValue,
		TriggerTime:    triggerTime,
		FirstTrigger:   firstTrigger,
		LastTrigger:    lastEval,
		LastEvalTime:   lastEval,
		IsRecovered:    isRecovered,
		RecoverTime:    lastEval,
		DurationSec:    durationSec,
		Location:       tags["location"],
		Cate:           "prometheus",
		TargetIdent:    ev.TargetIdent,
		Datasource:     "",
		Tags:           tags,
		TagsMap:        tags,
		TagsJSON:       formatTagsJSON(tags),
		AnnotationsJSON: map[string]string{},
	}
}

// parseSerializedTags 解析 "k=v,k2=v2" 格式的标签字符串
func parseSerializedTags(s string) map[string]string {
	out := map[string]string{}
	if s == "" {
		return out
	}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		idx := strings.Index(pair, "=")
		if idx < 0 {
			out[pair] = ""
		} else {
			out[pair[:idx]] = pair[idx+1:]
		}
	}
	return out
}

// formatTagsJSON 把 tag map 格式化成紧凑 JSON（供模板 {{$event.TagsJSON}} 使用）
func formatTagsJSON(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(`"`)
		b.WriteString(k)
		b.WriteString(`": "`)
		b.WriteString(strings.ReplaceAll(tags[k], `"`, `\"`))
		b.WriteString(`"`)
	}
	b.WriteByte('}')
	return b.String()
}

// parseIntSafe 辅助
func parseIntSafe(s string) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return 0
}
