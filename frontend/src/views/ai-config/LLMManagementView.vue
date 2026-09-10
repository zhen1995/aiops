<template>
  <div>
    <PageHeader :title="$t('llm.header.title')" :desc="$t('llm.header.desc')">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ {{ $t('llm.header.add') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('llm.list.title') }}</h3>
          <p class="card-sub">{{ $t('llm.list.sub') }}</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? $t('llm.list.loading') : $t('llm.list.refresh') }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('llm.table.name') }}</th>
            <th>{{ $t('llm.table.modelType') }}</th>
            <th>{{ $t('llm.table.supplierCategory') }}</th>
            <th>{{ $t('llm.table.model') }}</th>
            <th>{{ $t('llm.table.endpoint') }}</th>
            <th>{{ $t('llm.table.apiKey') }}</th>
            <th>{{ $t('llm.table.status') }}</th>
            <th>{{ $t('llm.table.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in llmConfigs" :key="m.id">
            <td>
              <b>{{ m.name }}</b>
              <span v-if="m.is_default === 1" class="default-tag">{{ $t('llm.table.default') }}</span>
            </td>
            <td>{{ modelTypeLabel(m.model_type) }}</td>
            <td>{{ supplierLabel(m.supplier_category) }}</td>
            <td class="mono">{{ m.model }}</td>
            <td class="muted mono" style="max-width: 220px; overflow: hidden; text-overflow: ellipsis">{{ m.base_url }}</td>
            <td class="mono">{{ maskApiKey(m.api_key) }}</td>
            <td>
              <LevelTag :level="m.is_enabled === 1 ? 'running' : 'info'">
                {{ m.is_enabled === 1 ? $t('llm.table.statusRunning') : $t('llm.table.statusStopped') }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(m)">{{ $t('llm.table.edit') }}</button>
                <button class="btn btn-sm" :class="m.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(m)">
                  {{ m.is_enabled === 1 ? $t('llm.table.disable') : $t('llm.table.enable') }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteConfig(m)">{{ $t('llm.table.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && llmConfigs.length === 0">
            <td colspan="8" class="empty-row">{{ $t('llm.table.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">{{ $t('llm.list.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('llm.modal.editTitle') : $t('llm.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item">
              <label class="form-label required">{{ $t('llm.modal.name') }}</label>
              <input v-model="form.name" class="form-input" :placeholder="$t('llm.modal.namePlaceholder')" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-toggles">
              <div class="toggle-item">
                <span class="toggle-label">{{ $t('llm.modal.enabled') }}</span>
                <label class="switch">
                  <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
                  <span class="slider"></span>
                </label>
              </div>
              <div class="toggle-item">
                <span class="toggle-label">{{ $t('llm.modal.isDefault') }}</span>
                <label class="switch">
                  <input type="checkbox" v-model="form.is_default" :true-value="1" :false-value="0" />
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('llm.modal.description') }}</label>
            <textarea v-model="form.description" class="form-input form-textarea" :placeholder="$t('llm.modal.descriptionPlaceholder')" rows="3"></textarea>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('llm.modal.modelType') }}</label>
              <select v-model="form.model_type" class="form-input">
                <option value="chat">{{ $t('llm.modal.typeChat') }}</option>
                <option value="embedding">{{ $t('llm.modal.typeEmbedding') }}</option>
              </select>
              <span v-if="errors.model_type" class="form-error">{{ errors.model_type }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('llm.modal.supplierCategory') }}</label>
              <select v-model="form.supplier_category" class="form-input">
                <option value="">{{ $t('llm.modal.supplierPlaceholder') }}</option>
                <option v-for="s in supplierOptions" :key="s" :value="s">{{ supplierLabel(s) }}</option>
              </select>
              <span v-if="errors.supplier_category" class="form-error">{{ errors.supplier_category }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">{{ $t('llm.modal.model') }}</label>
            <input v-model="form.model" class="form-input" :placeholder="$t('llm.modal.modelPlaceholder')" />
            <span v-if="errors.model" class="form-error">{{ errors.model }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">API URL</label>
            <input v-model="form.base_url" class="form-input" :placeholder="$t('llm.modal.apiUrlPlaceholder')" />
            <span v-if="errors.base_url" class="form-error">{{ errors.base_url }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">API Key</label>
            <div class="input-with-action">
              <input
                v-model="form.api_key"
                class="form-input"
                :type="showApiKey ? 'text' : 'password'"
                :placeholder="$t('llm.modal.apiKeyPlaceholder')"
              />
              <span class="input-action" @click="showApiKey = !showApiKey">
                {{ showApiKey ? $t('llm.modal.hide') : $t('llm.modal.show') }}
              </span>
            </div>
            <span v-if="errors.api_key" class="form-error">{{ errors.api_key }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('llm.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="saveConfig" :disabled="saving">
            {{ saving ? $t('llm.modal.saving') : $t('llm.modal.save') }}
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
import { llmConfigApi } from '../../api/llmConfig.js'

const { t } = useI18n({ useScope: 'global' })

// 供应商类型列表（key -> i18n 键，未知 key 原样展示）
const supplierKeys = ['openai', 'claude', 'gemini', 'qwen', 'deepseek', 'kimi', 'other']

const supplierOptions = ref([...supplierKeys])

const supplierLabel = (key) => {
  const label = t(`llm.supplier.${key}`)
  return label === `llm.supplier.${key}` ? key : label
}

// 模型类型映射（key -> i18n 键，未知 key 原样展示）
const modelTypeLabel = (key) => {
  const label = t(`llm.modelTypeOption.${key}`)
  return label === `llm.modelTypeOption.${key}` ? key : label
}

// 数据状态
const llmConfigs = ref([])
const loading = ref(false)
const saving = ref(false)

// 弹窗状态
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const showApiKey = ref(false)
const errors = reactive({})

const form = reactive({
  name: '',
  description: '',
  model_type: 'chat',
  supplier_category: '',
  model: '',
  base_url: '',
  api_key: '',
  is_enabled: 1,
  is_default: 0
})

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const data = await llmConfigApi.list()
    llmConfigs.value = data || []
  } catch (err) {
    alert(t('llm.message.loadFailed', { msg: err.message }))
    llmConfigs.value = []
  } finally {
    loading.value = false
  }
}

// 加载供应商选项
async function loadSuppliers() {
  try {
    const data = await llmConfigApi.suppliers()
    if (data && data.length > 0) {
      // 合并后端返回的和前端默认的，优先使用后端的
      const merged = [...new Set([...data, ...supplierKeys])]
      supplierOptions.value = merged
    }
  } catch (e) {
    // 使用本地默认选项
  }
}

function maskApiKey(key) {
  if (!key) return '-'
  if (key.length <= 8) return key.slice(0, 2) + '****'
  return key.slice(0, 7) + '****' + key.slice(-4)
}

function resetForm() {
  form.name = ''
  form.description = ''
  form.model_type = 'chat'
  form.supplier_category = ''
  form.model = ''
  form.base_url = ''
  form.api_key = ''
  form.is_enabled = 1
  form.is_default = 0
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  showApiKey.value = false
  modalVisible.value = true
}

function openEditModal(m) {
  isEdit.value = true
  editingId.value = m.id
  resetForm()
  form.name = m.name
  form.description = m.description || ''
  form.model_type = m.model_type || 'chat'
  form.supplier_category = m.supplier_category
  form.model = m.model
  form.base_url = m.base_url
  form.api_key = m.api_key || ''
  form.is_enabled = m.is_enabled
  form.is_default = m.is_default || 0
  showApiKey.value = false
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
    errors.name = t('llm.validation.nameRequired')
    valid = false
  }
  if (!form.model_type) {
    errors.model_type = t('llm.validation.modelTypeRequired')
    valid = false
  }
  if (!form.supplier_category) {
    errors.supplier_category = t('llm.validation.supplierRequired')
    valid = false
  }
  if (!form.model.trim()) {
    errors.model = t('llm.validation.modelRequired')
    valid = false
  }
  if (!form.base_url.trim()) {
    errors.base_url = t('llm.validation.apiUrlRequired')
    valid = false
  }
  if (!form.api_key.trim()) {
    errors.api_key = t('llm.validation.apiKeyRequired')
    valid = false
  }

  return valid
}

async function saveConfig() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      model_type: form.model_type,
      supplier_category: form.supplier_category,
      model: form.model,
      base_url: form.base_url,
      api_key: form.api_key,
      is_enabled: form.is_enabled,
      is_default: form.is_default
    }

    if (isEdit.value) {
      await llmConfigApi.update(editingId.value, payload)
    } else {
      await llmConfigApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert(isEdit.value ? t('llm.message.updateFailed', { msg: err.message }) : t('llm.message.createFailed', { msg: err.message }))
  } finally {
    saving.value = false
  }
}

async function toggleEnabled(m) {
  try {
    const result = await llmConfigApi.toggle(m.id)
    // 更新本地数据
    const idx = llmConfigs.value.findIndex(c => c.id === m.id)
    if (idx !== -1 && result) {
      llmConfigs.value[idx] = result
    }
  } catch (err) {
    alert(t('llm.message.toggleFailed', { msg: err.message }))
  }
}

async function deleteConfig(m) {
  if (!confirm(t('llm.message.deleteConfirm', { name: m.name }))) return
  try {
    await llmConfigApi.remove(m.id)
    await loadData()
  } catch (err) {
    alert(t('llm.message.deleteFailed', { msg: err.message }))
  }
}

// 初始化
onMounted(async () => {
  await Promise.all([loadData(), loadSuppliers()])
})
</script>

<style scoped>
.ops { display: flex; gap: 8px; }

.default-tag {
  display: inline-block;
  margin-left: 6px;
  padding: 1px 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--c-primary);
  background: var(--c-primary-tint);
  border-radius: var(--radius-tag);
  vertical-align: middle;
}
.btn-danger { color: var(--c-danger); }
.btn-danger:hover { border-color: var(--c-danger); color: var(--c-danger); }

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
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
  width: 560px;
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

.form-textarea {
  resize: vertical;
  min-height: 60px;
}

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

/* 开关 */
.form-toggles {
  display: flex;
  gap: 20px;
  flex-shrink: 0;
  padding-top: 22px;
}

.toggle-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-label {
  font-size: 13px;
  color: var(--c-text-2);
}

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