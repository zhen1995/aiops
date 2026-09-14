package inspection

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"aiops/internal/agent"
	"aiops/internal/chat"
	"aiops/models"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// maxDimensions 多维度巡检支持的维度数量上限（控制 LLM 调用成本）
const maxDimensions = 8

// InspectionDimension 多维度巡检的单个维度配置
type InspectionDimension struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

// ValidateModeAndDimensions 校验任务执行模式与维度配置（控制器创建/更新时使用）
func ValidateModeAndDimensions(mode, dimensions string) error {
	if mode == "" || mode == ModeSingle {
		return nil
	}
	if mode != ModeMulti {
		return fmt.Errorf("执行模式必须是 single 或 multi")
	}
	_, err := parseDimensions(dimensions)
	return err
}

// parseDimensions 解析任务维度配置；multi 模式必有值（控制器已校验）
func parseDimensions(raw string) ([]InspectionDimension, error) {
	var dims []InspectionDimension
	if err := json.Unmarshal([]byte(raw), &dims); err != nil {
		return nil, fmt.Errorf("巡检维度不是合法 JSON: %w", err)
	}
	if len(dims) == 0 || len(dims) > maxDimensions {
		return nil, fmt.Errorf("巡检维度数量必须是 1-%d 个", maxDimensions)
	}
	for i, d := range dims {
		if strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.Prompt) == "" {
			return nil, fmt.Errorf("第 %d 个维度的名称/提示词不能为空", i+1)
		}
	}
	return dims, nil
}

// runMultiDimension 多维度巡检：各维度并发由子巡检 Agent 产出小结，主编 Agent 汇总成最终报告
func runMultiDimension(ctx context.Context, db *gorm.DB, task models.InspectionTask, report models.InspectionReport, cm model.ChatModel, frontendBaseURL string) (models.InspectionReport, error) {
	fail := func(err error, msg string) (models.InspectionReport, error) {
		report.Status = "failed"
		report.Error = msg + ": " + err.Error()
		_ = db.Create(&report).Error
		return report, err
	}

	dims, err := parseDimensions(task.Dimensions)
	if err != nil {
		return fail(err, "解析巡检维度失败")
	}
	today := time.Now().Format("2006-01-02")

	// 1. 各维度并发出子巡检 Agent（每个维度独立 ChatModel，避免并发共享状态）
	cfg, err := chat.DefaultConfig(db, models.LLMModelTypeChat)
	if err != nil {
		return fail(err, "加载 LLM 配置失败")
	}
	factory := func(c context.Context) (model.ChatModel, error) {
		return chat.NewClient(c, cfg)
	}

	summaries := make([]string, len(dims))
	var wg sync.WaitGroup
	for i, d := range dims {
		wg.Add(1)
		go func(idx int, dim InspectionDimension) {
			defer wg.Done()
			summaries[idx] = runDimension(ctx, db, task, dim, today, factory)
		}(i, d)
	}
	wg.Wait()

	// 2. 统计失败维度：全部失败则整体失败，部分失败由主编说明
	failed := 0
	for _, s := range summaries {
		if strings.HasPrefix(s, "【巡检失败】") {
			failed++
		}
	}
	if failed == len(dims) {
		return fail(fmt.Errorf("所有维度巡检均失败"), "多维度巡检失败")
	}

	// 3. 主编 Agent 汇总各维度小结为最终报告
	editorPrompt := fmt.Sprintf("你是运维巡检主编。下面是任务「%s」在 %s 各巡检维度的分析小结。请汇总为一份结构清晰的巡检报告（Markdown），要求：去重合并各维度的重复发现；按风险严重程度排序；保留各维度的关键数据证据；最后给出整体健康状态与整改建议。若某维度巡检失败，在报告中注明该维度未覆盖，不要编造。", task.Name, today)
	var b strings.Builder
	b.WriteString(editorPrompt + "\n\n【任务要求】\n" + task.Prompt + "\n\n【各维度小结】\n")
	for i, d := range dims {
		fmt.Fprintf(&b, "--- %s ---\n%s\n\n", d.Name, summaries[i])
	}

	editorCM, err := factory(ctx)
	if err != nil {
		return fail(err, "初始化主编模型失败")
	}
	messages := []*schema.Message{
		schema.SystemMessage(editorPrompt),
		schema.UserMessage(b.String()),
	}
	resp, err := editorCM.Generate(ctx, messages)
	if err != nil {
		return fail(err, "主编汇总失败")
	}
	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return fail(fmt.Errorf("主编返回为空"), "多维度巡检失败")
	}

	// 4. 落库 + 通知（与单 Agent 路径一致）
	finalizeReport(db, task, &report, content, frontendBaseURL)
	return report, nil
}

// runDimension 执行单个维度的巡检，返回该维度 Markdown 小结（失败返回"【巡检失败】原因"）
func runDimension(ctx context.Context, db *gorm.DB, task models.InspectionTask, dim InspectionDimension, today string, factory func(context.Context) (model.ChatModel, error)) string {
	cm, err := factory(ctx)
	if err != nil {
		return "【巡检失败】初始化大模型失败: " + err.Error()
	}

	systemPrompt := fmt.Sprintf("你是运维巡检分析师，负责巡检维度【%s】。请结合工具查询到的真实指标/日志/剖析数据完成本维度分析，输出该维度的 Markdown 小结（发现的问题、数据证据、风险等级、建议），只基于真实数据，禁止编造。", dim.Name)
	userPrompt := fmt.Sprintf("[巡检任务] %s\n[巡检维度] %s\n[执行日期] %s\n[任务整体要求]\n%s\n\n[本维度要求]\n%s",
		task.Name, dim.Name, today, task.Prompt, dim.Prompt)

	ag := agent.New(cm, agent.NewRegistry(db), agent.Options{
		Instructions: systemPrompt,
	})
	resp, err := ag.Run(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}, nil)
	if err != nil {
		return "【巡检失败】" + err.Error()
	}
	summary := strings.TrimSpace(resp.Content)
	if summary == "" {
		return "【巡检失败】该维度未产出小结"
	}
	return summary
}
