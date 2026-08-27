package chat

import (
	"errors"

	"aiops/models"

	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

const (
	SystemPrompt = "你是 AIOPS 智能运维助手，擅长告警解读、异常检测、根因分析与巡检报告问答。请用中文回答，简洁专业。"
	ContextLimit = 10
)

// CreateSession 创建新会话
func CreateSession(db *gorm.DB, userID string, title string) (*models.ChatSession, error) {
	if title == "" {
		title = "新会话"
	}
	session := &models.ChatSession{
		UserID: userID,
		Title:  title,
		Status: "active",
	}
	if err := db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

// ListSessions 查询用户的会话列表（含最后一条消息）
func ListSessions(db *gorm.DB, userID string) ([]map[string]interface{}, error) {
	var sessions []models.ChatSession
	if err := db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}

	// 批量查询最后一条 assistant 消息，避免 N+1
	sessionIDs := make([]string, 0, len(sessions))
	for _, s := range sessions {
		sessionIDs = append(sessionIDs, s.ID)
	}

	lastMsgs := make(map[string]string, len(sessions))
	if len(sessionIDs) > 0 {
		var messages []models.ChatMessage
		if err := db.Where("session_id IN ? AND role = ?", sessionIDs, "assistant").
			Order("created_at DESC").Find(&messages).Error; err != nil {
			return nil, err
		}
		seen := make(map[string]bool, len(sessions))
		for _, m := range messages {
			if seen[m.SessionID] {
				continue
			}
			seen[m.SessionID] = true
			lastMsg := m.Content
			if len([]rune(lastMsg)) > 40 {
				lastMsg = string([]rune(lastMsg)[:40]) + "..."
			}
			lastMsgs[m.SessionID] = lastMsg
		}
	}

	items := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, map[string]interface{}{
			"id":         s.ID,
			"title":      s.Title,
			"status":     s.Status,
			"updated_at": s.UpdatedAt,
			"last_msg":   lastMsgs[s.ID],
		})
	}
	return items, nil
}

// DeleteSession 软删除会话及其消息
func DeleteSession(db *gorm.DB, sessionID, userID string) error {
	var session models.ChatSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会话不存在")
		}
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&models.ChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&session).Error
	})
}

// ListMessages 查询会话历史消息（user/assistant，按时间升序）
func ListMessages(db *gorm.DB, sessionID, userID string) ([]models.ChatMessage, error) {
	var session models.ChatSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在")
		}
		return nil, err
	}

	var messages []models.ChatMessage
	if err := db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// SaveUserMessage 保存用户消息（会先校验会话归属）
func SaveUserMessage(db *gorm.DB, sessionID, userID, content string) (*models.ChatMessage, error) {
	var session models.ChatSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在")
		}
		return nil, err
	}

	msg := &models.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// SaveAssistantMessage 保存助手消息（会先校验会话归属）
func SaveAssistantMessage(db *gorm.DB, sessionID, userID, content string, promptTokens, completionTokens, totalTokens int) (*models.ChatMessage, error) {
	var session models.ChatSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在")
		}
		return nil, err
	}

	msg := &models.ChatMessage{
		SessionID:        sessionID,
		Role:             "assistant",
		Content:          content,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

// BuildContext 组装 system + 最近 N 轮历史
// 调用方需要先保存当前用户问题到 chat_message，因此 messages 中已包含当前问题
func BuildContext(db *gorm.DB, sessionID, userID string) ([]*schema.Message, error) {
	messages, err := ListMessages(db, sessionID, userID)
	if err != nil {
		return nil, err
	}

	var history []*schema.Message
	history = append(history, &schema.Message{Role: schema.System, Content: SystemPrompt})

	// 取最近 ContextLimit 条 user/assistant
	start := 0
	if len(messages) > ContextLimit {
		start = len(messages) - ContextLimit
	}
	for i := start; i < len(messages); i++ {
		m := messages[i]
		var role schema.RoleType
		switch m.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		default:
			continue
		}
		history = append(history, &schema.Message{Role: role, Content: m.Content})
	}

	return history, nil
}

// UpdateSessionTitle 根据首句问题更新会话标题（可选，首次发送后调用）
func UpdateSessionTitle(db *gorm.DB, sessionID, userID, title string) error {
	var session models.ChatSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("会话不存在")
		}
		return err
	}
	return db.Model(&session).Update("title", title).Error
}

// SummaryTitle 截取问题前 20 字作为标题
func SummaryTitle(question string) string {
	r := []rune(question)
	if len(r) <= 20 {
		return question
	}
	return string(r[:20]) + "..."
}
