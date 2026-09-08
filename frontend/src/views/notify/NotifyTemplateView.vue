<template>
  <div>
    <PageHeader title="消息模板" desc="定义告警通知的内容模板，使用 Go template 语法，支持告警事件变量">
      <button class="btn btn-primary" @click="openForm(null)">+ 新建模板</button>
    </PageHeader>

    <div v-if="showForm" class="modal-mask" @click.self="close">
      <div class="modal modal-xl">
        <div class="modal-head">
          <h3>{{ editing ? '编辑消息模板' : '新建消息模板' }}</h3>
          <button class="close" @click="close">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item">
              <label>模板名称 <span class="req">*</span></label>
              <input v-model="form.name" placeholder="例如：钉钉-通用告警模板" />
            </div>
            <div class="form-item">
              <label>适用场景</label>
              <select v-model="form.type">
                <option value="all">通用（触发 + 恢复）</option>
                <option value="firing">仅告警触发</option>
                <option value="recovered">仅告警恢复</option>
              </select>
            </div>
            <div class="form-item">
              <label>媒介类型</label>
              <select v-model="form.media_type">
                <option value="dingtalk">钉钉</option>
                <option value="webhook">Webhook</option>
                <option value="email">邮件</option>
                <option value="wecom">企业微信</option>
              </select>
            </div>
          </div>
          <div class="form-item">
            <label>描述</label>
            <input v-model="form.description" placeholder="模板说明，帮助识别用途" />
          </div>

          <!-- 模板编辑器（占满宽度） -->
          <div class="tpl-edit-wrap">
            <div class="tpl-edit-head">
              <span>模板内容 (Go template)</span>
            </div>
            <textarea
              v-model="form.content"
              rows="14"
              class="tpl-textarea"
              placeholder='{{$event.RuleName}} 告警&#10;触发值: {{$event.TriggerValue}}&#10;站点: {{$.domain}}'></textarea>
          </div>

          <!-- 预览 + 变量参考 -->
          <div class="tpl-bottom">
            <div class="tpl-preview">
              <div class="tpl-panel-title">实时预览
                <span v-if="!previewContent && !previewError" class="tpl-hint">（模板合法后将显示渲染结果）</span>
              </div>
              <div v-if="previewError" class="tpl-error">❌ {{ previewError }}</div>
              <pre v-else-if="previewContent" class="preview-text">{{ previewContent }}</pre>
              <div v-else class="preview-placeholder">无内容</div>
            </div>

            <div class="tpl-ref">
              <div class="tpl-panel-title">变量参考
                <span class="tpl-hint">（点击可插入模板）</span>
              </div>
              <div class="ref-tabs">
                <button class="ref-tab" :class="{active: refTab === 'vars'}" @click="refTab = 'vars'">告警事件变量</button>
                <button class="ref-tab" :class="{active: refTab === 'funcs'}" @click="refTab = 'funcs'">内置函数</button>
              </div>
              <div v-if="refTab === 'vars'" class="ref-table">
                <div class="ref-row ref-head">
                  <span>变量</span>
                  <span>说明</span>
                </div>
                <div
                  v-for="item in alertVars" :key="item.key" class="ref-row clickable"
                  :title="'点击插入: ' + item.snippet"
                  @click="insertFunc(item.snippet)">
                  <code>{{ item.snippet }}</code>
                  <span class="ref-desc">{{ item.desc }}</span>
                </div>
              </div>
              <div v-else class="ref-table">
                <div class="ref-row ref-head">
                  <span>函数</span>
                  <span>说明</span>
                </div>
                <div
                  v-for="item in builtinFuncs" :key="item.key" class="ref-row clickable"
                  :title="'点击插入: ' + item.snippet"
                  @click="insertFunc(item.snippet)">
                  <code>{{ item.snippet }}</code>
                  <span class="ref-desc">{{ item.desc }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="close">取消</button>
          <button class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? '保存中...' : '保存模板' }}
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
            <button class="btn btn-sm" @click="openForm(t)">编辑</button>
            <button class="btn btn-sm btn-danger" @click="remove(t)">删除</button>
          </div>
        </div>
        <div v-if="templates.length === 0 && !loading" class="empty">暂无模板，点击右上角新建</div>
        <div v-if="loading" class="empty">加载中...</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import PageHeader from '../../components/PageHeader.vue'
import { notifyApi } from '../../api/notify.js'

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
  { key: 'event',     snippet: '{{$event}}',              desc: '整个告警事件对象，可用于调试查看全部字段' },
  { key: 'labels',    snippet: '{{$labels}}',             desc: '事件标签 map，等价于 $event.TagsMap' },
  { key: 'value',     snippet: '{{$value}}',              desc: '触发值，等价于 $event.TriggerValue' },
  { key: 'domain',    snippet: '{{$.domain}}',            desc: '站点地址，用于拼接详情链接' },

  // 基本信息
  { key: 'RuleName',      snippet: '{{$event.RuleName}}',     desc: '告警规则名称' },
  { key: 'RuleNote',      snippet: '{{$event.RuleNote}}',     desc: '告警规则备注/描述' },
  { key: 'Id',            snippet: '{{$event.Id}}',           desc: '告警事件唯一 ID' },
  { key: 'SeverityLabel', snippet: '{{$event.SeverityLabel}}',desc: '告警级别中文（P1-紧急 / P2-警告 / P3-提醒）' },
  { key: 'TriggerValue',  snippet: '{{$event.TriggerValue}}', desc: '触发值（PromQL 表达式计算结果）' },
  { key: 'BusiGroupName', snippet: '{{$event.BusiGroupName}}',desc: '夜莺业务组名称' },
  { key: 'Cluster',       snippet: '{{$event.Cluster}}',      desc: '告警集群标识' },

  // 触发相关
  { key: 'TriggerTime',    snippet: '{{timeformatCN $event.TriggerTime}}',    desc: '触发时间' },
  { key: 'LastEvalTime',   snippet: '{{timeformatCN $event.LastEvalTime}}',   desc: '最近一次 PromQL 命中时间' },
  { key: 'FirstTrigger',   snippet: '{{timeformatCN $event.FirstTrigger}}',   desc: '首次触发时间' },
  { key: 'IsRecovered',    snippet: '{{if $event.IsRecovered}}已恢复{{else}}告警{{end}}', desc: '是否已恢复（布尔）' },

  // 标签与注解
  { key: 'TagsMap',    snippet: '{{$event.TagsMap.instance}}',   desc: '事件标签 map，支持按 key 取值' },
  { key: 'TagsJSON',   snippet: '{{$event.TagsJSON}}',           desc: '事件标签 JSON 字符串' },
  { key: 'Annotations',snippet: '{{$event.AnnotationsJSON.dashboard}}', desc: '附加 annotations 字段' },
]

// -------- 内置函数（参考夜莺 tplx） --------
const builtinFuncs = [
  // 时间
  { key: 'timeformat',  snippet: '{{timeformat $event.TriggerTime "2006-01-02 15:04:05"}}', desc: '格式化时间，第二个参数为 Go layout' },
  { key: 'timeformatCN',snippet: '{{timeformatCN $event.TriggerTime}}',                   desc: '简化版，输出 yyyy-MM-dd HH:mm:ss' },
  { key: 'timestamp',   snippet: '{{timestamp}}',                                           desc: '当前时间字符串（常用于"发送时间"）' },
  { key: 'now',         snippet: '{{now}}',                                                 desc: '当前秒级时间戳（int64），可用来算持续时长' },

  // 算术
  { key: 'sub', snippet: '{{sub now $event.FirstTrigger.Unix}}',      desc: '减法，常用于计算告警持续秒数' },
  { key: 'add', snippet: '{{add $event.DurationSec 60}}',            desc: '加法' },
  { key: 'mul', snippet: '{{mul $event.DurationSec 1000}}',          desc: '乘法' },

  // 人类可读
  { key: 'humanizeDuration',    snippet: '{{humanizeDuration $time_duration}}',    desc: '秒数 → 人类可读（3分 / 2时 / 1.5天）' },
  { key: 'humanizeDurationIfc', snippet: '{{humanizeDurationInterface $time_duration}}', desc: 'humanizeDuration 别名，兼容夜莺模板' },
  { key: 'durationHuman',       snippet: '{{durationHuman $event.DurationSec}}',     desc: '简化版（旧模板用，已废弃建议用 humanizeDuration）' },
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
    alert('加载失败: ' + e.message)
  } finally { loading.value = false }
}

function openForm(t) {
  editing.value = t || null
  form.value = t ? { ...t } : defaultForm()
  previewContent.value = ''
  previewError.value = ''
  refTab.value = 'vars'
  showForm.value = true
}

function close() { showForm.value = false; editing.value = null }

async function save() {
  if (!form.value.name.trim() || !form.value.content.trim()) {
    return alert('请填写模板名称和内容')
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
    alert('保存失败: ' + e.message)
  } finally { saving.value = false }
}

async function remove(t) {
  if (!confirm(`确认删除模板「${t.name}」？`)) return
  try { await notifyApi.deleteTemplate(t.id); await load() } catch (e) { alert('删除失败: ' + e.message) }
}

function typeLabel(v) { return { all: '通用', firing: '触发', recovered: '恢复' }[v] || v }
function mediaLabel(v) { return { dingtalk: '钉钉', webhook: 'Webhook', email: '邮件', wecom: '企微' }[v] || v }
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
  font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 11.5px; color: var(--c-text-2);
  white-space: pre-wrap; line-height: 1.6; overflow: hidden; max-height: 110px;
}
.ci-actions { display: flex; gap: 6px; }

.empty { text-align: center; color: var(--c-text-3); padding: 40px 0; }

.btn-sm { padding: 3px 9px; font-size: 11.5px; border-radius: 5px; border: 1px solid var(--c-border); background: var(--c-surface); color: var(--c-text); cursor: pointer; }
.btn-sm:hover { border-color: var(--c-primary); color: var(--c-primary); }
</style>
