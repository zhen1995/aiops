<template>
  <div>
    <PageHeader title="巡检任务" desc="配置定时巡检任务，服务会按 Cron 表达式自动执行并生成报告">
      <button class="btn btn-primary" @click="openDialog()">+ 新建任务</button>
    </PageHeader>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th style="width: 44px">启用</th>
            <th>任务名称</th>
            <th style="width: 180px">Cron 表达式</th>
            <th>任务提示词</th>
            <th style="width: 170px">上次执行</th>
            <th style="width: 170px">预计下次</th>
            <th style="width: 170px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tasks" :key="t.id">
            <td>
              <label class="switch">
                <input type="checkbox" :checked="t.enabled" @change="toggleEnabled(t)" />
                <span class="slider"></span>
              </label>
            </td>
            <td><b>{{ t.name }}</b></td>
            <td><code class="cron">{{ t.cron_expr }}</code></td>
            <td class="muted prompt-cell">{{ truncate(t.prompt, 60) }}</td>
            <td class="muted">{{ fmtTime(t.last_run_at) || '-' }}</td>
            <td class="muted">{{ fmtTime(t.next_run_at) || '-' }}</td>
            <td>
              <button class="btn btn-sm" @click="openDialog(t)">编辑</button>
              <button class="btn btn-sm btn-outline" @click="runNow(t)">立即执行</button>
              <button class="btn btn-sm btn-danger-link" @click="remove(t)">删除</button>
            </td>
          </tr>
          <tr v-if="tasks.length === 0">
            <td colspan="7" class="empty">暂无巡检任务，点击右上角新建</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 任务配置弹窗 -->
    <div v-if="showDialog" class="modal-mask" @click.self="closeDialog">
      <div class="modal">
        <div class="modal-head">
          <h3>{{ editingTask ? '编辑巡检任务' : '新建巡检任务' }}</h3>
          <button class="close" @click="closeDialog">×</button>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label>任务名称</label>
            <input v-model="form.name" placeholder="例如：每日系统巡检" maxlength="128" />
          </div>

          <div class="form-item">
            <label class="row-label">
              Cron 表达式
              <span class="hint">支持 5 段（分 时 日 月 周）或 6 段（秒 分 时 日 月 周）</span>
            </label>
            <div class="cron-row">
              <input v-model="form.cron_expr" placeholder="点击右侧「配置」生成表达式" readonly />
              <button class="btn btn-sm" @click="showCronBuilder = !showCronBuilder">
                {{ showCronBuilder ? '收起' : '配置' }}
              </button>
              <button class="btn btn-sm" @click="previewCron" :disabled="!form.cron_expr">预览</button>
            </div>

            <!-- Cron 可视化选择器（内联展开） -->
            <CronBuilder
              v-if="showCronBuilder"
              v-model="form.cron_expr"
              @confirm="onCronConfirm"
              @cancel="showCronBuilder = false"
            />

            <div v-if="cronPreview.length > 0" class="cron-preview">
              <div class="cron-preview-title">预计下次执行时间：</div>
              <ul>
                <li v-for="(t, i) in cronPreview" :key="i">{{ t }}</li>
              </ul>
            </div>
            <div v-if="cronError" class="form-error">{{ cronError }}</div>
          </div>

          <div class="form-item">
            <label>任务提示词</label>
            <textarea
              v-model="form.prompt"
              rows="6"
              placeholder="描述你希望巡检任务分析的内容，例如：检查过去 24 小时内的异常告警、错误日志和服务健康状态..."
            />
          </div>

          <div class="form-item">
            <label class="row-label">
              通知媒介
              <span class="hint">报告生成后自动推送，可多选</span>
            </label>
            <div v-if="mediaOptions.length === 0" class="media-empty">
              暂无启用的通知媒介，请先在「通知管理 → 通知媒介」中创建
            </div>
            <div v-else class="media-options">
              <label v-for="m in mediaOptions" :key="m.id" class="media-option">
                <input type="checkbox" :value="m.id" v-model="form.notify_media_ids" />
                <span>{{ m.name }}</span>
                <em class="media-type">{{ m.type === 'dingtalk' ? '钉钉机器人' : 'Webhook' }}</em>
              </label>
            </div>
          </div>

          <div class="form-item switch-item">
            <label>启用任务</label>
            <label class="switch">
              <input type="checkbox" v-model="form.enabled" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="closeDialog">取消</button>
          <button class="btn btn-primary" :disabled="submitting" @click="save">
            {{ submitting ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import CronBuilder from '../components/CronBuilder.vue'
import {
  listTasks, createTask, updateTask, deleteTask,
  toggleTask, triggerTask, previewCron as apiPreviewCron
} from '../api/inspection.js'
import { notifyMediaApi } from '../api/notifyMedia.js'

const tasks = ref([])
const mediaOptions = ref([])
const showDialog = ref(false)
const editingTask = ref(null)
const submitting = ref(false)
const cronPreview = ref([])
const cronError = ref('')
const showCronBuilder = ref(false)

const defaultForm = () => ({
  name: '',
  cron_expr: '0 0 9 * * ?',
  prompt: '',
  notify_media_ids: [],
  enabled: true
})
const form = ref(defaultForm())

onMounted(() => {
  load()
  loadMediaOptions()
})

async function loadMediaOptions() {
  try {
    const data = await notifyMediaApi.list()
    mediaOptions.value = (data || []).filter(m => m.is_enabled === 1)
  } catch (e) {
    mediaOptions.value = []
  }
}

async function load() {
  try {
    tasks.value = await listTasks()
  } catch (e) {
    alert('加载任务失败: ' + e.message)
  }
}

function openDialog(task) {
  editingTask.value = task || null
  if (task) {
    let mediaIds = []
    try {
      mediaIds = JSON.parse(task.notify_media_ids || '[]')
    } catch (e) {
      mediaIds = []
    }
    form.value = { name: task.name, cron_expr: task.cron_expr, prompt: task.prompt, notify_media_ids: mediaIds, enabled: task.enabled }
  } else {
    form.value = defaultForm()
  }
  cronPreview.value = []
  cronError.value = ''
  showCronBuilder.value = false
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editingTask.value = null
  showCronBuilder.value = false
  cronPreview.value = []
  cronError.value = ''
}

function onCronConfirm() {
  showCronBuilder.value = false
  // 自动更新预览
  previewCron()
}

async function previewCron() {
  cronError.value = ''
  try {
    cronPreview.value = await apiPreviewCron(form.value.cron_expr, 5)
  } catch (e) {
    cronError.value = e.message
    cronPreview.value = []
  }
}

async function save() {
  if (!form.value.name.trim()) return alert('请填写任务名称')
  if (!form.value.cron_expr.trim()) return alert('请填写 Cron 表达式')
  if (!form.value.prompt.trim()) return alert('请填写任务提示词')

  submitting.value = true
  try {
    const payload = {
      name: form.value.name,
      cron_expr: form.value.cron_expr,
      prompt: form.value.prompt,
      notify_media_ids: JSON.stringify(form.value.notify_media_ids || []),
      enabled: form.value.enabled
    }

    if (editingTask.value) {
      await updateTask(editingTask.value.id, payload)
    } else {
      await createTask(payload)
    }
    closeDialog()
    await load()
  } catch (e) {
    alert('保存失败: ' + e.message)
  } finally {
    submitting.value = false
  }
}

async function toggleEnabled(task) {
  try {
    await toggleTask(task.id)
    await load()
  } catch (e) {
    alert('操作失败: ' + e.message)
  }
}

async function runNow(task) {
  if (!confirm(`立即执行巡检任务 "${task.name}"？（调用 LLM 可能需要几十秒）`)) return
  try {
    await triggerTask(task.id)
    alert('执行完成，已生成巡检报告')
    await load()
  } catch (e) {
    alert('执行失败: ' + e.message)
  }
}

async function remove(task) {
  if (!confirm(`确认删除巡检任务 "${task.name}"？`)) return
  try {
    await deleteTask(task.id)
    await load()
  } catch (e) {
    alert('删除失败: ' + e.message)
  }
}

function truncate(s, n) {
  if (!s) return ''
  return s.length > n ? s.slice(0, n) + '…' : s
}

function fmtTime(v) {
  if (!v) return ''
  const d = new Date(v)
  if (isNaN(d.getTime())) return v
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.prompt-cell { max-width: 300px; }

.cron {
  background: var(--c-bg);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12.5px;
}

/* Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 38px;
  height: 20px;
  cursor: pointer;
}
.switch input { display: none; }
.switch .slider {
  position: absolute;
  inset: 0;
  background: #ccc;
  border-radius: 20px;
  transition: 0.2s;
}
.switch .slider::before {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  left: 2px;
  top: 2px;
  background: #fff;
  border-radius: 50%;
  transition: 0.2s;
}
.switch input:checked + .slider { background: var(--c-primary); }
.switch input:checked + .slider::before { transform: translateX(18px); }

.btn-danger-link {
  background: transparent;
  color: var(--c-danger, #c93b3b);
  border: none;
  padding: 4px 8px;
  font-size: 12px;
  cursor: pointer;
}
.btn-danger-link:hover { text-decoration: underline; }

.btn-outline {
  background: transparent;
  border: 1px solid var(--c-border);
}

/* Modal */
.modal-mask {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}
.modal {
  background: var(--c-surface);
  border-radius: 12px;
  width: 560px;
  max-height: 85vh;
  display: flex; flex-direction: column;
  box-shadow: 0 12px 40px rgba(0,0,0,0.2);
}
.modal-head {
  padding: 18px 22px;
  border-bottom: 1px solid var(--c-border);
  display: flex; justify-content: space-between; align-items: center;
}
.modal-head h3 { margin: 0; font-size: 16px; }
.modal-head .close {
  border: none; background: transparent;
  font-size: 22px; color: var(--c-text-3); cursor: pointer;
}
.modal-body {
  padding: 20px 22px;
  overflow-y: auto;
  flex: 1;
}
.modal-foot {
  padding: 14px 22px;
  border-top: 1px solid var(--c-border);
  display: flex; justify-content: flex-end; gap: 10px;
}

.form-item { margin-bottom: 18px; }
.form-item label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--c-text-2);
  margin-bottom: 8px;
}
.row-label { display: flex; align-items: center; justify-content: space-between; }
.hint {
  font-size: 11px;
  color: var(--c-text-3);
  font-weight: normal;
}
.switch-item { display: flex; align-items: center; justify-content: space-between; }
.switch-item label { margin-bottom: 0; }

.form-item input,
.form-item textarea {
  width: 100%;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13.5px;
  background: var(--c-bg);
  color: var(--c-text);
  outline: none;
  box-sizing: border-box;
  font-family: inherit;
}
.form-item input:focus,
.form-item textarea:focus { border-color: var(--c-primary); }

.cron-row {
  display: flex; gap: 10px;
}
.cron-row input { flex: 1; }

.cron-preview {
  margin-top: 10px;
  padding: 10px 12px;
  background: var(--c-primary-soft);
  border-radius: 6px;
  font-size: 12px;
}
.cron-preview-title { color: var(--c-text-2); margin-bottom: 4px; }
.cron-preview ul { margin: 0; padding-left: 18px; }
.cron-preview li { color: var(--c-text); font-family: monospace; }

.form-error {
  margin-top: 8px;
  color: var(--c-danger, #c93b3b);
  font-size: 12px;
}

.media-empty {
  padding: 10px 12px;
  background: var(--c-bg);
  border-radius: 8px;
  font-size: 12.5px;
  color: var(--c-text-3);
}

.media-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.media-option {
  display: flex !important;
  align-items: center;
  gap: 6px;
  padding: 7px 12px;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px !important;
  font-weight: normal !important;
  color: var(--c-text) !important;
  margin-bottom: 0 !important;
}

.media-option:hover { border-color: var(--c-primary); }

.media-option input {
  width: auto !important;
  margin: 0;
  accent-color: var(--c-primary);
}

.media-type {
  font-style: normal;
  font-size: 11px;
  color: var(--c-text-3);
}

.table .empty {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 20px;
}
</style>
