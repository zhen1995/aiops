<template>
  <div>
    <PageHeader title="服务注册" desc="注册业务服务及其关联的日志 / 指标 / 性能剖析数据源">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增服务</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">服务列表</h3>
          <p class="card-sub">已注册的服务及数据源关联配置</p>
        </div>
        <div class="search-bar">
          <input v-model="keyword" class="form-input search-input" placeholder="搜索名称 / 编码 / 负责人" @keyup.enter="handleSearch" />
          <select v-model="statusFilter" class="form-input search-select" @change="handleSearch">
            <option value="">全部状态</option>
            <option value="1">已启用</option>
            <option value="0">已禁用</option>
          </select>
          <button class="btn btn-sm" @click="handleSearch" :disabled="loading">搜索</button>
          <button class="btn btn-sm" @click="handleReset">重置</button>
        </div>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>编码</th>
            <th>ES 索引模式</th>
            <th>Prom 标签</th>
            <th>负责人</th>
            <th>已验证</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="svc in services" :key="svc.id">
            <td><b>{{ svc.name }}</b></td>
            <td class="mono">{{ svc.code || '-' }}</td>
            <td class="muted mono" style="max-width: 200px; overflow: hidden; text-overflow: ellipsis" :title="patternsText(svc)">
              {{ patternsText(svc) }}
            </td>
            <td class="muted mono">{{ promLabelsText(svc) }}</td>
            <td>{{ svc.owner || '-' }}</td>
            <td>
              <LevelTag v-if="isVerified(svc) === true" level="online">是</LevelTag>
              <LevelTag v-else-if="isVerified(svc) === false" level="error">否</LevelTag>
              <LevelTag v-else level="info">未验证</LevelTag>
            </td>
            <td>
              <label class="switch">
                <input type="checkbox" :checked="svc.status === 1" @change="toggleStatus(svc)" />
                <span class="slider"></span>
              </label>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(svc)">编辑</button>
                <button class="btn btn-sm" @click="verifyService(svc)" :disabled="verifyingId === svc.id">
                  {{ verifyingId === svc.id ? '验证中...' : '验证' }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteService(svc)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && services.length === 0">
            <td colspan="8" class="empty-row">暂无数据，请点击"新增服务"添加配置</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>

      <div class="pagination" v-if="total > 0">
        <span class="muted">共 {{ total }} 条</span>
        <div class="page-ops">
          <button class="btn btn-sm" :disabled="page === 1 || loading" @click="changePage(page - 1)">上一页</button>
          <span class="page-info">第 {{ page }} 页</span>
          <button class="btn btn-sm" :disabled="page * pageSize >= total || loading" @click="changePage(page + 1)">下一页</button>
        </div>
      </div>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑服务' : '新增服务' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">名称</label>
              <input v-model="form.name" class="form-input" placeholder="请输入服务名称" />
              <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">服务编码</label>
              <input v-model="form.code" class="form-input" placeholder="唯一标识，例如：order-service" />
              <span v-if="errors.code" class="form-error">{{ errors.code }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">ES 数据源</label>
            <select v-model="form.es_datasource_id" class="form-input">
              <option value="">请选择数据源</option>
              <option v-for="ds in esOptions" :key="ds.id" :value="ds.id">{{ ds.name }}</option>
            </select>
          </div>

          <div class="form-item">
            <label class="form-label">ES 索引模式</label>
            <div class="tag-input">
              <span v-for="(p, i) in form.es_index_patterns" :key="p + i" class="tag-chip">
                {{ p }}
                <em class="tag-remove" @click="removePattern(i)">×</em>
              </span>
              <input
                v-model="patternInput"
                class="tag-text-input"
                placeholder="输入索引模式，回车添加，例如：app-log-*"
                @keyup.enter="addPattern"
                @blur="addPattern"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label">Pyroscope 应用名</label>
              <input v-model="form.pyroscope_app" class="form-input" placeholder="选填" />
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">Prometheus 数据源</label>
            <select v-model="form.prom_datasource_id" class="form-input">
              <option value="">请选择数据源</option>
              <option v-for="ds in promOptions" :key="ds.id" :value="ds.id">{{ ds.name }}</option>
            </select>
          </div>

          <div class="form-item">
            <label class="form-label">Prom 标签（键值对）</label>
            <div v-for="(label, i) in form.prom_labels" :key="i" class="kv-row">
              <input v-model="label.key" class="form-input kv-key" placeholder="标签键，例如：namespace" />
              <input v-model="label.value" class="form-input kv-value" placeholder="标签值" />
              <button class="btn btn-sm" @click="removeLabel(i)">移除</button>
            </div>
            <div class="label-actions">
              <button class="btn btn-sm" @click="addLabel">+ 添加标签</button>
              <span v-if="previewError" class="form-error preview-error-inline">{{ previewError }}</span>
            </div>
            <div class="preview-action">
              <button class="btn btn-sm btn-primary" @click="previewMatchedJobs" :disabled="previewLoading">
                {{ previewLoading ? '预览中...' : '预览 job' }}
              </button>
            </div>
            <div v-if="previewVisible" class="preview-result">
              <span v-if="previewLoading" class="muted">正在查询匹配的 job...</span>
              <span v-else-if="previewFailed" class="form-error">{{ previewMessage }}</span>
              <template v-else>
                <span v-if="!previewJobs.length" class="muted">无匹配的 job</span>
                <span v-for="job in previewJobs" :key="job" class="tag-chip">{{ job }}</span>
              </template>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">负责人</label>
            <input v-model="form.owner" class="form-input" placeholder="选填" />
          </div>

          <div class="form-item">
            <label class="form-label">描述</label>
            <textarea v-model="form.description" class="form-input form-textarea" placeholder="选填" rows="2"></textarea>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">启用</label>
            <label class="switch">
              <input type="checkbox" v-model="form.status" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveService" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 验证结果 弹窗 -->
    <div v-if="verifyVisible" class="modal-mask" @click.self="verifyVisible = false">
      <div class="modal verify-modal">
        <div class="modal-header">
          <h3>验证结果 - {{ verifyingName }}</h3>
          <span class="modal-close" @click="verifyVisible = false">×</span>
        </div>
        <div class="modal-body">
          <div class="verify-item">
            <span class="verify-label">ElasticSearch</span>
            <span class="verify-status" :class="verifyResult.es_ok ? 'ok' : 'fail'">
              {{ verifyResult.es_ok ? '连接正常' : '连接失败' }}
            </span>
            <p v-if="verifyResult.es_message" class="verify-message">{{ verifyResult.es_message }}</p>
          </div>
          <div class="verify-item">
            <span class="verify-label">Prometheus</span>
            <span class="verify-status" :class="verifyResult.prom_ok ? 'ok' : 'fail'">
              {{ verifyResult.prom_ok ? '连接正常' : '连接失败' }}
            </span>
            <p v-if="verifyResult.prom_message" class="verify-message">{{ verifyResult.prom_message }}</p>
          </div>
          <div class="verify-summary">
            综合结果：<b :class="verifyResult.verified ? 'ok' : 'fail'">{{ verifyResult.verified ? '验证通过' : '验证未通过' }}</b>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="verifyVisible = false">知道了</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { serviceApi } from '../api/service.js'
import { datasourceApi } from '../api/datasource.js'

// 数据状态
const services = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)
const saving = ref(false)

// 数据源下拉
const dsList = ref([])
const esOptions = computed(() => dsList.value.filter(d => d.type === 'ElasticSearch'))
const promOptions = computed(() => dsList.value.filter(d => d.type === 'Prometheus'))

// 弹窗状态
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const errors = reactive({})
const patternInput = ref('')

const form = reactive({
  name: '',
  code: '',
  es_datasource_id: '',
  es_index_patterns: [],
  prom_datasource_id: '',
  prom_labels: [],
  pyroscope_app: '',
  owner: '',
  description: '',
  status: 1
})

// 预览 job 状态（仅点击按钮时刷新，不随标签编辑实时刷新）
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewFailed = ref(false)
const previewMessage = ref('')
const previewJobs = ref([])
const previewError = ref('')

// 验证状态
const verifyVisible = ref(false)
const verifyingId = ref('')
const verifyingName = ref('')
const verifyResult = reactive({ es_ok: false, es_message: '', prom_ok: false, prom_message: '', verified: false })

// 已验证徽标：true=是 false=否 null=未验证
function isVerified(svc) {
  if (svc.verified === 1 || svc.verified === true) return true
  if (svc.verified === 0 || svc.verified === false) return false
  return null
}

function patternsText(svc) {
  const patterns = svc.es_index_patterns
  if (Array.isArray(patterns) && patterns.length > 0) return patterns.join(', ')
  return '-'
}

// Prom 标签展示：对象拼接为 k=v, k=v
function promLabelsText(svc) {
  const labels = svc.prom_labels
  if (labels && typeof labels === 'object' && Object.keys(labels).length > 0) {
    return Object.entries(labels)
      .map(([k, v]) => `${k}=${v}`)
      .join(', ')
  }
  return '-'
}

// 加载服务列表
async function loadData() {
  loading.value = true
  try {
    const data = await serviceApi.list({
      keyword: keyword.value,
      status: statusFilter.value,
      page: page.value,
      page_size: pageSize.value
    })
    services.value = data?.list || []
    total.value = data?.total || 0
  } catch (err) {
    alert('加载服务列表失败：' + err.message)
    services.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  loadData()
}

function handleReset() {
  keyword.value = ''
  statusFilter.value = ''
  handleSearch()
}

function changePage(p) {
  page.value = p
  loadData()
}

// 加载数据源下拉选项
async function loadDatasources() {
  try {
    const data = await datasourceApi.list()
    dsList.value = data || []
  } catch (e) {
    dsList.value = []
  }
}

function resetForm() {
  form.name = ''
  form.code = ''
  form.es_datasource_id = ''
  form.es_index_patterns = []
  form.prom_datasource_id = ''
  form.prom_labels = []
  form.pyroscope_app = ''
  form.owner = ''
  form.description = ''
  form.status = 1
  patternInput.value = ''
  previewVisible.value = false
  previewLoading.value = false
  previewFailed.value = false
  previewMessage.value = ''
  previewJobs.value = []
  previewError.value = ''
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

async function openEditModal(svc) {
  isEdit.value = true
  editingId.value = svc.id
  resetForm()
  try {
    const detail = await serviceApi.get(svc.id)
    fillForm(detail || svc)
  } catch (e) {
    fillForm(svc)
  }
  modalVisible.value = true
}

function fillForm(detail) {
  form.name = detail.name || ''
  form.code = detail.code || ''
  form.es_datasource_id = detail.es_datasource_id || ''
  form.es_index_patterns = Array.isArray(detail.es_index_patterns) ? [...detail.es_index_patterns] : []
  form.prom_datasource_id = detail.prom_datasource_id || ''
  // 标签键值对：对象转数组行
  const labels = detail.prom_labels || {}
  form.prom_labels = Object.entries(labels).map(([key, value]) => ({ key, value: String(value) }))
  form.pyroscope_app = detail.pyroscope_app || ''
  form.owner = detail.owner || ''
  form.description = detail.description || ''
  form.status = detail.status === 0 ? 0 : 1
}

function closeModal() {
  if (saving.value) return
  modalVisible.value = false
}

// ES 索引模式 tag 输入
function addPattern() {
  const p = patternInput.value.trim()
  if (p && !form.es_index_patterns.includes(p)) {
    form.es_index_patterns.push(p)
  }
  patternInput.value = ''
}

function removePattern(i) {
  form.es_index_patterns.splice(i, 1)
}

// Prom 标签键值对编辑
function addLabel() {
  form.prom_labels.push({ key: '', value: '' })
}

function removeLabel(i) {
  form.prom_labels.splice(i, 1)
}

// 按当前标签选择器预览匹配的 job（仅点击按钮时刷新）
async function previewMatchedJobs() {
  previewError.value = ''
  const labels = Object.fromEntries(
    form.prom_labels.filter(l => l.key.trim()).map(l => [l.key.trim(), l.value])
  )
  if (!form.prom_datasource_id || Object.keys(labels).length === 0) {
    previewError.value = '请先选择 Prometheus 数据源并添加标签'
    return
  }
  previewVisible.value = true
  previewLoading.value = true
  previewFailed.value = false
  try {
    const data = await serviceApi.previewJobs({
      prom_datasource_id: form.prom_datasource_id,
      prom_labels: labels
    })
    previewJobs.value = data?.jobs || []
    previewMessage.value = data?.message || ''
  } catch (err) {
    previewFailed.value = true
    previewMessage.value = err.message
    previewJobs.value = []
  } finally {
    previewLoading.value = false
  }
}

function validateForm() {
  // 提交前把输入框中未回车确认的索引模式也补进数组，避免用户误以为已填写
  addPattern()

  Object.keys(errors).forEach(k => delete errors[k])
  let valid = true

  if (!form.name.trim()) {
    errors.name = '请输入名称'
    valid = false
  }
  if (!form.code.trim()) {
    errors.code = '请输入服务编码'
    valid = false
  }

  return valid
}

async function saveService() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      code: form.code.trim(),
      es_datasource_id: form.es_datasource_id || '',
      es_index_patterns: form.es_index_patterns,
      prom_datasource_id: form.prom_datasource_id || '',
      prom_labels: Object.fromEntries(
        form.prom_labels.filter(l => l.key.trim()).map(l => [l.key.trim(), l.value])
      ),
      pyroscope_app: form.pyroscope_app.trim(),
      owner: form.owner.trim(),
      description: form.description.trim(),
      status: form.status
    }

    if (isEdit.value) {
      await serviceApi.update(editingId.value, payload)
    } else {
      await serviceApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
    saving.value = false
  }
}

async function toggleStatus(svc) {
  try {
    const result = await serviceApi.toggle(svc.id)
    const idx = services.value.findIndex(s => s.id === svc.id)
    if (idx !== -1 && result) {
      services.value[idx] = result
    } else if (idx !== -1) {
      services.value[idx].status = svc.status === 1 ? 0 : 1
    }
  } catch (err) {
    alert('切换状态失败：' + err.message)
  }
}

async function deleteService(svc) {
  if (!confirm(`确定要删除服务"${svc.name}"吗？`)) return
  try {
    await serviceApi.remove(svc.id)
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
  }
}

async function verifyService(svc) {
  verifyingId.value = svc.id
  try {
    const result = await serviceApi.verify(svc.id)
    Object.assign(verifyResult, {
      es_ok: !!result?.es_ok,
      es_message: result?.es_message || '',
      prom_ok: !!result?.prom_ok,
      prom_message: result?.prom_message || '',
      verified: !!result?.verified
    })
    verifyingName.value = svc.name
    verifyVisible.value = true
    await loadData()
  } catch (err) {
    alert('验证失败：' + err.message)
  } finally {
    verifyingId.value = ''
  }
}

onMounted(() => {
  loadData()
  loadDatasources()
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

.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input { width: 220px; }
.search-select { width: 120px; }

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-top: 1px solid var(--c-border);
}

.page-ops { display: flex; align-items: center; gap: 10px; }
.page-info { font-size: 13px; color: var(--c-text-2); }

/* 弹窗样式 */
.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--c-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--c-surface);
  border-radius: var(--radius-card);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  width: 640px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.verify-modal { width: 480px; }

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

/* tag 输入（ES 索引模式） */
.tag-input {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  padding: 6px 8px;
  background: var(--c-surface);
}
.tag-input:focus-within { border-color: var(--c-primary); }

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--c-primary-tint);
  color: var(--c-primary-dark);
  border-radius: var(--radius-tag);
  font-size: 12px;
  padding: 2px 8px;
}

.tag-remove {
  font-style: normal;
  cursor: pointer;
  color: var(--c-text-3);
  line-height: 1;
}
.tag-remove:hover { color: var(--c-danger); }

.tag-text-input {
  flex: 1;
  min-width: 160px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--c-text);
  font-family: inherit;
  padding: 3px 4px;
}
.tag-text-input::placeholder { color: var(--c-text-3); }

/* 键值对编辑行 */
.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.kv-key { flex: 1; }
.kv-value { flex: 1; }

/* 标签操作与预览结果 */
.label-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.preview-error-inline { margin-top: 0; }
.preview-action {
  display: flex;
  align-items: center;
  margin-top: 10px;
}
.preview-result {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 10px;
  border: 1px dashed var(--c-border);
  border-radius: 6px;
  font-size: 12px;
}
.preview-result .muted { color: var(--c-text-3); }

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

/* 验证结果 */
.verify-item {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--c-border);
}
.verify-item:last-of-type { border-bottom: none; }

.verify-label {
  display: inline-block;
  font-size: 13px;
  font-weight: 600;
  color: var(--c-text-2);
  margin-right: 10px;
}

.verify-status {
  font-size: 12px;
  font-weight: 600;
  padding: 1px 8px;
  border-radius: var(--radius-tag);
}
.verify-status.ok { color: var(--c-success); background: var(--c-success-bg); }
.verify-status.fail { color: var(--c-danger); background: var(--c-p0-bg); }

.verify-message {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--c-text-3);
  word-break: break-all;
}

.verify-summary {
  font-size: 13px;
  color: var(--c-text-2);
}
.verify-summary b.ok { color: var(--c-success); }
.verify-summary b.fail { color: var(--c-danger); }
</style>
