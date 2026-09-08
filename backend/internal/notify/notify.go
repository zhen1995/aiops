package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// 媒介类型
const (
	TypeWebhook  = "webhook"
	TypeDingtalk = "dingtalk"
)

// MediaConfig 媒介配置（NotifyMedia.Config 字段的 JSON 内容）
type MediaConfig struct {
	URL     string `json:"url"`     // webhook 回调地址（webhook 类型必填）
	Webhook string `json:"webhook"` // 钉钉机器人 webhook 地址（dingtalk 类型必填）
	Secret  string `json:"secret"`  // 钉钉加签密钥（选填）
	Method  string `json:"method"`  // 请求方法（webhook 类型，默认 POST）
	Timeout int    `json:"timeout"` // 超时时间 ms，默认 5000
}

// ParseConfig 解析媒介配置 JSON
func ParseConfig(rawConfig string) (*MediaConfig, error) {
	if rawConfig == "" {
		return nil, errors.New("媒介配置不能为空")
	}
	var cfg MediaConfig
	if err := json.Unmarshal([]byte(rawConfig), &cfg); err != nil {
		return nil, errors.New("媒介配置不是合法的 JSON")
	}
	return &cfg, nil
}

// Send 按媒介类型发送消息。msgtype: text(默认) / markdown / html。
// 钉钉/企微机器人支持 markdown；Webhook 默认发 text。
func Send(mediaType, rawConfig, content string) error {
	return SendTyped(mediaType, rawConfig, content, "text")
}

// SendTyped 带消息类型的发送
func SendTyped(mediaType, rawConfig, content, msgType string) error {
	cfg, err := ParseConfig(rawConfig)
	if err != nil {
		return err
	}
	if msgType == "" {
		msgType = "text"
	}

	timeout := time.Duration(cfg.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	switch mediaType {
	case TypeWebhook:
		method := cfg.Method
		if method == "" {
			method = http.MethodPost
		}
		if cfg.URL == "" {
			return errors.New("webhook 回调地址不能为空")
		}
		body, _ := json.Marshal(map[string]interface{}{
			"msgtype": msgType,
			msgType:   map[string]string{"content": content},
		})
		req, err := http.NewRequest(method, cfg.URL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		return doRequest(client, req)

	case TypeDingtalk:
		if cfg.Webhook == "" {
			return errors.New("钉钉机器人 Webhook 地址不能为空")
		}
		webhook := cfg.Webhook
		if cfg.Secret != "" {
			webhook = addDingtalkSign(webhook, cfg.Secret)
		}
		var bodyBytes []byte
		if msgType == "markdown" {
			body, _ := json.Marshal(map[string]interface{}{
				"msgtype":  "markdown",
				"markdown": map[string]string{"title": "AIOps 告警通知", "text": content},
			})
			bodyBytes = body
		} else {
			body, _ := json.Marshal(map[string]interface{}{
				"msgtype": "text",
				"text":    map[string]string{"content": content},
			})
			bodyBytes = body
		}
		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		return doDingtalkRequest(client, req)
	}
	return errors.New("未知的媒介类型")
}

// doRequest 执行请求，非 2xx 视为失败
func doRequest(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("目标返回状态码 %d", resp.StatusCode)
	}
	return nil
}

// doDingtalkRequest 执行钉钉请求并解析响应体，errcode 非 0 视为失败
func doDingtalkRequest(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("目标返回状态码 %d", resp.StatusCode)
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("无法解析钉钉响应: %s", string(body))
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("钉钉返回错误(%d): %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// addDingtalkSign 按钉钉加签规则为 webhook 拼接 timestamp 与 sign
func addDingtalkSign(webhook, secret string) string {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := timestamp + "\n" + secret
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return webhook + "&timestamp=" + timestamp + "&sign=" + url.QueryEscape(sign)
}
