<template>
  <div>
    <PageHeader title="通知媒介" desc="维护钉钉机器人、Webhook 回调等通知通道配置">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增媒介</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">媒介列表</h3>
          <p class="card-sub">告警通知将通过以下媒介进行分发</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>媒介名称</th>
            <th>类型</th>
            <th>配置摘要</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in media" :key="m.id">
            <td><b>{{ m.name }}</b></td>
            <td>
              <span class="type-badge" :class="m.type === 'dingtalk' ? 'badge-dingtalk' : 'badge-webhook'">
                {{ typeText(m.type) }}
              </span>
            </td>
            <td class="muted mono" style="max-width: 280px; overflow: hidden; text-overflow: ellipsis">{{ configSummary(m) }}</td>
            <td>
              <LevelTag :level="m.is_enabled === 1 ? 'running' : 'info'">
                {{ m.is_enabled === 1 ? '运行中' : '已停用' }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(m)">编辑</button>
                <button class="btn btn-sm" :disabled="testingId === m.id" @click="openTestModal(m)">
                  {{ testingId === m.id ? '测试中...' : '测试' }}
                </button>
                <button class="btn btn-sm" :class="m.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(m)">
                  {{ m.is_enabled === 1 ? '停用' : '启用' }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteMedia(m)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && media.length === 0">
            <td colspan="5" class="empty-row">暂无数据，请点击"新增媒介"添加配置</td>
          </tr>
          <tr v-if="loading">
            <td colspan="5" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑媒介' : '新增媒介' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">媒介名称</label>
              <input v-model="form.name" class="form-input" placeholder="例如：SRE 值班钉钉群" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">媒介类型</label>
              <select v-model="form.type" class="form-input" :disabled="isEdit">
                <option value="">请选择类型</option>
                <option value="dingtalk">钉钉机器人</option>
                <option value="webhook">Webhook 回调</option>
              </select>
              <span v-if="errors.type" class="form-error">{{ errors.type }}</span>
            </div>
          </div>

          <!-- Webhook 回调配置 -->
          <template v-if="form.type === 'webhook'">
            <div class="form-item">
              <label class="form-label required">回调地址</label>
              <input v-model="form.webhookUrl" class="form-input" placeholder="请输入回调 URL，例如：https://example.com/api/notify" />
              <span v-if="errors.webhookUrl" class="form-error">{{ errors.webhookUrl }}</span>
            </div>
            <div class="form-row">
              <div class="form-item form-item-half">
                <label class="form-label">请求方法</label>
                <select v-model="form.method" class="form-input">
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                </select>
              </div>
              <div class="form-item form-item-half">
                <label class="form-label">超时(ms)</label>
                <input v-model.number="form.timeout" type="number" class="form-input" placeholder="默认 5000" />
              </div>
            </div>
          </template>

          <!-- 钉钉机器人配置 -->
          <template v-if="form.type === 'dingtalk'">
            <div class="form-item">
              <label class="form-label required">机器人 Webhook</label>
              <input v-model="form.dingWebhook" class="form-input" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
              <span v-if="errors.dingWebhook" class="form-error">{{ errors.dingWebhook }}</span>
            </div>
            <div class="form-row">
              <div class="form-item form-item-half">
                <label class="form-label">加签 Secret</label>
                <div class="input-with-action">
                  <input
                    v-model="form.secret"
                    class="form-input"
                    :type="showSecret ? 'text' : 'password'"
                    placeholder="选填，开启加签时填写"
                  />
                  <span class="input-action" @click="showSecret = !showSecret">
                    {{ showSecret ? '隐藏' : '显示' }}
                  </span>
                </div>
              </div>
              <div class="form-item form-item-half">
                <label class="form-label">超时(ms)</label>
                <input v-model.number="form.timeout" type="number" class="form-input" placeholder="默认 5000" />
              </div>
            </div>
          </template>

          <div class="form-item">
            <label class="form-label">备注</label>
            <textarea v-model="form.remark" class="form-input form-textarea" placeholder="选填" rows="2"></textarea>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">启用</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveMedia" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 测试消息弹窗 -->
    <div v-if="testModalVisible" class="modal-mask" @click.self="closeTestModal">
      <div class="modal">
        <div class="modal-header">
          <h3>发送测试消息{{ testTarget ? ` — ${testTarget.name}` : '' }}</h3>
          <span class="modal-close" @click="closeTestModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">消息内容</label>
            <textarea
              v-model="testForm.content"
              class="form-input form-textarea"
              rows="4"
              placeholder="请输入测试消息内容"
            ></textarea>
            <span v-if="testErrors.content" class="form-error">{{ testErrors.content }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeTestModal" :disabled="testingId !== ''">取消</button>
          <button class="btn btn-primary" @click="sendTest" :disabled="testingId !== ''">
            {{ testingId !== '' ? '发送中...' : '发送测试' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import PageHeader from '../../components/PageHeader.vue'
import LevelTag from '../../components/LevelTag.vue'
import { notifyMediaApi } from '../../api/notifyMedia.js'

const typeText = (type) => {
  const map = { dingtalk: '钉钉机器人', webhook: 'Webhook 回调' }
  return map[type] || type
}

// 解析媒介配置 JSON，返回列表展示摘要
const configSummary = (m) => {
  try {
    const cfg = JSON.parse(m.config || '{}')
    if (m.type === 'dingtalk') {
      return (cfg.webhook || '') + (cfg.secret ? '（已加签）' : '')
    }
    return (cfg.url || '') + (cfg.method ? ` [${cfg.method}]` : '')
  } catch (e) {
    return ''
  }
}

// 数据状态
const media = ref([])
const loading = ref(false)
const saving = ref(false)
const testingId = ref('')

// 弹窗状态
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const showSecret = ref(false)
const errors = reactive({})

const form = reactive({
  name: '',
  type: '',
  webhookUrl: '',
  dingWebhook: '',
  secret: '',
  method: 'POST',
  timeout: 5000,
  remark: '',
  is_enabled: 1
})

async function loadData() {
  loading.value = true
  try {
    const data = await notifyMediaApi.list()
    media.value = data || []
  } catch (err) {
    alert('加载媒介列表失败：' + err.message)
    media.value = []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  form.type = ''
  form.webhookUrl = ''
  form.dingWebhook = ''
  form.secret = ''
  form.method = 'POST'
  form.timeout = 5000
  form.remark = ''
  form.is_enabled = 1
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  showSecret.value = false
  modalVisible.value = true
}

function openEditModal(m) {
  isEdit.value = true
  editingId.value = m.id
  resetForm()

  let cfg = {}
  try {
    cfg = JSON.parse(m.config || '{}')
  } catch (e) {
    cfg = {}
  }

  form.name = m.name
  form.type = m.type
  form.webhookUrl = cfg.url || ''
  form.dingWebhook = cfg.webhook || ''
  form.secret = cfg.secret || ''
  form.method = cfg.method || 'POST'
  form.timeout = cfg.timeout || 5000
  form.remark = m.remark || ''
  form.is_enabled = m.is_enabled
  showSecret.value = false
  modalVisible.value = true
}

function closeModal() {
  if (saving.value) return
  modalVisible.value = false
}

function validateForm() {
  Object.keys(errors).forEach(k => delete errors[k])
  let valid = true

  if (!form.name.trim()) {
    errors.name = '请输入媒介名称'
    valid = false
  }
  if (!form.type) {
    errors.type = '请选择媒介类型'
    valid = false
  }
  if (form.type === 'webhook' && !form.webhookUrl.trim()) {
    errors.webhookUrl = '请输入回调地址'
    valid = false
  }
  if (form.type === 'dingtalk' && !form.dingWebhook.trim()) {
    errors.dingWebhook = '请输入机器人 Webhook 地址'
    valid = false
  }

  return valid
}

async function saveMedia() {
  if (!validateForm()) return

  saving.value = true
  try {
    // 按媒介类型组装 config JSON
    let config = {}
    if (form.type === 'webhook') {
      config = { url: form.webhookUrl, method: form.method, timeout: form.timeout }
    } else if (form.type === 'dingtalk') {
      config = { webhook: form.dingWebhook, secret: form.secret, timeout: form.timeout }
    }

    const payload = {
      name: form.name,
      type: form.type,
      config: JSON.stringify(config),
      remark: form.remark,
      is_enabled: form.is_enabled
    }

    if (isEdit.value) {
      await notifyMediaApi.update(editingId.value, payload)
    } else {
      await notifyMediaApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
    saving.value = false
  }
}

// 生成默认测试文案，与后端默认格式保持一致
const defaultTestContent = () => {
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  const ts = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
  return `【AIOPS】通知媒介「测试」消息，时间：${ts}`
}

// 测试弹窗状态
const testModalVisible = ref(false)
const testTarget = ref(null)
const testForm = reactive({ content: '' })
const testErrors = reactive({})

function openTestModal(m) {
  testTarget.value = m
  testForm.content = defaultTestContent()
  Object.keys(testErrors).forEach(k => delete testErrors[k])
  testModalVisible.value = true
}

function closeTestModal() {
  if (testingId.value !== '') return
  testModalVisible.value = false
  testTarget.value = null
}

async function sendTest() {
  Object.keys(testErrors).forEach(k => delete testErrors[k])
  if (!testForm.content.trim()) {
    testErrors.content = '请输入测试消息内容'
    return
  }

  const m = testTarget.value
  testingId.value = m.id
  try {
    await notifyMediaApi.test(m.id, { content: testForm.content.trim() })
    testModalVisible.value = false
    testTarget.value = null
    alert(`测试消息已发送至「${m.name}」`)
  } catch (err) {
    // 保留弹窗内容，方便修改后重发
    alert(`「${m.name}」测试失败：${err.message}`)
  } finally {
    testingId.value = ''
  }
}

async function toggleEnabled(m) {
  try {
    const result = await notifyMediaApi.toggle(m.id)
    const idx = media.value.findIndex(item => item.id === m.id)
    if (idx !== -1 && result) {
      media.value[idx] = result
    }
  } catch (err) {
    alert('切换状态失败：' + err.message)
  }
}

async function deleteMedia(m) {
  if (!confirm(`确定要删除媒介"${m.name}"吗？`)) return
  try {
    await notifyMediaApi.remove(m.id)
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.ops { display: flex; gap: 8px; }
.btn-danger { color: var(--c-danger); }
.btn-danger:hover { border-color: var(--c-danger); color: var(--c-danger); }

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}

.type-badge {
  display: inline-block;
  padding: 2px 10px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  font-weight: 600;
}
.badge-dingtalk {
  background: #e6f4ff;
  color: #1677ff;
}
.badge-webhook {
  background: var(--c-primary-soft);
  color: var(--c-primary);
}

/* 弹窗样式 */
.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--c-surface);
  border-radius: var(--radius-card);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  width: 580px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  border-bottom: 1px solid var(--c-border);
}
.modal-header h3 { margin: 0; font-size: 16px; font-weight: 600; }
.modal-close {
  cursor: pointer;
  font-size: 24px;
  color: var(--c-text-3);
  line-height: 1;
  transition: color 0.15s;
}
.modal-close:hover { color: var(--c-text); }

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 24px;
  border-top: 1px solid var(--c-border);
}

/* 表单样式 */
.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 14px;
}

.form-item {
  margin-bottom: 14px;
  flex: 1;
}

.form-item-half {
  flex: 1;
}

.form-item-toggle {
  display: flex;
  align-items: center;
  gap: 12px;
}

.form-label {
  display: block;
  font-size: 13px;
  color: var(--c-text-2);
  margin-bottom: 6px;
  font-weight: 500;
}
.form-label.required::before {
  content: '*';
  color: var(--c-danger);
  margin-right: 3px;
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  transition: border-color 0.15s;
  outline: none;
  font-family: inherit;
}
.form-input:focus { border-color: var(--c-primary); }
.form-input::placeholder { color: var(--c-text-3); }
.form-input:disabled {
  background: var(--c-bg);
  color: var(--c-text-3);
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
  min-height: 50px;
}

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

/* 开关 */
.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
}
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--c-border);
  transition: .3s;
  border-radius: 22px;
}
.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
}
input:checked + .slider {
  background-color: var(--c-primary);
}
input:checked + .slider:before {
  transform: translateX(18px);
}

/* 带操作按钮的输入框 */
.input-with-action {
  position: relative;
}
.input-with-action .form-input {
  padding-right: 60px;
}
.input-action {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  cursor: pointer;
  color: var(--c-primary);
  font-size: 12px;
  padding: 2px 6px;
  user-select: none;
}
.input-action:hover { color: var(--c-primary-dark); }
</style>
