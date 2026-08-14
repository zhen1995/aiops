<template>
  <div class="chat-page">
    <div class="chat-card card">
      <div class="chat-header">
        <div>
          <h3 class="card-title">AI 运维助手</h3>
          <p class="card-sub">基于大模型的智能运维问答、告警解读与根因分析</p>
        </div>
        <div class="model-selector">
          <span class="muted">当前模型</span>
          <select v-model="currentModel" class="select">
            <option value="gpt-4o">GPT-4o</option>
            <option value="qwen-max">Qwen-Max</option>
            <option value="deepseek-coder">DeepSeek</option>
          </select>
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
              <span class="muted">{{ fmtTime(msg.time) }}</span>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
        </div>
      </div>

      <div class="chat-input-area">
        <div class="chat-toolbar">
          <button class="tool-btn" title="快捷提问" @click="quickAsk('今天有哪些 P0 告警？')">今日 P0 告警</button>
          <button class="tool-btn" title="快捷提问" @click="quickAsk('分析一下当前最高优先级的根因')">根因分析</button>
          <button class="tool-btn" title="快捷提问" @click="quickAsk('生成今日巡检摘要')">巡检摘要</button>
        </div>
        <div class="chat-input">
          <textarea
            v-model="input"
            rows="2"
            placeholder="输入问题，例如：今天系统状态如何？"
            @keydown.enter.prevent="send"
          />
          <button class="btn btn-primary send-btn" :disabled="!input.trim()" @click="send">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="22" y1="2" x2="11" y2="13" />
              <polygon points="22 2 15 22 11 13 2 9 22 2" />
            </svg>
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { chatMessages, fmtTime } from '../mock/data'

const messages = ref([...chatMessages])
const input = ref('')
const currentModel = ref('gpt-4o')
const messagesRef = ref(null)

const quickAsk = (text) => {
  input.value = text
}

const send = () => {
  const text = input.value.trim()
  if (!text) return

  messages.value.push({
    role: 'user',
    content: text,
    time: new Date().toISOString()
  })
  input.value = ''

  nextTick(() => {
    messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  })

  setTimeout(() => {
    messages.value.push({
      role: 'assistant',
      content: '已收到你的问题，正在基于当前监控数据生成回答…\n（此为前端原型演示，真实回复将由后端 LLM 服务返回）',
      time: new Date().toISOString()
    })
    nextTick(() => {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    })
  }, 800)
}
</script>

<style scoped>
.chat-page { max-width: 960px; margin: 0 auto; }

.chat-card {
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

.select {
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

.chat-input textarea:focus { border-color: var(--c-primary); }

.send-btn {
  align-self: flex-end;
  padding: 10px 18px;
}

.send-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
