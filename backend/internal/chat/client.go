package chat

import (
	"context"
	"errors"
	"io"

	"aiops/models"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// NewClient 根据 LLMConfig 初始化 Eino OpenAI 兼容客户端
func NewClient(ctx context.Context, cfg *models.LLMConfig) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
}

// StreamMessages 将 Eino 流读取器转换为字符串片段通道
func StreamMessages(ctx context.Context, streamReader *schema.StreamReader[*schema.Message]) (<-chan string, <-chan error) {
	chunkCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer streamReader.Close()

		for {
			msg, err := streamReader.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				errCh <- err
				return
			}
			if msg != nil && msg.Content != "" {
				select {
				case chunkCh <- msg.Content:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return chunkCh, errCh
}
