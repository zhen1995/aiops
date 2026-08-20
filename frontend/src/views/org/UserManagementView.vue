<template>
  <div>
    <PageHeader title="用户管理" desc="维护平台用户账号信息">
      <div class="header-actions">
        <div class="search-box">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchKeyword"
            type="text"
            placeholder="搜索用户名或姓名"
            @keyup.enter="loadData"
          />
          <button v-if="searchKeyword" class="clear-btn" @click="clearSearch">×</button>
        </div>
        <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ 新增用户</button>
      </div>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">用户列表</h3>
          <p class="card-sub">平台全部账号信息</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>用户名</th>
            <th>姓名</th>
            <th>创建时间</th>
            <th>更新时间</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td><b>{{ u.username }}</b></td>
            <td>{{ u.name }}</td>
            <td class="muted">{{ formatTime(u.created_at) }}</td>
            <td class="muted">{{ formatTime(u.update_at) }}</td>
            <td>
              <LevelTag :level="isDeleted(u) ? 'info' : 'running'">
                {{ isDeleted(u) ? '已删除' : '正常' }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(u)">编辑</button>
                <button class="btn btn-sm" @click="openResetPwdModal(u)">重置密码</button>
                <button class="btn btn-sm btn-danger" @click="deleteUser(u)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && users.length === 0">
            <td colspan="6" class="empty-row">暂无用户数据，请点击"新增用户"添加</td>
          </tr>
          <tr v-if="loading">
            <td colspan="6" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑 用户 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? '编辑用户' : '新增用户' }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">用户名</label>
            <input
              v-model="form.username"
              class="form-input"
              placeholder="请输入用户名"
              :disabled="isEdit"
              maxlength="30"
            />
            <span v-if="errors.username" class="form-error">{{ errors.username }}</span>
          </div>

          <div v-if="!isEdit" class="form-item">
            <label class="form-label required">密码</label>
            <div class="input-with-action">
              <input
                v-model="form.password"
                class="form-input"
                :type="showPassword ? 'text' : 'password'"
                placeholder="请输入初始密码"
                maxlength="30"
              />
              <span class="input-action" @click="showPassword = !showPassword">
                {{ showPassword ? '隐藏' : '显示' }}
              </span>
            </div>
            <span v-if="errors.password" class="form-error">{{ errors.password }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">姓名</label>
            <input
              v-model="form.name"
              class="form-input"
              placeholder="请输入姓名"
              maxlength="10"
            />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">取消</button>
          <button class="btn btn-primary" @click="saveUser" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 重置密码 弹窗 -->
    <div v-if="resetPwdVisible" class="modal-mask" @click.self="closeResetPwdModal">
      <div class="modal">
        <div class="modal-header">
          <h3>重置密码</h3>
          <span class="modal-close" @click="closeResetPwdModal">×</span>
        </div>
        <div class="modal-body">
          <p class="form-tip">正在为用户 <b>{{ resetPwdUser?.username }}</b> 重置密码</p>
          <div class="form-item">
            <label class="form-label required">新密码</label>
            <div class="input-with-action">
              <input
                v-model="resetPwdForm.password"
                class="form-input"
                :type="resetPwdShowPassword ? 'text' : 'password'"
                placeholder="请输入新密码"
                maxlength="30"
              />
              <span class="input-action" @click="resetPwdShowPassword = !resetPwdShowPassword">
                {{ resetPwdShowPassword ? '隐藏' : '显示' }}
              </span>
            </div>
            <span v-if="resetPwdErrors.password" class="form-error">{{ resetPwdErrors.password }}</span>
          </div>
          <div class="form-item">
            <label class="form-label required">确认密码</label>
            <div class="input-with-action">
              <input
                v-model="resetPwdForm.confirmPassword"
                class="form-input"
                :type="resetPwdShowPassword ? 'text' : 'password'"
                placeholder="请再次输入新密码"
                maxlength="30"
              />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeResetPwdModal" :disabled="resetPwdSaving">取消</button>
          <button class="btn btn-primary" @click="submitResetPwd" :disabled="resetPwdSaving">
            {{ resetPwdSaving ? '提交中...' : '确认重置' }}
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
import { userApi } from '../../api/user.js'

// 数据
const users = ref([])
const loading = ref(false)
const saving = ref(false)

// 搜索
const searchKeyword = ref('')

// 新增/编辑弹窗
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const showPassword = ref(false)
const errors = reactive({})
const form = reactive({
  username: '',
  password: '',
  name: ''
})

// 重置密码弹窗
const resetPwdVisible = ref(false)
const resetPwdSaving = ref(false)
const resetPwdShowPassword = ref(false)
const resetPwdUser = ref(null)
const resetPwdErrors = reactive({})
const resetPwdForm = reactive({
  password: '',
  confirmPassword: ''
})

// 工具函数
const formatTime = (t) => {
  if (!t) return '-'
  const d = new Date(t)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const isDeleted = (u) => {
  return u.deleted_at && new Date(u.deleted_at).getTime() > 0
}

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const data = await userApi.list(searchKeyword.value.trim())
    users.value = data || []
  } catch (err) {
    alert('加载用户列表失败：' + err.message)
    users.value = []
  } finally {
    loading.value = false
  }
}

function clearSearch() {
  searchKeyword.value = ''
  loadData()
}

// 表单重置
function resetForm() {
  form.username = ''
  form.password = ''
  form.name = ''
  Object.keys(errors).forEach(k => delete errors[k])
}

// 新增
function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  showPassword.value = false
  modalVisible.value = true
}

// 编辑
function openEditModal(u) {
  isEdit.value = true
  editingId.value = u.id
  resetForm()
  form.username = u.username
  form.name = u.name
  form.password = ''
  showPassword.value = false
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

  if (!form.username.trim()) {
    errors.username = '请输入用户名'
    valid = false
  } else if (form.username.length > 30) {
    errors.username = '用户名最长 30 个字符'
    valid = false
  }

  if (!isEdit.value) {
    if (!form.password) {
      errors.password = '请输入密码'
      valid = false
    } else if (form.password.length < 6) {
      errors.password = '密码至少 6 位'
      valid = false
    } else if (form.password.length > 30) {
      errors.password = '密码最长 30 个字符'
      valid = false
    }
  }

  if (!form.name.trim()) {
    errors.name = '请输入姓名'
    valid = false
  } else if (form.name.length > 10) {
    errors.name = '姓名最长 10 个字符'
    valid = false
  }

  return valid
}

// 保存
async function saveUser() {
  if (!validateForm()) return

  saving.value = true
  try {
    if (isEdit.value) {
      await userApi.update(editingId.value, {
        username: form.username.trim(),
        name: form.name.trim()
      })
    } else {
      await userApi.create({
        username: form.username.trim(),
        password: form.password,
        name: form.name.trim()
      })
    }
    // 关闭弹窗并刷新
    saving.value = false
    modalVisible.value = false
    await loadData()
  } catch (err) {
    alert('保存失败：' + err.message)
  } finally {
    saving.value = false
  }
}

// 删除
async function deleteUser(u) {
  if (!confirm(`确定要删除用户「${u.username}」吗？此操作将进行软删除。`)) return

  try {
    await userApi.remove(u.id)
    alert('删除成功')
    await loadData()
  } catch (err) {
    alert('删除失败：' + err.message)
  }
}

// 打开重置密码弹窗
function openResetPwdModal(u) {
  resetPwdUser.value = u
  resetPwdForm.password = ''
  resetPwdForm.confirmPassword = ''
  resetPwdShowPassword.value = false
  Object.keys(resetPwdErrors).forEach(k => delete resetPwdErrors[k])
  resetPwdVisible.value = true
}

function closeResetPwdModal() {
  if (resetPwdSaving.value) return
  resetPwdVisible.value = false
}

async function submitResetPwd() {
  Object.keys(resetPwdErrors).forEach(k => delete resetPwdErrors[k])

  if (!resetPwdForm.password) {
    resetPwdErrors.password = '请输入新密码'
    return
  }
  if (resetPwdForm.password.length < 6) {
    resetPwdErrors.password = '密码至少 6 位'
    return
  }
  if (resetPwdForm.password !== resetPwdForm.confirmPassword) {
    resetPwdErrors.password = '两次输入的密码不一致'
    return
  }

  resetPwdSaving.value = true
  try {
    await userApi.resetPassword(resetPwdUser.value.id, resetPwdForm.password)
    alert('密码重置成功')
    resetPwdVisible.value = false
  } catch (err) {
    alert('重置密码失败：' + err.message)
  } finally {
    resetPwdSaving.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 7px 12px;
  color: var(--c-text-3);
  width: 240px;
  position: relative;
}

.search-box input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  width: 100%;
  color: var(--c-text);
}

.search-box .clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: none;
  color: var(--c-text-3);
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 0 4px;
}

.search-box .clear-btn:hover {
  color: var(--c-text);
}

.ops {
  display: flex;
  gap: 8px;
}

.btn-danger {
  color: var(--c-p0);
  border-color: rgba(201, 59, 59, 0.4);
}

.btn-danger:hover {
  background: var(--c-p0-bg);
  border-color: var(--c-p0);
}

.form-tip {
  color: var(--c-text-2);
  font-size: 13px;
  margin-bottom: 16px;
}
</style>
