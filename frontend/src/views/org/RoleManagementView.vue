<template>
  <div>
    <PageHeader title="角色管理" desc="维护平台角色、权限范围与角色成员">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增角色</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">角色列表</h3>
          <p class="card-sub">基于 RBAC 的角色权限配置</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>角色名称</th>
            <th>成员数</th>
            <th>权限范围</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in roleList" :key="r.id">
            <td><b>{{ r.name }}</b></td>
            <td>{{ r.userCount }} 人</td>
            <td>
              <span v-for="perm in r.permissions" :key="perm" class="perm-tag">{{ perm }}</span>
              <span v-if="!r.permissions || r.permissions.length === 0" class="muted">暂未分配</span>
            </td>
            <td class="muted">{{ formatTime(r.createdAt) }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(r)">编辑</button>
                <button class="btn btn-sm" @click="openAuthModal(r)">权限</button>
                <button class="btn btn-sm btn-danger" @click="deleteRole(r)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && roleList.length === 0">
            <td colspan="5" class="empty-row">暂无角色数据，请点击"新增角色"添加</td>
          </tr>
          <tr v-if="loading">
            <td colspan="5" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑 角色 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑角色' : '新增角色' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">角色名称</label>
            <input
              v-model="form.name"
              class="form-input"
              placeholder="请输入角色名称"
              maxlength="30"
            />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveRole" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 权限分配 弹窗 -->
    <div v-if="authModalVisible" class="modal-mask" @click.self="closeAuthModal">
      <div class="modal modal-large">
        <div class="modal-header">
          <h3>权限分配 - {{ currentRole?.name }}</h3>
          <span class="modal-close" @click="closeAuthModal">×</span>
        </div>
        <div class="modal-body">
          <div v-if="loadingAuths" class="muted">加载权限列表中...</div>
          <div v-else-if="authList.length === 0" class="muted">暂无权限数据</div>
          <div v-else class="auth-grid">
            <label v-for="auth in authList" :key="auth.id" class="auth-item">
              <input
                type="checkbox"
                :value="auth.id"
                v-model="selectedAuthIds"
              />
              <span class="auth-name">{{ auth.name }}</span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeAuthModal" :disabled="authSaving">取消</button>
          <button class="btn btn-primary" @click="saveAuths" :disabled="authSaving">
            {{ authSaving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import PageHeader from '../../components/PageHeader.vue'
import { roleApi } from '../../api/role.js'

// 数据
const roleList = ref([])
const loading = ref(false)
const saving = ref(false)

// 新增/编辑弹窗
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const errors = reactive({})
const form = reactive({ name: '' })

// 权限分配弹窗
const authModalVisible = ref(false)
const authSaving = ref(false)
const loadingAuths = ref(false)
const currentRole = ref(null)
const authList = ref([])
const selectedAuthIds = ref([])

// 工具函数
const formatTime = (t) => {
  if (!t) return '-'
  const d = new Date(t)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 加载角色列表
async function loadData() {
  loading.value = true
  try {
    roleList.value = await roleApi.list() || []
  } catch (err) {
    alert('加载角色列表失败：' + err.message)
    roleList.value = []
  } finally {
    loading.value = false
  }
}

// 表单重置
function resetForm() {
  form.name = ''
  Object.keys(errors).forEach(k => delete errors[k])
}

// 新增
function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

// 编辑
function openEditModal(r) {
  isEdit.value = true
  editingId.value = r.id
  resetForm()
  form.name = r.name
  modalVisible.value = true
}

// 关闭弹窗
function closeModal() {
  if (saving.value) return
  modalVisible.value = false
}

// 表单校验
function validateForm() {
  Object.keys(errors).forEach(k => delete errors[k])
  let valid = true

  if (!form.name.trim()) {
    errors.name = '请输入角色名称'
    valid = false
  } else if (form.name.length > 30) {
    errors.name = '角色名称最长 30 个字符'
    valid = false
  }

  return valid
}

// 保存角色
async function saveRole() {
  if (!validateForm()) return

  saving.value = true
  try {
    if (isEdit.value) {
      await roleApi.update(editingId.value, { name: form.name.trim() })
    } else {
      await roleApi.create({ name: form.name.trim() })
    }
    saving.value = false
    modalVisible.value = false
    await loadData()
  } catch (err) {
    alert('保存失败：' + err.message)
  } finally {
    saving.value = false
  }
}

// 删除角色
async function deleteRole(r) {
  if (!confirm(`确定要删除角色「${r.name}」吗？`)) return

  try {
    await roleApi.remove(r.id)
    alert('删除成功')
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
  }
}

// 打开权限分配弹窗
async function openAuthModal(r) {
  currentRole.value = r
  selectedAuthIds.value = []
  authList.value = []
  loadingAuths.value = true
  authModalVisible.value = true

  try {
    // 并行加载权限列表和角色已分配的权限
    const [allAuths, roleDetail] = await Promise.all([
      roleApi.listAuths(),
      roleApi.get(r.id)
    ])
    authList.value = allAuths || []
    selectedAuthIds.value = roleDetail?.authIds || []
  } catch (err) {
    alert('加载权限数据失败：' + err.message)
    authModalVisible.value = false
  } finally {
    loadingAuths.value = false
  }
}

function closeAuthModal() {
  if (authSaving.value) return
  authModalVisible.value = false
}

// 保存权限分配
async function saveAuths() {
  authSaving.value = true
  try {
    await roleApi.setAuths(currentRole.value.id, selectedAuthIds.value)
    alert('权限设置成功')
    authSaving.value = false
    authModalVisible.value = false
    await loadData()
  } catch (err) {
    alert('保存失败：' + err.message)
  } finally {
    authSaving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.ops {
  display: flex;
  gap: 8px;
}

.perm-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  background: var(--c-primary-soft);
  color: var(--c-primary);
  margin-right: 6px;
  margin-bottom: 4px;
}

.btn-danger {
  color: var(--c-p0);
  border-color: rgba(201, 59, 59, 0.4);
}

.btn-danger:hover {
  background: var(--c-p0-bg);
  border-color: var(--c-p0);
}

.modal-large {
  max-width: 700px;
}

.auth-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 8px 0;
}

.auth-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.auth-item:hover {
  border-color: var(--c-primary);
  background: var(--c-primary-soft);
}

.auth-item input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--c-primary);
}

.auth-name {
  font-size: 14px;
  color: var(--c-text);
}
</style>
