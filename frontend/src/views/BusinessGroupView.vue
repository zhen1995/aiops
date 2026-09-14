<template>
  <div>
    <PageHeader :title="$t('group.title')" :desc="$t('group.desc')">
      <button class="btn btn-primary" @click="openCreateModal" :disabled="loading">+ {{ $t('group.addGroup') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('group.listTitle') }}</h3>
          <p class="card-sub">{{ $t('group.listSub') }}</p>
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? $t('group.loading') : $t('group.refresh') }}
        </button>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('group.colName') }}</th>
            <th>{{ $t('group.colRuleCount') }}</th>
            <th>{{ $t('group.colCreatedAt') }}</th>
            <th>{{ $t('group.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="g in groups" :key="g.id">
            <td><b>{{ g.name }}</b></td>
            <td>
              <span v-if="g.rule_count > 0" class="count-badge">{{ g.rule_count }}</span>
              <span v-else class="muted">0</span>
            </td>
            <td class="muted">{{ fmtTime(g.created_at) }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="openEditModal(g)">{{ $t('group.edit') }}</button>
                <button class="btn btn-sm btn-danger" @click="deleteGroup(g)">{{ $t('group.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && groups.length === 0">
            <td colspan="4" class="empty-row">{{ $t('group.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="4" class="empty-row">{{ $t('group.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑 弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('group.modal.editTitle') : $t('group.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('group.form.nameLabel') }}</label>
            <input
              v-model="form.name"
              class="form-input"
              :placeholder="$t('group.form.namePlaceholder')"
              maxlength="128"
              @keyup.enter="saveGroup"
            />
            <span v-if="error" class="form-error">{{ error }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('group.form.cancel') }}</button>
          <button class="btn btn-primary" @click="saveGroup" :disabled="saving">
            {{ saving ? $t('group.form.saving') : $t('group.form.save') }}
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
import { businessGroupApi } from '../api/businessGroup.js'

const { t } = useI18n({ useScope: 'global' })

const groups = ref([])
const loading = ref(false)
const saving = ref(false)

const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const form = reactive({ name: '' })
const error = ref('')

const fmtTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

async function loadData() {
  loading.value = true
  try {
    groups.value = await businessGroupApi.list() || []
  } catch (err) {
    alert(t('group.msg.loadFailed', { msg: err.message }))
    groups.value = []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.name = ''
  error.value = ''
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

function openEditModal(g) {
  isEdit.value = true
  editingId.value = g.id
  resetForm()
  form.name = g.name
  modalVisible.value = true
}

function closeModal() {
  if (saving.value) return
  modalVisible.value = false
}

async function saveGroup() {
  error.value = ''
  if (!form.name.trim()) {
    error.value = t('group.error.nameRequired')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await businessGroupApi.update(editingId.value, { name: form.name })
    } else {
      await businessGroupApi.create({ name: form.name })
    }
    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert(t(isEdit.value ? 'group.msg.updateFailed' : 'group.msg.createFailed', { msg: err.message }))
    saving.value = false
  }
}

async function deleteGroup(g) {
  if (!confirm(t('group.msg.confirmDelete', { name: g.name }))) return
  try {
    await businessGroupApi.remove(g.id)
    await loadData()
  } catch (err) {
    alert(t('group.msg.deleteFailed', { msg: err.message }))
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

.count-badge {
  font-size: 12px;
  font-weight: 600;
  background: var(--c-primary-tint);
  color: var(--c-primary);
  border-radius: 10px;
  padding: 1px 8px;
}

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}

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
  width: 480px;
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

.modal-body { padding: 20px 24px; overflow-y: auto; }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 24px;
  border-top: 1px solid var(--c-border);
}

.form-item { margin-bottom: 14px; }
.form-label { display: block; font-size: 13px; color: var(--c-text-2); margin-bottom: 6px; font-weight: 500; }
.form-label.required::before { content: '*'; color: var(--c-danger); margin-right: 3px; }

.form-input {
  width: 100%; padding: 8px 12px; border: 1px solid var(--c-border); border-radius: 6px;
  background: var(--c-surface); color: var(--c-text); font-size: 13px;
  transition: border-color 0.15s; outline: none; font-family: inherit; box-sizing: border-box;
}
.form-input:focus { border-color: var(--c-primary); }
.form-input::placeholder { color: var(--c-text-3); }

.form-error { display: block; color: var(--c-danger); font-size: 12px; margin-top: 4px; }
</style>
