package nightingale

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"aiops/models"
)

type Client struct {
	cfg *models.AlertEngineConfig
	cli *http.Client
}

func NewClient(cfg *models.AlertEngineConfig) *Client {
	return &Client{
		cfg: cfg,
		cli: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) baseURL() string {
	u := strings.TrimRight(c.cfg.BaseURL, "/")
	return u + "/api/n9e"
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out interface{}) error {
	reqURL := c.baseURL() + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-Token", c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("夜莺接口返回 %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		Dat json.RawMessage `json:"dat"`
		Err string          `json:"err"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if envelope.Err != "" {
		return fmt.Errorf("夜莺接口错误: %s", envelope.Err)
	}
	return json.Unmarshal(envelope.Dat, out)
}

// AlertRule 只保留前端需要的字段
type AlertRule struct {
	ID               int64  `json:"id"`
	GroupID          int64  `json:"group_id"`
	Name             string `json:"name"`
	Note             string `json:"note"`
	Cate             string `json:"cate"`
	Disabled         int    `json:"disabled"`
	Severity         int    `json:"severity"`
	Severities       []int  `json:"severities"`
	PromQL           string `json:"prom_ql"`
	PromForDuration  int    `json:"prom_for_duration"`
	PromEvalInterval int    `json:"prom_eval_interval"`
	CurEventCount    int    `json:"cur_event_count"`
	CreateAt         int64  `json:"create_at"`
	UpdateAt         int64  `json:"update_at"`
}

// ruleQuery 对应 rule_config.queries 中的单个查询
type ruleQuery struct {
	PromQL string `json:"prom_ql"`
}

// ruleConfig 对应夜莺规则配置，实际 PromQL 存放在 queries 数组中
type ruleConfig struct {
	Queries []ruleQuery `json:"queries"`
}

// rawAlertRule 用于接收夜莺原始字段，字段与 AlertRule 一致，但不带自定义反序列化
type rawAlertRule struct {
	ID               int64      `json:"id"`
	GroupID          int64      `json:"group_id"`
	Name             string     `json:"name"`
	Note             string     `json:"note"`
	Cate             string     `json:"cate"`
	Disabled         int        `json:"disabled"`
	Severity         int        `json:"severity"`
	Severities       []int      `json:"severities"`
	PromQL           string     `json:"prom_ql"`
	PromForDuration  int        `json:"prom_for_duration"`
	PromEvalInterval int        `json:"prom_eval_interval"`
	CurEventCount    int        `json:"cur_event_count"`
	CreateAt         int64      `json:"create_at"`
	UpdateAt         int64      `json:"update_at"`
	RuleConfig       ruleConfig `json:"rule_config"`
}

// UnmarshalJSON 在反序列化后，如果顶层 prom_ql 为空，则从 rule_config.queries 中补回第一个 PromQL
func (r *AlertRule) UnmarshalJSON(data []byte) error {
	var raw rawAlertRule
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = AlertRule{
		ID:               raw.ID,
		GroupID:          raw.GroupID,
		Name:             raw.Name,
		Note:             raw.Note,
		Cate:             raw.Cate,
		Disabled:         raw.Disabled,
		Severity:         raw.Severity,
		Severities:       raw.Severities,
		PromQL:           raw.PromQL,
		PromForDuration:  raw.PromForDuration,
		PromEvalInterval: raw.PromEvalInterval,
		CurEventCount:    raw.CurEventCount,
		CreateAt:         raw.CreateAt,
		UpdateAt:         raw.UpdateAt,
	}
	if r.PromQL == "" && len(raw.RuleConfig.Queries) > 0 {
		r.PromQL = raw.RuleConfig.Queries[0].PromQL
	}
	return nil
}

func (c *Client) ListRules(ctx context.Context, gids string) ([]AlertRule, error) {
	params := url.Values{}
	if strings.TrimSpace(gids) != "" {
		params.Set("gids", gids)
	}
	var list []AlertRule
	err := c.get(ctx, "/busi-groups/alert-rules", params, &list)
	return list, err
}

// EventListRequest 事件列表请求参数
type EventListRequest struct {
	Hours    int64  `json:"hours"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Query    string `json:"query"`
	Severity int    `json:"severity"`
}

// EventList 统一返回结构
type EventList struct {
	List  []AlertEvent `json:"list"`
	Total int64        `json:"total"`
}

// AlertEvent 统一事件字段，后端归一化 bool/int 差异
type AlertEvent struct {
	ID             int64  `json:"id"`
	RuleID         int64  `json:"rule_id"`
	RuleName       string `json:"rule_name"`
	RuleNote       string `json:"rule_note"`
	Severity       int    `json:"severity"`
	Status         int    `json:"status"`
	IsRecovered    bool   `json:"is_recovered"`
	TargetIdent    string `json:"target_ident"`
	TargetNote     string `json:"target_note"`
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	TriggerTime    int64  `json:"trigger_time"`
	LastEvalTime   int64  `json:"last_eval_time"`
	Tags           string `json:"tags"`
	TriggerValue   string `json:"trigger_value"`
	NotifyChannels string `json:"notify_channels"`
}

func (c *Client) ListActiveEvents(ctx context.Context, req EventListRequest) (*EventList, error) {
	params := eventParams(req)
	var raw struct {
		List  []activeEvent `json:"list"`
		Total int64         `json:"total"`
	}
	if err := c.get(ctx, "/alert-cur-events/list", params, &raw); err != nil {
		return nil, err
	}
	list := make([]AlertEvent, 0, len(raw.List))
	for _, e := range raw.List {
		list = append(list, e.normalize())
	}
	return &EventList{List: list, Total: raw.Total}, nil
}

func (c *Client) ListHistoryEvents(ctx context.Context, req EventListRequest) (*EventList, error) {
	params := eventParams(req)
	var raw struct {
		List  []historyEvent `json:"list"`
		Total int64          `json:"total"`
	}
	if err := c.get(ctx, "/alert-his-events/list", params, &raw); err != nil {
		return nil, err
	}
	list := make([]AlertEvent, 0, len(raw.List))
	for _, e := range raw.List {
		list = append(list, e.normalize())
	}
	return &EventList{List: list, Total: raw.Total}, nil
}

func eventParams(req EventListRequest) url.Values {
	p := url.Values{}
	if req.Hours > 0 {
		p.Set("hours", strconv.FormatInt(req.Hours, 10))
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	p.Set("p", strconv.Itoa(req.Page))
	p.Set("limit", strconv.Itoa(req.Limit))
	if req.Query != "" {
		p.Set("query", req.Query)
	}
	if req.Severity > 0 {
		p.Set("severity", strconv.Itoa(req.Severity))
	}
	return p
}

// activeEvent 为夜莺 /alert-cur-events/list 原始行
type activeEvent struct {
	ID             int64  `json:"id"`
	RuleID         int64  `json:"rule_id"`
	RuleName       string `json:"rule_name"`
	RuleNote       string `json:"rule_note"`
	Severity       int    `json:"severity"`
	Status         int    `json:"status"`
	IsRecovered    bool   `json:"is_recovered"`
	TargetIdent    string `json:"target_ident"`
	TargetNote     string `json:"target_note"`
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	TriggerTime    int64  `json:"trigger_time"`
	LastEvalTime   int64  `json:"last_eval_time"`
	Tags           []string `json:"tags"`
	TriggerValue   string   `json:"trigger_value"`
	NotifyChannels []string `json:"notify_channels"`
}

func (e activeEvent) normalize() AlertEvent {
	return AlertEvent{
		ID:             e.ID,
		RuleID:         e.RuleID,
		RuleName:       e.RuleName,
		RuleNote:       e.RuleNote,
		Severity:       e.Severity,
		Status:         e.Status,
		IsRecovered:    e.IsRecovered,
		TargetIdent:    e.TargetIdent,
		TargetNote:     e.TargetNote,
		GroupID:        e.GroupID,
		GroupName:      e.GroupName,
		TriggerTime:    e.TriggerTime,
		LastEvalTime:   e.LastEvalTime,
		Tags:           strings.Join(e.Tags, ","),
		TriggerValue:   e.TriggerValue,
		NotifyChannels: strings.Join(e.NotifyChannels, ","),
	}
}

// historyEvent 为夜莺 /alert-his-events/list 原始行，字段与 activeEvent 相同，但 is_recovered 为 int
type historyEvent struct {
	ID             int64    `json:"id"`
	RuleID         int64    `json:"rule_id"`
	RuleName       string   `json:"rule_name"`
	RuleNote       string   `json:"rule_note"`
	Severity       int      `json:"severity"`
	Status         int      `json:"status"`
	IsRecoveredInt int      `json:"is_recovered"`
	TargetIdent    string   `json:"target_ident"`
	TargetNote     string   `json:"target_note"`
	GroupID        int64    `json:"group_id"`
	GroupName      string   `json:"group_name"`
	TriggerTime    int64    `json:"trigger_time"`
	LastEvalTime   int64    `json:"last_eval_time"`
	Tags           []string `json:"tags"`
	TriggerValue   string   `json:"trigger_value"`
	NotifyChannels []string `json:"notify_channels"`
}

func (e historyEvent) normalize() AlertEvent {
	return AlertEvent{
		ID:             e.ID,
		RuleID:         e.RuleID,
		RuleName:       e.RuleName,
		RuleNote:       e.RuleNote,
		Severity:       e.Severity,
		Status:         e.Status,
		IsRecovered:    e.IsRecoveredInt == 1,
		TargetIdent:    e.TargetIdent,
		TargetNote:     e.TargetNote,
		GroupID:        e.GroupID,
		GroupName:      e.GroupName,
		TriggerTime:    e.TriggerTime,
		LastEvalTime:   e.LastEvalTime,
		Tags:           strings.Join(e.Tags, ","),
		TriggerValue:   e.TriggerValue,
		NotifyChannels: strings.Join(e.NotifyChannels, ","),
	}
}

func (c *Client) TestConnection(ctx context.Context) error {
	var list []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	return c.get(ctx, "/busi-groups", url.Values{"limit": []string{"1"}}, &list)
}
