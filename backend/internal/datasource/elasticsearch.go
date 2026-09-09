package datasource

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiops/models"
)

// ElasticsearchClient Elasticsearch 只读查询客户端
type ElasticsearchClient struct {
	ds models.Datasource
	hc *http.Client
}

// NewElasticsearchClient 创建 Elasticsearch 客户端
func NewElasticsearchClient(ds *models.Datasource) *ElasticsearchClient {
	timeout := ds.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout * 1000
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if ds.IsSkipSSL == 1 {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &ElasticsearchClient{
		ds: *ds,
		hc: &http.Client{
			Timeout:   time.Duration(timeout) * time.Millisecond,
			Transport: transport,
		},
	}
}

// Search 执行日志检索，返回精简后的命中列表
func (c *ElasticsearchClient) Search(ctx context.Context, indexPattern, queryString, start, end string, size int) (*ESQueryResult, error) {
	if indexPattern == "" {
		return nil, fmt.Errorf("索引模式不能为空")
	}
	if err := validateIndexPattern(indexPattern); err != nil {
		return nil, err
	}

	if size <= 0 || size > MaxESHits {
		size = MaxESHits
	}

	now := time.Now()
	startT, err := ParseTime(start, now)
	if err != nil {
		return nil, fmt.Errorf("解析 start 失败: %w", err)
	}
	endT, err := ParseTime(end, now)
	if err != nil {
		return nil, fmt.Errorf("解析 end 失败: %w", err)
	}
	if endT.Before(startT) {
		endT = now
	}

	body := map[string]interface{}{
		"size": size,
		"sort": []map[string]string{{"@timestamp": "desc"}},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					buildQueryString(queryString),
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]string{
								"gte": startT.Format(time.RFC3339),
								"lte": endT.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.ds.URL, "/") + "/" + indexPattern + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.ds.Username != "" {
		req.SetBasicAuth(c.ds.Username, c.ds.Password)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Elasticsearch 失败: %w", err)
	}
	defer resp.Body.Close()

	var payload esResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 响应失败: %w", err)
	}
	if payload.Error != nil {
		return nil, fmt.Errorf("Elasticsearch 查询失败: %v", payload.Error)
	}

	result := &ESQueryResult{
		DataSource: c.ds.Name,
		Index:      indexPattern,
		Total:      payload.Hits.Total.Value,
		Hits:       make([]*ESLogEntry, 0, len(payload.Hits.Hits)),
	}
	for _, h := range payload.Hits.Hits {
		result.Hits = append(result.Hits, &ESLogEntry{
			ID:     h.ID,
			Source: trimSource(h.Source),
		})
	}
	return result, nil
}

// Count 执行索引文档计数查询（match_all，不带时间过滤），用于服务注册的连通性探测
func (c *ElasticsearchClient) Count(ctx context.Context, indexPattern string) (int, error) {
	if indexPattern == "" {
		return 0, fmt.Errorf("索引模式不能为空")
	}
	if err := validateIndexPattern(indexPattern); err != nil {
		return 0, err
	}

	url := strings.TrimRight(c.ds.URL, "/") + "/" + indexPattern + "/_count"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewReader([]byte(`{"query":{"match_all":{}}}`)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.ds.Username != "" {
		req.SetBasicAuth(c.ds.Username, c.ds.Password)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求 Elasticsearch 失败: %w", err)
	}
	defer resp.Body.Close()

	var payload struct {
		Error map[string]interface{} `json:"error,omitempty"`
		Count int                    `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, fmt.Errorf("解析 Elasticsearch 响应失败: %w", err)
	}
	if payload.Error != nil {
		return 0, fmt.Errorf("Elasticsearch 查询失败: %v", payload.Error)
	}
	return payload.Count, nil
}

func buildQueryString(queryString string) map[string]interface{} {
	if strings.TrimSpace(queryString) == "" {
		return map[string]interface{}{"match_all": map[string]interface{}{}}
	}
	return map[string]interface{}{"query_string": map[string]interface{}{"query": queryString}}
}

func validateIndexPattern(pattern string) error {
	forbidden := []string{"/", "\\", " ", "_delete_by_query", "_update_by_query", "_bulk", "_create"}
	for _, s := range forbidden {
		if strings.Contains(pattern, s) {
			return fmt.Errorf("非法的索引模式: %s", pattern)
		}
	}
	return nil
}

func trimSource(src map[string]interface{}) map[string]interface{} {
	// 保留最关键的字段，避免单条文档过大；LLM 通常只需要 message/level/status 等
	keys := []string{"@timestamp", "timestamp", "message", "msg", "level", "loglevel", "status", "path", "host", "service", "pod", "namespace", "trace_id"}
	out := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		if v, ok := src[k]; ok {
			out[k] = v
		}
	}
	if len(out) == 0 {
		// 没有命中关键字段时，返回原字段但限制数量
		for k, v := range src {
			out[k] = v
			if len(out) >= 10 {
				break
			}
		}
	}
	return out
}

type esResponse struct {
	Error map[string]interface{} `json:"error,omitempty"`
	Hits  struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
