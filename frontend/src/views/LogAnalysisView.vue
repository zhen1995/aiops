<template>
  <div>
    <PageHeader :title="$t('log.header.title')" :desc="$t('log.header.desc')">
      <div class="range-switch">
        <div class="svc-multi" ref="svcMultiRef">
          <button class="btn btn-sm svc-multi-btn" :title="svcLabel" @click="svcOpen = !svcOpen">
            {{ svcLabel }}<span class="caret">▾</span>
          </button>
          <div v-if="svcOpen" class="svc-panel">
            <div class="svc-panel-head">
              <button class="link-btn" @click="selectAllServices">{{ $t('log.header.selectAll') }}</button>
              <button class="link-btn" @click="clearServices">{{ $t('log.header.clear') }}</button>
            </div>
            <label v-for="s in serviceOptions" :key="s.id" class="svc-option">
              <input type="checkbox" :value="s.id" v-model="selectedServiceIds" />
              <span>{{ s.name }}</span>
            </label>
            <div v-if="!serviceOptions.length" class="svc-empty">{{ $t('log.common.empty') }}</div>
          </div>
        </div>
        <button
          v-for="r in rangeOptions" :key="r.value"
          class="btn btn-sm" :class="{ 'btn-primary': hours === r.value }"
          @click="hours = r.value"
        >{{ $t(r.label) }}</button>
        <button class="btn btn-primary btn-sm" :disabled="loading" @click="loadData">
          {{ loading ? $t('log.header.querying') : $t('log.header.query') }}
        </button>
      </div>
    </PageHeader>

    <!-- 日志处理流水线（静态示意） -->
    <div class="card pipeline-card">
      <h3 class="card-title">{{ $t('log.pipeline.title') }}</h3>
      <p class="card-sub">{{ $t('log.pipeline.sub') }}</p>
      <div class="pipeline">
        <template v-for="(step, i) in pipeSteps" :key="step">
          <div class="pipe-step">
            <span class="pipe-no">{{ i + 1 }}</span>
            <span class="pipe-name">{{ step }}</span>
            <span class="pipe-desc">{{ pipeDescs[i] }}</span>
          </div>
          <span v-if="i < pipeSteps.length - 1" class="pipe-arrow">→</span>
        </template>
      </div>
    </div>

    <!-- 日志量趋势 -->
    <div class="card chart-card">
      <h3 class="card-title">{{ $t('log.trend.title') }}</h3>
      <p class="card-sub">{{ $t('log.trend.sub') }}</p>
      <ChartBox v-if="hasTrend" :option="seriesOption" height="300px" />
      <div v-else class="empty-hint">{{ trendError ? $t('log.common.loadFailed') : $t('log.common.empty') }}</div>
    </div>

    <!-- 日志聚类（整行展示） -->
    <div class="card cluster-card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('log.cluster.title') }}</h3>
          <p class="card-sub">{{ $t('log.cluster.sub') }}</p>
        </div>
        <div class="cluster-head-right">
          <span class="cluster-total muted">{{ $t('log.cluster.total', { count: clustersTotal }) }}</span>
          <select v-model="clusterLevel" class="level-filter" @change="loadClusters">
            <option value="error,warn">{{ $t('log.cluster.filterAnomaly') }}</option>
            <option value="error">{{ $t('log.cluster.filterError') }}</option>
            <option value="warn">{{ $t('log.cluster.filterWarn') }}</option>
            <option value="info">{{ $t('log.cluster.filterInfo') }}</option>
            <option value="">{{ $t('log.cluster.filterAll') }}</option>
          </select>
        </div>
      </div>
      <table v-if="clusters.length" class="table">
        <thead>
          <tr>
            <th>{{ $t('log.cluster.colId') }}</th>
            <th>{{ $t('log.cluster.colPattern') }}</th>
            <th>{{ $t('log.cluster.colCount') }}</th>
            <th>{{ $t('log.cluster.colLevel') }}</th>
            <th>{{ $t('log.cluster.colServices') }}</th>
            <th>{{ $t('log.cluster.colTrend') }}</th>
            <th>{{ $t('log.cluster.colFirstSeen') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in clusters" :key="c.id">
            <td class="mono">{{ c.id }}</td>
            <td class="mono pattern-cell" :title="c.pattern">
              <span class="pattern-text">{{ c.pattern }}</span>
              <button class="copy-btn" @click="copyText(c.pattern, 'c-' + c.id)">
                {{ copiedKey === 'c-' + c.id ? $t('log.common.copied') : $t('log.common.copy') }}
              </button>
            </td>
            <td class="mono">{{ formatCount(c.count) }}</td>
            <td><LevelTag :level="c.level" /></td>
            <td><span v-for="s in c.services" :key="s" class="mono svc-badge">{{ s }}</span></td>
            <td><span class="trend-tag" :class="c.trend">{{ trendText[c.trend] || c.trend }}</span></td>
            <td class="muted">{{ formatTime(c.firstSeen) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-hint">{{ clusterError ? $t('log.common.loadFailed') : $t('log.common.empty') }}</div>
    </div>

    <!-- Drain 模板提取（整行展示，避免与聚类表并排导致文字遮挡） -->
    <div class="card">
      <h3 class="card-title">{{ $t('log.template.title') }}</h3>
      <p class="card-sub">{{ $t('log.template.sub') }}</p>
      <div v-if="templates.length" class="tpl-list">
        <div v-for="(t, i) in templates" :key="i" class="tpl-item">
          <div class="tpl-head">
            <LevelTag :level="t.level" />
            <span class="mono muted tpl-count">{{ $t('log.template.count', { count: formatCount(t.count) }) }}</span>
          </div>
          <div class="tpl-row">
            <span class="tpl-label">{{ $t('log.template.tpl') }}</span>
            <span class="mono tpl-tpl">
              <template v-for="(seg, j) in splitTemplate(t.template)" :key="j">
                <em v-if="seg.ph" class="tpl-ph">&lt;*&gt;</em>
                <span v-else>{{ seg.text }}</span>
              </template>
            </span>
            <button class="copy-btn" @click="copyText(t.template, 't-' + i)">
              {{ copiedKey === 't-' + i ? $t('log.common.copied') : $t('log.common.copy') }}
            </button>
          </div>
          <div v-for="(sample, k) in t.paramSamples" :key="k" class="tpl-sample">
            <div class="tpl-row">
              <span class="tpl-label">{{ $t('log.template.raw') }}</span>
              <span class="mono tpl-raw">{{ sample.raw }}</span>
            </div>
            <div class="tpl-row">
              <span class="tpl-label">{{ $t('log.template.params') }}</span>
              <span class="tpl-params">
                <span v-for="(p, j) in sample.params" :key="j" class="mono param-badge">{{ p }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-hint">{{ templateError ? $t('log.common.loadFailed') : $t('log.common.empty') }}</div>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { logApi } from '../api/log.js'
import { serviceApi } from '../api/service.js'

const { t } = useI18n({ useScope: 'global' })

const rangeOptions = [
  { value: 1, label: 'log.header.last1h' },
  { value: 6, label: 'log.header.last6h' },
  { value: 12, label: 'log.header.last12h' },
  { value: 24, label: 'log.header.last24h' }
]

const hours = ref(24)
// 聚类表默认聚焦异常级别（ERROR/WARN），可在卡片右上筛选切换
const clusterLevel = ref('error,warn')
// 服务多选过滤（来源：服务注册），空数组表示全部服务
const serviceOptions = ref([])
const selectedServiceIds = ref([])
const svcOpen = ref(false)
const svcMultiRef = ref(null)

const serviceParam = computed(() => selectedServiceIds.value.join(','))

const svcLabel = computed(() => {
  if (!selectedServiceIds.value.length) return t('log.header.serviceAll')
  if (selectedServiceIds.value.length === 1) {
    const hit = serviceOptions.value.find((s) => s.id === selectedServiceIds.value[0])
    return hit ? hit.name : t('log.header.serviceSelected', { count: 1 })
  }
  return t('log.header.serviceSelected', { count: selectedServiceIds.value.length })
})

const selectAllServices = () => {
  selectedServiceIds.value = serviceOptions.value.map((s) => s.id)
}

const clearServices = () => {
  selectedServiceIds.value = []
}

const onDocClick = (e) => {
  if (svcOpen.value && svcMultiRef.value && !svcMultiRef.value.contains(e.target)) {
    svcOpen.value = false
  }
}

const loadServiceOptions = async () => {
  try {
    serviceOptions.value = await serviceApi.options() || []
  } catch (e) {
    serviceOptions.value = []
  }
}
const trend = ref({ labels: [], total: [], error: [] })
const clusters = ref([])
const clustersTotal = ref(0)
const templates = ref([])
const trendError = ref(false)
const clusterError = ref(false)
const templateError = ref(false)
const loading = ref(false)
const copiedKey = ref('')

let copyTimer = null

const pipeSteps = computed(() => [
  t('log.pipeline.step1'),
  t('log.pipeline.step2'),
  t('log.pipeline.step3'),
  t('log.pipeline.step4'),
  t('log.pipeline.step5'),
  t('log.pipeline.step6')
])

const pipeDescs = computed(() => [
  t('log.pipeline.desc1'),
  t('log.pipeline.desc2'),
  t('log.pipeline.desc3'),
  t('log.pipeline.desc4'),
  t('log.pipeline.desc5'),
  t('log.pipeline.desc6')
])

const trendText = computed(() => ({
  spike: t('log.cluster.trendSpike'),
  rising: t('log.cluster.trendRising'),
  flat: t('log.cluster.trendFlat')
}))

const hasTrend = computed(() => (trend.value.labels || []).length > 0)

const formatCount = (n) => (typeof n === 'number' ? n.toLocaleString() : (n ?? 0))

// ISO 时间格式化为 HH:mm（与原 Mock 视觉一致）
// 完整展示年月日时分秒（本地时区）
const formatTime = (iso) => {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso || '-'
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

// 将模板字符串按 <*> 占位符切分，用于高亮渲染
const splitTemplate = (tpl) => {
  const parts = (tpl || '').split('<*>')
  const out = []
  parts.forEach((text, i) => {
    if (text) out.push({ text })
    if (i < parts.length - 1) out.push({ ph: true })
  })
  return out
}

// 复制文本到剪贴板，并短暂显示「已复制」反馈（参照 ChatView 的复制实现）
const copyText = async (text, key) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
  } catch (e) {
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
  copiedKey.value = key
  if (copyTimer) clearTimeout(copyTimer)
  copyTimer = setTimeout(() => {
    copiedKey.value = ''
  }, 1500)
}

const loadTrend = async () => {
  trendError.value = false
  try {
    trend.value = await logApi.getTrend({ hours: hours.value, service: serviceParam.value })
  } catch (e) {
    trend.value = { labels: [], total: [], error: [] }
    trendError.value = true
  }
}

const loadClusters = async () => {
  clusterError.value = false
  try {
    const data = await logApi.getClusters({ hours: hours.value, service: serviceParam.value, level: clusterLevel.value, page: 1, page_size: 20 })
    clusters.value = data.items || []
    clustersTotal.value = data.total ?? clusters.value.length
  } catch (e) {
    clusters.value = []
    clustersTotal.value = 0
    clusterError.value = true
  }
}

const loadTemplates = async () => {
  templateError.value = false
  try {
    const data = await logApi.getTemplates({ service: serviceParam.value, limit: 20 })
    const items = (data.items || []).slice()
    items.sort((a, b) => (b.count || 0) - (a.count || 0))
    templates.value = items
  } catch (e) {
    templates.value = []
    templateError.value = true
  }
}

// 点击「查询」按钮才刷新数据；时间按钮仅切换范围，避免每次切换都触发请求
const loadData = async () => {
  loading.value = true
  try {
    await Promise.all([loadTrend(), loadClusters(), loadTemplates()])
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
  loadServiceOptions()
  document.addEventListener('click', onDocClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
})

const seriesOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: [t('log.chart.total'), t('log.chart.error')], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 60, right: 16, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: trend.value.labels, axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    {
      name: t('log.chart.total'), type: 'line', smooth: true, data: trend.value.total, showSymbol: false,
      lineStyle: { color: '#0e7c72', width: 2.5 }, itemStyle: { color: '#0e7c72' },
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(14,124,114,0.14)' }, { offset: 1, color: 'rgba(14,124,114,0)' }] } }
    },
    { name: t('log.chart.error'), type: 'line', smooth: true, data: trend.value.error, showSymbol: false, lineStyle: { color: '#c93b3b', width: 2 }, itemStyle: { color: '#c93b3b' } }
  ]
}))
</script>

<style scoped>
.range-switch { display: flex; gap: 8px; align-items: center; }
.svc-multi { position: relative; }
.svc-multi-btn { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.svc-multi-btn .caret { margin-left: 4px; font-size: 11px; }
.svc-panel {
  position: absolute;
  right: 0;
  top: calc(100% + 4px);
  min-width: 180px;
  max-height: 260px;
  overflow-y: auto;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius-card);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  padding: 6px;
  z-index: 30;
}
.svc-panel-head { display: flex; justify-content: space-between; padding: 2px 6px 6px; border-bottom: 1px solid var(--c-border); margin-bottom: 4px; }
.link-btn { border: none; background: none; color: var(--c-primary); font-size: 12px; cursor: pointer; padding: 2px 4px; font-family: inherit; }
.link-btn:hover { color: var(--c-primary-dark); }
.svc-option { display: flex; align-items: center; gap: 8px; padding: 5px 6px; border-radius: 6px; font-size: 13px; color: var(--c-text-2); cursor: pointer; }
.svc-option:hover { background: var(--c-primary-soft); }
.svc-option input { accent-color: var(--c-primary); }
.svc-empty { padding: 12px; text-align: center; color: var(--c-text-3); font-size: 12px; }
.pipeline-card { margin-bottom: 16px; }
.pipeline { display: flex; align-items: stretch; gap: 6px; }
.pipe-step {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
  padding: 12px 8px;
  border-radius: var(--radius-card);
  background: var(--c-primary-soft);
  border: 1px solid var(--c-border);
}
.pipe-no {
  width: 22px; height: 22px; border-radius: 50%;
  background: var(--c-primary); color: #fff;
  font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.pipe-name { font-size: 13px; font-weight: 600; color: var(--c-primary-dark); }
.pipe-desc { font-size: 11px; color: var(--c-text-3); line-height: 1.4; }
.pipe-arrow { align-self: center; color: var(--c-primary); font-weight: 700; flex-shrink: 0; }
.chart-card { margin-bottom: 16px; }
.cluster-card { margin-bottom: 16px; }
.card-head-flex { display: flex; justify-content: space-between; align-items: flex-start; }
.cluster-head-right { display: flex; align-items: center; gap: 12px; }
.cluster-total { white-space: nowrap; }
.level-filter {
  padding: 4px 8px;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-tag);
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 12.5px;
  font-family: inherit;
  cursor: pointer;
}
.pattern-cell { max-width: 480px; }
.pattern-text { display: inline-block; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: middle; }
.copy-btn {
  flex-shrink: 0;
  margin-left: 6px;
  padding: 1px 8px;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-tag);
  background: var(--c-surface);
  color: var(--c-primary-dark);
  font-size: 12px;
  cursor: pointer;
  vertical-align: middle;
}
.copy-btn:hover { background: var(--c-primary-tint); }
.svc-badge { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); background: var(--c-primary-tint); color: var(--c-primary-dark); }
.trend-tag { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); font-size: 12px; font-weight: 600; white-space: nowrap; }
.trend-tag.spike { color: var(--c-p0); background: var(--c-p0-bg); }
.trend-tag.rising { color: var(--c-p1); background: var(--c-p1-bg); }
.trend-tag.flat { color: var(--c-text-3); background: var(--c-p4-bg); }
.empty-hint { padding: 32px 0; text-align: center; color: var(--c-text-3); font-size: 13px; }
.tpl-list { display: grid; grid-template-columns: 1fr; gap: 12px; }
.tpl-item { border: 1px solid var(--c-border); border-radius: var(--radius-card); padding: 12px 14px; display: flex; flex-direction: column; gap: 8px; background: var(--c-primary-soft); }
.tpl-head { display: flex; align-items: center; justify-content: space-between; }
.tpl-count { font-size: 12px; }
.tpl-sample { display: flex; flex-direction: column; gap: 8px; border-top: 1px dashed var(--c-border); padding-top: 8px; }
.tpl-row { display: flex; align-items: flex-start; gap: 10px; min-width: 0; }
.tpl-label { flex-shrink: 0; width: 56px; font-size: 12px; color: var(--c-text-3); padding-top: 2px; }
.tpl-raw { color: var(--c-text-2); word-break: break-all; overflow-wrap: anywhere; min-width: 0; }
.tpl-tpl { color: var(--c-text); word-break: break-all; overflow-wrap: anywhere; min-width: 0; }
.tpl-ph { font-style: normal; color: var(--c-primary); font-weight: 700; background: var(--c-primary-tint); padding: 0 3px; border-radius: 4px; }
.tpl-params { display: flex; flex-wrap: wrap; gap: 6px; min-width: 0; }
.param-badge { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); background: var(--c-surface); border: 1px solid var(--c-border); color: var(--c-primary-dark); word-break: break-all; overflow-wrap: anywhere; max-width: 100%; }
@media (max-width: 1200px) {
  .tpl-list { grid-template-columns: 1fr; }
  .pipeline { flex-wrap: wrap; }
  .pipe-step { flex: 1 1 28%; }
  .pipe-arrow { display: none; }
}
</style>
