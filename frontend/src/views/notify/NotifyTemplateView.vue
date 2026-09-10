<template>
  <div>
    <PageHeader :title="$t('notify.template.pageTitle')" :desc="$t('notify.template.pageDesc')">
      <button class="btn btn-primary" @click="openForm(null)">{{ $t('notify.template.create') }}</button>
    </PageHeader>

    <div v-if="showForm" class="modal-mask" @click.self="close">
      <div class="modal modal-xl">
        <div class="modal-head">
          <h3>{{ editing ? $t('notify.template.modal.editTitle') : $t('notify.template.modal.createTitle') }}</h3>
          <button class="close" @click="close">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item">
              <label>{{ $t('notify.template.modal.name') }} <span class="req">*</span></label>
              <input v-model="form.name" :placeholder="$t('notify.template.modal.namePlaceholder')" />
            </div>
            <div class="form-item">
              <label>{{ $t('notify.template.modal.scene') }}</label>
              <select v-model="form.type">
                <option value="all">{{ $t('notify.template.modal.sceneAll') }}</option>
                <option value="firing">{{ $t('notify.template.modal.sceneFiring') }}</option>
                <option value="recovered">{{ $t('notify.template.modal.sceneRecovered') }}</option>
              </select>
            </div>
            <div class="form-item">
              <label>{{ $t('notify.template.modal.mediaType') }}</label>
              <select v-model="form.media_type">
                <option value="dingtalk">{{ $t('notify.template.modal.mediaDingtalk') }}</option>
                <option value="webhook">{{ $t('notify.template.modal.mediaWebhook') }}</option>
                <option value="email">{{ $t('notify.template.modal.mediaEmail') }}</option>
                <option value="wecom">{{ $t('notify.template.modal.mediaWecom') }}</option>
              </select>
            </div>
          </div>
          <div class="form-item">
            <label>{{ $t('notify.template.modal.description') }}</label>
            <input v-model="form.description" :placeholder="$t('notify.template.modal.descriptionPlaceholder')" />
          </div>

          <!-- 模板编辑器（占满宽度） -->
          <div class="tpl-edit-wrap">
            <div class="tpl-edit-head">
              <span>{{ $t('notify.template.modal.contentLabel') }}</span>
            </div>
            <textarea
              v-model="form.content"
              rows="14"
              class="tpl-textarea"
              :placeholder="$t('notify.template.modal.contentPlaceholder')"></textarea>
          </div>

          <!-- 预览 + 变量参考 -->
          <div class="tpl-bottom">
            <div class="tpl-preview">
              <div class="tpl-panel-title">{{ $t('notify.template.modal.preview') }}
                <span v-if="!previewContent && !previewError" class="tpl-hint">{{ $t('notify.template.modal.previewHint') }}</span>
              </div>
              <div v-if="previewError" class="tpl-error">❌ {{ previewError }}</div>
              <pre v-else-if="previewContent" class="preview-text">{{ previewContent }}</pre>
              <div v-else class="preview-placeholder">{{ $t('notify.template.modal.emptyPreview') }}</div>
            </div>

            <div class="tpl-ref">
              <div class="tpl-panel-title">{{ $t('notify.template.modal.refTitle') }}
                <span class="tpl-hint">{{ $t('notify.template.modal.refHint') }}</span>
              </div>
              <div class="ref-tabs">
                <button class="ref-tab" :class="{active: refTab === 'vars'}" @click="refTab = 'vars'">{{ $t('notify.template.modal.tabVars') }}</button>
                <button class="ref-tab" :class="{active: refTab === 'funcs'}" @click="refTab = 'funcs'">{{ $t('notify.template.modal.tabFuncs') }}</button>
              </div>
              <div v-if="refTab === 'vars'" class="ref-table">
                <div class="ref-row ref-head">
                  <span>{{ $t('notify.template.modal.colVar') }}</span>
                  <span>{{ $t('notify.template.modal.colDesc') }}</span>
                </div>
                <div
                  v-for="item in alertVars" :key="item.key" class="ref-row clickable"
                  :title="$t('notify.template.modal.clickInsert') + item.snippet"
                  @click="insertFunc(item.snippet)">
                  <code>{{ item.snippet }}</code>
                  <span class="ref-desc">{{ item.desc }}</span>
                </div>
              </div>
              <div v-else class="ref-table">
                <div class="ref-row ref-head">
                  <span>{{ $t('notify.template.modal.colFunc') }}</span>
                  <span>{{ $t('notify.template.modal.colDesc') }}</span>
                </div>
                <div
                  v-for="item in builtinFuncs" :key="item.key" class="ref-row clickable"
                  :title="$t('notify.template.modal.clickInsert') + item.snippet"
                  @click="insertFunc(item.snippet)">
                  <code>{{ item.snippet }}</code>
                  <span class="ref-desc">{{ item.desc }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="close">{{ $t('notify.template.modal.cancel') }}</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? $t('notify.template.modal.saving') : $t('notify.template.modal.save') }}
          </button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-list">
        <div v-for="t in templates" :key="t.id" class="card-item">
          <div class="ci-head">
            <div>
              <b>{{ t.name }}</b>
              <span class="type-chip" :class="t.type">{{ typeLabel(t.type) }}</span>
              <span class="type-chip media-chip">{{ mediaLabel(t.media_type) }}</span>
            </div>
            <span class="ci-time">{{ fmtTime(t.updated_at || t.created_at) }}</span>
          </div>
          <div v-if="t.description" class="ci-desc">{{ t.description }}</div>
          <pre class="ci-preview">{{ truncate(contentOf(t.content), 160) }}</pre>
          <div class="ci-actions">
            <button class="btn btn-sm" @click="openForm(t)">{{ $t('notify.template.list.edit') }}</button>
            <button class="btn btn-sm btn-danger" @click="remove(t)">{{ $t('notify.template.list.delete') }}</button>
          </div>
        </div>
        <div v-if="templates.length === 0 && !loading" class="empty">{{ $t('notify.template.list.empty') }}</div>
        <div v-if="loading" class="empty">{{ $t('notify.template.list.loading') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import { notifyApi } from '../../api/notify.js'

const { t } = useI18n({ useScope: 'global' })

const templates = ref([])
const loading = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editing = ref(null)
const previewContent = ref('')
const previewError = ref('')
const refTab = ref('vars') // vars | funcs

const defaultContent = `#### {{if $event.IsRecovered}}<font color="#008800">💚{{$event.RuleName}}</font>{{else}}<font color="#FF0000">💔{{$event.RuleName}}</font>{{end}}
---
{{$time_duration := sub now $event.FirstTrigger.Unix }}{{if $event.IsRecovered}}{{$time_duration = sub $event.LastEvalTime.Unix $event.FirstTrigger.Unix }}{{end}}
- **告警级别**: {{$event.SeverityLabel}}
{{- if $event.RuleNote}}
	- **规则备注**: {{$event.RuleNote}}
{{- end}}
{{- if not $event.IsRecovered}}
- **触发时值**: {{$event.TriggerValue}}
- **触发时间**: {{timeformatCN $event.TriggerTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- else}}
- **恢复时间**: {{timeformatCN $event.LastEvalTime}}
- **告警持续时长**: {{humanizeDuration $time_duration}}
{{- end}}
- **业务组**: {{$event.BusiGroupName}}
- **告警事件标签**:
{{- range $key, $val := $event.TagsMap}}
	- {{$key}}: {{$val}}
{{- end}}`

// -------- 告警事件变量（按夜莺分类） --------
const alertVars = [
  // 整个事件
  { key: 'event',     snippet: '{{$event}}',              desc: t('notify.template.vars.event') },
  { key: 'labels',    snippet: '{{$labels}}',             desc: t('notify.template.vars.labels') },
  { key: 'value',     snippet: '{{$value}}',              desc: t('notify.template.vars.value') },
  { key: 'domain',    snippet: '{{$.domain}}',            desc: t('notify.template.vars.domain') },

  // 基本信息
  { key: 'RuleName',      snippet: '{{$event.RuleName}}',     desc: t('notify.template.vars.ruleName') },
  { key: 'RuleNote',      snippet: '{{$event.RuleNote}}',     desc: t('notify.template.vars.ruleNote') },
  { key: 'Id',            snippet: '{{$event.Id}}',           desc: t('notify.template.vars.id') },
  { key: 'SeverityLabel', snippet: '{{$event.SeverityLabel}}',desc: t('notify.template.vars.severityLabel') },
  { key: 'TriggerValue',  snippet: '{{$event.TriggerValue}}', desc: t('notify.template.vars.triggerValue') },
  { key: 'BusiGroupName', snippet: '{{$event.BusiGroupName}}',desc: t('notify.template.vars.busiGroupName') },
  { key: 'Cluster',       snippet: '{{$event.Cluster}}',      desc: t('notify.template.vars.cluster') },

  // 触发相关
  { key: 'TriggerTime',    snippet: '{{timeformatCN $event.TriggerTime}}',    desc: t('notify.template.vars.triggerTime') },
  { key: 'LastEvalTime',   snippet: '{{timeformatCN $event.LastEvalTime}}',   desc: t('notify.template.vars.lastEvalTime') },
  { key: 'FirstTrigger',   snippet: '{{timeformatCN $event.FirstTrigger}}',   desc: t('notify.template.vars.firstTrigger') },
  { key: 'IsRecovered',    snippet: '{{if $event.IsRecovered}}已恢复{{else}}告警{{end}}', desc: t('notify.template.vars.isRecovered') },

  // 标签与注解
  { key: 'TagsMap',    snippet: '{{$event.TagsMap.instance}}',   desc: t('notify.template.vars.tagsMap') },
  { key: 'TagsJSON',   snippet: '{{$event.TagsJSON}}',           desc: t('notify.template.vars.tagsJson') },
  { key: 'Annotations',snippet: '{{$event.AnnotationsJSON.dashboard}}', desc: t('notify.template.vars.annotations') },
]

// -------- 内置函数（参考夜莺 tplx） --------
const builtinFuncs = [
  // 时间
  { key: 'timeformat',  snippet: '{{timeformat $event.TriggerTime "2006-01-02 15:04:05"}}', desc: t('notify.template.funcs.timeformat') },
  { key: 'timeformatCN',snippet: '{{timeformatCN $event.TriggerTime}}',                   desc: t('notify.template.funcs.timeformatCN') },
  { key: 'timestamp',   snippet: '{{timestamp}}',                                           desc: t('notify.template.funcs.timestamp') },
  { key: 'now',         snippet: '{{now}}',                                                 desc: t('notify.template.funcs.now') },

  // 算术
  { key: 'sub', snippet: '{{sub now $event.FirstTrigger.Unix}}',      desc: t('notify.template.funcs.sub') },
  { key: 'add', snippet: '{{add $event.DurationSec 60}}',            desc: t('notify.template.funcs.add') },
  { key: 'mul', snippet: '{{mul $event.DurationSec 1000}}',          desc: t('notify.template.funcs.mul') },

  // 人类可读
  { key: 'humanizeDuration',    snippet: '{{humanizeDuration $time_duration}}',    desc: t('notify.template.funcs.humanizeDuration') },
  { key: 'humanizeDurationIfc', snippet: '{{humanizeDurationInterface $time_duration}}', desc: t('notify.template.funcs.humanizeDurationIfc') },
  { key: 'durationHuman',       snippet: '{{durationHuman $event.DurationSec}}',     desc: t('notify.template.funcs.durationHuman') },
]

const defaultForm = () => ({
  name: '', description: '', type: 'all', media_type: 'dingtalk',
  content: defaultContent, is_enabled: 1
})
const form = ref(defaultForm())

// 实时预览
let previewTimer = null
watch(() => form.value.content, () => {
  clearTimeout(previewTimer)
  previewTimer = setTimeout(doPreview, 400)
})
watch(() => showForm.value, (v) => { if (v) doPreview() })

async function doPreview() {
  if (!form.value.content) { previewContent.value = ''; previewError.value = ''; return }
  try {
    const data = await notifyApi.previewTemplate(form.value.content, null)
    previewContent.value = data.content || ''
    previewError.value = data.error || ''
  } catch (e) {
    previewError.value = e.message
  }
}

function insertFunc(snippet) {
  form.value.content = (form.value.content || '') + snippet
}

async function load() {
  loading.value = true
  try {
    templates.value = await notifyApi.listTemplates() || []
  } catch (e) {
    alert(t('notify.template.error.loadFailed') + e.message)
  } finally { loading.value = false }
}

function openForm(tpl) {
  editing.value = tpl || null
  form.value = tpl ? { ...tpl } : defaultForm()
  previewContent.value = ''
  previewError.value = ''
  refTab.value = 'vars'
  showForm.value = true
}

function close() { showForm.value = false; editing.value = null }

async function save() {
  if (!form.value.name.trim() || !form.value.content.trim()) {
    return alert(t('notify.template.error.nameContentRequired'))
  }
  saving.value = true
  try {
    if (editing.value) {
      await notifyApi.updateTemplate(editing.value.id, form.value)
    } else {
      await notifyApi.createTemplate(form.value)
    }
    showForm.value = false
    await load()
  } catch (e) {
    alert(t('notify.template.error.saveFailed') + e.message)
  } finally { saving.value = false }
}

async function remove(tpl) {
  if (!confirm(t('notify.template.error.confirmDelete', { name: tpl.name }))) return
  try { await notifyApi.deleteTemplate(tpl.id); await load() } catch (e) { alert(t('notify.template.error.deleteFailed') + e.message) }
}

function typeLabel(v) {
  return {
    all: t('notify.template.type.all'),
    firing: t('notify.template.type.firing'),
    recovered: t('notify.template.type.recovered')
  }[v] || v
}
function mediaLabel(v) {
  return {
    dingtalk: t('notify.template.mediaLabel.dingtalk'),
    webhook: t('notify.template.mediaLabel.webhook'),
    email: t('notify.template.mediaLabel.email'),
    wecom: t('notify.template.mediaLabel.wecom')
  }[v] || v
}
function fmtTime(s) { if (!s) return ''; return new Date(s).toLocaleString('zh-CN') }
function truncate(s, n) { if (!s) return ''; return s.length > n ? s.slice(0, n) + '...' : s }
const contentOf = (c) => c

onMounted(load)
</script>

<style scoped>
.modal-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal { background: var(--c-surface); border-radius: 12px; display: flex; flex-direction: column; box-shadow: 0 12px 40px rgba(0,0,0,0.2); }
.modal-xl { width: 1100px; max-height: 90vh; }
.modal-head { padding: 14px 20px; border-bottom: 1px solid var(--c-border); display: flex; justify-content: space-between; align-items: center; }
.modal-head h3 { margin: 0; font-size: 15px; }
.close { border: none; background: transparent; font-size: 22px; color: var(--c-text-3); cursor: pointer; }
.modal-body { padding: 18px 20px; overflow-y: auto; flex: 1; }
.modal-foot { padding: 12px 20px; border-top: 1px solid var(--c-border); display: flex; justify-content: flex-end; gap: 10px; }

.form-row { display: flex; gap: 12px; }
.form-row .form-item { flex: 1; }
.form-item { margin-bottom: 14px; }
.form-item label { display: block; font-size: 12.5px; color: var(--c-text-2); margin-bottom: 5px; }
.req { color: var(--c-danger, #c93b3b); }
.form-item input, .form-item select {
  width: 100%; padding: 7px 10px; border: 1px solid var(--c-border); border-radius: 7px;
  font-size: 13px; background: var(--c-bg); color: var(--c-text); outline: none; box-sizing: border-box;
}
.form-item input:focus, .form-item select:focus { border-color: var(--c-primary); }

/* 编辑器（占满宽度） */
.tpl-edit-wrap { border: 1px solid var(--c-border); border-radius: 8px; overflow: hidden; margin-bottom: 14px; }
.tpl-edit-head {
  padding: 7px 12px; background: var(--c-surface); border-bottom: 1px solid var(--c-border);
  font-size: 12px; color: var(--c-text-3);
}
.tpl-textarea {
  width: 100%; padding: 10px 12px; border: none; outline: none; resize: vertical; min-height: 240px;
  font-family: ui-monospace, Menlo, Monaco, Consolas, monospace; font-size: 12.5px;
  background: var(--c-bg); color: var(--c-text); box-sizing: border-box; line-height: 1.65;
}

/* 下部分栏：预览 + 变量参考 */
.tpl-bottom { display: grid; grid-template-columns: 1fr 380px; gap: 14px; }

.tpl-preview, .tpl-ref {
  border: 1px solid var(--c-border); border-radius: 8px; overflow: hidden;
  background: var(--c-surface);
}
.tpl-panel-title {
  padding: 7px 12px; background: var(--c-surface); border-bottom: 1px solid var(--c-border);
  font-size: 12px; color: var(--c-text-2); font-weight: 500;
  display: flex; justify-content: space-between; align-items: center;
}
.tpl-hint { color: var(--c-text-3); font-weight: normal; font-size: 11px; }
.tpl-error { padding: 10px 12px; font-size: 12px; color: var(--c-danger, #c93b3b); background: #fff3f3; }
.preview-text { margin: 0; padding: 10px 12px; font-size: 12.5px; line-height: 1.7; white-space: pre-wrap; word-break: break-word; color: var(--c-text); overflow-y: auto; max-height: 320px; }
.preview-placeholder { padding: 30px 12px; text-align: center; color: var(--c-text-3); font-size: 12px; }

/* 变量参考表格 */
.ref-tabs { display: flex; gap: 4px; padding: 8px 12px; background: var(--c-bg); border-bottom: 1px solid var(--c-border); }
.ref-tab {
  padding: 3px 10px; font-size: 11.5px; border-radius: 5px; border: 1px solid var(--c-border);
  background: transparent; color: var(--c-text-2); cursor: pointer;
}
.ref-tab.active { background: var(--c-primary); color: #fff; border-color: var(--c-primary); }

.ref-table { max-height: 360px; overflow-y: auto; }
.ref-row {
  display: grid; grid-template-columns: 170px 1fr; gap: 10px;
  padding: 6px 12px; font-size: 12px; line-height: 1.5;
  border-bottom: 1px solid var(--c-border);
}
.ref-row.ref-head {
  background: var(--c-primary-soft, rgba(22,119,255,0.06)); font-weight: 600; color: var(--c-text-2); font-size: 11.5px;
}
.ref-row code {
  font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 11px;
  background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary);
  padding: 2px 6px; border-radius: 3px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.ref-row.clickable { cursor: pointer; transition: background .15s; }
.ref-row.clickable:hover { background: var(--c-primary-soft, rgba(22,119,255,0.06)); }
.ref-desc { color: var(--c-text-2); }

/* 列表卡片 */
.type-chip { padding: 1px 8px; font-size: 10.5px; border-radius: 10px; margin-left: 6px; }
.type-chip.all { background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); }
.type-chip.firing { background: #ffe5e5; color: var(--c-danger, #c93b3b); }
.type-chip.recovered { background: #e8f5ee; color: #1f7a45; }
.type-chip.media-chip { background: #f0f4ff; color: #5b67c6; }

.btn-danger { background: transparent; border: 1px solid transparent; color: var(--c-danger, #c93b3b); cursor: pointer; padding: 4px 8px; font-size: 12px; }
.btn-danger:hover { text-decoration: underline; }

.card-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.card-item { border: 1px solid var(--c-border); border-radius: 10px; padding: 14px 16px; background: var(--c-bg); }
.ci-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.ci-time { font-size: 11.5px; color: var(--c-text-3); }
.ci-desc { font-size: 12px; color: var(--c-text-2); margin-bottom: 8px; }
.ci-preview {
  background: var(--c-surface); border-radius: 6px; padding: 8px 10px; margin: 6px 0 10px;
  font-family: ui-monospace, Menlo, Monaco, Consolas, monospace; font-size: 11.5px; color: var(--c-text-2);
  white-space: pre-wrap; line-height: 1.6; overflow: hidden; max-height: 110px;
}
.ci-actions { display: flex; gap: 6px; }

.empty { text-align: center; color: var(--c-text-3); padding: 40px 0; }

.btn-sm { padding: 3px 9px; font-size: 11.5px; border-radius: 5px; border: 1px solid var(--c-border); background: var(--c-surface); color: var(--c-text); cursor: pointer; }
.btn-sm:hover { border-color: var(--c-primary); color: var(--c-primary); }
</style>
