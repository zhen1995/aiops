package loganalysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ParseResult 单条日志的 Drain 解析结果（与 Python /api/v1/logs/parse 契约一致）
type ParseResult struct {
	Template  string   `json:"template"`
	Params    []string `json:"params"`
	ClusterID int      `json:"cluster_id"`
	Level     string   `json:"level"`
}

// Client Python 日志解析服务（drain3）客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient 创建客户端（超时 30s，Python 不可用时报明确错误）
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTPClient: &http.Client{Timeout: 30 * time.Second}}
}

// Parse 批量调用 Python Drain 解析，返回与 logs 等长的结果（结果缺失时按原文兜底）
func (c *Client) Parse(ctx context.Context, serviceID string, logs []string) ([]ParseResult, error) {
	if len(logs) == 0 {
		return nil, nil
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"service_id": serviceID,
		"logs":       logs,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/logs/parse", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Python 日志解析服务不可用: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python 日志解析失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	var out struct {
		Results []ParseResult `json:"results"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("解析 Python 响应失败: %w", err)
	}
	// 结果缺失兜底：按原文行构造，保证与输入对齐
	for len(out.Results) < len(logs) {
		out.Results = append(out.Results, ParseResult{Template: logs[len(out.Results)], Level: "info"})
	}
	return out.Results, nil
}

// extractDetail 从错误响应体中提取 detail 字段
func extractDetail(body []byte) string {
	var e struct {
		Detail string `json:"detail"`
	}
	if json.Unmarshal(body, &e) == nil && e.Detail != "" {
		return e.Detail
	}
	return string(body)
}
