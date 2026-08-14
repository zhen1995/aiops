<template>
  <div>
    <PageHeader title="数据源接入" desc="指标与日志双通道接入 · Prometheus / ELK · 采集状态实时监控">
      <button class="btn">连接测试</button>
      <button class="btn btn-primary">新增数据源</button>
    </PageHeader>

    <!-- 数据源大卡片 -->
    <div class="ds-grid">
      <div v-for="ds in dataSources" :key="ds.name" class="card ds-card">
        <div class="ds-head">
          <div class="ds-logo">{{ ds.name.charAt(0) }}</div>
          <div class="ds-title">
            <div class="ds-name-row">
              <span class="ds-name">{{ ds.name }}</span>
              <span class="ds-type-badge">{{ ds.type }}</span>
              <LevelTag :level="ds.status" />
            </div>
            <span class="muted">{{ ds.method }}</span>
          </div>
          <div class="ds-rate">
            <span class="ds-rate-value">{{ rateNumber(ds.metricsRate) }}</span>
            <span class="ds-rate-unit">{{ rateUnit(ds.metricsRate) }}</span>
          </div>
        </div>
        <div class="ds-endpoint">
          <span class="muted">Endpoint</span>
          <span class="mono">{{ ds.endpoint }}</span>
        </div>
        <p class="ds-desc">{{ ds.desc }}</p>
      </div>
    </div>

    <!-- 采集速率趋势 -->
    <div class="card chart-card">
      <h3 class="card-title">采集速率趋势</h3>
      <p class="card-sub">指标采集（k samples/s）与日志采集（k docs/s）· 近 24 小时</p>
      <ChartBox :option="rateOption" height="300px" />
    </div>

    <div class="row-2">
      <!-- 日志结构化字段 -->
      <div class="card">
        <h3 class="card-title">日志结构化字段</h3>
        <p class="card-sub">日志解析后写入 Elasticsearch 的标准字段映射</p>
        <table class="table">
          <thead>
            <tr><th>字段名</th><th>类型</th><th>说明</th><th>示例</th></tr>
          </thead>
          <tbody>
            <tr v-for="f in logFields" :key="f.field">
              <td class="mono">{{ f.field }}</td>
              <td><span class="type-badge mono">{{ f.type }}</span></td>
              <td>{{ f.desc }}</td>
              <td class="mono muted">{{ f.example }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 指标类型覆盖 -->
      <div class="card">
        <h3 class="card-title">指标类型覆盖</h3>
        <p class="card-sub">Prometheus 四类指标全量接入及异常检测适用场景</p>
        <table class="table">
          <thead>
            <tr><th>指标类型</th><th>说明</th><th>异常检测适用场景</th></tr>
          </thead>
          <tbody>
            <tr v-for="m in metricTypes" :key="m.type">
              <td><span class="type-badge mono">{{ m.type }}</span></td>
              <td>{{ m.desc }}</td>
              <td class="scene-cell">{{ m.scene }}</td>
            </tr>
          </tbody>
        </table>
        <p class="coverage-note">
          当前共接入指标 <b>2,418</b> 个、采集目标 <b>136</b> 个，全部通过服务发现自动注册，
          新增服务接入后 1 分钟内纳入异常检测范围。
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { dataSources, collectorStats, logFields, metricTypes } from '../mock/data'

// '42k samples/s' → ('42k', 'samples/s')
const rateNumber = (rate) => rate.split(' ')[0]
const rateUnit = (rate) => rate.split(' ').slice(1).join(' ')

const rateOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['指标采集 (k samples/s)', '日志采集 (k docs/s)'], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 44, right: 48, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: collectorStats.labels, axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: [
    { type: 'value', name: 'k samples/s', nameTextStyle: { color: '#93a19d' }, splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
    { type: 'value', name: 'k docs/s', nameTextStyle: { color: '#93a19d' }, splitLine: { show: false }, axisLabel: { color: '#93a19d' } }
  ],
  series: [
    {
      name: '指标采集 (k samples/s)', type: 'line', smooth: true, data: collectorStats.metrics, showSymbol: false,
      lineStyle: { color: '#0e7c72', width: 2.5 }, itemStyle: { color: '#0e7c72' },
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(14,124,114,0.14)' }, { offset: 1, color: 'rgba(14,124,114,0)' }] } }
    },
    { name: '日志采集 (k docs/s)', type: 'line', yAxisIndex: 1, smooth: true, data: collectorStats.logs, showSymbol: false, lineStyle: { color: '#d97b29', width: 2 }, itemStyle: { color: '#d97b29' } }
  ]
}))
</script>

<style scoped>
.ds-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.ds-card { display: flex; flex-direction: column; gap: 12px; }
.ds-head { display: flex; align-items: center; gap: 14px; }
.ds-logo {
  flex-shrink: 0; width: 46px; height: 46px; border-radius: var(--radius-card);
  background: var(--c-primary-tint); color: var(--c-primary);
  font-size: 20px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.ds-title { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.ds-name-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.ds-name { font-size: 16px; font-weight: 700; }
.ds-type-badge {
  padding: 1px 8px; border-radius: var(--radius-tag);
  background: var(--c-primary-tint); color: var(--c-primary-dark);
  font-size: 12px; font-weight: 600; white-space: nowrap;
}
.ds-rate { margin-left: auto; text-align: right; flex-shrink: 0; }
.ds-rate-value { font-size: 26px; font-weight: 700; color: var(--c-primary-dark); letter-spacing: -0.5px; }
.ds-rate-unit { display: block; font-size: 12px; color: var(--c-text-3); }
.ds-endpoint {
  display: flex; align-items: center; gap: 10px;
  background: var(--c-primary-soft); border: 1px solid var(--c-border);
  border-radius: var(--radius-card); padding: 8px 12px;
}
.ds-desc { margin: 0; font-size: 13px; color: var(--c-text-2); }
.chart-card { margin-bottom: 16px; }
.row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.type-badge {
  display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag);
  background: var(--c-primary-tint); color: var(--c-primary-dark); font-weight: 600;
}
.scene-cell { color: var(--c-text-2); }
.coverage-note { margin: 14px 0 0; font-size: 12.5px; color: var(--c-text-2); background: var(--c-primary-tint); border-radius: var(--radius-card); padding: 10px 14px; line-height: 1.7; }
.coverage-note b { color: var(--c-primary-dark); }
@media (max-width: 1200px) {
  .ds-grid, .row-2 { grid-template-columns: 1fr; }
}
</style>
