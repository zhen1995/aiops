<template>
  <div>
    <PageHeader :title="$t('dashboard.header.title')" :desc="$t('dashboard.header.desc')">
      <div class="range-switch">
        <button class="btn btn-sm" :class="{ 'btn-primary': rangeMode === '24h' }" @click="switchMode('24h')">
          {{ $t('dashboard.header.range24h') }}
        </button>
        <button class="btn btn-sm" :class="{ 'btn-primary': rangeMode === 'custom' }" @click="switchMode('custom')">
          {{ $t('dashboard.header.rangeCustom') }}
        </button>
      </div>
      <div v-if="rangeMode === 'custom'" class="custom-range">
        <input v-model="customStart" type="datetime-local" />
        <span>{{ $t('dashboard.header.to') }}</span>
        <input v-model="customEnd" type="datetime-local" />
        <button class="btn btn-sm btn-primary" :disabled="loading || !customStart || !customEnd" @click="applyCustom">
          {{ $t('dashboard.header.confirm') }}
        </button>
      </div>
    </PageHeader>

    <div v-if="error" class="card error-bar">
      <span>{{ error }}</span>
      <button class="btn btn-sm" @click="loadData">{{ $t('dashboard.header.retry') }}</button>
    </div>

    <div class="kpi-grid">
      <StatCard v-for="k in kpiStats" :key="k.label" v-bind="k" />
    </div>

    <div class="row-2">
      <div class="card">
        <h3 class="card-title">{{ $t('dashboard.trend.title') }}</h3>
        <p class="card-sub">{{ $t('dashboard.trend.sub', { rate: trendRate }) }}</p>
        <ChartBox v-if="hasTrend" :option="trendOption" height="280px" />
        <div v-else class="chart-empty">{{ $t('dashboard.header.noData') }}</div>
      </div>
      <div class="card">
        <h3 class="card-title">{{ $t('dashboard.severity.title') }}</h3>
        <p class="card-sub">{{ $t('dashboard.severity.sub') }}</p>
        <ChartBox v-if="severityDist.length" :option="pieOption" height="280px" />
        <div v-else class="chart-empty">{{ $t('dashboard.header.noData') }}</div>
      </div>
    </div>

    <div class="row-2b">
      <div class="card">
        <h3 class="card-title">{{ $t('dashboard.health.title') }}</h3>
        <p class="card-sub">{{ $t('dashboard.health.sub') }}</p>
        <div v-if="serviceHealth.length" class="health-list">
          <div v-for="s in serviceHealth" :key="s.service" class="health-item">
            <div class="health-head">
              <span class="mono">{{ s.service }}</span>
              <span class="health-score" :class="healthStatus(s.score)">{{ s.score }}</span>
            </div>
            <div class="progress"><i :style="{ width: s.score + '%', background: barColor(s.score) }"></i></div>
            <div class="health-foot muted">
              <span>{{ $t('dashboard.health.abnormalMetrics', { count: s.abnormal_metrics }) }}</span>
              <span>{{ trendText(s.trend) }}</span>
            </div>
          </div>
        </div>
        <div v-else class="chart-empty">{{ $t('dashboard.header.noData') }}</div>
      </div>

      <div class="card">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">{{ $t('dashboard.alerts.title') }}</h3>
            <p class="card-sub">{{ $t('dashboard.alerts.sub') }}</p>
          </div>
          <button class="btn btn-sm" @click="goAlertEvents">{{ $t('dashboard.alerts.view') }}</button>
        </div>
        <table v-if="latestAlerts.length" class="table">
          <thead>
            <tr>
              <th>{{ $t('dashboard.alerts.colLevel') }}</th>
              <th>{{ $t('dashboard.alerts.colContent') }}</th>
              <th>{{ $t('dashboard.alerts.colService') }}</th>
              <th>{{ $t('dashboard.alerts.colTime') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in latestAlerts" :key="a.id">
              <td>
                <LevelTag :level="severityTag(a.severity)">{{ severityText[a.severity] || ('P' + a.severity) }}</LevelTag>
              </td>
              <td class="alert-title">{{ a.content }}</td>
              <td class="mono muted">{{ a.service }}</td>
              <td class="muted">{{ a.time }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="chart-empty">{{ $t('dashboard.header.noData') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { dashboardApi } from '../api/dashboard.js'

const router = useRouter()
const { t } = useI18n({ useScope: 'global' })

const rangeMode = ref('24h')
const customStart = ref('')
const customEnd = ref('')
const loading = ref(false)
const error = ref('')
const kpi = ref({})
const trend = ref({ buckets: [], raw: [], denoised: [] })
const severityDist = ref([])
const serviceHealth = ref([])
const latestAlerts = ref([])

// deltaType：up(青)/down(红)。告警量、异常数、MTTR 下降为好转；压缩率上升为好转
const deltaConf = (pct, goodWhenNegative) => {
  if (pct == null || isNaN(pct)) return { delta: '', deltaType: 'flat' }
  const negative = pct < 0
  const good = goodWhenNegative ? negative : !negative
  return { delta: (pct > 0 ? '+' : '') + pct + '%', deltaType: good ? 'up' : 'down' }
}

const kpiStats = computed(() => {
  const k = kpi.value || {}
  return [
    {
      label: t('dashboard.kpi.activeAlerts.label'), value: k.active_alerts ?? '-',
      ...deltaConf(k.active_alerts_delta_pct, true), hint: t('dashboard.kpi.activeAlerts.hint'),
      help: t('dashboard.kpi.activeAlerts.help')
    },
    {
      label: t('dashboard.kpi.anomalyToday.label'), value: k.anomaly_today ?? '-',
      delta: k.anomaly_delta != null ? ((k.anomaly_delta > 0 ? '+' : '') + k.anomaly_delta) : '',
      deltaType: k.anomaly_delta > 0 ? 'down' : (k.anomaly_delta < 0 ? 'up' : 'flat'),
      hint: t('dashboard.kpi.anomalyToday.hint'),
      help: t('dashboard.kpi.anomalyToday.help')
    },
    {
      label: t('dashboard.kpi.compressionRate.label'), value: k.compression_rate ?? '-', unit: '%',
      ...deltaConf(k.compression_delta_pct, false), hint: t('dashboard.kpi.compressionRate.hint'),
      help: t('dashboard.kpi.compressionRate.help')
    },
    {
      label: t('dashboard.kpi.mttr.label'), value: k.mttr_minutes ?? '-', unit: 'min',
      ...deltaConf(k.mttr_delta_pct, true), hint: t('dashboard.kpi.mttr.hint'),
      help: t('dashboard.kpi.mttr.help')
    }
  ]
})

const trendRate = computed(() => trend.value.compression_rate ?? '-')
const hasTrend = computed(() => (trend.value.buckets || []).length > 0)

const severityColor = { P1: '#c93b3b', P2: '#d97b29', P3: '#d9a400' }
const severityText = computed(() => ({
  1: t('dashboard.alerts.p1'),
  2: t('dashboard.alerts.p2'),
  3: t('dashboard.alerts.p3')
}))
const severityTag = (s) => (s === 1 ? 'critical' : s === 2 ? 'major' : 'warning')

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: [t('dashboard.trend.legendRaw'), t('dashboard.trend.legendDenoised')], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 40, right: 16, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: trend.value.buckets || [], axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    { name: t('dashboard.trend.legendRaw'), type: 'line', smooth: true, data: trend.value.raw || [], showSymbol: false, lineStyle: { color: '#c9d4d1', width: 2 }, itemStyle: { color: '#c9d4d1' } },
    {
      name: t('dashboard.trend.legendDenoised'), type: 'line', smooth: true, data: trend.value.denoised || [], showSymbol: false,
      lineStyle: { color: '#0e7c72', width: 2.5 }, itemStyle: { color: '#0e7c72' },
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(14,124,114,0.16)' }, { offset: 1, color: 'rgba(14,124,114,0)' }] } }
    }
  ]
}))

const pieOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0, textStyle: { color: '#5c6b68' } },
  series: [{
    type: 'pie', radius: ['48%', '72%'], center: ['50%', '44%'],
    label: { show: false },
    data: severityDist.value.map((d) => ({
      name: d.level,
      value: d.count,
      itemStyle: { color: severityColor[d.level] || '#8a9693' }
    })),
    itemStyle: { borderColor: '#fff', borderWidth: 2 }
  }]
}))

const healthStatus = (score) => (score < 80 ? 'warning' : 'online')
const barColor = (score) => (score < 80 ? 'var(--c-p1)' : 'var(--c-primary)')
const trendText = (trend) => (trend === '恶化' ? t('dashboard.health.trendWorse') : trend === '好转' ? t('dashboard.health.trendBetter') : t('dashboard.health.trendStable'))

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const params = rangeMode.value === 'custom' && customStart.value && customEnd.value
      ? { start: new Date(customStart.value).toISOString(), end: new Date(customEnd.value).toISOString() }
      : { hours: 24 }
    const data = await dashboardApi.overview(params)
    kpi.value = data.kpi || {}
    trend.value = data.trend || { buckets: [], raw: [], denoised: [] }
    severityDist.value = data.severity_dist || []
    serviceHealth.value = data.service_health || []
    latestAlerts.value = (data.latest_alerts || []).slice(0, 10)
  } catch (e) {
    error.value = e.message || t('dashboard.error.loadFailed')
  } finally {
    loading.value = false
  }
}

function switchMode(mode) {
  if (rangeMode.value === mode) return
  rangeMode.value = mode
  if (mode === '24h') loadData()
}

function applyCustom() {
  if (!customStart.value || !customEnd.value) return
  loadData()
}

function goAlertEvents() {
  router.push('/alerts/events')
}

onMounted(loadData)
</script>

<style scoped>
.range-switch { display: flex; gap: 8px; }
.custom-range { display: flex; align-items: center; gap: 8px; }
.custom-range input { padding: 4px 8px; border: 1px solid var(--c-border); border-radius: 6px; font-size: 12px; }
.error-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; color: var(--c-p0); }
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.row-2 { display: grid; grid-template-columns: 1.7fr 1fr; gap: 16px; margin-bottom: 16px; }
.row-2b { display: grid; grid-template-columns: 1fr 1.7fr; gap: 16px; }
.card-head-flex { display: flex; justify-content: space-between; align-items: flex-start; }
.alert-title { max-width: 380px; }
.chart-empty { display: flex; align-items: center; justify-content: center; height: 280px; color: var(--c-text-3); font-size: 13px; }
.health-list { display: grid; grid-template-columns: 1fr 1fr; gap: 14px 18px; }
.health-item { display: flex; flex-direction: column; gap: 5px; }
.health-head { display: flex; justify-content: space-between; font-size: 13px; }
.health-score { font-weight: 700; }
.health-score.online { color: var(--c-primary); }
.health-score.warning { color: var(--c-p1); }
.health-foot { display: flex; justify-content: space-between; }
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .row-2, .row-2b { grid-template-columns: 1fr; }
}
</style>
