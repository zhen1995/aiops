<template>
  <div class="chat-page">
    <!-- 左侧会话列表 -->
    <aside class="session-list">
      <div class="session-header">
        <button class="btn btn-primary" @click="createNewSession">+ 新会话</button>
      </div>
      <div class="session-items">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.id === currentSessionId }"
          @click="switchSession(session.id)"
        >
          <div class="session-title">{{ session.title }}</div>
          <div class="session-meta">{{ session.last_msg || '暂无消息' }}</div>
          <button class="session-delete" @click.stop="removeSession(session.id)">×</button>
        </div>
      </div>
    </aside>

    <!-- 右侧主区域 -->
    <div class="chat-card card">
      <div class="chat-header">
        <div>
          <h3 class="card-title">AI 运维助手</h3>
          <p class="card-sub">基于大模型的智能运维问答、告警解读与根因分析</p>
        </div>
        <div class="model-selector">
          <span class="muted">当前模型</span>
          <span class="model-name">{{ defaultModelName }}</span>
        </div>
      </div>

      <div ref="messagesRef" class="chat-messages">
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="message"
          :class="msg.role"
        >
          <div class="message-avatar">
            <span v-if="msg.role === 'assistant'">AI</span>
            <span v-else>我</span>
          </div>
          <div class="message-body">
            <div class="message-meta">
              <b>{{ msg.role === 'assistant' ? 'AIOPS 助手' : '运维管理员' }}</b>
              <span class="muted">{{ fmtTime(msg.created_at) }}</span>
            </div>
            <div class="message-content">
              <template v-if="isThinkingMessage(msg)">
                <span class="thinking-indicator">
                  <span class="thinking-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </span>
                  <span class="thinking-text">正在思考中</span>
                </span>
              </template>
              <MarkdownContent v-else-if="msg.role === 'assistant'" :content="msg.content" />
              <template v-else>{{ msg.content }}</template>
            </div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <div class="chat-toolbar">
          <button class="tool-btn" @click="quickAsk('今天有哪些 P0 告警？')">今日 P0 告警</button>
          <button class="tool-btn" @click="quickAsk('分析一下当前最高优先级的根因')">根因分析</button>
          <button class="tool-btn" @click="quickAsk('生成今日巡检摘要')">巡检摘要</button>
        </div>
        <div class="chat-input">
          <textarea
            v-model="input"
            rows="2"
            :disabled="isStreaming"
            placeholder="输入问题，例如：今天系统状态如何？"
            @keydown.enter.prevent="send"
          />
          <button class="btn btn-primary send-btn" :disabled="!input.trim() || isStreaming" @click="send">
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fmtTime } from '../mock/data'
import { getSessions, createSession, deleteSession, getMessages, streamChat } from '../api/chat'
import { llmConfigApi } from '../api/llmConfig'
import MarkdownContent from '../components/MarkdownContent.vue'

const route = useRoute()
const router = useRouter()

const sessions = ref([])
const currentSessionId = ref('')
const messages = ref([])
const input = ref('')
const isStreaming = ref(false)
const isThinking = ref(false)
const messagesRef = ref(null)
const defaultModelName = ref('默认模型')

function isThinkingMessage(msg) {
  return msg.role === 'assistant' && msg.content === '' && isThinking.value
}

let currentEventSource = null

onMounted(() => {
  loadSessions()
  loadDefaultModelName()
  handleRcaParam()
})

async function loadDefaultModelName() {
  try {
    const configs = await llmConfigApi.list()
    const defaultConfig = configs.find(c => c.is_default === 1 && c.is_enabled === 1)
    if (defaultConfig) {
      defaultModelName.value = defaultConfig.name
    } else {
      defaultModelName.value = '未配置默认模型'
    }
  } catch (e) {
    defaultModelName.value = '默认模型'
  }
}

onUnmounted(() => {
  if (currentEventSource) {
    currentEventSource.close()
    currentEventSource = null
  }
})

function resetStreaming() {
  if (currentEventSource) {
    currentEventSource.close()
    currentEventSource = null
  }
  isStreaming.value = false
  isThinking.value = false
}

async function loadSessions() {
  try {
    const data = await getSessions()
    sessions.value = data || []
    if (sessions.value.length === 0) {
      await handleCreateSession()
    } else if (!currentSessionId.value) {
      currentSessionId.value = sessions.value[0].id
      await loadMessages(currentSessionId.value)
    }
  } catch (e) {
    console.error('加载会话失败', e)
  }
}

async function handleCreateSession() {
  resetStreaming()
  try {
    const data = await createSession('新会话')
    sessions.value.unshift(data)
    currentSessionId.value = data.id
    messages.value = []
  } catch (e) {
    console.error('创建会话失败', e)
  }
}

function createNewSession() {
  handleCreateSession()
}

async function switchSession(id) {
  resetStreaming()
  if (id === currentSessionId.value) return
  currentSessionId.value = id
  await loadMessages(id)
}

async function loadMessages(id) {
  try {
    const data = await getMessages(id)
    messages.value = data || []
    scrollToBottom()
  } catch (e) {
    console.error('加载消息失败', e)
  }
}

async function removeSession(id) {
  try {
    await deleteSession(id)
    sessions.value = sessions.value.filter(s => s.id !== id)
    if (currentSessionId.value === id) {
      if (sessions.value.length > 0) {
        await switchSession(sessions.value[0].id)
      } else {
        await handleCreateSession()
      }
    }
  } catch (e) {
    console.error('删除会话失败', e)
  }
}

function quickAsk(text) {
  input.value = text
}

function decodeEvent(encoded) {
  try {
    return JSON.parse(decodeURIComponent(atob(encoded)))
  } catch (e) {
    console.error('解析 RCA 事件参数失败', e)
    return null
  }
}

function buildRcaPrompt(event) {
  const statusText = event.is_recovered ? '已恢复' : '告警中'
  const severityText = { 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }[event.severity] || event.severity
  const timeStr = event.trigger_time ? new Date(event.trigger_time * 1000).toLocaleString() : '-'
  return `请对以下告警事件进行根因分析：
- 规则名称：${event.rule_name || '-'}
- 告警对象：${event.target_ident || '-'}
- 触发时间：${timeStr}
- 级别：${severityText}
- 状态：${statusText}
- 标签：${event.tags || '-'}
- 触发值：${event.trigger_value || '-'}

请查询相关 Prometheus 指标和 Elasticsearch 日志，排查宕机原因，并给出根因、证据链和修复建议。`
}

async function handleRcaParam() {
  if (route.query.rca !== '1' || !route.query.event) return
  const event = decodeEvent(route.query.event)
  if (!event) {
    alert('根因分析参数无效')
    clearRcaQuery()
    return
  }
  const prompt = buildRcaPrompt(event)
  input.value = prompt
  clearRcaQuery()
  await nextTick()
  send()
}

function clearRcaQuery() {
  router.replace({ path: '/chat', query: {} })
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

async function send() {
  const text = input.value.trim()
  if (!text || isStreaming.value) return

  let sessionId = currentSessionId.value
  if (!sessionId) {
    try {
      const data = await createSession('新会话')
      sessions.value.unshift(data)
      currentSessionId.value = data.id
      sessionId = data.id
      messages.value = []
    } catch (e) {
      console.error('创建会话失败', e)
      return
    }
  }

  messages.value.push({
    role: 'user',
    content: text,
    created_at: new Date().toISOString()
  })
  input.value = ''
  scrollToBottom()

  isStreaming.value = true
  isThinking.value = true
  const assistantIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    created_at: new Date().toISOString()
  })

  currentEventSource = streamChat(sessionId, text, {
    onChunk: (chunk) => {
      isThinking.value = false
      messages.value[assistantIndex].content += chunk
      scrollToBottom()
    },
    onDone: () => {
      isStreaming.value = false
      isThinking.value = false
      loadSessions()
    },
    onError: (err) => {
      isStreaming.value = false
      isThinking.value = false
      messages.value[assistantIndex].content += '\n[错误：' + err + ']'
      scrollToBottom()
    }
  })
}
</script>

<style scoped>
.chat-page {
  display: flex;
  gap: 16px;
  max-width: 1200px;
  margin: 0 auto;
}

.session-list {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 12px;
  overflow: hidden;
}

.session-header {
  padding: 14px;
  border-bottom: 1px solid var(--c-border);
}

.session-header button {
  width: 100%;
}

.session-items {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  position: relative;
  padding: 10px 28px 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
}

.session-item:hover,
.session-item.active {
  background: var(--c-primary-tint);
}

.session-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--c-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-meta {
  font-size: 11px;
  color: var(--c-text-2);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-delete {
  position: absolute;
  right: 6px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: var(--c-text-2);
  cursor: pointer;
  display: none;
}

.session-item:hover .session-delete {
  display: block;
}

.session-delete:hover {
  color: var(--c-danger, #c93b3b);
}

.chat-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: calc(100vh - var(--topbar-h) - 44px);
  padding: 0;
  overflow: hidden;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 20px 14px;
  border-bottom: 1px solid var(--c-border);
}

.model-selector {
  display: flex;
  align-items: center;
  gap: 10px;
}

.model-name {
  padding: 6px 10px;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: var(--c-bg);
}

.message {
  display: flex;
  gap: 12px;
  margin-bottom: 18px;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--c-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.message.user .message-avatar { background: var(--c-text-3); }

.message-body {
  flex: 1;
  min-width: 0;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.message-meta b { font-size: 13px; }

.message-content {
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 10px;
  padding: 12px 14px;
  font-size: 13.5px;
  line-height: 1.7;
  white-space: pre-wrap;
}

.message.user .message-content { background: var(--c-primary-tint); border-color: #cfe6e2; }

.thinking-indicator {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--c-text-2);
  font-size: 13.5px;
}

.thinking-dots {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  width: 28px;
  height: 8px;
}

.thinking-dots span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--c-primary);
  animation: thinking-bounce 1.4s infinite ease-in-out both;
}

.thinking-dots span:nth-child(1) { animation-delay: -0.32s; }
.thinking-dots span:nth-child(2) { animation-delay: -0.16s; }
.thinking-dots span:nth-child(3) { animation-delay: 0s; }

@keyframes thinking-bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

.chat-input-area {
  padding: 14px 20px 18px;
  border-top: 1px solid var(--c-border);
  background: var(--c-surface);
}

.chat-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}

.tool-btn {
  padding: 4px 10px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-bg);
  color: var(--c-text-2);
  font-size: 12px;
  cursor: pointer;
}

.tool-btn:hover { border-color: var(--c-primary); color: var(--c-primary); }

.chat-input {
  display: flex;
  gap: 10px;
}

.chat-input textarea {
  flex: 1;
  resize: none;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 14px;
  font-family: inherit;
  outline: none;
}

.chat-input textarea:disabled {
  background: var(--c-bg);
  opacity: 0.7;
  cursor: not-allowed;
}

.chat-input textarea:focus { border-color: var(--c-primary); }

.send-btn {
  align-self: flex-end;
  padding: 10px 18px;
}

.send-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
