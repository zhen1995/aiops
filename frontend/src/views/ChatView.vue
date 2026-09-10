<template>
  <div class="chat-page">
    <!-- 左侧会话列表 -->
    <aside class="session-list">
      <div class="session-header">
        <button class="btn btn-primary" @click="createNewSession">{{ $t('chat.session.newButton') }}</button>
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
          <div class="session-meta">{{ session.last_msg || $t('chat.session.noMessage') }}</div>
          <button class="session-delete" @click.stop="removeSession(session.id)">×</button>
        </div>
      </div>
    </aside>

    <!-- 右侧主区域 -->
    <div class="chat-card card">
      <div class="chat-header">
        <div>
          <h3 class="card-title">{{ $t('chat.header.title') }}</h3>
          <p class="card-sub">{{ $t('chat.header.subtitle') }}</p>
        </div>
        <div class="model-selector">
          <span class="muted">{{ $t('chat.header.currentModel') }}</span>
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
            <span v-else>{{ $t('chat.message.userAvatar') }}</span>
          </div>
          <div class="message-body">
            <div class="message-meta">
              <b>{{ msg.role === 'assistant' ? $t('chat.message.assistantName') : $t('chat.message.userName') }}</b>
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
                  <span class="thinking-text">{{ $t('chat.message.thinking') }}</span>
                </span>
              </template>
              <MarkdownContent v-else-if="msg.role === 'assistant'" :content="msg.content" />
              <template v-else>{{ msg.content }}</template>
            </div>
            <div v-if="msg.role === 'user'" class="message-actions">
              <button
                class="copy-btn"
                :class="{ copied: copyState[idx] }"
                @click="copyMessage(msg, idx)"
              >{{ copyState[idx] ? $t('chat.message.copied') : $t('chat.message.copy') }}</button>
            </div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <div class="chat-toolbar">
          <button class="tool-btn" @click="quickAsk('今天有哪些 P0 告警？')">{{ $t('chat.toolbar.p0Alerts') }}</button>
          <button class="tool-btn" @click="quickAsk('分析一下当前最高优先级的根因')">{{ $t('chat.toolbar.rootCause') }}</button>
          <button class="tool-btn" @click="quickAsk('生成今日巡检摘要')">{{ $t('chat.toolbar.inspectionSummary') }}</button>
        </div>
        <div class="chat-input">
          <textarea
            v-model="input"
            rows="2"
            :disabled="isStreaming"
            :placeholder="$t('chat.input.placeholder')"
            @keydown.enter.prevent="send"
          />
          <button
            v-if="!isStreaming"
            class="btn btn-primary send-btn"
            :disabled="!input.trim()"
            @click="send"
          >{{ $t('chat.input.send') }}</button>
          <button
            v-else
            class="btn btn-danger send-btn"
            @click="abortStreaming"
          >
            <span class="stop-icon"></span>{{ $t('chat.input.abort') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { fmtTime } from '../mock/data'
import { getSessions, createSession, deleteSession, getMessages, streamChat } from '../api/chat'
import { llmConfigApi } from '../api/llmConfig'
import MarkdownContent from '../components/MarkdownContent.vue'

const { t } = useI18n({ useScope: 'global' })

const route = useRoute()
const router = useRouter()

const sessions = ref([])
const currentSessionId = ref('')
const messages = ref([])
const input = ref('')
const isStreaming = ref(false)
const isThinking = ref(false)
const messagesRef = ref(null)
const defaultModelName = ref(t('chat.header.defaultModel'))
const currentAssistantIndex = ref(-1)
const manuallyAborting = ref(false)
const copyState = ref({})

async function copyMessage(msg, idx) {
  try {
    await navigator.clipboard.writeText(msg.content)
  } catch (e) {
    const ta = document.createElement('textarea')
    ta.value = msg.content
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
  copyState.value = { ...copyState.value, [idx]: true }
  setTimeout(() => {
    copyState.value = { ...copyState.value, [idx]: false }
  }, 1500)
}

function isThinkingMessage(msg) {
  return msg.role === 'assistant' && msg.content === '' && isThinking.value
}

let currentEventSource = null

onMounted(() => {
  loadSessions()
  loadDefaultModelName()
  handleRcaParam()
  handleSearchParam()
})

async function loadDefaultModelName() {
  try {
    const configs = await llmConfigApi.list()
    const defaultConfig = configs.find(c => c.is_default === 1 && c.is_enabled === 1)
    if (defaultConfig) {
      defaultModelName.value = defaultConfig.name
    } else {
      defaultModelName.value = t('chat.header.noDefaultModel')
    }
  } catch (e) {
    defaultModelName.value = t('chat.header.defaultModel')
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
    const data = await createSession(t('chat.session.defaultTitle'))
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
  const statusText = event.is_recovered ? t('chat.rca.statusRecovered') : t('chat.rca.statusFiring')
  const severityText = {
    1: t('chat.rca.severityP1'),
    2: t('chat.rca.severityP2'),
    3: t('chat.rca.severityP3')
  }[event.severity] || event.severity
  const timeStr = event.trigger_time ? new Date(event.trigger_time * 1000).toLocaleString() : '-'
  return t('chat.rca.prompt', {
    ruleName: event.rule_name || '-',
    targetIdent: event.target_ident || '-',
    time: timeStr,
    severity: severityText,
    status: statusText,
    tags: event.tags || '-',
    triggerValue: event.trigger_value || '-'
  })
}

async function handleRcaParam() {
  if (route.query.rca !== '1' || !route.query.event) return
  const event = decodeEvent(route.query.event)
  if (!event) {
    alert(t('chat.rca.invalidParam'))
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

// 全局搜索「问 AI」入口：/chat?q=关键词，代填并发送
async function handleSearchParam() {
  const q = route.query.q
  if (!q || typeof q !== 'string') return
  input.value = q
  router.replace({ path: '/chat', query: {} })
  await nextTick()
  send()
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
      const data = await createSession(t('chat.session.defaultTitle'))
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
  manuallyAborting.value = false
  currentAssistantIndex.value = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    created_at: new Date().toISOString()
  })

  currentEventSource = streamChat(sessionId, text, {
    onChunk: (chunk) => {
      isThinking.value = false
      messages.value[currentAssistantIndex.value].content += chunk
      scrollToBottom()
    },
    onDone: () => {
      if (manuallyAborting.value) return
      isStreaming.value = false
      isThinking.value = false
      currentEventSource = null
      loadSessions()
    },
    onError: (err) => {
      if (manuallyAborting.value) {
        // 用户主动中止，不显示错误
        manuallyAborting.value = false
        isStreaming.value = false
        isThinking.value = false
        currentEventSource = null
        return
      }
      isStreaming.value = false
      isThinking.value = false
      currentEventSource = null
      messages.value[currentAssistantIndex.value].content += t('chat.message.error', { err })
      scrollToBottom()
    }
  })
}

function abortStreaming() {
  manuallyAborting.value = true
  if (currentEventSource) {
    currentEventSource.close()
    currentEventSource = null
  }
  isStreaming.value = false
  isThinking.value = false

  // 在当前助手消息末尾追加中止标记
  if (currentAssistantIndex.value >= 0 && messages.value[currentAssistantIndex.value]) {
    const msg = messages.value[currentAssistantIndex.value]
    msg.content += (msg.content ? '\n\n' : '') + t('chat.message.aborted')
    scrollToBottom()
  }

  // 保存当前已生成内容到数据库（异步，不阻塞 UI）
  loadSessions()
}
</script>

<style scoped>
.chat-page {
  display: flex;
  gap: 16px;
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
  height: calc(100vh - var(--topbar-h) - 54px);
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
  background: var(--c-surface);
}

.message {
  display: flex;
  gap: 12px;
  margin-bottom: 18px;
}

.message.user { flex-direction: row-reverse; }

.message.user .message-body {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.message.user .message-meta { flex-direction: row-reverse; }

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

.message-actions {
  display: flex;
  justify-content: flex-end;
  width: 100%;
  margin-top: 4px;
}

.copy-btn {
  border: none;
  background: transparent;
  color: var(--c-text-2);
  font-size: 12px;
  cursor: pointer;
  padding: 0 4px;
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.15s, color 0.15s;
}

.message.user:hover .copy-btn { opacity: 1; }

.copy-btn:hover { color: var(--c-primary); }

.copy-btn.copied { opacity: 1; color: var(--c-primary); cursor: default; }

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

.btn-danger {
  background: var(--c-danger, #e04040);
  color: #fff;
  border: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-danger:hover { background: #c93535; }

.stop-icon {
  display: inline-block;
  width: 10px;
  height: 10px;
  background: #fff;
  border-radius: 2px;
}
</style>
