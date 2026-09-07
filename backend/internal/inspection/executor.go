package inspection

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aiops/internal/agent"
	"aiops/internal/chat"
	"aiops/internal/notify"
	"aiops/models"

	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// Executor 根据巡检任务的提示词调用默认 LLM 生成报告
// frontendBaseURL 为前端访问地址，用于报告通知中的"完整报告"链接
func Executor(ctx context.Context, db *gorm.DB, task models.InspectionTask, frontendBaseURL string) (models.InspectionReport, error) {
	report := models.InspectionReport{
		TaskID:   task.ID,
		TaskName: task.Name,
		Status:   "generating",
	}

	// 1. 加载默认 LLM 配置
	cfg, err := chat.DefaultConfig(db, models.LLMModelTypeChat)
	if err != nil {
		report.Status = "failed"
		report.Error = "加载 LLM 配置失败: " + err.Error()
		_ = db.Create(&report).Error
		return report, err
	}

	// 2. 初始化 ChatModel
	cm, err := chat.NewClient(ctx, cfg)
	if err != nil {
		report.Status = "failed"
		report.Error = "初始化 LLM 失败: " + err.Error()
		_ = db.Create(&report).Error
		return report, err
	}

	// 3. 组装消息（巡检专用的 system prompt + 用户提示词）
	systemPrompt := "你是一个专业的运维巡检助手。请根据用户提供的巡检任务要求，结合你所掌握的运维知识与系统状态信息，生成一份结构清晰的巡检报告。报告应包含：巡检时间、整体健康状态、关键指标摘要、发现的问题与建议、风险评估等部分。使用 Markdown 格式输出。报告中的关键结论应尽量基于工具查询到的真实数据。"
	today := time.Now().Format("2006-01-02")
	userPrompt := fmt.Sprintf("[巡检任务] %s\n\n[执行时间] %s\n\n[任务要求]\n%s", task.Name, today, task.Prompt)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 4. 使用公共 Agent 执行（与对话 Agent 共享数据源工具集，可查询真实指标/日志/火焰图）
	ag := agent.New(cm, agent.NewRegistry(db), agent.Options{
		Instructions: systemPrompt,
	})
	resp, err := ag.Run(ctx, messages, nil)
	if err != nil {
		report.Status = "failed"
		report.Error = "LLM 调用失败: " + err.Error()
		_ = db.Create(&report).Error
		return report, err
	}

	content := resp.Content
	if content == "" {
		report.Status = "failed"
		report.Error = "LLM 返回为空"
		_ = db.Create(&report).Error
		return report, fmt.Errorf("LLM 返回为空")
	}

	// 5. 解析结果
	report.Title = fmt.Sprintf("%s - %s", task.Name, today)
	report.Content = content
	report.Summary = extractSummary(content, 120)
	report.Score = estimateScore(content)
	report.Status = "completed"

	if err := db.Create(&report).Error; err != nil {
		report.Error = "保存报告失败: " + err.Error()
		report.Status = "failed"
		_ = db.Save(&report).Error
		return report, err
	}

	// 6. 通过任务配置的通知媒介推送报告（失败仅记录日志，不影响报告本身）
	notifyReport(db, task, report, frontendBaseURL)

	return report, nil
}

// notifyReport 将巡检报告通过任务关联的通知媒介发送出去
func notifyReport(db *gorm.DB, task models.InspectionTask, report models.InspectionReport, frontendBaseURL string) {
	if len(task.NotifyMediaIDs) == 0 {
		return
	}
	var ids []string
	if err := json.Unmarshal([]byte(task.NotifyMediaIDs), &ids); err != nil || len(ids) == 0 {
		return
	}

	var mediaList []models.NotifyMedia
	if err := db.Where("id IN ? AND is_enabled = ?", ids, 1).Find(&mediaList).Error; err != nil {
		fmt.Printf("[巡检通知] 查询通知媒介失败: %v\n", err)
		return
	}

	content := fmt.Sprintf("【AIOPS 巡检报告】\n任务：%s\n标题：%s\n评分：%d\n时间：%s\n完整报告： %s/#/inspection/reports/%d",
		report.TaskName, report.Title, report.Score,
		report.CreatedAt.Format("2006-01-02 15:04:05"),
		strings.TrimRight(frontendBaseURL, "/"), report.ID)

	for _, m := range mediaList {
		if err := notify.Send(m.Type, m.Config, content); err != nil {
			fmt.Printf("[巡检通知] 通过媒介 %s(%s) 发送报告失败: %v\n", m.Name, m.ID, err)
		} else {
			fmt.Printf("[巡检通知] 已通过媒介 %s 发送报告\n", m.Name)
		}
	}
}

// extractSummary 从报告内容里取前几个非空段落的前 maxLen 个字符作为摘要
func extractSummary(content string, maxLen int) string {
	runes := []rune(content)
	out := make([]rune, 0, maxLen)
	count := 0
	for _, r := range runes {
		if r == '\n' || r == '\r' {
			count++
			continue
		}
		out = append(out, r)
		if len(out) >= maxLen {
			out = append(out, '…')
			break
		}
	}
	_ = count
	return string(out)
}

// estimateScore 基于内容关键字粗略估算一个分数（0-100）
func estimateScore(content string) int {
	score := 80
	badKeywords := []string{"严重", "critical", "紧急", "故障", "down", "宕机", "error", "failed", "失败"}
	goodKeywords := []string{"健康", "正常", "ok", "stable", "平稳", "成功", "已恢复"}
	lower := content
	for _, kw := range badKeywords {
		if containsCI(lower, kw) {
			score -= 8
		}
	}
	for _, kw := range goodKeywords {
		if containsCI(lower, kw) {
			score += 4
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func containsCI(s, substr string) bool {
	if len(substr) == 0 {
		return false
	}
	// 简单实现：用 toLower 的 rune 比较
	toLower := func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}
	lowerS := ""
	for _, r := range s {
		lowerS += string(toLower(r))
	}
	lowerSub := ""
	for _, r := range substr {
		lowerSub += string(toLower(r))
	}
	return len(lowerS) >= len(lowerSub) && indexOf(lowerS, lowerSub) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// randomID 用于区分并发执行时的日志
func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
