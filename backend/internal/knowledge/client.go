package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"aiops/models"
)

// ChunkHit 检索命中的一条知识片段（含溯源信息）
type ChunkHit struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
}

// Client Python 知识库服务客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTPClient: &http.Client{Timeout: 120 * time.Second}}
}

// Index 上传文档到 Python 服务解析、向量化并入 Qdrant，返回分块数
func (c *Client) Index(ctx context.Context, documentID, title string, emb *models.LLMConfig, fileName string, file io.Reader) (int, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("document_id", documentID)
	_ = w.WriteField("title", title)
	_ = w.WriteField("base_url", emb.BaseURL)
	_ = w.WriteField("api_key", emb.APIKey)
	_ = w.WriteField("model", emb.Model)
	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return 0, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return 0, err
	}
	if err := w.Close(); err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/knowledge/index", &buf)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("知识库索引失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	var out struct {
		Chunks int `json:"chunks"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, err
	}
	return out.Chunks, nil
}

// Retrieve 语义检索知识库
func (c *Client) Retrieve(ctx context.Context, query string, topK int, emb *models.LLMConfig) ([]ChunkHit, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"query": query,
		"top_k": topK,
		"embedding": map[string]string{
			"base_url": emb.BaseURL,
			"api_key":  emb.APIKey,
			"model":    emb.Model,
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/knowledge/retrieve", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("知识库检索失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	var out struct {
		Results []ChunkHit `json:"results"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// DeleteDocument 删除文档的全部向量
func (c *Client) DeleteDocument(ctx context.Context, documentID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/api/v1/knowledge/documents/"+documentID, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除知识库向量失败(%d): %s", resp.StatusCode, extractDetail(body))
	}
	return nil
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
