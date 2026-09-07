package rca

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aiops/internal/chat"
	"aiops/models"

	"gorm.io/gorm"
)

// analysisTimeout 单次根因分析的最大执行时长
const analysisTimeout = 10 * time.Minute

// severityName 告警级别数字到中文名
var severityName = map[int]string{
	1: "P1紧急",
	2: "P2警告",
	3: "P3提醒",
}

// StartAnalysis 为指定告警事件启动根因分析：
// 事件不存在返回 error；已有 running 状态的记录则直接返回其 ID（防重复触发）；
// 否则创建 running 记录并异步执行，立即返回分析记录 ID。
func StartAnalysis(db *gorm.DB, eventID string) (string, error) {
	var event models.AlertEvent
	if err := db.First(&event, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("告警事件不存在: %s", eventID)
		}
		return "", fmt.Errorf("加载告警事件失败: %w", err)
	}

	// 防重复触发：同一事件已有分析中的记录则直接复用
	var existing models.RootCauseAnalysis
	err := db.Where("alert_event_id = ? AND status = ?", eventID, models.RootCauseStatusRunning).
		First(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("查询根因分析记录失败: %w", err)
	}

	analysis := models.RootCauseAnalysis{
		AlertEventID: event.ID,
		RuleName:     event.RuleName,
		TargetIdent:  event.TargetIdent,
		Tags:         event.Tags,
		Severity:     event.Severity,
		TriggerTime:  event.TriggerTime,
		Status:       models.RootCauseStatusRunning,
	}
	if err := db.Create(&analysis).Error; err != nil {
		return "", fmt.Errorf("创建根因分析记录失败: %w", err)
	}

	go runAnalysis(db, analysis.ID, event)

	return analysis.ID, nil
}

// runAnalysis 在后台 goroutine 中执行根因分析工作流并落库结果
func runAnalysis(db *gorm.DB, analysisID string, event models.AlertEvent) {
	defer func() {
		// panic 兜底：任何未 recovered 的异常都置为 failed，避免记录停留在 running
		if r := recover(); r != nil {
			failAnalysis(db, analysisID, fmt.Sprintf("分析过程发生未预期错误: %v", r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	// 1. 加载默认 LLM 配置
	cfg, err := chat.DefaultConfig(db, models.LLMModelTypeChat)
	if err != nil {
		failAnalysis(db, analysisID, "加载 LLM 配置失败: "+err.Error())
		return
	}

	// 2. 初始化 ChatModel
	cm, err := chat.NewClient(ctx, cfg)
	if err != nil {
		failAnalysis(db, analysisID, "初始化 LLM 失败: "+err.Error())
		return
	}

	// 3. 构建告警事件上下文字符串
	eventCtx := buildEventContext(event)

	// 4. 编译并执行 Graph 工作流
	var evidence string
	runner, err := buildGraph(cm, db, eventCtx, &evidence)
	if err != nil {
		failAnalysis(db, analysisID, "编译根因分析工作流失败: "+err.Error())
		return
	}
	report, err := runner.Invoke(ctx, eventCtx)
	if err != nil {
		failAnalysis(db, analysisID, "执行根因分析失败: "+err.Error())
		return
	}

	// 5. 成功落库：Result 为 JSON 报告，Content 为证据 + 报告
	updates := map[string]any{
		"status":  models.RootCauseStatusCompleted,
		"result":  report,
		"content": fmt.Sprintf("## 证据汇总\n\n%s\n\n## 分析报告\n\n%s", evidence, report),
		"error":   "",
	}
	if err := db.Model(&models.RootCauseAnalysis{}).Where("id = ?", analysisID).Updates(updates).Error; err != nil {
		failAnalysis(db, analysisID, "保存分析结果失败: "+err.Error())
	}
}

// failAnalysis 将分析记录置为失败
func failAnalysis(db *gorm.DB, analysisID, errMsg string) {
	fmt.Printf("[根因分析] 记录 %s 失败: %s\n", analysisID, errMsg)
	updates := map[string]any{
		"status": models.RootCauseStatusFailed,
		"error":  errMsg,
	}
	if err := db.Model(&models.RootCauseAnalysis{}).Where("id = ?", analysisID).Updates(updates).Error; err != nil {
		fmt.Printf("[根因分析] 更新失败状态落库失败: %v\n", err)
	}
}

// buildEventContext 将告警事件关键字段序列化为事件上下文字符串
func buildEventContext(event models.AlertEvent) string {
	level, ok := severityName[event.Severity]
	if !ok {
		level = fmt.Sprintf("P%d", event.Severity)
	}
	return fmt.Sprintf("[规则名称] %s\n[告警级别] %s\n[告警对象] %s\n[标签] %s\n[触发值] %s\n[触发时间] %s",
		event.RuleName,
		level,
		event.TargetIdent,
		event.Tags,
		event.TriggerValue,
		event.TriggerTime.Format("2006-01-02 15:04:05"),
	)
}
