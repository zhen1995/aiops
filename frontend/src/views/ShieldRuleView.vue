<template>
  <div>
    <PageHeader :title="$t('alert.shield.title')" :desc="$t('alert.shield.desc')">
      <button class="btn btn-primary" @click="openModal(null)">+ {{ $t('alert.shield.add') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('alert.shield.title') }}</h3>
          <p class="card-sub">{{ $t('alert.shield.desc') }}</p>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-sm" @click="loadData" :disabled="loading">
            {{ loading ? $t('group.loading') : $t('group.refresh') }}
          </button>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('alert.shield.colName') }}</th>
            <th>{{ $t('alert.shield.colTags') }}</th>
            <th>{{ $t('alert.shield.colStart') }}</th>
            <th>{{ $t('alert.shield.colEnd') }}</th>
            <th>{{ $t('alert.shield.colStatus') }}</th>
            <th>{{ $t('alert.shield.colCreator') }}</th>
            <th>{{ $t('alert.shield.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in rules" :key="r.id">
            <td>
              <a class="rule-name-link" :title="$t('alert.rules.edit')" @click="openModal(r)">{{ r.name }}</a>
              <div v-if="r.remark" class="muted remark-cell">{{ r.remark }}</div>
            </td>
            <td>
              <div class="tag-list">
                <span v-for="tag in parseTags(r.tags)" :key="tag" class="tag-chip">{{ tag }}</span>
              </div>
            </td>
            <td class="muted col-time">{{ fmtTime(r.start_at) }}</td>
            <td class="muted col-time">{{ fmtTime(r.end_at) }}</td>
            <td>
              <div class="status-cell">
                <label class="switch" :title="r.is_enabled === 1 ? $t('alert.shield.disable') : $t('alert.shield.enable')">
                  <input type="checkbox" :checked="r.is_enabled === 1" @change="toggleRule(r)" />
                  <span class="slider"></span>
                </label>
                <span v-if="statusClass(r) !== 'active'" class="status-tag" :class="statusClass(r)">{{ statusText(r) }}</span>
              </div>
            </td>
            <td class="muted">{{ r.created_by || '-' }}</td>
            <td>
              <button class="gp-op row-del-op" :title="$t('alert.rules.delete')" @click="removeRule(r)">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                  <path d="M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                </svg>
              </button>
            </td>
          </tr>
          <tr v-if="!loading && rules.length === 0">
            <td colspan="7" class="empty">{{ $t('alert.shield.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="7" class="empty">{{ $t('group.loading') }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑弹窗 -->
    <div v-if="modal.visible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ modal.isEdit ? $t('alert.shield.edit') : $t('alert.shield.add') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('alert.shield.nameLabel') }}</label>
            <input v-model="form.name" class="form-input" :placeholder="$t('alert.shield.namePlaceholder')" maxlength="128" />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">{{ $t('alert.shield.tagsLabel') }}</label>
            <div class="tag-editor">
              <div v-for="(tag, idx) in form.tagRows" :key="idx" class="tag-row">
                <input
                  v-model="form.tagRows[idx]"
                  class="form-input tag-input"
                  :placeholder="$t('alert.shield.tagPlaceholder')"
                  @input="errors.tags = ''"
                />
                <button class="tag-remove" title="×" @click="form.tagRows.splice(idx, 1)">×</button>
              </div>
              <button class="btn btn-sm" @click="form.tagRows.push('')">+ {{ $t('alert.shield.addTag') }}</button>
            </div>
            <span class="form-hint">{{ $t('alert.shield.tagsHint') }}</span>
            <span v-if="errors.tags" class="form-error">{{ errors.tags }}</span>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('alert.shield.quick') }}</label>
            <div class="quick-row">
              <button
                v-for="d in quickDurations"
                :key="d.label"
                class="quick-btn"
                :class="{ active: activeQuick === d.ms }"
                @click="applyQuick(d.ms)"
              >{{ d.label }}</button>
            </div>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('alert.shield.startLabel') }}</label>
              <input v-model="form.startAt" type="datetime-local" step="1" class="form-input" />
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('alert.shield.endLabel') }}</label>
              <input v-model="form.endAt" type="datetime-local" step="1" class="form-input" />
            </div>
          </div>
          <div v-if="durationText" class="form-hint duration-hint">{{ $t('alert.shield.duration') }}: {{ durationText }}</div>
          <span v-if="errors.time" class="form-error">{{ errors.time }}</span>

          <div class="form-item">
            <label class="form-label">{{ $t('alert.shield.remarkLabel') }}</label>
            <textarea v-model="form.remark" class="form-input form-textarea" rows="2" :placeholder="$t('alert.shield.remarkPlaceholder')" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('knowledge.failDialog.close') }}</button>
          <button class="btn btn-primary" @click="saveRule" :disabled="saving">
            {{ saving ? $t('system.actions.saving') : $t('system.actions.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import { shieldRuleApi } from '../api/shieldRule.js'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute()
const router = useRouter()

const rules = ref([])
const loading = ref(false)
const saving = ref(false)

const modal = reactive({ visible: false, isEdit: false, editingId: null })
const form = reactive({ name: '', tagRows: [''], startAt: '', endAt: '', remark: '' })
const errors = reactive({ name: '', tags: '', time: '' })
const activeQuick = ref(null)

const quickDurations = [
  { label: '30m', ms: 30 * 60 * 1000 },
  { label: '1h', ms: 60 * 60 * 1000 },
  { label: '2h', ms: 2 * 60 * 60 * 1000 },
  { label: '6h', ms: 6 * 60 * 60 * 1000 },
  { label: '12h', ms: 12 * 60 * 60 * 1000 },
  { label: '1d', ms: 24 * 60 * 60 * 1000 },
  { label: '3d', ms: 3 * 24 * 60 * 60 * 1000 },
  { label: '7d', ms: 7 * 24 * 60 * 60 * 1000 },
  { label: '30d', ms: 30 * 24 * 60 * 60 * 1000 }
]

const fmtTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN', { hour12: false })
}

// datetime-local 值格式化（本地时区，秒精度）
function toLocalInputValue(d) {
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function parseTags(raw) {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

// 规则状态：生效中 / 未开始 / 已过期 / 已停用
function statusClass(r) {
  if (r.is_enabled !== 1) return 'off'
  const now = Date.now()
  if (now < new Date(r.start_at).getTime()) return 'pending'
  if (now > new Date(r.end_at).getTime()) return 'expired'
  return 'active'
}

function statusText(r) {
  const map = {
    active: t('alert.shield.enable'),
    pending: t('knowledge.status.pending'),
    expired: t('alert.shield.colEnd'),
    off: t('alert.shield.disable')
  }
  const cls = statusClass(r)
  if (cls === 'pending') return t('knowledge.status.pending')
  if (cls === 'expired') return t('alert.rules.enabledOff')
  return map[cls]
}

const durationText = computed(() => {
  if (!form.startAt || !form.endAt) return ''
  const ms = new Date(form.endAt).getTime() - new Date(form.startAt).getTime()
  if (isNaN(ms) || ms <= 0) return ''
  const minutes = Math.round(ms / 60000)
  if (minutes < 60) return `${minutes}m`
  if (minutes < 60 * 24) return `${(minutes / 60).toFixed(1)}h`
  return `${(minutes / 60 / 24).toFixed(1)}d`
})

async function loadData() {
  loading.value = true
  try {
    rules.value = await shieldRuleApi.list({ includeExpired: true }) || []
  } catch (err) {
    alert(t('alert.shield.msg.loadFailed', { msg: err.message }))
  } finally {
    loading.value = false
  }
}

// 快捷时长：开始时间=现在，结束时间=现在+时长
function applyQuick(ms) {
  const now = new Date()
  form.startAt = toLocalInputValue(now)
  form.endAt = toLocalInputValue(new Date(now.getTime() + ms))
  activeQuick.value = ms
}

function openModal(rule) {
  modal.isEdit = !!rule
  modal.editingId = rule ? rule.id : null
  form.name = rule ? rule.name : ''
  form.tagRows = rule && parseTags(rule.tags).length > 0 ? parseTags(rule.tags) : ['']
  form.remark = rule ? (rule.remark || '') : ''
  activeQuick.value = null
  if (rule) {
    form.startAt = toLocalInputValue(new Date(rule.start_at))
    form.endAt = toLocalInputValue(new Date(rule.end_at))
  } else {
    applyQuick(60 * 60 * 1000)
  }
  errors.name = ''
  errors.tags = ''
  errors.time = ''
  modal.visible = true
}

function closeModal() {
  if (saving.value) return
  modal.visible = false
}

function validate() {
  errors.name = ''
  errors.tags = ''
  errors.time = ''
  let valid = true
  if (!form.name.trim()) {
    errors.name = t('alert.shield.err.nameRequired')
    valid = false
  }
  const tags = form.tagRows.map(s => s.trim()).filter(Boolean)
  if (tags.length === 0) {
    errors.tags = t('alert.shield.err.tagsRequired')
    valid = false
  } else if (tags.some(tag => !tag.includes('='))) {
    errors.tags = t('alert.shield.err.tagFormat')
    valid = false
  }
  if (!form.startAt || !form.endAt) {
    errors.time = t('alert.shield.err.timeRequired')
    valid = false
  } else if (new Date(form.endAt).getTime() <= new Date(form.startAt).getTime()) {
    errors.time = t('alert.shield.err.timeRange')
    valid = false
  }
  return valid ? tags : null
}

async function saveRule() {
  const tags = validate()
  if (!tags) return
  saving.value = true
  const payload = {
    name: form.name,
    tags: JSON.stringify(tags),
    start_at: form.startAt,
    end_at: form.endAt,
    is_enabled: 1,
    remark: form.remark
  }
  try {
    if (modal.isEdit) {
      await shieldRuleApi.update(modal.editingId, payload)
    } else {
      await shieldRuleApi.create(payload)
    }
    saving.value = false
    closeModal()
    await loadData()
  } catch (err) {
    alert(t('alert.shield.msg.saveFailed', { msg: err.message }))
    saving.value = false
  }
}

async function toggleRule(r) {
  try {
    await shieldRuleApi.toggle(r.id)
    await loadData()
  } catch (err) {
    alert(t('alert.shield.msg.toggleFailed', { msg: err.message }))
  }
}

async function removeRule(r) {
  if (!confirm(t('alert.shield.confirmDelete', { name: r.name }))) return
  try {
    await shieldRuleApi.remove(r.id)
    await loadData()
  } catch (err) {
    alert(t('alert.shield.msg.deleteFailed', { msg: err.message }))
  }
}

onMounted(async () => {
  await loadData()
  // 从告警事件弹窗「屏蔽」跳转而来：预填事件标签并打开新建弹窗
  const { create, tags, rule_name: ruleName } = route.query
  if (create === '1') {
    openModal(null)
    const tagArr = tags ? String(tags).split(',').map(s => s.trim()).filter(Boolean) : []
    form.tagRows = tagArr.length > 0 ? tagArr : ['']
    if (ruleName && !form.name) {
      form.name = `${ruleName}`
    }
    router.replace({ path: '/alerts/shields', query: {} })
  }
})
</script>

<style scoped>
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-head-flex {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 14px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  max-width: 320px;
}

.tag-chip {
  font-size: 11px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  background: var(--c-primary-soft, rgba(22, 119, 255, 0.08));
  color: var(--c-primary);
  padding: 1px 7px;
  border-radius: 10px;
  white-space: nowrap;
}

.remark-cell {
  font-size: 12px;
  margin-top: 2px;
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-time { white-space: nowrap; }

/* 规则名称：主色高亮可点击，点击打开编辑弹窗 */
.rule-name-link {
  color: var(--c-primary);
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.rule-name-link:hover {
  opacity: 0.8;
  text-decoration: underline;
}

.status-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 删除图标按钮（常显，悬停变红） */
.gp-op {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--c-text-3);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: all 0.15s;
}

.gp-op:hover { background: var(--c-primary-tint); color: var(--c-primary); }
.gp-op.row-del-op:hover { background: #ffe5e5; color: var(--c-danger); }

/* 启停开关 */
.switch { position: relative; display: inline-block; width: 40px; height: 22px; flex-shrink: 0; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--c-border); transition: .3s; border-radius: 22px;
}
.slider:before {
  position: absolute; content: ""; height: 16px; width: 16px; left: 3px; bottom: 3px;
  background-color: white; transition: .3s; border-radius: 50%;
}
input:checked + .slider { background-color: var(--c-primary); }
input:checked + .slider:before { transform: translateX(18px); }

.status-tag {
  font-size: 11px;
  font-weight: 600;
  border-radius: 10px;
  padding: 2px 9px;
  white-space: nowrap;
}

.status-tag.active { background: var(--c-primary-tint); color: var(--c-primary); }
.status-tag.pending { background: #fdf6dc; color: #d9a400; }
.status-tag.expired { background: var(--c-bg); color: var(--c-text-3); border: 1px solid var(--c-border); }
.status-tag.off { background: #f0f0f0; color: var(--c-text-3); }

.ops { display: flex; gap: 8px; }
.btn-danger { color: var(--c-danger); }
.btn-danger:hover { border-color: var(--c-danger); color: var(--c-danger); }

.empty {
  text-align: center;
  color: var(--c-text-3);
  padding: 32px 12px;
}

/* 弹窗表单 */
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
  width: 640px;
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

.form-item { margin-bottom: 14px; }

.form-label {
  display: block;
  font-size: 13px;
  color: var(--c-text-2);
  margin-bottom: 6px;
  font-weight: 500;
}
.form-label.required::before { content: '*'; color: var(--c-danger); margin-right: 3px; }

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
.form-textarea { resize: vertical; min-height: 60px; font-family: inherit; }

.form-hint {
  display: block;
  color: var(--c-text-3);
  font-size: 12px;
  margin-top: 4px;
}

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

.form-row { display: flex; gap: 16px; }
.form-item-half { flex: 1; }

.duration-hint { margin: -6px 0 12px; }

/* 标签编辑器 */
.tag-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}

.tag-row {
  display: flex;
  gap: 6px;
  width: 100%;
  align-items: center;
}

.tag-input {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12.5px;
}

.tag-remove {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--c-text-3);
  cursor: pointer;
  font-size: 14px;
  flex-shrink: 0;
  transition: all 0.15s;
}
.tag-remove:hover { background: #ffe5e5; color: var(--c-danger); }

/* 快捷时长 */
.quick-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.quick-btn {
  padding: 5px 12px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.quick-btn:hover {
  border-color: var(--c-primary);
  color: var(--c-primary);
}

.quick-btn.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
}
</style>
