<template>
  <div>
    <PageHeader title="LLM 管理" desc="管理大语言模型接入配置，支持多厂商、多模型切换">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增模型</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">模型列表</h3>
          <p class="card-sub">已接入的 LLM 服务与运行状态</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>提供商类型</th>
            <th>模型</th>
            <th>接入端点</th>
            <th>API Key</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in llmConfigs" :key="m.id">
            <td>
              <b>{{ m.name }}</b>
              <span v-if="m.is_default === 1" class="default-tag">默认</span>
            </td>
            <td>{{ supplierLabel(m.supplier_category) }}</td>
            <td class="mono">{{ m.model }}</td>
            <td class="muted mono" style="max-width: 220px; overflow: hidden; text-overflow: ellipsis">{{ m.base_url }}</td>
            <td class="mono">{{ maskApiKey(m.api_key) }}</td>
            <td>
              <LevelTag :level="m.is_enabled === 1 ? 'running' : 'info'">
                {{ m.is_enabled === 1 ? '运行中' : '已停用' }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(m)">编辑</button>
                <button class="btn btn-sm" :class="m.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(m)">
                  {{ m.is_enabled === 1 ? '停用' : '启用' }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteConfig(m)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && llmConfigs.length === 0">
            <td colspan="7" class="empty-row">暂无数据，请点击"新增模型"添加配置</td>
          </tr>
          <tr v-if="loading">
            <td colspan="7" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑 LLM 配置' : '新建 LLM 配置' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item">
              <label class="form-label required">名称</label>
              <input v-model="form.name" class="form-input" placeholder="请输入 LLM 配置的名称" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-toggles">
              <div class="toggle-item">
                <span class="toggle-label">启用</span>
                <label class="switch">
                  <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
                  <span class="slider"></span>
                </label>
              </div>
              <div class="toggle-item">
                <span class="toggle-label">默认</span>
                <label class="switch">
                  <input type="checkbox" v-model="form.is_default" :true-value="1" :false-value="0" />
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">描述</label>
            <textarea v-model="form.description" class="form-input form-textarea" placeholder="请输入 LLM 配置的描述信息" rows="3"></textarea>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">提供商类型</label>
              <select v-model="form.supplier_category" class="form-input">
                <option value="">请选择提供商类型</option>
                <option v-for="s in supplierOptions" :key="s" :value="s">{{ supplierLabel(s) }}</option>
              </select>
              <span v-if="errors.supplier_category" class="form-error">{{ errors.supplier_category }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">模型</label>
              <input v-model="form.model" class="form-input" placeholder="请输入模型名称，例如：gpt-4o" />
              <span v-if="errors.model" class="form-error">{{ errors.model }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">API URL</label>
            <input v-model="form.base_url" class="form-input" placeholder="请输入接口地址，例如：https://api.openai.com/v1" />
            <span v-if="errors.base_url" class="form-error">{{ errors.base_url }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">API Key</label>
            <div class="input-with-action">
              <input
                v-model="form.api_key"
                class="form-input"
                :type="showApiKey ? 'text' : 'password'"
                placeholder="请输入 API Key"
              />
              <span class="input-action" @click="showApiKey = !showApiKey">
                {{ showApiKey ? '隐藏' : '显示' }}
              </span>
            </div>
            <span v-if="errors.api_key" class="form-error">{{ errors.api_key }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveConfig" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
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
import { llmConfigApi } from '../../api/llmConfig.js'

// 供应商类型映射（key -> 中文标签）
const supplierMap = {
  openai: 'OpenAI 兼容',
  claude: 'Claude',
  gemini: 'Gemini',
  qwen: '阿里云千问',
  deepseek: 'DeepSeek',
  kimi: 'Kimi',
  other: '其他'
}

const supplierOptions = ref(Object.keys(supplierMap))

const supplierLabel = (key) => supplierMap[key] || key

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
    alert('加载 LLM 配置列表失败：' + err.message)
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
      // 合并后端返回的和前端映射的，优先使用后端的
      const merged = [...new Set([...data, ...Object.keys(supplierMap)])]
      // 确保映射表中有对应标签
      merged.forEach(key => {
        if (!supplierMap[key]) {
          supplierMap[key] = key
        }
      })
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
    errors.name = '请输入名称'
    valid = false
  }
  if (!form.supplier_category) {
    errors.supplier_category = '请选择提供商类型'
    valid = false
  }
  if (!form.model.trim()) {
    errors.model = '请输入模型名称'
    valid = false
  }
  if (!form.base_url.trim()) {
    errors.base_url = '请输入 API URL'
    valid = false
  }
  if (!form.api_key.trim()) {
    errors.api_key = '请输入 API Key'
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
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
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
    alert('切换状态失败：' + err.message)
  }
}

async function deleteConfig(m) {
  if (!confirm(`确定要删除配置"${m.name}"吗？`)) return
  try {
    await llmConfigApi.remove(m.id)
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
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