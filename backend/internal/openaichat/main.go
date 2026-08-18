package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 从环境变量读取 OpenAI 兼容服务的配置
	apiKey := "sk-kimi-87JDBrdkbwV6Hf9IfFPJbVtuxTjhYK4NJ17Kn592qoua0c0mcLBwcWFFkLdoNRKz"
	baseURL := "https://api.kimi.com/coding/v1"
	modelName := "k3-256k"

	// 初始化 OpenAI 协议的 ChatModel
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		panic("初始化模型失败: " + err.Error())
	}

	// 构造对话消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are a helpful assistant.",
		},
		{
			Role:    schema.User,
			Content: "你好，请用一句话介绍自己",
		},
	}

	// 调用模型生成回复
	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		panic("对话调用失败: " + err.Error())
	}

	fmt.Println("Assistant:", resp.Content)
}
