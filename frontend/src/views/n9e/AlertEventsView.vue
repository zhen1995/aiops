<template>
  <div>
    <PageHeader title="夜莺告警事件" desc="活跃告警与历史告警事件查询">
      <button class="btn btn-sm" @click="load" :disabled="loading">刷新</button>
    </PageHeader>

    <div class="card-filter">
      <div class="filter-tabs">
        <button :class="['tab', { active: tab === 'cur' }]" @click="switchTab('cur')">
          活跃告警
          <span v-if="curCount > 0" class="tab-badge">{{ curCount }}</span>
        </button>
        <button :class="['tab', { active: tab === 'his' }]" @click="switchTab('his')">历史告警</button>
      </div>

      <template v-if="tab === 'his'">
        <div class="filter-row">
          <label>起始时间</label>
          <input type="datetime-local" v-model="hisStart" />
          <label>结束时间</label>
          <input type="datetime-local" v-model="hisEnd" />
          <button class="btn btn-sm btn-primary" @click="loadHis">查询</button>
          <button class="btn btn-sm" @click="presetRange(7)">近 7 天</button>
          <button class="btn btn-sm" @click="presetRange(30)">近 30 天</button>
        </div>
      </template>

      <div v-if="message" class="banner error" style="margin-top:10px">{{ message }}</div>
    </div>

    <!-- 活跃告警 -->
    <div class="card" v-if="tab === 'cur' && curEvents.length > 0">
      <table class="table">
        <thead>
          <tr>
            <th style="width:90px">级别</th>
            <th style="width:160px">告警名称</th>
            <th>标签</th>
            <th style="width:180px">触发时间</th>
            <th style="width:120px">业务组</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(e, i) in curEvents" :key="keyOf(e, i)">
            <td><span class="level level-cur">{{ severityOf(e) }}</span></td>
            <td><b>{{ ruleNameOf(e) }}</b></td>
            <td class="tags-cell">
              <span v-for="(v, k) in tagsOf(e)" :key="k" class="tag-chip">{{ k }}={{ v }}</span>
            </td>
            <td class="muted">{{ fmtTime(e.trigger_time || e.triggerTime || e.first_trigger_time) }}</td>
            <td class="muted">{{ groupOf(e) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else-if="tab === 'cur' && !loading" class="empty-card">
      <p class="empty-title">🎉 当前没有活跃告警</p>
      <p class="empty-sub">夜莺引擎运行正常</p>
    </div>

    <!-- 历史告警 -->
    <div class="card" v-if="tab === 'his' && hisEvents.length > 0">
      <table class="table">
        <thead>
          <tr>
            <th style="width:90px">级别</th>
            <th style="width:160px">告警名称</th>
            <th>标签</th>
            <th style="width:170px">触发时间</th>
            <th style="width:170px">恢复时间</th>
            <th style="width:80px">持续</th>
            <th style="width:120px">业务组</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(e, i) in hisEvents" :key="keyOf(e, i)" :class="{ closed: !isActive(e) }">
            <td><span class="level" :class="isActive(e) ? 'level-cur' : 'level-resolved'">{{ severityOf(e) }}</span></td>
            <td><b>{{ ruleNameOf(e) }}</b></td>
            <td class="tags-cell">
              <span v-for="(v, k) in tagsOf(e)" :key="k" class="tag-chip">{{ k }}={{ v }}</span>
            </td>
            <td class="muted">{{ fmtTime(e.trigger_time || e.triggerTime || e.first_trigger_time) }}</td>
            <td class="muted">{{ fmtTime(e.recover_time || e.recoverTime || e.last_trigger_time) }}</td>
            <td class="muted">{{ durationOf(e) }}</td>
            <td class="muted">{{ groupOf(e) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else-if="tab === 'his' && !loading" class="empty-card">
      <p class="empty-title">该时间段内无历史告警</p>
      <p class="empty-sub">调整时间范围或刷新重试</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import PageHeader from '../../components/PageHeader.vue'
import { n9eApi } from '../../api/n9e.js'

const tab = ref('cur')
const loading = ref(false)
const curEvents = ref([])
const hisEvents = ref([])
const message = ref('')
const curCount = computed(() => curEvents.value.length)

// 业务组 id → name 映射
const groupMap = reactive({})

// 历史时间范围（默认近 30 天）
const now = new Date()
const thirty = new Date(now.getTime() - 30 * 24 * 3600 * 1000)
const fmtInput = (d) => d.toISOString().slice(0, 16)
const hisStart = ref(fmtInput(thirty))
const hisEnd = ref(fmtInput(now))

onMounted(async () => {
  await Promise.all([loadGroups(), load()])
})

async function loadGroups() {
  try {
    const raw = await n9eApi.getBusiGroups()
    const list = extractList(raw)
    for (const g of list) {
      if (g.id !== undefined && g.name !== undefined) groupMap[g.id] = g.name
    }
  } catch (e) {
    console.warn('加载业务组失败:', e.message)
  }
}

async function load() {
  message.value = ''
  if (tab.value === 'cur') {
    await loadCur()
  } else {
    await loadHis()
  }
}

function switchTab(t) {
  tab.value = t
  load()
}

async function loadCur() {
  loading.value = true
  message.value = ''
  try {
    const raw = await n9eApi.getCurEvents({ p: 1, limit: 50, my_groups: 'true' })
    curEvents.value = extractList(raw)
  } catch (e) {
    message.value = e.message
    curEvents.value = []
  } finally {
    loading.value = false
  }
}

async function loadHis() {
  loading.value = true
  message.value = ''
  try {
    const raw = await n9eApi.getHisEvents({
      p: 1, limit: 50,
      stime: toUnix(hisStart.value),
      etime: toUnix(hisEnd.value),
    })
    hisEvents.value = extractList(raw)
  } catch (e) {
    message.value = e.message
    hisEvents.value = []
  } finally {
    loading.value = false
  }
}

function presetRange(days) {
  const end = new Date()
  const start = new Date(end.getTime() - days * 24 * 3600 * 1000)
  hisStart.value = fmtInput(start)
  hisEnd.value = fmtInput(end)
  loadHis()
}

function extractList(raw) {
  if (Array.isArray(raw)) return raw
  if (raw && typeof raw === 'object') {
    // 夜莺历史/活跃事件返回 {"dat": {"list": [...]}}
    for (const topKey of ['data', 'dat', 'list', 'items']) {
      const v = raw[topKey]
      if (Array.isArray(v)) return v
      if (v && typeof v === 'object') {
        // 嵌套一层找 list
        for (const subKey of ['list', 'data', 'items', 'records']) {
          const sv = v[subKey]
          if (Array.isArray(sv)) return sv
        }
      }
    }
    // 最后兜底：遍历所有值找第一个数组
    for (const v of Object.values(raw)) {
      if (Array.isArray(v)) return v
      if (v && typeof v === 'object') {
        for (const sv of Object.values(v)) {
          if (Array.isArray(sv)) return sv
        }
      }
    }
  }
  return []
}

function toUnix(inputVal) {
  if (!inputVal) return ''
  const d = new Date(inputVal)
  if (isNaN(d.getTime())) return ''
  return Math.floor(d.getTime() / 1000)
}

function keyOf(e, i) {
  return e.id ?? e.event_id ?? e.group_key ?? e.groupKey ?? e.ident ?? i
}

function severityOf(e) {
  const lv = e.severity
  if (lv === undefined || lv === null) return '-'
  if (typeof lv === 'number') {
    return { 0: 'P3', 1: 'P2', 2: 'P1' }[lv] ?? `P${lv}`
  }
  return String(lv)
}

function isActive(e) {
  // 夜莺历史事件中：status 或 is_recovered 或 recovered
  if (e.status !== undefined) return e.status === 0 || e.status === 'ok' || e.status === 'firing'
  if (e.is_recovered !== undefined) return !e.is_recovered
  if (e.recovered !== undefined) return !e.recovered
  return false
}

function ruleNameOf(e) { return e.rule_name || e.ruleName || e.rname || '-' }

function groupOf(e) {
  const gid = e.group_id ?? e.busi_group_id ?? e.groupId
  if (gid !== undefined && groupMap[gid]) return groupMap[gid]
  return e.busi_group_name ?? e.busi_group ?? e.group_name ?? '-'
}

function tagsOf(e) {
  const raw = e.tags || e.labels || e.rule_labels
  if (Array.isArray(raw)) {
    // 事件里 tags 是 ["__name__=xxx", "alertName=yyy", ...] 数组格式
    const obj = {}
    for (const t of raw) {
      if (typeof t === 'string') {
        const idx = t.indexOf('=')
        if (idx > 0) obj[t.slice(0, idx)] = t.slice(idx + 1)
      }
    }
    return obj
  }
  if (raw && typeof raw === 'object') return raw
  return {}
}

function fmtTime(v) {
  if (!v) return '-'
  // 夜莺 timestamp 可能是秒级 unix
  if (typeof v === 'number' && v > 1e9 && v < 1e11) {
    const d = new Date(v * 1000)
    return d.toLocaleString('zh-CN')
  }
  // 毫秒
  if (typeof v === 'number' && v > 1e12) {
    return new Date(v).toLocaleString('zh-CN')
  }
  // ISO 字符串
  const d = new Date(v)
  if (!isNaN(d.getTime())) return d.toLocaleString('zh-CN')
  return String(v)
}

function durationOf(e) {
  const start = e.trigger_time || e.first_trigger_time || e.triggerTime
  const end = e.recover_time || e.last_trigger_time || e.recoverTime || Date.now() / 1000
  if (!start) return '-'
  const startSec = typeof start === 'number' && start > 1e12 ? start / 1000 : start
  const endSec = typeof end === 'number' && end > 1e12 ? end / 1000 : end
  const secs = Math.max(0, Number(endSec) - Number(startSec))
  if (secs < 60) return secs + '秒'
  if (secs < 3600) return Math.round(secs / 60) + '分'
  if (secs < 86400) return (secs / 3600).toFixed(1) + '时'
  return (secs / 86400).toFixed(1) + '天'
}
</script>

<style scoped>
.card-filter {
  padding: 14px 18px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 10px;
  margin-bottom: 14px;
}

.filter-tabs { display: flex; gap: 8px; }
.tab {
  padding: 6px 18px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 13px;
  cursor: pointer;
  display: flex; align-items: center; gap: 6px;
}
.tab:hover { color: var(--c-primary); }
.tab.active { background: var(--c-primary); border-color: var(--c-primary); color: #fff; }
.tab-badge {
  background: var(--c-danger); color: #fff; border-radius: 10px;
  padding: 0 7px; font-size: 11px; min-width: 18px; text-align: center;
}

.filter-row {
  display: flex; align-items: center; gap: 10px;
  margin-top: 12px; flex-wrap: wrap;
}
.filter-row label { font-size: 12.5px; color: var(--c-text-3); }
.filter-row input[type="datetime-local"] {
  padding: 5px 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  font-size: 12.5px;
  color: var(--c-text);
}

.level { padding: 2px 10px; border-radius: 10px; font-size: 11.5px; font-weight: 600; }
.level-cur { background: var(--c-p0-bg, #ffe5e5); color: var(--c-danger, #c93b3b); }
.level-resolved { background: #e8f5ee; color: #1f7a45; }

.tags-cell { max-width: 340px; }
.tag-chip {
  display: inline-block;
  padding: 1px 7px;
  margin: 1px 3px 1px 0;
  border-radius: 9px;
  background: var(--c-primary-soft, rgba(22,119,255,0.08));
  color: var(--c-primary);
  font-size: 11px;
  font-family: ui-monospace, Menlo, Monaco, Consolas, monospace;
}

.muted { color: var(--c-text-3); font-size: 12.5px; }
.banner.error {
  padding: 10px 16px; border-radius: 8px; font-size: 13px;
  background: var(--c-p0-bg, #ffe5e5); color: var(--c-danger, #c93b3b);
}

.empty-card {
  padding: 60px 20px; text-align: center; color: var(--c-text-3);
  border: 1px dashed var(--c-border); border-radius: 10px;
}
.empty-title { font-size: 15px; font-weight: 500; color: var(--c-text-2); margin-bottom: 6px; }
.empty-sub { font-size: 12.5px; }

.table th { font-size: 12px; color: var(--c-text-3); font-weight: 500; text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--c-border); }
.table td { padding: 10px 12px; font-size: 13px; border-bottom: 1px solid var(--c-border); }
.table tr:last-child td { border-bottom: none; }
.table tr.closed td { color: var(--c-text-3); }
</style>
