<template>
  <div>
    <PageHeader title="运维总览大盘" desc="多源监控数据融合 · 异常检测 / 根因分析 / 告警降噪 实时概览">
      <div class="range-switch">
        <button class="btn btn-sm" :class="{ 'btn-primary': rangeMode === '24h' }" @click="switchMode('24h')">
          近 24 小时
        </button>
        <button class="btn btn-sm" :class="{ 'btn-primary': rangeMode === 'custom' }" @click="switchMode('custom')">
          自定义范围
        </button>
      </div>
      <div v-if="rangeMode === 'custom'" class="custom-range">
        <input v-model="customStart" type="datetime-local" />
        <span>至</span>
        <input v-model="customEnd" type="datetime-local" />
        <button class="btn btn-sm btn-primary" :disabled="loading || !customStart || !customEnd" @click="applyCustom">
          确定
        </button>
      </div>
    </PageHeader>

    <div v-if="error" class="card error-bar">
      <span>{{ error }}</span>
      <button class="btn btn-sm" @click="loadData">重试</button>
    </div>

    <div class="kpi-grid">
      <StatCard v-for="k in kpiStats" :key="k.label" v-bind="k" />
    </div>

    <div class="row-2">
      <div class="card">
        <h3 class="card-title">告警趋势（原始 vs 降噪后）</h3>
        <p class="card-sub">智能降噪实时生效，压缩率 {{ trendRate }}%</p>
        <ChartBox v-if="hasTrend" :option="trendOption" height="280px" />
        <div v-else class="chart-empty">暂无数据</div>
      </div>
      <div class="card">
        <h3 class="card-title">告警级别分布</h3>
        <p class="card-sub">范围内全部告警按级别分级</p>
        <ChartBox v-if="severityDist.length" :option="pieOption" height="280px" />
        <div v-else class="chart-empty">暂无数据</div>
      </div>
    </div>

    <div class="row-2b">
      <div class="card">
        <h3 class="card-title">服务健康度</h3>
        <p class="card-sub">基于多指标关联检测综合评分</p>
        <div v-if="serviceHealth.length" class="health-list">
          <div v-for="s in serviceHealth" :key="s.service" class="health-item">
            <div class="health-head">
              <span class="mono">{{ s.service }}</span>
              <span class="health-score" :class="healthStatus(s.score)">{{ s.score }}</span>
            </div>
            <div class="progress"><i :style="{ width: s.score + '%', background: barColor(s.score) }"></i></div>
            <div class="health-foot muted">
              <span>异常指标 {{ s.abnormal_metrics }} 个</span>
              <span>{{ trendText(s.trend) }}</span>
            </div>
          </div>
        </div>
        <div v-else class="chart-empty">暂无数据</div>
      </div>

      <div class="card">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">最新高危告警</h3>
            <p class="card-sub">范围内最新告警事件（按级别排序）</p>
          </div>
          <button class="btn btn-sm" @click="goAlertEvents">查看告警事件</button>
        </div>
        <table v-if="latestAlerts.length" class="table">
          <thead>
            <tr><th>级别</th><th>告警内容</th><th>服务</th><th>时间</th></tr>
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
        <div v-else class="chart-empty">暂无数据</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { dashboardApi } from '../api/dashboard.js'

const router = useRouter()

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
      label: '活动告警', value: k.active_alerts ?? '-',
      ...deltaConf(k.active_alerts_delta_pct, true), hint: '较昨日变化'
    },
    {
      label: '今日异常检测', value: k.anomaly_today ?? '-',
      delta: k.anomaly_delta != null ? ((k.anomaly_delta > 0 ? '+' : '') + k.anomaly_delta) : '',
      deltaType: k.anomaly_delta > 0 ? 'down' : (k.anomaly_delta < 0 ? 'up' : 'flat'),
      hint: '命中异常点'
    },
    {
      label: '告警压缩率', value: k.compression_rate ?? '-', unit: '%',
      ...deltaConf(k.compression_delta_pct, false), hint: '目标 ≥80%'
    },
    {
      label: '平均 MTTR', value: k.mttr_minutes ?? '-', unit: 'min',
      ...deltaConf(k.mttr_delta_pct, true), hint: '目标 <12min'
    }
  ]
})

const trendRate = computed(() => trend.value.compression_rate ?? '-')
const hasTrend = computed(() => (trend.value.buckets || []).length > 0)

const severityColor = { P1: '#c93b3b', P2: '#d97b29', P3: '#0e7c72' }
const severityText = { 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }
const severityTag = (s) => (s === 1 ? 'critical' : s === 2 ? 'warning' : 'info')

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['原始告警', '降噪后告警'], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 40, right: 16, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: trend.value.buckets || [], axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    { name: '原始告警', type: 'line', smooth: true, data: trend.value.raw || [], showSymbol: false, lineStyle: { color: '#c9d4d1', width: 2 }, itemStyle: { color: '#c9d4d1' } },
    {
      name: '降噪后告警', type: 'line', smooth: true, data: trend.value.denoised || [], showSymbol: false,
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
const trendText = (t) => (t === '恶化' ? '↓ 恶化' : t === '好转' ? '↑ 好转' : '→ 平稳')

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
    error.value = e.message || '加载总览数据失败'
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
