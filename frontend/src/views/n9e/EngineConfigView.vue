<template>
  <div>
    <PageHeader :title="$t('n9e.engine.title')" :desc="$t('n9e.engine.desc')">
      <button class="btn btn-primary" @click="openForm(null)">{{ $t('n9e.engine.create') }}</button>
    </PageHeader>

    <div class="card">
      <div class="list-wrap">
        <div v-for="c in configs" :key="c.id" class="cfg-item">
          <div class="cfg-main">
            <div class="cfg-title">
              <b>{{ c.name }}</b>
              <span class="tag" :class="c.is_enabled === 1 ? 'ok' : 'off'">
                {{ $t(c.is_enabled === 1 ? 'n9e.engine.enabled' : 'n9e.engine.disabled') }}
              </span>
            </div>
            <div class="cfg-row">
              <span class="cfg-label">{{ $t('n9e.engine.address') }}</span>
              <code>{{ c.address }}</code>
            </div>
            <div class="cfg-row">
              <span class="cfg-label">Token</span>
              <code>{{ c.token || $t('n9e.engine.notConfigured') }}</code>
            </div>
          </div>
          <div class="cfg-actions">
            <button class="btn btn-sm" @click="openForm(c)">{{ $t('n9e.engine.edit') }}</button>
            <button class="btn btn-sm" @click="test(c)">{{ $t('n9e.engine.testConnection') }}</button>
            <button class="btn btn-sm" @click="toggle(c)">{{ $t(c.is_enabled === 1 ? 'n9e.engine.disable' : 'n9e.engine.enable') }}</button>
            <button class="btn btn-sm btn-danger-link" @click="remove(c)">{{ $t('n9e.engine.delete') }}</button>
          </div>
        </div>
        <div v-if="configs.length === 0 && !loading" class="empty">
          {{ $t('n9e.engine.empty') }}
        </div>
        <div v-if="loading" class="empty">{{ $t('n9e.engine.loading') }}</div>
      </div>
    </div>

    <!-- 新建/编辑 -->
    <div v-if="showForm" class="modal-mask" @click.self="showForm = false">
      <div class="modal">
        <div class="modal-head">
          <h3>{{ $t(editing ? 'n9e.engine.modalEditTitle' : 'n9e.engine.modalCreateTitle') }}</h3>
          <button class="close" @click="showForm = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label>{{ $t('n9e.engine.name') }}</label>
            <input v-model="form.name" :placeholder="$t('n9e.engine.namePlaceholder')" maxlength="64" />
          </div>
          <div class="form-item">
            <label>{{ $t('n9e.engine.addressLabel') }}</label>
            <input v-model="form.address" :placeholder="$t('n9e.engine.addressPlaceholder')" />
          </div>
          <div class="form-item">
            <label>{{ $t('n9e.engine.tokenLabel') }}</label>
            <input v-model="form.token" :placeholder="$t('n9e.engine.tokenPlaceholder')" />
          </div>
          <div class="form-item switch-item">
            <label>{{ $t('n9e.engine.enabledLabel') }}</label>
            <label class="switch">
              <input type="checkbox" v-model.number="form.is_enabled" true-value="1" false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="showForm = false">{{ $t('n9e.engine.cancel') }}</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ $t(saving ? 'n9e.engine.saving' : 'n9e.engine.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import { n9eApi } from '../../api/n9e.js'

const { t } = useI18n({ useScope: 'global' })

const configs = ref([])
const loading = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editing = ref(null)

const defaultForm = () => ({ name: t('n9e.engine.defaultName'), address: '', token: '', is_enabled: 1 })
const form = ref(defaultForm())

onMounted(load)

async function load() {
  loading.value = true
  try {
    configs.value = await n9eApi.listConfigs() || []
  } catch (e) {
    alert(t('n9e.engine.loadFailed', { message: e.message }))
  } finally {
    loading.value = false
  }
}

function openForm(c) {
  editing.value = c || null
  form.value = c
    ? { id: c.id, name: c.name, address: c.address, token: '', is_enabled: c.is_enabled }
    : defaultForm()
  showForm.value = true
}

async function save() {
  if (!form.value.name.trim() || !form.value.address.trim() || !form.value.token.trim()) {
    return alert(t('n9e.engine.fillAll'))
  }
  saving.value = true
  try {
    await n9eApi.saveConfig(form.value)
    showForm.value = false
    await load()
  } catch (e) {
    alert(t('n9e.engine.saveFailed', { message: e.message }))
  } finally {
    saving.value = false
  }
}

async function test(cfg) {
  try {
    const data = await n9eApi.testConfig({ address: cfg.address, token: cfg.token })
    if (data.ok) {
      alert(t('n9e.engine.testSuccess', { status: data.status }))
    } else {
      alert(t('n9e.engine.testFailed', { message: data.message || t('n9e.engine.statusCode', { status: data.status }) }))
    }
  } catch (e) {
    alert(t('n9e.engine.testError', { message: e.message }))
  }
}

async function toggle(c) {
  const payload = { id: c.id, name: c.name, address: c.address, token: c.token, is_enabled: c.is_enabled === 1 ? 0 : 1 }
  try {
    await n9eApi.saveConfig(payload)
    await load()
  } catch (e) {
    alert(t('n9e.engine.operateFailed', { message: e.message }))
  }
}

async function remove(c) {
  if (!confirm(t('n9e.engine.confirmDelete', { name: c.name }))) return
  try {
    await n9eApi.deleteConfig(c.id)
    await load()
  } catch (e) {
    alert(t('n9e.engine.deleteFailed', { message: e.message }))
  }
}
</script>

<style scoped>
.list-wrap { display: flex; flex-direction: column; gap: 12px; }

.cfg-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  background: var(--c-bg);
}
.cfg-main { flex: 1; }
.cfg-title { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
.cfg-row {
  display: flex; align-items: center; gap: 8px;
  font-size: 12.5px; color: var(--c-text-2);
  margin-top: 3px;
}
.cfg-label { color: var(--c-text-3); width: 50px; }
.cfg-row code {
  background: var(--c-surface); padding: 2px 8px; border-radius: 4px;
  font-family: monospace; font-size: 12px;
}

.tag { padding: 2px 8px; border-radius: 10px; font-size: 11.5px; }
.tag.ok { background: var(--c-success-bg, #e8f5ee); color: #1f7a45; }
.tag.off { background: #f0f0f0; color: var(--c-text-3); }

.cfg-actions { display: flex; gap: 8px; flex-shrink: 0; }

.empty { text-align: center; color: var(--c-text-3); padding: 40px 0; }

.btn-danger-link { background: transparent; color: var(--c-danger); border: none; padding: 4px 8px; font-size: 12px; cursor: pointer; }
.btn-danger-link:hover { text-decoration: underline; }

.switch-item { display: flex; align-items: center; justify-content: space-between; }
.switch-item label:first-child { margin-bottom: 0; }

.switch { position: relative; display: inline-block; width: 38px; height: 20px; }
.switch input { display: none; }
.switch .slider {
  position: absolute; inset: 0; background: #ccc; border-radius: 20px; transition: 0.2s;
}
.switch .slider::before {
  content: ''; position: absolute; width: 16px; height: 16px; left: 2px; top: 2px;
  background: #fff; border-radius: 50%; transition: 0.2s;
}
.switch input:checked + .slider { background: var(--c-primary); }
.switch input:checked + .slider::before { transform: translateX(18px); }

/* Modal */
.modal-mask {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}
.modal { background: var(--c-surface); border-radius: 12px; width: 480px; display: flex; flex-direction: column; box-shadow: 0 12px 40px rgba(0,0,0,0.2); }
.modal-head { padding: 18px 22px; border-bottom: 1px solid var(--c-border); display: flex; justify-content: space-between; align-items: center; }
.modal-head h3 { margin: 0; font-size: 15px; }
.modal-head .close { border: none; background: transparent; font-size: 22px; color: var(--c-text-3); cursor: pointer; }
.modal-body { padding: 20px 22px; }
.modal-foot { padding: 14px 22px; border-top: 1px solid var(--c-border); display: flex; justify-content: flex-end; gap: 10px; }

.form-item { margin-bottom: 16px; }
.form-item label { display: block; font-size: 13px; color: var(--c-text-2); margin-bottom: 6px; font-weight: 500; }
.form-item input {
  width: 100%; border: 1px solid var(--c-border); border-radius: 8px;
  padding: 9px 12px; font-size: 13px; background: var(--c-bg); color: var(--c-text);
  outline: none; box-sizing: border-box; font-family: inherit;
}
.form-item input:focus { border-color: var(--c-primary); }
</style>
