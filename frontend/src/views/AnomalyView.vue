<template>
  <div>
    <PageHeader title="异常检测" desc="多算法融合的指标 / 日志异常检测 · 可解释性输出">
      <button class="btn">检测配置</button>
      <button class="btn btn-primary">新建检测任务</button>
    </PageHeader>

    <div class="kpi-grid">
      <StatCard label="运行中模型" :value="runningModels" deltaType="flat" :hint="`共 ${anomalyAlgorithms.length} 种算法`" />
      <StatCard label="今日检测任务" :value="186" delta="+12" deltaType="flat" hint="覆盖全部核心服务" />
      <StatCard label="命中异常" :value="47" delta="+8" deltaType="down" hint="今日异常点" />
      <StatCard label="平均置信度" :value="avgConfidence" unit="%" delta="+2.1%" deltaType="up" hint="Top 4 命中结果" />
    </div>

    <div class="row-main">
      <!-- 指标时序图 -->
      <div class="card">
        <h3 class="card-title">指标时序检测：node_cpu_seconds_total</h3>
        <p class="card-sub">order-db-primary · Prophet 时序分解 · 红点为命中的异常点（30 分钟粒度）</p>
        <ChartBox :option="seriesOption" height="380px" />
      </div>

      <!-- 检测结果详情 -->
      <div class="card result-card">
        <h3 class="card-title">检测结果详情</h3>
        <p class="card-sub">最近命中的异常结果与可解释性说明</p>
        <div class="result-list">
          <div v-for="r in anomalyResults" :key="r.metric + r.time" class="result-item">
            <div class="result-head">
              <div>
                <div class="mono result-metric">{{ r.metric }}</div>
                <div class="muted">{{ r.service }} · {{ r.algorithm }} · {{ fmtTime(r.time) }}</div>
              </div>
              <LevelTag :level="r.severity" />
            </div>

            <div class="score-row">
              <div class="score-item">
                <div class="score-label"><span>异常评分</span><b>{{ r.score.toFixed(2) }}</b></div>
                <div class="progress"><i :style="{ width: r.score * 100 + '%' }"></i></div>
              </div>
              <div class="score-item">
                <div class="score-label"><span>置信度</span><b>{{ Math.round(r.confidence * 100) }}%</b></div>
                <div class="progress"><i :style="{ width: r.confidence * 100 + '%' }"></i></div>
              </div>
            </div>

            <div class="value-row">
              <div class="value-item"><span class="muted">实际值</span><b class="actual">{{ fmtNum(r.value) }}</b></div>
              <div class="value-item"><span class="muted">期望值</span><b>{{ fmtNum(r.expected) }}</b></div>
              <div class="value-item">
                <span class="muted">偏差</span>
                <b :class="r.deviation >= 0 ? 'dev-up' : 'dev-down'">{{ r.deviation >= 0 ? '+' : '' }}{{ fmtNum(r.deviation) }}</b>
              </div>
            </div>

            <div class="explain">
              <div class="explain-row"><span class="explain-key">检测方法</span><span>{{ r.explanation.method }}</span></div>
              <div class="explain-row"><span class="explain-key">趋势</span><span>{{ r.explanation.trend }}</span></div>
              <div class="explain-row"><span class="explain-key">周期性</span><span>{{ r.explanation.seasonality }}</span></div>
              <div class="explain-row"><span class="explain-key">判定原因</span><span>{{ r.explanation.reason }}</span></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 算法选型矩阵 -->
    <div class="card">
      <h3 class="card-title">算法选型矩阵</h3>
      <p class="card-sub">按检测场景划分的算法能力与运行状态</p>
      <table class="table">
        <thead>
          <tr><th>检测场景</th><th>算法</th><th>模型类型</th><th>优势</th><th>适用数据</th><th>运行状态</th></tr>
        </thead>
        <tbody>
          <tr v-for="a in anomalyAlgorithms" :key="a.scene + a.algo">
            <td>{{ a.scene }}</td>
            <td class="mono">{{ a.algo }}</td>
            <td><span class="type-badge">{{ a.type }}</span></td>
            <td class="muted">{{ a.advantage }}</td>
            <td class="muted">{{ a.data }}</td>
            <td>
              <LevelTag v-if="a.status === 'running'" level="running" />
              <span v-else class="paused-tag">已暂停</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { anomalyAlgorithms, anomalyResults, metricSeries, fmtTime } from '../mock/data'

const runningModels = anomalyAlgorithms.filter((a) => a.status === 'running').length
const avgConfidence = Math.round(
  (anomalyResults.reduce((s, r) => s + r.confidence, 0) / anomalyResults.length) * 100
)

const fmtNum = (v) => (Math.abs(v) >= 1000 ? v.toLocaleString('en-US') : v)

// 预测区间带：期望值 ± 8
const bandLow = metricSeries.expected.map((v) => Math.round((v - 8) * 10) / 10)
const bandRange = metricSeries.expected.map((v) => 16)

const anomalyPoints = metricSeries.anomalyIndex.map((i) => ({
  coord: [metricSeries.labels[i], Math.round(metricSeries.actual[i] * 10) / 10],
  value: Math.round(metricSeries.actual[i] * 10) / 10
}))

const seriesOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['实际值', '期望值', '预测区间'], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 48, right: 16, top: 36, bottom: 28 },
  xAxis: { type: 'category', data: metricSeries.labels, axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d', interval: 5 } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    {
      name: 'band-low', type: 'line', data: bandLow, stack: 'band', showSymbol: false,
      lineStyle: { opacity: 0 }, itemStyle: { color: 'transparent' }, silent: true, legendHoverLink: false
    },
    {
      name: '预测区间', type: 'line', data: bandRange, stack: 'band', showSymbol: false,
      lineStyle: { opacity: 0 }, areaStyle: { color: 'rgba(14,124,114,0.10)' }, silent: true,
      itemStyle: { color: 'rgba(14,124,114,0.25)' }
    },
    {
      name: '期望值', type: 'line', smooth: true, data: metricSeries.expected, showSymbol: false,
      lineStyle: { color: '#8a9693', width: 1.8, type: 'dashed' }, itemStyle: { color: '#8a9693' }
    },
    {
      name: '实际值', type: 'line', smooth: true, data: metricSeries.actual.map((v) => Math.round(v * 10) / 10), showSymbol: false,
      lineStyle: { color: '#0e7c72', width: 2.5 }, itemStyle: { color: '#0e7c72' },
      markPoint: {
        symbol: 'circle', symbolSize: 11,
        itemStyle: { color: '#c93b3b', borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '异常', color: '#c93b3b', fontSize: 11, offset: [0, -14] },
        data: anomalyPoints
      }
    }
  ]
}))
</script>

<style scoped>
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.row-main { display: grid; grid-template-columns: 1.4fr 1fr; gap: 16px; margin-bottom: 16px; }
.card:last-child { margin-bottom: 0; }

.result-card { display: flex; flex-direction: column; }
.result-list { display: flex; flex-direction: column; gap: 14px; overflow: auto; max-height: 470px; padding-right: 4px; }
.result-item { border: 1px solid var(--c-border); border-radius: var(--radius-card); padding: 12px 14px; display: flex; flex-direction: column; gap: 10px; }
.result-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 10px; }
.result-metric { font-weight: 600; color: var(--c-text); }

.score-row { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.score-label { display: flex; justify-content: space-between; font-size: 12px; color: var(--c-text-2); margin-bottom: 4px; }
.score-label b { color: var(--c-text); }

.value-row { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; }
.value-item { display: flex; flex-direction: column; gap: 1px; font-size: 12px; }
.value-item b { font-size: 15px; }
.value-item .actual { color: var(--c-p0); }
.dev-up { color: var(--c-p0); }
.dev-down { color: var(--c-p1); }

.explain { background: var(--c-primary-soft); border-radius: var(--radius-tag); padding: 8px 12px; display: flex; flex-direction: column; gap: 3px; }
.explain-row { display: flex; gap: 10px; font-size: 12px; line-height: 1.55; }
.explain-key { flex: none; width: 56px; color: var(--c-text-3); }

.type-badge {
  display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag);
  background: var(--c-p4-bg); color: var(--c-text-2); font-size: 12px;
}
.paused-tag {
  display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag);
  background: var(--c-p4-bg); color: var(--c-p4); font-size: 12px; font-weight: 600;
}

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .row-main { grid-template-columns: 1fr; }
  .result-list { max-height: none; }
}
</style>
