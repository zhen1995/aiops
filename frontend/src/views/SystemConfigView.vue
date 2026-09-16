<template>
  <div>
    <PageHeader :title="$t('system.header.title')" :desc="$t('system.header.desc')" />

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('system.card.title') }}</h3>
          <p class="card-sub">{{ $t('system.card.sub') }}</p>
        </div>
      </div>

      <div class="form-item">
        <label class="form-label required">{{ $t('system.form.qdrantUrl') }}</label>
        <input
          v-model="qdrantUrl"
          class="form-input"
          :placeholder="$t('system.form.qdrantUrlPlaceholder')"
          :disabled="loading"
        />
        <span v-if="urlError" class="form-error">{{ urlError }}</span>
      </div>

      <div class="test-result" v-if="testResult" :class="testResult.ok ? 'test-ok' : 'test-fail'">
        {{ testResult.ok ? $t('system.result.success') : $t('system.result.failed') }}：{{ testResult.message }}
      </div>

      <div class="actions">
        <button class="btn btn-primary" @click="saveConfig" :disabled="loading || saving">
          {{ saving ? $t('system.actions.saving') : $t('system.actions.save') }}
        </button>
        <button class="btn" @click="testConnection" :disabled="loading || testing">
          {{ testing ? $t('system.actions.testing') : $t('system.actions.testConnection') }}
        </button>
      </div>
    </div>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('system.frontend.title') }}</h3>
          <p class="card-sub">{{ $t('system.frontend.sub') }}</p>
        </div>
      </div>

      <div class="form-item">
        <label class="form-label required">{{ $t('system.frontend.urlLabel') }}</label>
        <input
          v-model="frontendBaseUrl"
          class="form-input"
          :placeholder="$t('system.frontend.urlPlaceholder')"
          :disabled="loading"
        />
        <span v-if="frontendUrlError" class="form-error">{{ frontendUrlError }}</span>
      </div>

      <div class="actions">
        <button class="btn btn-primary" @click="saveFrontendUrl" :disabled="loading || savingFrontend">
          {{ savingFrontend ? $t('system.actions.saving') : $t('system.actions.save') }}
        </button>
      </div>
    </div>

    <div class="card tips-card">
      <h3 class="card-title">{{ $t('system.tips.title') }}</h3>
      <ul class="tips-list">
        <li>{{ $t('system.tips.item1') }}</li>
        <li>{{ $t('system.tips.item2') }}</li>
        <li>{{ $t('system.tips.item3') }}</li>
        <li>{{ $t('system.tips.item4') }}</li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import { systemConfigApi } from '../api/systemConfig.js'

const { t } = useI18n({ useScope: 'global' })

const qdrantUrl = ref('')
const frontendBaseUrl = ref('')
const loading = ref(false)
const saving = ref(false)
const savingFrontend = ref(false)
const testing = ref(false)
const urlError = ref('')
const frontendUrlError = ref('')
const testResult = ref(null)

async function loadConfig() {
  loading.value = true
  try {
    const data = await systemConfigApi.getSystemConfig()
    qdrantUrl.value = (data && data.qdrant_url) || ''
    frontendBaseUrl.value = (data && data.frontend_base_url) || ''
  } catch (err) {
    alert(t('system.messages.loadFailed', { message: err.message }))
  } finally {
    loading.value = false
  }
}

function validateUrl() {
  urlError.value = ''
  if (!qdrantUrl.value.trim()) {
    urlError.value = t('system.messages.urlRequired')
    return false
  }
  return true
}

async function saveConfig() {
  testResult.value = null
  if (!validateUrl()) return
  saving.value = true
  try {
    const data = await systemConfigApi.updateSystemConfig({ qdrant_url: qdrantUrl.value.trim() })
    if (data && data.qdrant_url) {
      qdrantUrl.value = data.qdrant_url
    }
    alert(t('system.messages.saveSuccess'))
  } catch (err) {
    alert(t('system.messages.saveFailed', { message: err.message }))
  } finally {
    saving.value = false
  }
}

async function testConnection() {
  if (!validateUrl()) return
  testing.value = true
  testResult.value = null
  try {
    testResult.value = await systemConfigApi.testQdrant({ qdrant_url: qdrantUrl.value.trim() })
  } catch (err) {
    testResult.value = { ok: false, message: err.message }
  } finally {
    testing.value = false
  }
}

function validateFrontendUrl() {
  frontendUrlError.value = ''
  const url = frontendBaseUrl.value.trim()
  if (!url) {
    frontendUrlError.value = t('system.messages.frontendUrlRequired')
    return false
  }
  if (!/^https?:\/\//.test(url)) {
    frontendUrlError.value = t('system.messages.frontendUrlInvalid')
    return false
  }
  return true
}

async function saveFrontendUrl() {
  if (!validateFrontendUrl()) return
  savingFrontend.value = true
  try {
    const data = await systemConfigApi.updateSystemConfig({ frontend_base_url: frontendBaseUrl.value.trim() })
    if (data && data.frontend_base_url) {
      frontendBaseUrl.value = data.frontend_base_url
    }
    alert(t('system.messages.saveSuccess'))
  } catch (err) {
    alert(t('system.messages.saveFailed', { message: err.message }))
  } finally {
    savingFrontend.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.form-item {
  margin-bottom: 14px;
  max-width: 520px;
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

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

.actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.test-result {
  max-width: 520px;
  padding: 10px 14px;
  border-radius: 6px;
  font-size: 13px;
  margin-top: 4px;
}
.test-ok {
  background: var(--c-success-bg);
  color: var(--c-success);
}
.test-fail {
  background: var(--c-surface);
  border: 1px solid var(--c-danger);
  color: var(--c-danger);
}

.tips-card { margin-top: 16px; }
.tips-list {
  margin: 8px 0 0;
  padding-left: 18px;
  color: var(--c-text-2);
  font-size: 13px;
  line-height: 1.8;
}
</style>
