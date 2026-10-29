// Package anomaly Go 侧异常检测客户端：调用 Python 算法服务的 /api/v1/anomaly/detect。
// 原则：原始时序只在程序侧流转，进模型上下文的只有异常信号摘要。
package anomaly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SeriesInput 单条待检测时序；HistoryPoints 为上一周期同时段序列（可选，用于周期对比）
type SeriesInput struct {
	Metric        string            `json:"metric"`
	Labels        map[string]string `json:"labels,omitempty"`
	Points        [][2]interface{}  `json:"points"` // [ts(毫秒或ISO), value]
	HistoryPoints [][2]interface{}  `json:"history_points,omitempty"`
}

// DetectRequest 异常检测请求
type DetectRequest struct {
	Series  []SeriesInput  `json:"series"`
	Options *DetectOptions `json:"options,omitempty"`
}

// DetectOptions 检测参数（零值时用 Python 侧默认）
type DetectOptions struct {
	Sigma       float64 `json:"sigma,omitempty"`
	IQRK        float64 `json:"iqr_k,omitempty"`
	RecentRatio float64 `json:"recent_ratio,omitempty"`
	ShiftZ      float64 `json:"shift_z,omitempty"`
	MaxSignals  int     `json:"max_signals,omitempty"`
}

// Signal 单条异常信号
type Signal struct {
	Metric      string            `json:"metric"`
	Labels      map[string]string `json:"labels"`
	Type        string            `json:"type"` // spike|drop|level_shift|outlier
	Deviation   float64           `json:"deviation"`
	Window      [2]string         `json:"window"`
	RecentMean  float64           `json:"recent_mean"`
	OverallMean float64           `json:"overall_mean"`
	Severity    string            `json:"severity"` // high|medium|low
}

// DetectResponse 异常检测响应
type DetectResponse struct {
	Signals []Signal `json:"signals"`
	Summary string   `json:"summary"`
}

// Client 异常检测客户端
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient 创建客户端；baseURL 为空时 Detect 返回明确错误
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 60 * time.Second}}
}

// Detect 调用 Python 异常检测服务
func (c *Client) Detect(ctx context.Context, req *DetectRequest) (*DetectResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("Python 算法服务地址未配置")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/anomaly/detect", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("异常检测服务不可用: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("异常检测服务返回 %s: %s", resp.Status, string(data))
	}
	var out DetectResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("解析异常检测结果失败: %w", err)
	}
	if out.Signals == nil {
		out.Signals = []Signal{}
	}
	return &out, nil
}
