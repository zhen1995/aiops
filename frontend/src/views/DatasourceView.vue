<template>
  <div>
    <PageHeader title="数据源接入" desc="管理 Prometheus / ElasticSearch / Pyroscope 等数据源接入配置">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增数据源</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">数据源列表</h3>
          <p class="card-sub">已配置的数据源及运行状态</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>接入地址</th>
            <th>超时(ms)</th>
            <th>状态</th>
            <th>操作</th>
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
                {{ ds.is_enabled === 1 ? '已启用' : '已禁用' }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(ds)">编辑</button>
                <button class="btn btn-sm" :class="ds.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(ds)">
                  {{ ds.is_enabled === 1 ? '禁用' : '启用' }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteDatasource(ds)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && datasources.length === 0">
            <td colspan="6" class="empty-row">暂无数据，请点击"新增数据源"添加配置</td>
          </tr>
          <tr v-if="loading">
            <td colspan="6" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑数据源' : '新增数据源' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">名称</label>
              <input v-model="form.name" class="form-input" placeholder="请输入数据源名称" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">类型</label>
              <select v-model="form.type" class="form-input">
                <option value="">请选择类型</option>
                <option v-for="t in typeOptions" :key="t" :value="t">{{ t }}</option>
              </select>
              <span v-if="errors.type" class="form-error">{{ errors.type }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">接入地址</label>
            <input v-model="form.url" class="form-input" placeholder="请输入 HTTP 地址，例如：http://localhost:9090" />
            <span v-if="errors.url" class="form-error">{{ errors.url }}</span>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label">超时(ms)</label>
              <input v-model.number="form.timeout" type="number" class="form-input" placeholder="默认 5000" />
            </div>
            <div class="form-item form-item-half form-item-toggle">
              <label class="form-label">跳过 SSL 验证</label>
              <label class="switch">
                <input type="checkbox" v-model="form.is_skip_ssl" :true-value="1" :false-value="0" />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label">用户名</label>
              <input v-model="form.username" class="form-input" placeholder="选填" />
            </div>
            <div class="form-item form-item-half">
              <label class="form-label">密码</label>
              <div class="input-with-action">
                <input
                  v-model="form.password"
                  class="form-input"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="选填"
                />
                <span class="input-action" @click="showPassword = !showPassword">
                  {{ showPassword ? '隐藏' : '显示' }}
                </span>
              </div>
            </div>
          </div>

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
          <button class="btn btn-primary" @click="saveDatasource" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { datasourceApi } from '../api/datasource.js'

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
    alert('加载数据源列表失败：' + err.message)
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
    errors.name = '请输入名称'
    valid = false
  }
  if (!form.type) {
    errors.type = '请选择类型'
    valid = false
  }
  if (!form.url.trim()) {
    errors.url = '请输入接入地址'
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
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
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
    alert('切换状态失败：' + err.message)
  }
}

async function deleteDatasource(ds) {
  if (!confirm(`确定要删除数据源"${ds.name}"吗？`)) return
  try {
    await datasourceApi.remove(ds.id)
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
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