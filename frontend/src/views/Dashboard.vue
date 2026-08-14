<template>
  <div>
    <PageHeader title="运维总览大盘" desc="多源监控数据融合 · 异常检测 / 根因分析 / 告警降噪 实时概览">
      <button class="btn">近 24 小时</button>
      <button class="btn btn-primary">自定义范围</button>
    </PageHeader>

    <div class="kpi-grid">
      <StatCard v-for="k in kpiStats" :key="k.label" v-bind="k" />
    </div>

    <div class="row-2">
      <div class="card">
        <h3 class="card-title">告警趋势（原始 vs 降噪后）</h3>
        <p class="card-sub">智能降噪实时生效，压缩率 86.4%</p>
        <ChartBox :option="trendOption" height="280px" />
      </div>
      <div class="card">
        <h3 class="card-title">告警级别分布</h3>
        <p class="card-sub">今日全部告警按 P0-P4 分级</p>
        <ChartBox :option="pieOption" height="280px" />
      </div>
    </div>

    <div class="row-2b">
      <div class="card">
        <h3 class="card-title">服务健康度</h3>
        <p class="card-sub">基于多指标关联检测（VAE）综合评分</p>
        <div class="health-list">
          <div v-for="s in serviceHealth" :key="s.service" class="health-item">
            <div class="health-head">
              <span class="mono">{{ s.service }}</span>
              <span class="health-score" :class="s.status">{{ s.health }}</span>
            </div>
            <div class="progress"><i :style="{ width: s.health + '%', background: barColor(s.status) }"></i></div>
            <div class="health-foot muted">
              <span>异常指标 {{ s.anomaly }} 个</span>
              <span>{{ s.trend === 'up' ? '↑ 好转' : s.trend === 'down' ? '↓ 恶化' : '→ 平稳' }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">最新高危告警</h3>
            <p class="card-sub">P0 / P1 级活动告警</p>
          </div>
          <RouterLink to="/alerts/console" class="btn btn-sm">进入告警控制台</RouterLink>
        </div>
        <table class="table">
          <thead>
            <tr><th>级别</th><th>告警内容</th><th>服务</th><th>时间</th></tr>
          </thead>
          <tbody>
            <tr v-for="a in topAlerts" :key="a.id">
              <td><LevelTag :level="a.level" /></td>
              <td class="alert-title">{{ a.title }}</td>
              <td class="mono muted">{{ a.service }}</td>
              <td class="muted">{{ fmtTime(a.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { kpiStats, alertTrend, severityDist, serviceHealth, alerts, fmtTime } from '../mock/data'

const topAlerts = computed(() => alerts.filter((a) => ['p0', 'p1'].includes(a.level)).slice(0, 5))

const barColor = (status) =>
  status === 'warning' ? 'var(--c-p1)' : 'var(--c-primary)'

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['原始告警', '降噪后告警'], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 40, right: 16, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: alertTrend.labels, axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    { name: '原始告警', type: 'line', smooth: true, data: alertTrend.raw, showSymbol: false, lineStyle: { color: '#c9d4d1', width: 2 }, itemStyle: { color: '#c9d4d1' } },
    {
      name: '降噪后告警', type: 'line', smooth: true, data: alertTrend.effective, showSymbol: false,
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
    label: { show: false }, data: severityDist,
    itemStyle: { borderColor: '#fff', borderWidth: 2 }
  }]
}))
</script>

<style scoped>
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.row-2 { display: grid; grid-template-columns: 1.7fr 1fr; gap: 16px; margin-bottom: 16px; }
.row-2b { display: grid; grid-template-columns: 1fr 1.7fr; gap: 16px; }
.card-head-flex { display: flex; justify-content: space-between; align-items: flex-start; }
.alert-title { max-width: 380px; }
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
