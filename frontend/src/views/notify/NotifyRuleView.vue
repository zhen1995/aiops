<template>
  <div>
    <PageHeader :title="$t('notify.rule.pageTitle')" :desc="$t('notify.rule.pageDesc')">
      <button class="btn btn-primary" @click="openForm(null)">{{ $t('notify.rule.create') }}</button>
    </PageHeader>

    <!-- 编辑弹窗 -->
    <div v-if="showForm" class="modal-mask" @click.self="close">
      <div class="modal modal-lg">
        <div class="modal-head">
          <h3>{{ editing ? $t('notify.rule.modal.editTitle') : $t('notify.rule.modal.createTitle') }}</h3>
          <button class="close" @click="close">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item">
              <label>{{ $t('notify.rule.modal.name') }} <span class="req">*</span></label>
              <input v-model="form.name" :placeholder="$t('notify.rule.modal.namePlaceholder')" />
            </div>
          </div>
          <div class="form-item">
            <label>{{ $t('notify.rule.modal.remark') }}</label>
            <input v-model="form.remark" :placeholder="$t('notify.rule.modal.remarkPlaceholder')" />
          </div>

          <div class="form-section">
            <label class="section-label">{{ $t('notify.rule.modal.media') }} <span class="hint">{{ $t('notify.rule.modal.pickOne') }}</span></label>
            <div v-if="medias.length === 0" class="hint-tip">{{ $t('notify.rule.modal.mediaEmpty') }}</div>
            <select v-model="form.media_id" class="form-input" v-if="medias.length > 0">
              <option value="">{{ $t('notify.rule.modal.selectMedia') }}</option>
              <option v-for="m in medias" :key="m.id" :value="m.id">{{ m.name }}（{{ mediaTypeLabel(m.type) }}）</option>
            </select>
          </div>

          <div class="form-section">
            <label class="section-label">{{ $t('notify.rule.modal.template') }} <span class="hint">{{ $t('notify.rule.modal.pickOne') }}</span></label>
            <div v-if="templates.length === 0" class="hint-tip">{{ $t('notify.rule.modal.templateEmpty') }}</div>
            <select v-model="form.template_id" class="form-input" v-if="templates.length > 0">
              <option value="">{{ $t('notify.rule.modal.selectTemplate') }}</option>
              <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}（{{ mediaTypeLabel(t.media_type) || t.media_type }}）</option>
            </select>
          </div>

          <div class="form-item form-item-toggle" style="margin-top: 14px;">
            <label class="form-label">{{ $t('notify.rule.modal.enabled') }}</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span></span>
            </label>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="close">{{ $t('notify.rule.modal.cancel') }}</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? $t('notify.rule.modal.saving') : $t('notify.rule.modal.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 列表 -->
    <div class="card">
      <div class="tb-head">
        <span class="tb-count">{{ $t('notify.rule.list.total', { count: rules.length }) }}</span>
      </div>
      <div class="tb-wrap">
        <table class="tb">
          <thead>
            <tr>
              <th style="width:20%">{{ $t('notify.rule.list.name') }}</th>
              <th style="width:14%">{{ $t('notify.rule.list.triggerScene') }}</th>
              <th style="width:22%">{{ $t('notify.rule.list.media') }}</th>
              <th style="width:22%">{{ $t('notify.rule.list.template') }}</th>
              <th style="width:8%">{{ $t('notify.rule.list.status') }}</th>
              <th style="width:14%">{{ $t('notify.rule.list.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in rules" :key="r.id">
              <td>
                <div>{{ r.name }}</div>
                <div v-if="r.remark" class="tb-remark">{{ r.remark }}</div>
              </td>
              <td>
                <span class="type-chip" :class="r.trigger_types">{{ triggerLabel(r.trigger_types) }}</span>
              </td>
              <td>
                <template v-if="r.media_id">
                  <span class="mini-chip media">{{ (medias.find(m => m.id === r.media_id) || {}).name || r.media_id }}</span>
                </template>
                <span v-else class="empty-s">—</span>
              </td>
              <td>
                <template v-if="r.template_id">
                  <span class="mini-chip tpl">{{ (templates.find(t => t.id === r.template_id) || {}).name || r.template_id }}</span>
                </template>
                <span v-else class="empty-s">—</span>
              </td>
              <td>
                <label class="switch">
                  <input type="checkbox" :checked="r.is_enabled === 1" @change="toggleEnabled(r)" />
                  <span></span>
                </label>
              </td>
              <td class="op-col">
                <button class="btn-tiny" @click="openForm(r)">{{ $t('notify.rule.list.edit') }}</button>
                <button class="btn-tiny danger" @click="remove(r)">{{ $t('notify.rule.list.delete') }}</button>
              </td>
            </tr>
            <tr v-if="rules.length === 0 && !loading">
              <td colspan="6" class="tb-empty">{{ $t('notify.rule.list.empty') }}</td>
            </tr>
            <tr v-if="loading"><td colspan="6" class="tb-empty">{{ $t('notify.rule.list.loading') }}</td></tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import { notifyApi } from '../../api/notify.js'

const { t } = useI18n({ useScope: 'global' })

const rules = ref([])
const templates = ref([])
const medias = ref([])
const loading = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editing = ref(null)

const defaultForm = () => ({
  name: '', remark: '',
  media_id: '', template_id: '',
  trigger_types: 'all', is_enabled: 1
})
const form = ref(defaultForm())

async function loadAll() {
  loading.value = true
  try {
    const [r, t, m] = await Promise.all([
      notifyApi.listRules(),
      notifyApi.listTemplates(),
      notifyApi.listMedia()
    ])
    rules.value = r || []
    templates.value = t || []
    medias.value = m || []
  } finally { loading.value = false }
}

function openForm(r) {
  editing.value = r || null
  if (r) {
    form.value = {
      ...defaultForm(),
      ...r,
      media_id: r.media_id || '',
      template_id: r.template_id || ''
    }
  } else {
    form.value = defaultForm()
  }
  showForm.value = true
}

function close() { showForm.value = false; editing.value = null }

async function save() {
  if (!form.value.name.trim()) return alert(t('notify.rule.error.nameRequired'))
  saving.value = true
  const payload = {
    name: form.value.name,
    remark: form.value.remark,
    media_id: form.value.media_id || '',
    template_id: form.value.template_id || '',
    trigger_types: form.value.trigger_types,
    is_enabled: form.value.is_enabled
  }
  try {
    if (editing.value) {
      await notifyApi.updateRule(editing.value.id, payload)
    } else {
      await notifyApi.createRule(payload)
    }
    showForm.value = false
    await loadAll()
  } catch (e) {
    alert(t('notify.rule.error.saveFailed') + e.message)
  } finally { saving.value = false }
}

async function remove(r) {
  if (!confirm(t('notify.rule.error.confirmDelete', { name: r.name }))) return
  try { await notifyApi.deleteRule(r.id); await loadAll() } catch (e) { alert(t('notify.rule.error.deleteFailed') + e.message) }
}

async function toggleEnabled(r) {
  try {
    const updated = await notifyApi.toggleRule(r.id)
    r.is_enabled = updated.is_enabled
  } catch (e) { alert(t('notify.rule.error.toggleFailed') + e.message) }
}

function mediaTypeLabel(type) {
  return {
    dingtalk: t('notify.rule.mediaType.dingtalk'),
    webhook: t('notify.rule.mediaType.webhook'),
    email: t('notify.rule.mediaType.email'),
    wecom: t('notify.rule.mediaType.wecom')
  }[type] || type
}
function triggerLabel(v) {
  return {
    all: t('notify.rule.trigger.all'),
    firing: t('notify.rule.trigger.firing'),
    recovered: t('notify.rule.trigger.recovered')
  }[v] || v
}

onMounted(loadAll)
</script>

<style scoped>
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal { background: var(--c-surface); border-radius: 12px; display: flex; flex-direction: column; box-shadow: 0 12px 40px rgba(0,0,0,0.2); }
.modal-lg { width: 820px; max-height: 88vh; }
.modal-head { padding: 14px 20px; border-bottom: 1px solid var(--c-border); display: flex; justify-content: space-between; align-items: center; }
.modal-head h3 { margin: 0; font-size: 15px; }
.close { border: none; background: transparent; font-size: 22px; color: var(--c-text-3); cursor: pointer; }
.modal-body { padding: 16px 20px; overflow-y: auto; flex: 1; }
.modal-foot { padding: 12px 20px; border-top: 1px solid var(--c-border); display: flex; justify-content: flex-end; gap: 10px; }

.form-row { display: flex; gap: 12px; }
.form-row .form-item { flex: 1; }
.form-item { margin-bottom: 12px; }
.form-item label { display: block; font-size: 12.5px; color: var(--c-text-2); margin-bottom: 5px; }
.req { color: var(--c-danger, #c93b3b); }
.form-item input, .form-item select {
  width: 100%; padding: 7px 10px; border: 1px solid var(--c-border); border-radius: 7px;
  font-size: 13px; background: var(--c-bg); color: var(--c-text); outline: none; box-sizing: border-box;
}

.form-section { margin-top: 14px; padding-top: 14px; border-top: 1px dashed var(--c-border); }
.section-label { display: block; font-size: 12.5px; font-weight: 500; color: var(--c-text); margin-bottom: 8px; }
.section-label .hint { color: var(--c-text-3); font-weight: 400; font-size: 11.5px; margin-left: 4px; }
.hint-tip { font-size: 11.5px; color: var(--c-text-3); padding: 6px 0; }

.chip-group { display: flex; flex-wrap: wrap; gap: 6px; }
.chip {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 5px 10px; border: 1px solid var(--c-border); border-radius: 16px;
  background: var(--c-bg); font-size: 12px; color: var(--c-text); cursor: pointer; user-select: none;
  transition: all 0.15s;
}
.chip:hover { border-color: var(--c-primary); }
.chip.selected { border-color: var(--c-primary); background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); }
.chip.disabled { opacity: 0.5; cursor: not-allowed; }
.chip-type { font-size: 10px; padding: 1px 6px; border-radius: 8px; background: var(--c-surface); color: var(--c-text-3); }
.chip-type.dingtalk { background: #1a6cff22; color: #1677ff; }
.chip-type.webhook { background: #ff8c1a22; color: #d86a00; }
.chip-type.email { background: #2ea44f22; color: #1f7a45; }
.chip-type.tpl { background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); }
.chip-type.n9e { background: #ffd91a22; color: #a07200; }
.chip-name { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.sev-group { display: flex; gap: 10px; flex-wrap: wrap; }
.sev-box {
  display: inline-flex; align-items: center; gap: 5px; cursor: pointer;
  padding: 5px 12px; border: 1px solid var(--c-border); border-radius: 16px;
  font-size: 12px; color: var(--c-text-2); background: var(--c-bg);
}
.sev-box input { display: none; }
.sev-box.on { border-color: var(--c-primary); background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); }

.tb-head { display: flex; justify-content: space-between; padding: 12px 16px; border-bottom: 1px solid var(--c-border); align-items: center; }
.tb-count { font-size: 12px; color: var(--c-text-3); }
.tb-wrap { overflow-x: auto; }
.tb { width: 100%; border-collapse: collapse; font-size: 13px; }
.tb thead th { padding: 10px 12px; font-size: 12px; color: var(--c-text-3); font-weight: 500; text-align: left; background: var(--c-surface); border-bottom: 1px solid var(--c-border); }
.tb tbody td { padding: 11px 12px; border-bottom: 1px solid var(--c-border); vertical-align: middle; }
.tb tbody tr:hover { background: var(--c-primary-soft, rgba(22,119,255,0.03)); }
.tb-remark { font-size: 11.5px; color: var(--c-text-3); margin-top: 2px; }
.tb-empty { text-align: center; color: var(--c-text-3); padding: 32px !important; }
.op-col { display: flex; gap: 8px; }

.type-chip { padding: 1px 8px; font-size: 10.5px; border-radius: 10px; }
.type-chip.all { background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); }
.type-chip.firing { background: #ffe5e5; color: var(--c-danger, #c93b3b); }
.type-chip.recovered { background: #e8f5ee; color: #1f7a45; }
.sev-label { font-size: 11px; color: var(--c-text-3); margin-left: 6px; }

.mini-chip {
  display: inline-block; padding: 1px 7px; margin: 2px 3px 2px 0; font-size: 11px; border-radius: 9px;
  background: var(--c-surface); color: var(--c-text-2); border: 1px solid var(--c-border);
}
.mini-chip.media { background: #ffe5e5; color: var(--c-danger, #c93b3b); border-color: #ffd2d2; }
.mini-chip.tpl { background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); border-color: #c2d9ff; }
.empty-s { color: var(--c-text-3); font-size: 12px; }

.btn-tiny { background: transparent; border: none; color: var(--c-primary); font-size: 12px; cursor: pointer; padding: 2px 4px; }
.btn-tiny:hover { text-decoration: underline; }
.btn-tiny.danger { color: var(--c-danger, #c93b3b); }

.switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.switch input { display: none; }
.switch span {
  position: absolute; cursor: pointer; inset: 0; background: var(--c-border); border-radius: 20px; transition: 0.2s;
}
.switch span::before {
  content: ''; position: absolute; height: 14px; width: 14px; left: 3px; bottom: 3px; background: white; border-radius: 50%; transition: 0.2s;
}
.switch input:checked + span { background: var(--c-primary); }
.switch input:checked + span::before { transform: translateX(16px); }

.btn-sm { padding: 3px 9px; font-size: 11.5px; border-radius: 5px; border: 1px solid var(--c-border); background: var(--c-surface); color: var(--c-text); cursor: pointer; }
.btn-sm:hover { border-color: var(--c-primary); color: var(--c-primary); }
</style>
