package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aiops/internal/n9e"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// n9eTools 返回所有夜莺工具的 builder
func (r *Registry) n9eTools() []func() (tool.InvokableTool, error) {
	return []func() (tool.InvokableTool, error){
		r.queryN9eAlertRulesTool,
		r.queryN9eCurEventsTool,
		r.queryN9eHisEventsTool,
	}
}

// buildN9eClient 从 DB 加载启用的夜莺引擎配置
func (r *Registry) buildN9eClient(ctx context.Context) (*n9e.Client, error) {
	if r.db == nil {
		return nil, fmt.Errorf("工具注册表未初始化数据库连接")
	}
	return n9e.NewClientFromDB(r.db)
}

// ---------- query_n9e_alert_rules ----------

type QueryN9eAlertRulesInput struct {
	RuleName    string `json:"rule_name,omitempty" jsonschema:"description=按规则名称关键词过滤，例如 磁盘、CPU、MySQL。空表示返回全部"`
	BusiGroupID int64  `json:"busi_group_id,omitempty" jsonschema:"description=按业务组 ID 过滤，空表示不过滤"`
}

type N9eAlertRuleBrief struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	BusiGroupID   int      `json:"busi_group_id"`
	BusiGroupName string   `json:"busi_group_name,omitempty"`
	Severity      int      `json:"severity"`
	PromQL        string   `json:"prom_ql"`
	DurationSec   int      `json:"prom_for_duration"`
	Enabled       bool     `json:"enabled"`
	Cluster       string   `json:"cluster,omitempty"`
	NotifyRecover bool     `json:"notify_recovered"`
	NotifyChan    []string `json:"notify_channels,omitempty"`
}

func (r *Registry) queryN9eAlertRulesTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryN9eAlertRulesInput, []N9eAlertRuleBrief](
		"query_n9e_alert_rules",
		"查询夜莺告警规则列表。可以按规则名称关键词（例如 磁盘、CPU、MySQL）和业务组 ID 过滤。返回规则名称、PromQL 表达式、持续时长、告警级别、是否启用等核心字段。当用户询问当前配置了哪些告警规则、某个系统有哪些告警规则、需要排查告警规则配置时调用此工具。",
		func(ctx context.Context, in QueryN9eAlertRulesInput) ([]N9eAlertRuleBrief, error) {
			client, err := r.buildN9eClient(ctx)
			if err != nil {
				return nil, err
			}
			body, _, err := client.GetAlertRules(ctx)
			if err != nil {
				return nil, err
			}
			raw, err := extractN9eList(body)
			if err != nil {
				return nil, err
			}
			out := make([]N9eAlertRuleBrief, 0, len(raw))
			for _, m := range raw {
				if in.RuleName != "" && !stringsContains(strVal(m, "name"), in.RuleName) {
					continue
				}
				gid := int64Val(m, "group_id")
				if in.BusiGroupID > 0 && gid != in.BusiGroupID {
					continue
				}
				out = append(out, N9eAlertRuleBrief{
					ID:            intVal(m, "id"),
					Name:          strVal(m, "name"),
					BusiGroupID:   int(gid),
					BusiGroupName: strVal(m, "group_name"),
					Severity:      firstSeverity(m["severities"]),
					PromQL:        extractPromQL(m),
					DurationSec:   intVal(m, "prom_for_duration"),
					Enabled:       !boolVal(m, "disabled"),
					Cluster:       strVal(m, "cluster"),
					NotifyRecover: boolVal(m, "notify_recovered"),
					NotifyChan:    strArrVal(m, "notify_channels"),
				})
			}
			return out, nil
		},
	)
}

// ---------- query_n9e_cur_events ----------

type QueryN9eCurEventsInput struct {
	Limit       int    `json:"limit,omitempty" jsonschema:"description=最多返回多少条，默认 20，最大 50"`
	RuleName    string `json:"rule_name,omitempty" jsonschema:"description=按告警规则名称关键词过滤"`
	BusiGroupID int64  `json:"busi_group_id,omitempty" jsonschema:"description=按业务组 ID 过滤"`
}

type N9eEventBrief struct {
	ID           int      `json:"id"`
	RuleID       int      `json:"rule_id"`
	RuleName     string   `json:"rule_name"`
	GroupID      int      `json:"group_id"`
	GroupName    string   `json:"group_name,omitempty"`
	Cluster      string   `json:"cluster"`
	Severity     int      `json:"severity"`
	TriggerTime  int64    `json:"trigger_time_unix"`
	FirstTrigger int64    `json:"first_trigger_unix"`
	TriggerValue string   `json:"trigger_value"`
	IsRecovered  bool     `json:"is_recovered"`
	Tags         []string `json:"tags"`
}

func (r *Registry) queryN9eCurEventsTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryN9eCurEventsInput, []N9eEventBrief](
		"query_n9e_cur_events",
		"查询夜莺当前活跃（未恢复）的告警事件。返回告警规则名称、触发时间（Unix 秒）、业务组、告警级别、标签（含实例名、IP、服务名等）。当用户询问当前有哪些告警在发生、P1/P2 告警情况、某个服务是否在告警时调用此工具。",
		func(ctx context.Context, in QueryN9eCurEventsInput) ([]N9eEventBrief, error) {
			client, err := r.buildN9eClient(ctx)
			if err != nil {
				return nil, err
			}
			limit := in.Limit
			if limit <= 0 || limit > 50 {
				limit = 20
			}
			body, _, err := client.GetCurEvents(ctx, n9e.CurEventsParams{P: 1, Limit: limit, MyGroups: true})
			if err != nil {
				return nil, err
			}
			return parseEvents(body, in.RuleName, in.BusiGroupID)
		},
	)
}

// ---------- query_n9e_his_events ----------

type QueryN9eHisEventsInput struct {
	StartTimeUnix int64  `json:"start_time,omitempty" jsonschema:"description=查询起始时间，Unix 秒级时间戳。默认 24 小时前"`
	EndTimeUnix   int64  `json:"end_time,omitempty" jsonschema:"description=查询结束时间，Unix 秒级时间戳。默认现在"`
	Limit         int    `json:"limit,omitempty" jsonschema:"description=最多返回多少条，默认 20，最大 100"`
	RuleName      string `json:"rule_name,omitempty" jsonschema:"description=按告警规则名称关键词过滤"`
	BusiGroupID   int64  `json:"busi_group_id,omitempty" jsonschema:"description=按业务组 ID 过滤"`
}

func (r *Registry) queryN9eHisEventsTool() (tool.InvokableTool, error) {
	return utils.InferTool[QueryN9eHisEventsInput, []N9eEventBrief](
		"query_n9e_his_events",
		"查询夜莺历史告警事件（已恢复或已触发过的）。需要提供时间范围（Unix 秒级时间戳）。返回告警规则名称、首次触发时间、恢复时间（如果已恢复）、业务组、告警级别。当用户询问过去一段时间发生过哪些告警、某条告警的完整生命周期、某个告警规则最近触发频率时调用此工具。",
		func(ctx context.Context, in QueryN9eHisEventsInput) ([]N9eEventBrief, error) {
			client, err := r.buildN9eClient(ctx)
			if err != nil {
				return nil, err
			}
			now := time.Now().Unix()
			if in.StartTimeUnix == 0 {
				in.StartTimeUnix = now - 24*3600
			}
			if in.EndTimeUnix == 0 {
				in.EndTimeUnix = now
			}
			limit := in.Limit
			if limit <= 0 || limit > 100 {
				limit = 20
			}
			body, _, err := client.GetHisEvents(ctx, n9e.HisEventsParams{
				P: 1, Limit: limit,
				Stime: in.StartTimeUnix, Etime: in.EndTimeUnix,
			})
			if err != nil {
				return nil, err
			}
			return parseEvents(body, in.RuleName, in.BusiGroupID)
		},
	)
}

// ---------- 解析辅助 ----------

func extractN9eList(body []byte) ([]map[string]interface{}, error) {
	var root interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	return extractListDeep(root), nil
}

func extractListDeep(v interface{}) []map[string]interface{} {
	switch t := v.(type) {
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	case map[string]interface{}:
		for _, key := range []string{"list", "data", "dat", "items", "records"} {
			if sub := extractListDeep(t[key]); len(sub) > 0 {
				return sub
			}
		}
	}
	return nil
}

func parseEvents(body []byte, ruleName string, busiGroupID int64) ([]N9eEventBrief, error) {
	raw, err := extractN9eList(body)
	if err != nil {
		return nil, err
	}
	out := make([]N9eEventBrief, 0, len(raw))
	for _, m := range raw {
		if ruleName != "" && !stringsContains(strVal(m, "rule_name"), ruleName) {
			continue
		}
		gid := int64Val(m, "group_id")
		if busiGroupID > 0 && gid != busiGroupID {
			continue
		}
		out = append(out, N9eEventBrief{
			ID:           intVal(m, "id"),
			RuleID:       intVal(m, "rule_id"),
			RuleName:     strVal(m, "rule_name"),
			GroupID:      int(gid),
			GroupName:    strVal(m, "group_name"),
			Cluster:      strVal(m, "cluster"),
			Severity:     intVal(m, "severity"),
			TriggerTime:  int64Val(m, "trigger_time"),
			FirstTrigger: int64Val(m, "first_trigger_time"),
			TriggerValue: strVal(m, "trigger_value"),
			IsRecovered:  boolVal(m, "is_recovered"),
			Tags:         strArrVal(m, "tags"),
		})
	}
	return out, nil
}

// ---------- 类型转换 ----------

func strVal(m map[string]interface{}, k string) string {
	if s, ok := m[k].(string); ok {
		return s
	}
	return ""
}

func intVal(m map[string]interface{}, k string) int {
	return int(int64Val(m, k))
}

func int64Val(m map[string]interface{}, k string) int64 {
	switch v := m[k].(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

func boolVal(m map[string]interface{}, k string) bool {
	switch v := m[k].(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case int:
		return v != 0
	}
	return false
}

func strArrVal(m map[string]interface{}, k string) []string {
	switch v := m[k].(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	}
	return nil
}

func stringsContains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func extractPromQL(m map[string]interface{}) string {
	if rc, ok := m["rule_config"].(map[string]interface{}); ok {
		if qs, ok := rc["queries"].([]interface{}); ok && len(qs) > 0 {
			if q0, ok := qs[0].(map[string]interface{}); ok {
				if p, ok := q0["prom_ql"].(string); ok && p != "" {
					return p
				}
			}
		}
	}
	if p, ok := m["prom_ql"].(string); ok {
		return p
	}
	return ""
}

func firstSeverity(v interface{}) int {
	if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
		return int64ToInt(arr[0])
	}
	return int64ToInt(v)
}

func int64ToInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	}
	return 0
}
