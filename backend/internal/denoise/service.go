// Package denoise 告警降噪：在告警事件落库前按策略拦截，
// 支持时间窗口聚合（同规则同标签在窗口期内合并为一条）与拓扑抑制（父服务故障时抑制子服务告警）。
// 本包不依赖 internal/alerting，由 main.go 创建后注入告警引擎。
package denoise

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aiops/models"

	"gorm.io/gorm"
)

// 默认值：策略表为空或配置解析失败时使用
const (
	defaultWindowMinutes = 5
	maxAncestors         = 20 // 上溯祖先的最大深度，防止脏数据成环导致死循环
)

// DenoiseService 降噪服务：读取策略配置并执行拦截判定
type DenoiseService struct {
	db *gorm.DB
}

// NewService 创建降噪服务
func NewService(db *gorm.DB) *DenoiseService {
	return &DenoiseService{db: db}
}

// windowConfig 窗口聚合策略的配置结构
type windowConfig struct {
	WindowMinutes int `json:"window_minutes"`
}

// policy 读取指定策略；表空或记录不存在时返回默认配置（启用、窗口 5 分钟）
func (s *DenoiseService) policy(strategy string) models.DenoisePolicy {
	p := models.DenoisePolicy{Strategy: strategy, Enabled: true, Config: `{"window_minutes":5}`}
	if s.db == nil {
		return p
	}
	var row models.DenoisePolicy
	if err := s.db.Where("strategy = ?", strategy).First(&row).Error; err != nil {
		return p
	}
	return row
}

// Enabled 指定策略是否启用（默认启用）
func (s *DenoiseService) Enabled(strategy string) bool {
	return s.policy(strategy).Enabled
}

// WindowMinutes 窗口聚合的窗口分钟数（默认 5 分钟）
func (s *DenoiseService) WindowMinutes() int {
	p := s.policy(models.DenoiseStrategyWindowAggregation)
	if p.Config == "" {
		return defaultWindowMinutes
	}
	var cfg windowConfig
	if err := json.Unmarshal([]byte(p.Config), &cfg); err != nil || cfg.WindowMinutes <= 0 {
		return defaultWindowMinutes
	}
	return cfg.WindowMinutes
}

// SuppressWindow 时间窗口聚合：同规则同标签的最近一条事件距 now 小于窗口分钟数则拦截。
// 返回 (是否拦截, 拦截原因)；未拦截时原因为空。
func (s *DenoiseService) SuppressWindow(ruleID, tags string, now time.Time) (bool, string) {
	if !s.Enabled(models.DenoiseStrategyWindowAggregation) {
		return false, ""
	}
	var ev models.AlertEvent
	err := s.db.Where("rule_id = ? AND tags = ?", ruleID, tags).
		Order("trigger_time DESC").First(&ev).Error
	if err != nil {
		return false, ""
	}
	if now.Sub(ev.TriggerTime) < time.Duration(s.WindowMinutes())*time.Minute {
		reason := fmt.Sprintf("命中时间窗口聚合策略：同规则同标签在 %d 分钟窗口内已有告警事件", s.WindowMinutes())
		return true, reason
	}
	return false, ""
}

// SuppressTopology 拓扑抑制：告警标签中的 service 对应服务的任一祖先服务当前存在 firing 告警则拦截。
// 返回 (是否拦截, 拦截原因)；未拦截时原因为空。
func (s *DenoiseService) SuppressTopology(tags string) (bool, string) {
	if !s.Enabled(models.DenoiseStrategyTopologySuppression) {
		return false, ""
	}
	svcName := parseTags(tags)["service"]
	if svcName == "" {
		return false, ""
	}

	// 按 code 或 name 定位服务
	var svc models.Service
	err := s.db.Where("code = ? OR name = ?", svcName, svcName).First(&svc).Error
	if err != nil {
		return false, ""
	}

	// 沿 ParentID 上溯收集祖先的 code/name
	ancestors := make([]string, 0, 4)
	ancestorNames := make([]string, 0, 4)
	visited := map[string]bool{svc.ID: true}
	cur := svc.ParentID
	for cur != nil && *cur != "" && len(ancestors) < maxAncestors && !visited[*cur] {
		visited[*cur] = true
		var parent models.Service
		if err := s.db.First(&parent, "id = ?", *cur).Error; err != nil {
			break
		}
		ancestors = append(ancestors, parent.Code, parent.Name)
		if parent.Name != "" {
			ancestorNames = append(ancestorNames, parent.Name)
		} else if parent.Code != "" {
			ancestorNames = append(ancestorNames, parent.Code)
		}
		cur = parent.ParentID
	}
	if len(ancestors) == 0 {
		return false, ""
	}

	// 任一祖先服务有 firing 告警（其 code/name 出现在某条 firing 事件标签中）则抑制。
	// 参数化查询，避免 OR 条件拼接注入。
	var count int64
	if err := s.db.Model(&models.AlertEvent{}).
		Where("type = ? AND status = ?", models.AlertEventTypeAlert, models.AlertEventStatusFiring).
		Where("("+orLikeClauses("tags", len(ancestors))+")", likeArgs(ancestors)...).
		Count(&count).Error; err != nil {
		return false, ""
	}
	if count > 0 {
		reason := fmt.Sprintf("命中拓扑抑制策略：父服务 %s 存在未恢复告警，级联告警已抑制", strings.Join(ancestorNames, "、"))
		return true, reason
	}
	return false, ""
}

// orLikeClauses 生成 tags LIKE ? OR tags LIKE ? ... 的参数占位串
func orLikeClauses(column string, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = column + " LIKE ?"
	}
	return strings.Join(parts, " OR ")
}

// likeArgs 生成 ["%v1%", "%v2%", ...] 的参数列表
func likeArgs(values []string) []interface{} {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = "%" + v + "%"
	}
	return args
}

// Record 写入一条降噪拦截记录
func (s *DenoiseService) Record(strategy string, ruleID, ruleName string, severity int, targetIdent, tags string) {
	rec := models.DenoiseRecord{
		Strategy:    strategy,
		RuleID:      ruleID,
		RuleName:    ruleName,
		Severity:    severity,
		TargetIdent: targetIdent,
		Tags:        tags,
	}
	if err := s.db.Create(&rec).Error; err != nil {
		// 拦截记录失败不影响主流程，仅打印日志
		println("[告警降噪] 写入拦截记录失败: " + err.Error())
	}
}

// dayStart 本地时区当日 00:00
func dayStart(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

// Stats 今日降噪统计：alert_events 数（有效通知）、按策略分组的拦截数
type Stats struct {
	AlertEvents int64          // 今日告警事件总数（有效通知）
	Suppressed  map[string]int64 // 今日各策略拦截数
}

// TodayStats 统计今日 0 点至今的告警事件数与各策略拦截数
func (s *DenoiseService) TodayStats(now time.Time) (*Stats, error) {
	st := &Stats{Suppressed: map[string]int64{
		models.DenoiseStrategyWindowAggregation:   0,
		models.DenoiseStrategyTopologySuppression: 0,
	}}
	start := dayStart(now)
	if err := s.db.Model(&models.AlertEvent{}).
		Where("trigger_time >= ?", start).
		Count(&st.AlertEvents).Error; err != nil {
		return nil, err
	}
	type row struct {
		Strategy string
		Count    int64
	}
	var rows []row
	if err := s.db.Model(&models.DenoiseRecord{}).
		Select("strategy, COUNT(*) AS count").
		Where("created_at >= ?", start).
		Group("strategy").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		st.Suppressed[r.Strategy] = r.Count
	}
	return st, nil
}

// parseTags 解析 "k=v,k2=v2" 格式的标签字符串（serializeTags 的逆解析）
func parseTags(s string) map[string]string {
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
