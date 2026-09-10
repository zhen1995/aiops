<template>
  <div>
    <PageHeader :title="$t('org.user.title')" :desc="$t('org.user.desc')">
      <div class="header-actions">
        <div class="search-box">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchKeyword"
            type="text"
            :placeholder="$t('org.user.searchPlaceholder')"
            @keyup.enter="loadData"
          />
          <button v-if="searchKeyword" class="clear-btn" @click="clearSearch">×</button>
        </div>
        <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">{{ $t('org.user.addUser') }}</button>
      </div>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('org.user.listTitle') }}</h3>
          <p class="card-sub">{{ $t('org.user.listSub') }}</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? $t('org.user.loading') : $t('org.user.refresh') }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('org.user.table.username') }}</th>
            <th>{{ $t('org.user.table.name') }}</th>
            <th>{{ $t('org.user.table.roles') }}</th>
            <th>{{ $t('org.user.table.createdAt') }}</th>
            <th>{{ $t('org.user.table.updatedAt') }}</th>
            <th>{{ $t('org.user.table.status') }}</th>
            <th>{{ $t('org.user.table.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td><b>{{ u.username }}</b></td>
            <td>{{ u.name }}</td>
            <td>
              <span v-for="role in (u.roles || [])" :key="role" class="role-tag">{{ role }}</span>
              <span v-if="!u.roles || u.roles.length === 0" class="muted">-</span>
            </td>
            <td class="muted">{{ formatTime(u.created_at) }}</td>
            <td class="muted">{{ formatTime(u.update_at) }}</td>
            <td>
              <LevelTag :level="isDeleted(u) ? 'info' : 'running'">
                {{ isDeleted(u) ? $t('org.user.statusDeleted') : $t('org.user.statusNormal') }}
              </LevelTag>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(u)">{{ $t('org.user.edit') }}</button>
                <button class="btn btn-sm" @click="openResetPwdModal(u)">{{ $t('org.user.resetPassword') }}</button>
                <button class="btn btn-sm btn-danger" @click="deleteUser(u)">{{ $t('org.user.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && users.length === 0">
            <td colspan="7" class="empty-row">{{ $t('org.user.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="7" class="empty-row">{{ $t('org.user.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新增/编辑 用户 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('org.user.modal.editTitle') : $t('org.user.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('org.user.modal.username') }}</label>
            <input
              v-model="form.username"
              class="form-input"
              :placeholder="$t('org.user.modal.usernamePlaceholder')"
              :disabled="isEdit"
              maxlength="30"
            />
            <span v-if="errors.username" class="form-error">{{ errors.username }}</span>
          </div>

          <div v-if="!isEdit" class="form-item">
            <label class="form-label required">{{ $t('org.user.modal.password') }}</label>
            <div class="input-with-action">
              <input
                v-model="form.password"
                class="form-input"
                :type="showPassword ? 'text' : 'password'"
                :placeholder="$t('org.user.modal.passwordPlaceholder')"
                maxlength="30"
              />
              <span class="input-action" @click="showPassword = !showPassword">
                {{ showPassword ? $t('org.user.modal.hide') : $t('org.user.modal.show') }}
              </span>
            </div>
            <span v-if="errors.password" class="form-error">{{ errors.password }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">{{ $t('org.user.modal.name') }}</label>
            <input
              v-model="form.name"
              class="form-input"
              :placeholder="$t('org.user.modal.namePlaceholder')"
              maxlength="10"
            />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('org.user.modal.roles') }}</label>
            <div v-if="!rolesLoading && roleOptions.length === 0" class="muted" style="font-size: 12px;">{{ $t('org.user.modal.noRoles') }}</div>
            <div v-else-if="rolesLoading" class="muted" style="font-size: 12px;">{{ $t('org.user.modal.loadingRoles') }}</div>
            <div v-else class="role-select">
              <label v-for="role in roleOptions" :key="role.id" class="role-option">
                <input
                  type="checkbox"
                  :value="role.id"
                  v-model="form.roleIds"
                />
                <span>{{ role.name }}</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('org.user.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="saveUser" :disabled="saving">
            {{ saving ? $t('org.user.modal.saving') : $t('org.user.modal.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 重置密码 弹窗 -->
    <div v-if="resetPwdVisible" class="modal-mask" @click.self="closeResetPwdModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('org.user.resetPwd.title') }}</h3>
          <span class="modal-close" @click="closeResetPwdModal">×</span>
        </div>
        <div class="modal-body">
          <p class="form-tip">{{ $t('org.user.resetPwd.tipPrefix') }} <b>{{ resetPwdUser?.username }}</b> {{ $t('org.user.resetPwd.tipSuffix') }}</p>
          <div class="form-item">
            <label class="form-label required">{{ $t('org.user.resetPwd.newPassword') }}</label>
            <div class="input-with-action">
              <input
                v-model="resetPwdForm.password"
                class="form-input"
                :type="resetPwdShowPassword ? 'text' : 'password'"
                :placeholder="$t('org.user.resetPwd.newPasswordPlaceholder')"
                maxlength="30"
              />
              <span class="input-action" @click="resetPwdShowPassword = !resetPwdShowPassword">
                {{ resetPwdShowPassword ? $t('org.user.modal.hide') : $t('org.user.modal.show') }}
              </span>
            </div>
            <span v-if="resetPwdErrors.password" class="form-error">{{ resetPwdErrors.password }}</span>
          </div>
          <div class="form-item">
            <label class="form-label required">{{ $t('org.user.resetPwd.confirmPassword') }}</label>
            <div class="input-with-action">
              <input
                v-model="resetPwdForm.confirmPassword"
                class="form-input"
                :type="resetPwdShowPassword ? 'text' : 'password'"
                :placeholder="$t('org.user.resetPwd.confirmPasswordPlaceholder')"
                maxlength="30"
              />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeResetPwdModal" :disabled="resetPwdSaving">{{ $t('org.user.modal.cancel') }}</button>
          <button class="btn btn-primary" @click="submitResetPwd" :disabled="resetPwdSaving">
            {{ resetPwdSaving ? $t('org.user.resetPwd.submitting') : $t('org.user.resetPwd.confirm') }}
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
import { userApi } from '../../api/user.js'
import { roleApi } from '../../api/role.js'

const { t } = useI18n({ useScope: 'global' })

// 数据
const users = ref([])
const loading = ref(false)
const saving = ref(false)

// 角色选项
const roleOptions = ref([])
const rolesLoading = ref(false)

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
  name: '',
  roleIds: []
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

// 加载角色选项
async function loadRoleOptions() {
  rolesLoading.value = true
  try {
    roleOptions.value = await roleApi.listSimple() || []
  } catch (err) {
    console.error('加载角色列表失败', err)
    roleOptions.value = []
  } finally {
    rolesLoading.value = false
  }
}

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const data = await userApi.list(searchKeyword.value.trim())
    users.value = data || []
  } catch (err) {
    alert(t('org.user.loadFailed') + err.message)
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
  form.roleIds = []
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
async function openEditModal(u) {
  isEdit.value = true
  editingId.value = u.id
  resetForm()
  form.username = u.username
  form.name = u.name
  form.password = ''
  showPassword.value = false

  // 加载用户详情获取角色 ID
  try {
    const detail = await userApi.get(u.id)
    form.roleIds = detail?.roleIds || []
  } catch (err) {
    form.roleIds = []
  }

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
    errors.username = t('org.user.validation.usernameRequired')
    valid = false
  } else if (form.username.length > 30) {
    errors.username = t('org.user.validation.usernameMax')
    valid = false
  }

  if (!isEdit.value) {
    if (!form.password) {
      errors.password = t('org.user.validation.passwordRequired')
      valid = false
    } else if (form.password.length < 6) {
      errors.password = t('org.user.validation.passwordMin')
      valid = false
    } else if (form.password.length > 30) {
      errors.password = t('org.user.validation.passwordMax')
      valid = false
    }
  }

  if (!form.name.trim()) {
    errors.name = t('org.user.validation.nameRequired')
    valid = false
  } else if (form.name.length > 10) {
    errors.name = t('org.user.validation.nameMax')
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
        name: form.name.trim(),
        roleIds: form.roleIds
      })
    } else {
      await userApi.create({
        username: form.username.trim(),
        password: form.password,
        name: form.name.trim(),
        roleIds: form.roleIds
      })
    }
    saving.value = false
    modalVisible.value = false
    await loadData()
  } catch (err) {
    alert(t('org.user.saveFailed') + err.message)
  } finally {
    saving.value = false
  }
}

// 删除
async function deleteUser(u) {
  if (!confirm(t('org.user.deleteConfirm', { name: u.username }))) return

  try {
    await userApi.remove(u.id)
    alert(t('org.user.deleteSuccess'))
    await loadData()
  } catch (err) {
    alert(t('org.user.deleteFailed') + err.message)
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
    resetPwdErrors.password = t('org.user.resetPwd.newPasswordRequired')
    return
  }
  if (resetPwdForm.password.length < 6) {
    resetPwdErrors.password = t('org.user.resetPwd.passwordMin')
    return
  }
  if (resetPwdForm.password !== resetPwdForm.confirmPassword) {
    resetPwdErrors.password = t('org.user.resetPwd.mismatch')
    return
  }

  resetPwdSaving.value = true
  try {
    await userApi.resetPassword(resetPwdUser.value.id, resetPwdForm.password)
    alert(t('org.user.resetPwd.success'))
    resetPwdVisible.value = false
  } catch (err) {
    alert(t('org.user.resetPwdFailed') + err.message)
  } finally {
    resetPwdSaving.value = false
  }
}

onMounted(() => {
  loadData()
  loadRoleOptions()
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

.role-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  background: var(--c-primary-soft);
  color: var(--c-primary);
  margin-right: 4px;
  margin-bottom: 2px;
}

.role-select {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 0;
}

.role-option {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.role-option:hover {
  border-color: var(--c-primary);
  background: var(--c-primary-soft);
}

.role-option input[type="checkbox"] {
  width: 14px;
  height: 14px;
  accent-color: var(--c-primary);
}
</style>
