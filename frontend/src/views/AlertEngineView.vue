<template>
  <div>
    <PageHeader title="告警引擎配置" desc="管理 Nightingale 告警引擎接入配置">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增配置</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">引擎配置列表</h3>
          <p class="card-sub">已配置的 Nightingale 告警引擎实例</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>地址</th>
            <th>Token</th>
            <th>业务组ID</th>
            <th>启用状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cfg in engines" :key="cfg.id">
            <td><b>{{ cfg.name }}</b></td>
            <td class="muted mono cell-ellipsis">{{ cfg.base_url }}</td>
            <td class="mono">{{ maskToken(cfg.token) }}</td>
            <td class="muted">{{ cfg.gids || '-' }}</td>
            <td>
              <LevelTag :level="cfg.is_enabled === 1 ? 'running' : 'info'">
                {{ cfg.is_enabled === 1 ? '已启用' : '已禁用' }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(cfg)">编辑</button>
                <button class="btn btn-sm" :class="cfg.is_enabled === 1 ? '' : 'btn-primary'" @click="toggleEnabled(cfg)">
                  {{ cfg.is_enabled === 1 ? '禁用' : '启用' }}
                </button>
                <button class="btn btn-sm" @click="testConnection(cfg)">测试</button>
                <button class="btn btn-sm btn-danger" @click="deleteEngine(cfg)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && engines.length === 0">
            <td colspan="6" class="empty-row">暂无引擎配置，请点击"新增配置"添加</td>
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
          <h3>{{ isEdit ? '编辑引擎配置' : '新增引擎配置' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">名称</label>
            <input v-model="form.name" class="form-input" placeholder="请输入配置名称" />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">夜莺地址</label>
            <input v-model="form.base_url" class="form-input" placeholder="例如：http://10.2.209.145:17000/" />
            <span v-if="errors.base_url" class="form-error">{{ errors.base_url }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">Token</label>
            <input v-model="form.token" class="form-input" placeholder="请输入 Nightingale 用户 Token" />
            <span v-if="errors.token" class="form-error">{{ errors.token }}</span>
          </div>

          <div class="form-item">
            <label class="form-label">业务组ID</label>
            <input v-model="form.gids" class="form-input" placeholder="选填，多个用逗号分隔，例如：21" />
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
          <button class="btn btn-primary" @click="saveEngine" :disabled="saving">
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
import { alertEngineApi } from '../api/alertEngine.js'

const maskToken = (t) => t ? t.slice(0, 6) + '****' + t.slice(-4) : ''

const engines = ref([])
const loading = ref(false)
const saving = ref(false)

const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const errors = reactive({})

const form = reactive({
  name: '',
  base_url: '',
  token: '',
  gids: '',
  remark: '',
  is_enabled: 1
})

async function loadData() {
  loading.value = true
  try {
    const data = await alertEngineApi.list()
    engines.value = data || []
  } catch (err) {
    alert('加载引擎配置失败：' + err.message)
    engines.value = []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  form.base_url = ''
  form.token = ''
  form.gids = ''
  form.remark = ''
  form.is_enabled = 1
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

function openEditModal(cfg) {
  isEdit.value = true
  editingId.value = cfg.id
  resetForm()
  form.name = cfg.name
  form.base_url = cfg.base_url
  form.token = cfg.token
  form.gids = cfg.gids || ''
  form.remark = cfg.remark || ''
  form.is_enabled = cfg.is_enabled
  modalVisible.value = true
}

function closeModal() {
  modalVisible.value = false
}

function validateForm() {
  Object.keys(errors).forEach(k => delete errors[k])
  let valid = true

  if (!form.name.trim()) {
    errors.name = '请输入名称'
    valid = false
  }
  if (!form.base_url.trim()) {
    errors.base_url = '请输入夜莺地址'
    valid = false
  }
  if (!form.token.trim()) {
    errors.token = '请输入 Token'
    valid = false
  }

  return valid
}

async function saveEngine() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      base_url: form.base_url.trim(),
      token: form.token.trim(),
      gids: form.gids.trim(),
      remark: form.remark.trim(),
      is_enabled: form.is_enabled
    }

    if (isEdit.value) {
      await alertEngineApi.update(editingId.value, payload)
    } else {
      await alertEngineApi.create(payload)
    }

    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert((isEdit.value ? '更新' : '新建') + '失败：' + err.message)
    saving.value = false
  }
}

async function toggleEnabled(cfg) {
  try {
    const result = await alertEngineApi.toggle(cfg.id)
    const idx = engines.value.findIndex(c => c.id === cfg.id)
    if (idx !== -1 && result) {
      engines.value[idx] = result
    }
  } catch (err) {
    alert('切换状态失败：' + err.message)
  }
}

async function testConnection(cfg) {
  try {
    await alertEngineApi.test(cfg.id)
    alert('连接成功')
  } catch (err) {
    alert('测试连接失败：' + err.message)
  }
}

async function deleteEngine(cfg) {
  if (!confirm(`确定要删除引擎配置"${cfg.name}"吗？`)) return
  try {
    await alertEngineApi.remove(cfg.id)
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
.ops { display: flex; gap: 8px; flex-wrap: wrap; }
.btn-danger { color: var(--c-danger); }
.btn-danger:hover { border-color: var(--c-danger); color: var(--c-danger); }

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}
.cell-ellipsis {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

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
  box-shadow: var(--shadow-modal);
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
.form-item {
  margin-bottom: 14px;
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
  background-color: var(--c-surface);
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
