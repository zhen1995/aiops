# 异常检测列表根因分析按钮实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在异常检测列表为每条告警事件增加“根因分析”按钮，点击后跳转对话页并自动触发 Agent 对指标和日志的根因排查。

**Architecture:** 复用现有 Vue Router、ChatView、Eino Agent 与 Prometheus/Elasticsearch 工具。告警事件通过 URL 查询参数传递到对话页，对话页解析后自动生成 Prompt 并发送。无需新增后端接口。

**Tech Stack:** Vue 3 / Vue Router 4 / Vite / Go / Gin / Eino

## Global Constraints

- 保持中文用户可见文本。
- 不新增后端接口，复用 `/api/chat/sessions/:id/stream`。
- 事件数据通过 URL 查询参数传递，需做 Base64 + URL 编码。
- 刷新页面不应重复触发分析。
- 前端构建必须成功（`npm run build`）。
- 后端构建必须成功（`go build ./cmd/center && go vet ./...`）。

---

## 文件变更映射

- `frontend/src/views/AnomalyView.vue`：增加操作列、根因分析按钮、事件编码跳转逻辑。
- `frontend/src/views/ChatView.vue`：解析路由参数、自动生成并发送 RCA Prompt、清理参数。
- `frontend/src/router/index.js`：确认 `/chat` 路由支持查询参数（通常已支持，无需修改）。

---

### Task 1: AnomalyView.vue 增加操作列与根因分析按钮

**Files:**
- Modify: `frontend/src/views/AnomalyView.vue:58-94`（表头与表格行）
- Modify: `frontend/src/views/AnomalyView.vue:132-142`（增加辅助函数）

**Interfaces:**
- Consumes: `events` 数组中的事件对象字段：`id`, `rule_name`, `target_ident`, `tags`, `trigger_time`, `trigger_value`, `severity`, `is_recovered`。
- Produces: `buildRcaQuery(event)` 函数返回可用于 `window.location.hash` 的查询字符串；点击后跳转到 `/#/chat?rca=1&event=...`。

- [ ] **Step 1: 表头增加“操作”列**

  将表头从 7 列改为 8 列：

  ```vue
  <thead>
    <tr>
      <th>规则名称</th>
      <th>级别</th>
      <th>状态</th>
      <th>告警对象</th>
      <th>触发时间</th>
      <th>标签</th>
      <th>触发值</th>
      <th>操作</th>
    </tr>
  </thead>
  ```

- [ ] **Step 2: 每行增加操作单元格**

  在 `<tr v-for="ev in events" :key="ev.id">` 最后一列后追加：

  ```vue
  <td>
    <button class="btn btn-sm" @click="goToRca(ev)">根因分析</button>
  </td>
  ```

- [ ] **Step 3: 更新空状态与加载状态的 colspan**

  将两处 `colspan="7"` 改为 `colspan="8"`：

  ```vue
  <tr v-if="!loading && events.length === 0">
    <td colspan="8" class="empty-row">暂无告警事件数据</td>
  </tr>
  <tr v-if="loading">
    <td colspan="8" class="empty-row">加载中...</td>
  </tr>
  ```

- [ ] **Step 4: 实现 goToRca 与事件编码函数**

  在 `<script setup>` 中导入 `useRouter`：

  ```js
  import { useRouter } from 'vue-router'
  ```

  实例化路由：

  ```js
  const router = useRouter()
  ```

  增加辅助函数：

  ```js
  function encodeEvent(event) {
    const payload = {
      rule_name: event.rule_name,
      target_ident: event.target_ident,
      tags: event.tags,
      trigger_time: event.trigger_time,
      trigger_value: event.trigger_value,
      severity: event.severity,
      is_recovered: event.is_recovered
    }
    return btoa(encodeURIComponent(JSON.stringify(payload)))
  }

  function goToRca(event) {
    router.push({
      path: '/chat',
      query: { rca: '1', event: encodeEvent(event) }
    })
  }
  ```

- [ ] **Step 5: 验证 AnomalyView.vue 语法**

  运行：

  ```bash
  cd /e/aiops/frontend && npm run build
  ```

  预期：构建成功，无 AnomalyView 相关错误。

- [ ] **Step 6: 提交**

  ```bash
  cd /e/aiops
  git add frontend/src/views/AnomalyView.vue
  git commit -m "feat(anomaly): 告警事件列表增加根因分析操作列"
  ```

---

### Task 2: ChatView.vue 解析 RCA 参数并自动发送

**Files:**
- Modify: `frontend/src/views/ChatView.vue:93-118`（导入与初始化）
- Modify: `frontend/src/views/ChatView.vue:115-118`（onMounted）
- Modify: `frontend/src/views/ChatView.vue:226-268`（send 函数）

**Interfaces:**
- Consumes: Vue Router 查询参数 `rca` 和 `event`；事件对象字段同 Task 1。
- Produces: 自动生成的 RCA Prompt 字符串；调用 `send()` 发送后清除路由参数。

- [ ] **Step 1: 导入 useRoute**

  在 `<script setup>` 顶部：

  ```js
  import { useRoute } from 'vue-router'
  ```

  实例化：

  ```js
  const route = useRoute()
  ```

- [ ] **Step 2: 实现事件解码函数**

  在 `<script setup>` 中增加：

  ```js
  function decodeEvent(encoded) {
    try {
      return JSON.parse(decodeURIComponent(atob(encoded)))
    } catch (e) {
      console.error('解析 RCA 事件参数失败', e)
      return null
    }
  }
  ```

- [ ] **Step 3: 实现 RCA Prompt 生成函数**

  ```js
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
  ```

- [ ] **Step 4: 在 onMounted 中处理 RCA 参数**

  修改 `onMounted`：

  ```js
  onMounted(() => {
    loadSessions()
    loadDefaultModelName()
    handleRcaParam()
  })
  ```

  新增 `handleRcaParam`：

  ```js
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
  ```

  注意：`clearRcaQuery()` 要在发送前调用，避免刷新重复触发。

- [ ] **Step 5: 确保 send 函数在会话未创建时自动创建**

  修改 `send` 函数开头：

  ```js
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
    // ... 后续逻辑保持不变
  }
  ```

  将 `const sessionId = currentSessionId.value` 替换为上面的 `let sessionId` 块。

- [ ] **Step 6: 验证 ChatView.vue 语法与构建**

  运行：

  ```bash
  cd /e/aiops/frontend && npm run build
  ```

  预期：构建成功。

- [ ] **Step 7: 提交**

  ```bash
  cd /e/aiops
  git add frontend/src/views/ChatView.vue
  git commit -m "feat(chat): 支持通过 URL 参数自动触发告警事件根因分析"
  ```

---

### Task 3: 后端构建验证

**Files:**
- 无代码变更，仅验证构建。

- [ ] **Step 1: 运行后端构建**

  ```bash
  cd /e/aiops/backend && go build ./cmd/center && go vet ./...
  ```

  预期：无错误。

- [ ] **Step 2: 提交（如 center.exe 被覆盖需还原）**

  若 `backend/center.exe` 被修改：

  ```bash
  cd /e/aiops && git checkout -- backend/center.exe
  ```

---

### Task 4: 端到端验证

**Files:**
- 无代码变更。

- [ ] **Step 1: 启动前后端**

  终端 1：

  ```bash
  cd /e/aiops/backend && ./center.exe
  ```

  终端 2：

  ```bash
  cd /e/aiops/frontend && npm run dev
  ```

- [ ] **Step 2: 浏览器验证**

  1. 登录后进入 `#/anomaly`。
  2. 确认告警事件表格出现“操作”列和“根因分析”按钮。
  3. 点击按钮，页面跳转到 `#/chat`。
  4. 确认对话页自动出现 RCA Prompt 并发送。
  5. 确认 Agent 返回根因分析结果。
  6. 刷新对话页，确认不会再次自动发送 RCA Prompt。

- [ ] **Step 3: 回归常规聊天**

  创建新会话，发送普通问题，确认对话功能正常。

---

## Self-Review Checklist

- [x] Spec coverage：设计文档中所有要求（操作列、按钮、跳转、自动发送、刷新不重复、Prometheus/Elasticsearch 分析）均已对应到任务。
- [x] Placeholder scan：无 TBD/TODO/未填写代码。
- [x] Type consistency：事件字段名称在 Task 1 和 Task 2 中一致。
