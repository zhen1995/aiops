<template>
  <div>
    <PageHeader :title="$t('datasource.header.title')" :desc="$t('datasource.header.desc')">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ {{ $t('datasource.header.add') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('datasource.list.title') }}</h3>
          <p class="card-sub">{{ $t('datasource.list.sub') }}</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? $t('datasource.list.loading') : $t('datasource.list.refresh') }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('datasource.list.columns.name') }}</th>
            <th>{{ $t('datasource.list.columns.type') }}</th>
            <th>{{ $t('datasource.list.columns.url') }}</th>
            <th>{{ $t('datasource.list.columns.timeout') }}</th>
            <th>{{ $t('datasource.list.columns.status') }}</th>
            <th>{{ $t('datasource.list.columns.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ds in datasources" :key="ds.id">
            <td><b>{{ ds.name }}</b></td>
            <td>
              <span class="type-badge" :class="typeClass(ds.type)">{{ ds.type }}</span>
            </td>
            <td class="muted mono" style="max-width: 220px; overflow: hidden; text-overflow: ellipsis">{{ ds.url }}</td>
            <td>{{ ds.timeout }}</td>
            <td>
              <LevelTag :level="ds.is_enabled === 1 ? 'running' : 'info'">
                {{ ds.is_enabled === 1 ? $t('datasource.status.enabled') : $t('datasource.status.disabled') }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(ds)">{{ $t('datasource.actions.edit') }}</button>
                <button class="btn btn-sm" :class="ds.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(ds)">
                  {{ ds.is_enabled === 1 ? $t('datasource.actions.disable') : $t('datasource.actions.enable') }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteDatasource(ds)">{{ $t('datasource.actions.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && datasources.length === 0">
            <td colspan="6" class="empty-row">{{ $t('datasource.list.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="6" class="empty-row">{{ $t('datasource.list.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('datasource.modal.editTitle') : $t('datasource.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('datasource.modal.form.name') }}</label>
              <input v-model="form.name" class="form-input" :placeholder="$t('datasource.modal.form.namePlaceholder')" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('datasource.modal.form.type') }}</label>
              <select v-model="form.type" class="form-input">
                <option value="">{{ $t('datasource.modal.form.typePlaceholder') }}</option>
                <option v-for="t in typeOptions" :key="t" :value="t">{{ t }}</option>
              </select>
              <span v-if="errors.type" class="form-error">{{ errors.type }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">{{ $t('datasource.modal.form.url') }}</label>
            <input v-model="form.url" class="form-input" :placeholder="$t('datasource.modal.form.urlPlaceholder')" />
            <span v-if="errors.url" class="form-error">{{ errors.url }}</span>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label">{{ $t('datasource.modal.form.timeout') }}</label>
              <input v-model.number="form.timeout" type="number" class="form-input" :placeholder="$t('datasource.modal.form.timeoutPlaceholder')" />
            </div>
            <div class="form-item form-item-half form-item-toggle">
              <label class="form-label">{{ $t('datasource.modal.form.skipSsl') }}</label>
              <label class="switch">
                <input type="checkbox" v-model="form.is_skip_ssl" :true-value="1" :false-value="0" />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label">{{ $t('datasource.modal.form.username') }}</label>
              <input v-model="form.username" class="form-input" :placeholder="$t('datasource.modal.form.optional')" />
            </div>
            <div class="form-item form-item-half">
              <label class="form-label">{{ $t('datasource.modal.form.password') }}</label>
              <div class="input-with-action">
                <input
                  v-model="form.password"
                  class="form-input"
                  :type="showPassword ? 'text' : 'password'"
                  :placeholder="$t('datasource.modal.form.optional')"
                />
                <span class="input-action" @click="showPassword = !showPassword">
                  {{ showPassword ? $t('datasource.modal.form.hide') : $t('datasource.modal.form.show') }}
                </span>
              </div>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('datasource.modal.form.remark') }}</label>
            <textarea v-model="form.remark" class="form-input form-textarea" :placeholder="$t('datasource.modal.form.optional')" rows="2"></textarea>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">{{ $t('datasource.modal.form.enabled') }}</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('datasource.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="saveDatasource" :disabled="saving">
            {{ saving ? $t('datasource.modal.saving') : $t('datasource.modal.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { datasourceApi } from '../api/datasource.js'

const { t } = useI18n({ useScope: 'global' })

// 类型徽章颜色
const typeClass = (type) => {
  const map = {
    'Prometheus': 'badge-prometheus',
    'ElasticSearch': 'badge-elasticsearch',
    'Pyroscope': 'badge-pyroscope'
  }
  return map[type] || 'badge-default'
}

// 类型选项
const typeOptions = ref(['Prometheus', 'ElasticSearch', 'Pyroscope'])

// 数据状态
const datasources = ref([])
const loading = ref(false)
const saving = ref(false)

// 弹窗状态
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const showPassword = ref(false)
const errors = reactive({})

const form = reactive({
  name: '',
  type: '',
  url: '',
  timeout: 5000,
  username: '',
  password: '',
  is_skip_ssl: 0,
  remark: '',
  is_enabled: 1
})

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const data = await datasourceApi.list()
    datasources.value = data || []
  } catch (err) {
    alert(t('datasource.messages.loadFailed', { message: err.message }))
    datasources.value = []
  } finally {
    loading.value = false
  }
}

// 加载类型选项
async function loadTypes() {
  try {
    const data = await datasourceApi.types()
    if (data && data.length > 0) {
      typeOptions.value = data
    }
  } catch (e) {
    // 使用本地默认选项
  }
}

function resetForm() {
  form.name = ''
  form.type = ''
  form.url = ''
  form.timeout = 5000
  form.username = ''
  form.password = ''
  form.is_skip_ssl = 0
  form.remark = ''
  form.is_enabled = 1
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  showPassword.value = false
  modalVisible.value = true
}

function openEditModal(ds) {
  isEdit.value = true
  editingId.value = ds.id
  resetForm()
  form.name = ds.name
  form.type = ds.type
  form.url = ds.url
  form.timeout = ds.timeout || 5000
  form.username = ds.username || ''
  form.password = ds.password || ''
  form.is_skip_ssl = ds.is_skip_ssl || 0
  form.remark = ds.remark || ''
  form.is_enabled = ds.is_enabled
  showPassword.value = false
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
    errors.name = t('datasource.messages.nameRequired')
    valid = false
  }
  if (!form.type) {
    errors.type = t('datasource.messages.typeRequired')
    valid = false
  }
  if (!form.url.trim()) {
    errors.url = t('datasource.messages.urlRequired')
    valid = false
  }

  return valid
}

async function saveDatasource() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      type: form.type,
      url: form.url,
      timeout: form.timeout,
      username: form.username,
      password: form.password,
      is_skip_ssl: form.is_skip_ssl,
      remark: form.remark,
      is_enabled: form.is_enabled
    }

    if (isEdit.value) {
      await datasourceApi.update(editingId.value, payload)
    } else {
      await datasourceApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert(t(isEdit.value ? 'datasource.messages.updateFailed' : 'datasource.messages.createFailed', { message: err.message }))
    saving.value = false
  }
}

async function toggleEnabled(ds) {
  try {
    const result = await datasourceApi.toggle(ds.id)
    const idx = datasources.value.findIndex(c => c.id === ds.id)
    if (idx !== -1 && result) {
      datasources.value[idx] = result
    }
  } catch (err) {
    alert(t('datasource.messages.toggleFailed', { message: err.message }))
  }
}

async function deleteDatasource(ds) {
  if (!confirm(t('datasource.messages.deleteConfirm', { name: ds.name }))) return
  try {
    await datasourceApi.remove(ds.id)
    await loadData()
  } catch (err) {
    alert(t('datasource.messages.deleteFailed', { message: err.message }))
  }
}

// 初始化
onMounted(async () => {
  await Promise.all([loadData(), loadTypes()])
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
.badge-prometheus {
  background: #e7f2ef;
  color: #0e7c72;
}
.badge-elasticsearch {
  background: #fef3e2;
  color: #d97b29;
}
.badge-pyroscope {
  background: #e8eaf6;
  color: #3f51b5;
}
.badge-default {
  background: var(--c-primary-tint);
  color: var(--c-primary-dark);
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