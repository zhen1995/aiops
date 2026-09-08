// Package n9e 夜莺 (Nightingale) HTTP 客户端
package n9e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aiops/models"

	"gorm.io/gorm"
)

// Client 夜莺 API 客户端（单引擎配置）
type Client struct {
	BaseURL string
	Token   string
	httpCli *http.Client
}

// NewClientFromDB 从 DB 加载当前启用的夜莺引擎配置，构造 Client
func NewClientFromDB(db *gorm.DB) (*Client, error) {
	cfg, err := models.GetEnabledN9eConfig(db)
	if err != nil {
		return nil, fmt.Errorf("未配置启用的夜莺引擎: %w", err)
	}
	return &Client{
		BaseURL: strings.TrimRight(cfg.Address, "/"),
		Token:   cfg.Token,
		httpCli: &http.Client{Timeout: 20 * time.Second},
	}, nil
}

// NewClient 直接根据地址+token 构造
func NewClient(address, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(address, "/"),
		Token:   token,
		httpCli: &http.Client{Timeout: 20 * time.Second},
	}
}

// rawProxy 内部统一代理
func (c *Client) rawProxy(ctx context.Context, method, path string, body io.Reader) ([]byte, int, error) {
	target := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, 0, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Token", c.Token)
	req.Header.Set("Accept", "*/*")

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}

// GetBusiGroups 查询业务组
func (c *Client) GetBusiGroups(ctx context.Context) ([]byte, int, error) {
	return c.rawProxy(ctx, "GET", "/api/n9e/busi-groups?limit=5000", nil)
}

// GetAlertRules 查询告警规则列表
func (c *Client) GetAlertRules(ctx context.Context) ([]byte, int, error) {
	return c.rawProxy(ctx, "GET", "/api/n9e/busi-groups/alert-rules", nil)
}

// CurEventsParams 活跃告警参数
type CurEventsParams struct {
	P         int
	Limit     int
	MyGroups  bool
}

// GetCurEvents 查询活跃告警事件（POST）
func (c *Client) GetCurEvents(ctx context.Context, p CurEventsParams) ([]byte, int, error) {
	vs := url.Values{}
	if p.P > 0 {
		vs.Set("p", fmt.Sprintf("%d", p.P))
	}
	if p.Limit > 0 {
		vs.Set("limit", fmt.Sprintf("%d", p.Limit))
	} else {
		vs.Set("limit", "30")
	}
	vs.Set("my_groups", boolStr(p.MyGroups))
	path := "/api/n9e/alert-cur-events/list"
	if len(vs) > 0 {
		path += "?" + vs.Encode()
	}
	return c.rawProxy(ctx, "POST", path, strings.NewReader("{}"))
}

// HisEventsParams 历史告警参数
type HisEventsParams struct {
	P            int
	Limit        int
	Stime        int64
	Etime        int64
	RuleIDs      string
	BusiGroupIDs string
	Cate         string
	Tags         string
}

// GetHisEvents 查询历史告警事件
func (c *Client) GetHisEvents(ctx context.Context, p HisEventsParams) ([]byte, int, error) {
	vs := url.Values{}
	if p.P > 0 {
		vs.Set("p", fmt.Sprintf("%d", p.P))
	}
	if p.Limit > 0 {
		vs.Set("limit", fmt.Sprintf("%d", p.Limit))
	} else {
		vs.Set("limit", "30")
	}
	if p.Stime > 0 {
		vs.Set("stime", fmt.Sprintf("%d", p.Stime))
	}
	if p.Etime > 0 {
		vs.Set("etime", fmt.Sprintf("%d", p.Etime))
	}
	if p.RuleIDs != "" {
		vs.Set("rule_ids", p.RuleIDs)
	}
	if p.BusiGroupIDs != "" {
		vs.Set("busi_group_ids", p.BusiGroupIDs)
	}
	if p.Cate != "" {
		vs.Set("cate", p.Cate)
	}
	if p.Tags != "" {
		vs.Set("tags", p.Tags)
	}
	path := "/api/n9e/alert-his-events/list"
	if len(vs) > 0 {
		path += "?" + vs.Encode()
	}
	return c.rawProxy(ctx, "GET", path, nil)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
