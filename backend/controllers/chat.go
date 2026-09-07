package controllers

import (
	"context"
	"net/http"
	"time"

	"aiops/internal/agent"
	"aiops/internal/chat"
	"aiops/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ChatController 对话控制器
type ChatController struct {
	DB       *gorm.DB
	registry *agent.Registry
}

// NewChatController 创建控制器
func NewChatController(db *gorm.DB, registry *agent.Registry) *ChatController {
	return &ChatController{DB: db, registry: registry}
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

// sseJob 统一所有 SSE 输出请求，由单个 sender goroutine 串行写入 ctx.Writer
type sseJob struct {
	event string
	data  interface{}
	raw   []byte // 非 SSE 事件的原始字节（心跳注释）
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
	cfg, err := chat.DefaultConfig(c.DB, models.LLMModelTypeChat)
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

	// 4. 可取消 context：前端中止（EventSource.close）或代理断连时触发 cancel
	cancelCtx, cancelFn := context.WithCancel(ctx.Request.Context())
	defer cancelFn()

	// 5. SSE 响应头（必须在发送任何事件前设置）
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.Writer.WriteHeaderNow()

	// 检测客户端断连，触发 cancel
	clientGone := ctx.Writer.CloseNotify()
	go func() {
		select {
		case <-clientGone:
			cancelFn()
		case <-cancelCtx.Done():
		}
	}()

	// 6. 初始化 Eino 客户端（使用可取消的 context）
	cm, err := chat.NewClient(cancelCtx, cfg)
	if err != nil {
		c.writeSSEError(ctx, 500, "初始化大模型失败")
		return
	}

	// ---------- 核心修复：channel + 单一 sender goroutine 串行化所有写入 ----------

	// 带缓冲的 channel，足够容纳并发工具回调 + 流式文本 + 心跳
	jobCh := make(chan sseJob, 256)

	// sender 是唯一允许触摸 ctx.Writer 的 goroutine
	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()
		for {
			select {
			case <-clientGone:
				return
			case job, ok := <-jobCh:
				if !ok {
					return
				}
				if job.raw != nil {
					_, _ = ctx.Writer.Write(job.raw)
				} else {
					ctx.SSEvent(job.event, job.data)
				}
				ctx.Writer.Flush()
			case <-heartbeat.C:
				// SSE 注释心跳，EventSource 会忽略，但保持连接活跃
				_, _ = ctx.Writer.Write([]byte(": ping\n\n"))
				ctx.Writer.Flush()
			}
		}
	}()

	// 统一 sendEvent：任何 goroutine 调这个都是推 channel，不直接写 ctx.Writer
	sendEvent := func(event string, data interface{}) bool {
		select {
		case <-clientGone:
			return false
		case jobCh <- sseJob{event: event, data: data}:
			return true
		}
	}

	// streamText 同样走 channel
	streamText := func(text string) bool {
		runes := []rune(text)
		chunkSize := 4
		for i := 0; i < len(runes); i += chunkSize {
			select {
			case <-clientGone:
				return false
			default:
			}
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}
			select {
			case <-clientGone:
				return false
			case jobCh <- sseJob{event: "message", data: gin.H{"chunk": string(runes[i:end])}}:
			}
		}
		return true
	}

	// 发送初始状态，确保代理层立即收到数据
	if !sendEvent("status", gin.H{"status": "正在思考..."}) {
		close(jobCh)
		<-senderDone
		return
	}

	// ---------- 6. 使用 Function Calling Agent 执行对话 ----------
	ag := agent.New(cm, c.registry, agent.Options{
		Instructions:      chat.SystemPrompt,
		EnforceDataSource: true,
	})

	var fullContent string
	var promptTokens, completionTokens, totalTokens int

	resp, err := ag.Run(cancelCtx, messages, &agent.Callbacks{
		OnStatus: func(status string) {
			sendEvent("status", gin.H{"status": status})
		},
		OnToolCall: func(name, args string) {
			sendEvent("tool_call", gin.H{"name": name, "arguments": args})
		},
		OnToolResult: func(name, result string) {
			sendEvent("tool_result", gin.H{"name": name, "result": result})
		},
		OnChunk: func(chunk string) {
			fullContent = chunk
			streamText(chunk)
		},
	})
	if err != nil {
		sendEvent("error", gin.H{"code": 500, "message": "生成失败: " + err.Error()})
		close(jobCh)
		<-senderDone
		return
	}

	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		promptTokens = resp.ResponseMeta.Usage.PromptTokens
		completionTokens = resp.ResponseMeta.Usage.CompletionTokens
		totalTokens = resp.ResponseMeta.Usage.TotalTokens
	}

	// 保存助手消息
	assistantMsg, err := chat.SaveAssistantMessage(c.DB, sessionID, userID, fullContent, promptTokens, completionTokens, totalTokens)
	if err != nil {
		sendEvent("error", gin.H{"code": 500, "message": "保存回复失败"})
		close(jobCh)
		<-senderDone
		return
	}

	// 仅当会话仍为默认标题时才更新标题
	var session models.ChatSession
	if err := c.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err == nil {
		if session.Title == "新会话" {
			_ = chat.UpdateSessionTitle(c.DB, sessionID, userID, chat.SummaryTitle(question))
		}
	}

	sendEvent("done", gin.H{
		"message_id": assistantMsg.ID,
		"session_id": sessionID,
		"done":       true,
	})

	// 关闭 channel，让 sender goroutine 退出
	close(jobCh)
	<-senderDone
}

func (c *ChatController) writeSSEError(ctx *gin.Context, code int, message string) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.SSEvent("error", gin.H{"code": code, "message": message})
	ctx.Writer.Flush()
}
