<template>
  <div>
    <PageHeader title="告警规则" desc="创建和维护系统内的告警规则，基于 PromQL 表达式触发告警">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增规则</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">规则列表</h3>
          <p class="card-sub">已配置的告警规则及启用状态</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>规则名称</th>
            <th>告警级别</th>
            <th>PromQL</th>
            <th>执行频率</th>
            <th>持续时间（秒）</th>
            <th>启用状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in rules" :key="rule.id">
            <td><b>{{ rule.name }}</b></td>
            <td>
              <LevelTag :level="severityLevel(rule.severity)">
                {{ severityText(rule.severity) }}
              </LevelTag>
            </td>
            <td class="mono" :title="rule.prom_ql">{{ rule.prom_ql || '-' }}</td>
            <td>每 {{ rule.eval_interval }} 秒</td>
            <td>{{ rule.duration === 0 ? '0（立即）' : rule.duration }}</td>
            <td>
              <LevelTag :level="rule.is_enabled === 1 ? 'running' : 'info'">
                {{ rule.is_enabled === 1 ? '已启用' : '已停用' }}
              </LevelTag>
            </td>
            <td class="muted">{{ fmtTime(rule.created_at) }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(rule)">编辑</button>
                <button class="btn btn-sm" :class="rule.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(rule)">
                  {{ rule.is_enabled === 1 ? '停用' : '启用' }}
                </button>
                <button class="btn btn-sm btn-danger" @click="deleteRule(rule)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && rules.length === 0">
            <td colspan="8" class="empty-row">暂无告警规则，请点击"新增规则"创建</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑告警规则' : '新增告警规则' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">规则名称</label>
            <input v-model="form.name" class="form-input" placeholder="例如：CPU 使用率超过 90%" maxlength="128" />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">PromQL</label>
            <textarea
              v-model="form.prom_ql"
              class="form-input form-textarea mono-textarea"
              rows="3"
              placeholder='例如：100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 90'
            />
            <span v-if="errors.prom_ql" class="form-error">{{ errors.prom_ql }}</span>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">执行频率</label>
              <select v-model.number="form.eval_interval" class="form-input">
                <option :value="15">每 15 秒</option>
                <option :value="30">每 30 秒</option>
                <option :value="45">每 45 秒</option>
                <option :value="60">每 60 秒</option>
                <option :value="120">每 120 秒</option>
                <option :value="180">每 180 秒</option>
                <option :value="300">每 300 秒</option>
              </select>
              <span v-if="errors.eval_interval" class="form-error">{{ errors.eval_interval }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">持续时间（秒）</label>
              <input v-model.number="form.duration" type="number" min="0" class="form-input" placeholder="例如：300" />
              <span class="form-hint">0 表示只要有一次查询满足告警条件即触发</span>
              <span v-if="errors.duration" class="form-error">{{ errors.duration }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">告警级别</label>
            <select v-model.number="form.severity" class="form-input">
              <option :value="1">P1-紧急</option>
              <option :value="2">P2-警告</option>
              <option :value="3">P3-提醒</option>
            </select>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">启用状态</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveRule" :disabled="saving">
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
import { alertRuleApi } from '../api/alertRule.js'

const severityText = (s) => ({ 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }[s] || `P${s}`)

const severityLevel = (s) => {
  if (s === 1) return 'critical'
  if (s === 2) return 'warning'
  return 'info'
}

const fmtTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

// 数据状态
const rules = ref([])
const loading = ref(false)
const saving = ref(false)

// 弹窗状态
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const errors = reactive({})

const form = reactive({
  name: '',
  prom_ql: '',
  eval_interval: 60,
  duration: 300,
  severity: 2,
  is_enabled: 1
})

async function loadData() {
  loading.value = true
  try {
    const data = await alertRuleApi.list()
    rules.value = data || []
  } catch (err) {
    alert('加载告警规则失败：' + err.message)
    rules.value = []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  form.prom_ql = ''
  form.eval_interval = 60
  form.duration = 300
  form.severity = 2
  form.is_enabled = 1
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

function openEditModal(rule) {
  isEdit.value = true
  editingId.value = rule.id
  resetForm()
  form.name = rule.name
  form.prom_ql = rule.prom_ql
  form.eval_interval = rule.eval_interval || 60
  form.duration = rule.duration
  form.severity = rule.severity
  form.is_enabled = rule.is_enabled
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
    errors.name = '请输入规则名称'
    valid = false
  }
  if (!form.prom_ql.trim()) {
    errors.prom_ql = '请输入 PromQL 表达式'
    valid = false
  }
  if (form.eval_interval === null || form.eval_interval === undefined || form.eval_interval <= 0) {
    errors.eval_interval = '请选择执行频率'
    valid = false
  }
  if (form.duration === null || form.duration === undefined || form.duration < 0) {
    errors.duration = '持续时间不能为负数'
    valid = false
  }

  return valid
}

async function saveRule() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      prom_ql: form.prom_ql,
      eval_interval: form.eval_interval,
      duration: form.duration,
      severity: form.severity,
      is_enabled: form.is_enabled
    }

    if (isEdit.value) {
      await alertRuleApi.update(editingId.value, payload)
    } else {
      await alertRuleApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
    saving.value = false
  }
}

async function toggleEnabled(rule) {
  try {
    const result = await alertRuleApi.toggle(rule.id)
    const idx = rules.value.findIndex(r => r.id === rule.id)
    if (idx !== -1 && result) {
      rules.value[idx] = result
    }
  } catch (err) {
    alert('切换状态失败：' + err.message)
  }
}

async function deleteRule(rule) {
  if (!confirm(`确定要删除告警规则"${rule.name}"吗？`)) return
  try {
    await alertRuleApi.remove(rule.id)
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

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  box-sizing: border-box;
}
.form-input:focus { border-color: var(--c-primary); }
.form-input::placeholder { color: var(--c-text-3); }

.form-textarea {
  resize: vertical;
  min-height: 70px;
}

.mono-textarea {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12.5px;
}

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

.form-hint {
  display: block;
  color: var(--c-text-3);
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
</style>
