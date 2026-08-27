package controllers

import (
	"net/http"
	"strings"

	"aiops/internal/chat"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ChatController 对话控制器
type ChatController struct {
	DB *gorm.DB
}

// NewChatController 创建控制器
func NewChatController(db *gorm.DB) *ChatController {
	return &ChatController{DB: db}
}

// currentUserID 从 gin context 取当前用户 ID（由 JWT/Session middleware 注入）
func currentUserID(ctx *gin.Context) string {
	uid, _ := ctx.Get("user_id")
	if uid == nil {
		return ""
	}
	return uid.(string)
}

// ListSessions 获取当前用户会话列表
func (c *ChatController) ListSessions(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	items, err := chat.ListSessions(c.DB, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// CreateSession 创建新会话
func (c *ChatController) CreateSession(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	ctx.ShouldBindJSON(&req)

	session, err := chat.CreateSession(c.DB, userID, req.Title)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": session})
}

// DeleteSession 删除会话
func (c *ChatController) DeleteSession(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	sessionID := ctx.Param("id")
	if err := chat.DeleteSession(c.DB, sessionID, userID); err != nil {
		if err.Error() == "会话不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// ListMessages 获取会话历史消息
func (c *ChatController) ListMessages(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	sessionID := ctx.Param("id")
	messages, err := chat.ListMessages(c.DB, sessionID, userID)
	if err != nil {
		if err.Error() == "会话不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": messages})
}

// StreamChat SSE 流式对话
func (c *ChatController) StreamChat(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}

	sessionID := ctx.Param("id")
	question := ctx.Query("content")
	if question == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "问题内容不能为空"})
		return
	}

	// 1. 保存用户问题
	if _, err := chat.SaveUserMessage(c.DB, sessionID, userID, question); err != nil {
		if err.Error() == "会话不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存问题失败"})
		return
	}

	// 2. 加载默认 LLM 配置
	cfg, err := chat.DefaultConfig(c.DB)
	if err != nil {
		c.writeSSEError(ctx, 500, "未配置默认大模型")
		return
	}

	// 3. 组装上下文
	messages, err := chat.BuildContext(c.DB, sessionID, userID)
	if err != nil {
		c.writeSSEError(ctx, 500, "组装上下文失败")
		return
	}

	// 4. 初始化 Eino 客户端
	cm, err := chat.NewClient(ctx, cfg)
	if err != nil {
		c.writeSSEError(ctx, 500, "初始化大模型失败")
		return
	}

	// 5. 流式生成
	streamReader, err := cm.Stream(ctx, messages)
	if err != nil {
		c.writeSSEError(ctx, 500, "启动流式生成失败")
		return
	}

	// 6. SSE 输出
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.Writer.WriteHeaderNow()

	var fullContent strings.Builder
	var promptTokens, completionTokens, totalTokens int

	chunkCh, errCh := chat.StreamMessages(ctx, streamReader)
	for {
		select {
		case chunk, ok := <-chunkCh:
			if !ok {
				chunkCh = nil
				break
			}
			fullContent.WriteString(chunk)
			ctx.SSEvent("message", gin.H{"chunk": chunk})
			ctx.Writer.Flush()
		case err := <-errCh:
			if err != nil {
				ctx.SSEvent("error", gin.H{"code": 500, "message": "生成失败: " + err.Error()})
				ctx.Writer.Flush()
				return
			}
		case <-ctx.Request.Context().Done():
			return
		}
		if chunkCh == nil {
			break
		}
	}

	// 保存助手消息
	assistantMsg, err := chat.SaveAssistantMessage(c.DB, sessionID, userID, fullContent.String(), promptTokens, completionTokens, totalTokens)
	if err != nil {
		ctx.SSEvent("error", gin.H{"code": 500, "message": "保存回复失败"})
		ctx.Writer.Flush()
		return
	}

	// 仅当会话仍为默认标题时才更新标题
	var session models.ChatSession
	if err := c.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err == nil {
		if session.Title == "新会话" {
			if err := chat.UpdateSessionTitle(c.DB, sessionID, userID, chat.SummaryTitle(question)); err != nil {
				// 标题更新失败不影响主流程，仅记录
			}
		}
	}

	ctx.SSEvent("done", gin.H{
		"message_id": assistantMsg.ID,
		"session_id": sessionID,
		"done":       true,
	})
	ctx.Writer.Flush()
}

func (c *ChatController) writeSSEError(ctx *gin.Context, code int, message string) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.SSEvent("error", gin.H{"code": code, "message": message})
	ctx.Writer.Flush()
}
