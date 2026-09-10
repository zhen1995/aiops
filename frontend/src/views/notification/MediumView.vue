<template>
  <div>
    <PageHeader :title="$t('notify.medium.pageTitle')" :desc="$t('notify.medium.pageDesc')">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">{{ $t('notify.medium.create') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('notify.medium.list.title') }}</h3>
          <p class="card-sub">{{ $t('notify.medium.list.subtitle') }}</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? $t('notify.medium.list.loading') : $t('notify.medium.list.refresh') }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('notify.medium.list.name') }}</th>
            <th>{{ $t('notify.medium.list.type') }}</th>
            <th>{{ $t('notify.medium.list.summary') }}</th>
            <th>{{ $t('notify.medium.list.status') }}</th>
            <th>{{ $t('notify.medium.list.actions') }}</th>
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
                {{ m.is_enabled === 1 ? $t('notify.medium.list.running') : $t('notify.medium.list.disabled') }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(m)">{{ $t('notify.medium.list.edit') }}</button>
                <button class="btn btn-sm" :disabled="testingId === m.id" @click="openTestModal(m)">
                  {{ testingId === m.id ? $t('notify.medium.list.testing') : $t('notify.medium.list.test') }}
                </button>
                <button class="btn btn-sm" :class="m.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(m)">
                  {{ m.is_enabled === 1 ? $t('notify.medium.list.disable') : $t('notify.medium.list.enable') }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteMedia(m)">{{ $t('notify.medium.list.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && media.length === 0">
            <td colspan="5" class="empty-row">{{ $t('notify.medium.list.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="5" class="empty-row">{{ $t('notify.medium.list.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('notify.medium.modal.editTitle') : $t('notify.medium.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('notify.medium.modal.name') }}</label>
              <input v-model="form.name" class="form-input" :placeholder="$t('notify.medium.modal.namePlaceholder')" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('notify.medium.modal.type') }}</label>
              <select v-model="form.type" class="form-input" :disabled="isEdit">
                <option value="">{{ $t('notify.medium.modal.selectType') }}</option>
                <option value="dingtalk">{{ $t('notify.medium.type.dingtalk') }}</option>
                <option value="webhook">{{ $t('notify.medium.type.webhook') }}</option>
              </select>
              <span v-if="errors.type" class="form-error">{{ errors.type }}</span>
            </div>
          </div>

          <!-- Webhook 回调配置 -->
          <template v-if="form.type === 'webhook'">
            <div class="form-item">
              <label class="form-label required">{{ $t('notify.medium.modal.webhookUrl') }}</label>
              <input v-model="form.webhookUrl" class="form-input" :placeholder="$t('notify.medium.modal.webhookUrlPlaceholder')" />
              <span v-if="errors.webhookUrl" class="form-error">{{ errors.webhookUrl }}</span>
            </div>
            <div class="form-row">
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('notify.medium.modal.method') }}</label>
                <select v-model="form.method" class="form-input">
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                </select>
              </div>
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('notify.medium.modal.timeout') }}</label>
                <input v-model.number="form.timeout" type="number" class="form-input" :placeholder="$t('notify.medium.modal.timeoutPlaceholder')" />
              </div>
            </div>
          </template>

          <!-- 钉钉配置 -->
          <template v-if="form.type === 'dingtalk'">
            <div class="form-item">
              <label class="form-label required">{{ $t('notify.medium.modal.dingWebhook') }}</label>
              <input v-model="form.dingWebhook" class="form-input" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
              <span v-if="errors.dingWebhook" class="form-error">{{ errors.dingWebhook }}</span>
            </div>
            <div class="form-row">
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('notify.medium.modal.secret') }}</label>
                <div class="input-with-action">
                  <input
                    v-model="form.secret"
                    class="form-input"
                    :type="showSecret ? 'text' : 'password'"
                    :placeholder="$t('notify.medium.modal.secretPlaceholder')"
                  />
                  <span class="input-action" @click="showSecret = !showSecret">
                    {{ showSecret ? $t('notify.medium.modal.hide') : $t('notify.medium.modal.show') }}
                  </span>
                </div>
              </div>
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('notify.medium.modal.timeout') }}</label>
                <input v-model.number="form.timeout" type="number" class="form-input" :placeholder="$t('notify.medium.modal.timeoutPlaceholder')" />
              </div>
            </div>
          </template>

          <div class="form-item">
            <label class="form-label">{{ $t('notify.medium.modal.remark') }}</label>
            <textarea v-model="form.remark" class="form-input form-textarea" :placeholder="$t('notify.medium.modal.remarkPlaceholder')" rows="2"></textarea>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">{{ $t('notify.medium.modal.enabled') }}</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('notify.medium.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="saveMedia" :disabled="saving">
            {{ saving ? $t('notify.medium.modal.saving') : $t('notify.medium.modal.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 测试消息弹窗 -->
    <div v-if="testModalVisible" class="modal-mask" @click.self="closeTestModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ testTarget ? $t('notify.medium.test.titleWithName', { name: testTarget.name }) : $t('notify.medium.test.title') }}</h3>
          <span class="modal-close" @click="closeTestModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('notify.medium.test.content') }}</label>
            <textarea
              v-model="testForm.content"
              class="form-input form-textarea"
              rows="4"
              :placeholder="$t('notify.medium.test.contentPlaceholder')"
            ></textarea>
            <span v-if="testErrors.content" class="form-error">{{ testErrors.content }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeTestModal" :disabled="testingId !== ''">{{ $t('notify.medium.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="sendTest" :disabled="testingId !== ''">
            {{ testingId !== '' ? $t('notify.medium.test.sending') : $t('notify.medium.test.send') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import LevelTag from '../../components/LevelTag.vue'
import { notifyMediaApi } from '../../api/notifyMedia.js'

const { t } = useI18n({ useScope: 'global' })

const typeText = (type) => {
  const map = { dingtalk: t('notify.medium.type.dingtalk'), webhook: t('notify.medium.type.webhook') }
  return map[type] || type
}

// 解析媒介配置 JSON，返回列表展示摘要
const configSummary = (m) => {
  try {
    const cfg = JSON.parse(m.config || '{}')
    if (m.type === 'dingtalk') {
      return (cfg.webhook || '') + (cfg.secret ? t('notify.medium.signed') : '')
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
    alert(t('notify.medium.error.loadFailed') + err.message)
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
    errors.name = t('notify.medium.error.nameRequired')
    valid = false
  }
  if (!form.type) {
    errors.type = t('notify.medium.error.typeRequired')
    valid = false
  }
  if (form.type === 'webhook' && !form.webhookUrl.trim()) {
    errors.webhookUrl = t('notify.medium.error.webhookUrlRequired')
    valid = false
  }
  if (form.type === 'dingtalk' && !form.dingWebhook.trim()) {
    errors.dingWebhook = t('notify.medium.error.dingWebhookRequired')
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
    alert((isEdit.value ? t('notify.medium.error.updateFailed') : t('notify.medium.error.createFailed')) + err.message)
    saving.value = false
  }
}

// 生成默认测试文案，与后端默认格式保持一致
const defaultTestContent = () => {
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  const ts = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
  return t('notify.medium.defaultTestMessage', { time: ts })
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
    testErrors.content = t('notify.medium.error.testContentRequired')
    return
  }

  const m = testTarget.value
  testingId.value = m.id
  try {
    await notifyMediaApi.test(m.id, { content: testForm.content.trim() })
    testModalVisible.value = false
    testTarget.value = null
    alert(t('notify.medium.error.testSent', { name: m.name }))
  } catch (err) {
    // 保留弹窗内容，方便修改后重发
    alert(t('notify.medium.error.testFailed', { name: m.name, message: err.message }))
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
    alert(t('notify.medium.error.toggleFailed') + err.message)
  }
}

async function deleteMedia(m) {
  if (!confirm(t('notify.medium.error.confirmDelete', { name: m.name }))) return
  try {
    await notifyMediaApi.remove(m.id)
    await loadData()
  } catch (err) {
    alert(t('notify.medium.error.deleteFailed') + err.message)
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
