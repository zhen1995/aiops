package notify

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// ---------- 告警事件上下文（模板变量） ----------

// AlertEvent 传给 Go template 的告警事件结构体。
// 字段命名参考夜莺 models/AlertCurEvent，保证默认模板可直接复用夜莺语法。
type AlertEvent struct {
	Id             string            `json:"id"`               // 事件唯一 ID（夜莺事件 id / 本地告警事件 id）
	RuleID         int               `json:"rule_id"`
	RuleName       string            `json:"rule_name"`
	RuleNote       string            `json:"rule_note"`
	Cluster        string            `json:"cluster"`
	BusiGroupID    int               `json:"busi_group_id"`
	BusiGroupName  string            `json:"busi_group_name"`
	Severity       int               `json:"severity"`
	SeverityLabel  string            `json:"severity_label"` // P1-紧急 / P2-警告 / P3-提醒
	TriggerValue   string            `json:"trigger_value"`
	TriggerTime    time.Time         `json:"trigger_time"`
	FirstTrigger   time.Time         `json:"first_trigger_time"`
	LastTrigger    time.Time         `json:"last_trigger_time"` // 触发告警时最近一次 PromQL 命中时间
	LastEvalTime   time.Time         `json:"last_eval_time"`    // 别名，夜莺模板用
	IsRecovered    bool              `json:"is_recovered"`
	RecoverTime    time.Time         `json:"recover_time"`
	DurationSec    int64             `json:"duration_sec"` // 持续时长（秒）
	Location       string            `json:"location"`     // labels 中的 location 字段
	Cate           string            `json:"cate"`         // prometheus / host 等
	TargetIdent    string            `json:"target_ident"` // 被监控对象标识
	TargetNote     string            `json:"target_note"`
	Datasource     string            `json:"datasource"`
	Tags           map[string]string `json:"tags"`      // 标签 key→value（旧模板用）
	TagsMap        map[string]string `json:"tags_map"`  // 别名，夜莺模板用
	TagsJSON       string            `json:"tags_json"` // 标签 JSON 字符串
	AnnotationsJSON map[string]string `json:"annotations_json"` // 附加 annotations
	Extra          map[string]interface{} `json:"extra"` // 原始扩展字段
}

// ParseEventFromRaw 从夜莺事件原始 map 构造 AlertEvent
func ParseEventFromRaw(raw map[string]interface{}) *AlertEvent {
	e := &AlertEvent{
		Id:            strVal(raw, "id"),
		RuleID:        intVal(raw, "rule_id"),
		RuleName:      strVal(raw, "rule_name"),
		RuleNote:      strVal(raw, "rule_note"),
		Cluster:       strVal(raw, "cluster"),
		BusiGroupID:   intVal(raw, "group_id"),
		BusiGroupName: strVal(raw, "group_name"),
		Severity:      intVal(raw, "severity"),
		TriggerValue:  strVal(raw, "trigger_value"),
		IsRecovered:   boolVal(raw, "is_recovered"),
		Datasource:    strVal(raw, "datasource_id"),
		Cate:          strVal(raw, "cate"),
		TargetIdent:   strVal(raw, "target_ident"),
		TargetNote:    strVal(raw, "target_note"),
	}
	if e.Id == "" {
		// 夜莺的 id 是数字，会以 float64 进来
		if v, ok := raw["id"]; ok {
			switch t := v.(type) {
			case float64:
				e.Id = strconv.FormatInt(int64(t), 10)
			case int:
				e.Id = strconv.Itoa(t)
			case string:
				e.Id = t
			}
		}
	}
	e.SeverityLabel = severityLabel(e.Severity)

	if t, ok := parseUnixTime(raw, "trigger_time"); ok {
		e.TriggerTime = t
	}
	if t, ok := parseUnixTime(raw, "first_trigger_time"); ok {
		e.FirstTrigger = t
	}
	if t, ok := parseUnixTime(raw, "last_trigger_time"); ok {
		e.LastTrigger = t
		e.LastEvalTime = t
	} else if t, ok := parseUnixTime(raw, "last_eval_time"); ok {
		e.LastTrigger = t
		e.LastEvalTime = t
	}
	if t, ok := parseUnixTime(raw, "recover_time"); ok {
		e.RecoverTime = t
	}

	if e.FirstTrigger.IsZero() {
		e.DurationSec = 0
	} else {
		end := e.RecoverTime
		if e.IsRecovered || end.IsZero() {
			end = time.Now()
		}
		e.DurationSec = int64(end.Sub(e.FirstTrigger).Seconds())
	}

	// tags
	e.Tags = parseTags(raw)
	e.TagsMap = e.Tags // 别名
	if loc, ok := e.Tags["location"]; ok {
		e.Location = loc
	}
	e.TagsJSON = tagsJSON(e.Tags)
	e.AnnotationsJSON = parseAnnotations(raw)

	return e
}

// parseAnnotations 解析夜莺 annotations（map[string]interface{} 或 JSON string）
func parseAnnotations(raw map[string]interface{}) map[string]string {
	out := map[string]string{}
	if v, ok := raw["annotations"].(map[string]interface{}); ok {
		for k, val := range v {
			out[k] = fmt.Sprintf("%v", val)
		}
	}
	return out
}

// ---------- Template 引擎 ----------

// 默认全局函数（给模板用），命名对齐夜莺 tplx
var templateFuncMap = template.FuncMap{
	// ---- 时间 ----
	"timeformat": func(t time.Time, layout string) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format(layout)
	},
	"timeformatCN": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format("2006-01-02 15:04:05")
	},
	"timeformatDay": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format("01-02 15:04")
	},
	"timestamp": func() string { return time.Now().Format("2006-01-02 15:04:05") },

	// ---- 时间戳（秒） ----
	"now": func() int64 { return time.Now().Unix() },

	// ---- 算术 ----
	"add": func(a, b int64) int64 { return a + b },
	"sub": func(a, b int64) int64 { return a - b },
	"mul": func(a, b int64) int64 { return a * b },

	// ---- 严重级别 ----
	"severityLabel": func(level int) string { return severityLabel(level) },
	"severityEmoji": func(level int) string {
		switch level {
		case 3:
			return "🔴"
		case 2:
			return "🟠"
		case 1:
			return "🟡"
		}
		return "⚪"
	},
	"isRecoveredText": func(v bool) string {
		if v {
			return "恢复"
		}
		return "触发"
	},

	// ---- 人类可读 ----
	"durationHuman": func(sec int64) string {
		return humanizeDuration(sec)
	},
	"humanizeDurationInterface": func(sec int64) string {
		return humanizeDuration(sec)
	},
	"humanizeDuration": func(sec int64) string {
		return humanizeDuration(sec)
	},

	// ---- 其他 ----
	"tag": func(tags map[string]string, key string) string {
		if v, ok := tags[key]; ok {
			return v
		}
		return ""
	},
	"label": func(tags map[string]string, key string) string {
		if v, ok := tags[key]; ok {
			return v
		}
		return ""
	},
}

func humanizeDuration(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	if sec < 60 {
		return fmt.Sprintf("%d秒", sec)
	}
	if sec < 3600 {
		return fmt.Sprintf("%d分", sec/60)
	}
	if sec < 86400 {
		return fmt.Sprintf("%.1f时", float64(sec)/3600)
	}
	return fmt.Sprintf("%.1f天", float64(sec)/86400)
}

// 模板头部自动注入的 defs —— 让模板里可以直接用 {{$event.xxx}} / {{$labels}} / {{$value}}
const templateDefs = `{{ $events := .events }}
{{ $event := index $events 0 }}
{{ $labels := $event.TagsMap }}
{{ $value := $event.TriggerValue }}
`

// ValidateTemplate 校验模板语法是否合法（自动注入 defs 后再 parse）
func ValidateTemplate(content string) error {
	if content == "" {
		return fmt.Errorf("模板内容不能为空")
	}
	text := templateDefs + content
	_, err := template.New("").Funcs(templateFuncMap).Parse(text)
	if err != nil {
		return fmt.Errorf("模板语法错误: %w", err)
	}
	return nil
}

// RenderTemplate 用 AlertEvent 渲染模板内容，domain 为站点地址（如 http://localhost:5173）。
// 渲染数据结构对齐夜莺：根是 map[string]interface{}{events: []*AlertEvent, domain: string}
func RenderTemplate(content string, evt *AlertEvent, domain string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("模板内容不能为空")
	}
	renderData := map[string]interface{}{
		"events": []*AlertEvent{evt},
		"domain": domain,
	}
	text := templateDefs + content
	tpl, err := template.New("").Funcs(templateFuncMap).Parse(text)
	if err != nil {
		return "", fmt.Errorf("模板语法错误: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}
	return buf.String(), nil
}

// RenderTemplateJSON 返回一个"可预览"的渲染结果和错误信息，适合前端预览
func RenderTemplateJSON(content string, evt *AlertEvent, domain string) map[string]interface{} {
	result := map[string]interface{}{"content": "", "error": ""}
	out, err := RenderTemplate(content, evt, domain)
	if err != nil {
		result["error"] = err.Error()
	} else {
		result["content"] = out
	}
	return result
}

// ---------- 默认模板（夜莺风格，使用 $event 变量） ----------

// DefaultTemplateFiring 默认触发告警模板（标题 + 正文，钉钉 markdown 风格）
const DefaultTemplateFiring = `#### {{if $event.IsRecovered}}<font color="#008800">💚{{$event.RuleName}}</font>{{else}}<font color="#FF0000">💔{{$event.RuleName}}</font>{{end}}
---
{{$time_duration := sub now $event.FirstTrigger.Unix }}{{if $event.IsRecovered}}{{$time_duration = sub $event.LastEvalTime.Unix $event.FirstTrigger.Unix }}{{end}}
- **告警级别**: {{$event.SeverityLabel}}
{{- if $event.RuleNote}}
	- **规则备注**: {{$event.RuleNote}}
{{- end}}
{{- if not $event.IsRecovered}}
- **触发时值**: {{$event.TriggerValue}}
- **触发时间**: {{timeformatCN $event.TriggerTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- else}}
- **恢复时间**: {{timeformatCN $event.LastEvalTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- end}}
- **业务组**: {{$event.BusiGroupName}}
- **告警事件标签**:
{{- range $key, $val := $event.TagsMap}}
	- {{$key}}: {{$val}}
{{- end}}
{{if $event.AnnotationsJSON}}
- **附加信息**:
{{- range $key, $val := $event.AnnotationsJSON}}
	- {{$key}}: {{$val}}
{{- end}}
{{end}}
`

// ---------- 工具 ----------

func parseUnixTime(raw map[string]interface{}, key string) (time.Time, bool) {
	v, ok := raw[key]
	if !ok {
		return time.Time{}, false
	}
	var sec int64
	switch t := v.(type) {
	case float64:
		sec = int64(t)
	case int:
		sec = int64(t)
	case int64:
		sec = t
	case string:
		if n, err := strconv.ParseInt(t, 10, 64); err == nil {
			sec = n
		} else {
			return time.Time{}, false
		}
	default:
		return time.Time{}, false
	}
	// 毫秒级
	if sec > 1e12 {
		sec = sec / 1000
	}
	return time.Unix(sec, 0), true
}

func parseTags(raw map[string]interface{}) map[string]string {
	out := map[string]string{}
	switch t := raw["tags"].(type) {
	case []interface{}:
		for _, item := range t {
			if s, ok := item.(string); ok {
				if idx := strings.Index(s, "="); idx >= 0 {
					out[s[:idx]] = s[idx+1:]
				} else {
					out[s] = ""
				}
			}
		}
	case map[string]interface{}:
		for k, v := range t {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
	}
	return out
}

func tagsJSON(tags map[string]string) string {
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

func severityLabel(level int) string {
	switch level {
	case 3:
		return "P1-紧急"
	case 2:
		return "P2-警告"
	case 1:
		return "P3-提醒"
	}
	return fmt.Sprintf("P%d", level)
}

func intVal(m map[string]interface{}, k string) int {
	switch v := m[k].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

func strVal(m map[string]interface{}, k string) string {
	if s, ok := m[k].(string); ok {
		return s
	}
	return ""
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
