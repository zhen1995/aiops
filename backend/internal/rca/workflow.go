package rca

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"aiops/internal/agent"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// 图中两个节点的 key
const (
	nodeCollect    = "collect"
	nodeSynthesize = "synthesize"
)

// buildGraph 编排根因分析工作流：collect（证据收集）→ synthesize（综合分析）。
// eventCtx 为告警事件上下文字符串，会作为 collect 的输入并供 synthesize 组合 user prompt；
// evidenceOut 用于把 collect 节点产出的证据文本透传出来，供上层拼装完整报告。
func buildGraph(cm model.ChatModel, db *gorm.DB, eventCtx string, evidenceOut *string) (compose.Runnable[string, string], error) {
	g := compose.NewGraph[string, string]()

	// collect 节点：构建证据收集 Agent，调用只读数据源工具后返回证据汇总文本
	collect := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		ag := agent.New(cm, agent.NewRegistry(db), agent.Options{
			Instructions:  collectPrompt,
			MaxIterations: 15, // 证据收集涉及多轮工具调用，放宽迭代上限
		})
		resp, err := ag.Run(ctx, []*schema.Message{schema.UserMessage(input)}, nil)
		if err != nil {
			return "", fmt.Errorf("证据收集失败: %w", err)
		}
		if evidenceOut != nil {
			*evidenceOut = resp.Content
		}
		return resp.Content, nil
	})
	if err := g.AddLambdaNode(nodeCollect, collect); err != nil {
		return nil, err
	}

	// synthesize 节点：基于事件上下文 + 证据生成严格 JSON 报告
	synthesize := compose.InvokableLambda(func(ctx context.Context, evidence string) (string, error) {
		userPrompt := fmt.Sprintf("[告警事件]\n%s\n\n[证据汇总]\n%s", eventCtx, evidence)
		return generateJSONReport(ctx, cm, userPrompt)
	})
	if err := g.AddLambdaNode(nodeSynthesize, synthesize); err != nil {
		return nil, err
	}

	if err := g.AddEdge(compose.START, nodeCollect); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeCollect, nodeSynthesize); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeSynthesize, compose.END); err != nil {
		return nil, err
	}

	return g.Compile(context.Background())
}

// generateJSONReport 调用大模型生成 JSON 报告；
// 若输出不是合法 JSON，把解析错误反馈给模型修复重试一次；
// 两次均非法时原样返回文本并附加可识别的失败标记。
func generateJSONReport(ctx context.Context, cm model.ChatModel, userPrompt string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage(synthesizePrompt),
		schema.UserMessage(userPrompt),
	}

	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("综合分析失败: %w", err)
	}
	out := normalizeJSON(resp.Content)
	if json.Valid([]byte(out)) {
		return out, nil
	}
	firstRaw := out

	// 修复重试：将解析错误反馈给模型，要求重新输出合法 JSON
	messages = append(messages,
		schema.AssistantMessage(resp.Content, nil),
		schema.UserMessage(fmt.Sprintf("你的上一次输出不是合法 JSON（解析校验未通过），原文：\n%s\n\n请修复后重新输出，且仅输出符合 schema 要求的一个 JSON 对象，不要包含 markdown 代码块。", firstRaw)),
	)
	resp2, err := cm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("综合分析修复重试失败: %w", err)
	}
	out2 := normalizeJSON(resp2.Content)
	if json.Valid([]byte(out2)) {
		return out2, nil
	}

	// 两次均非法：原样返回文本并附加失败标记，便于上层与排查时识别
	return fmt.Sprintf("[INVALID_JSON] 模型两次输出均不是合法 JSON，以下为第二次原始输出：\n%s", out2), nil
}

// normalizeJSON 去除模型输出首尾空白及可能的 markdown 代码块包裹
func normalizeJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
