<template>
  <div>
    <PageHeader :title="$t('log.header.title')" :desc="$t('log.header.desc')">
      <button class="btn">{{ $t('log.header.last24h') }}</button>
      <button class="btn btn-primary">{{ $t('log.header.query') }}</button>
    </PageHeader>

    <!-- 日志处理流水线 -->
    <div class="card pipeline-card">
      <h3 class="card-title">{{ $t('log.pipeline.title') }}</h3>
      <p class="card-sub">{{ $t('log.pipeline.sub') }}</p>
      <div class="pipeline">
        <template v-for="(step, i) in logPipeline" :key="step">
          <div class="pipe-step">
            <span class="pipe-no">{{ i + 1 }}</span>
            <span class="pipe-name">{{ step }}</span>
            <span class="pipe-desc">{{ pipeDescs[i] }}</span>
          </div>
          <span v-if="i < logPipeline.length - 1" class="pipe-arrow">→</span>
        </template>
      </div>
    </div>

    <!-- 日志量趋势 -->
    <div class="card chart-card">
      <h3 class="card-title">{{ $t('log.trend.title') }}</h3>
      <p class="card-sub">{{ $t('log.trend.sub') }}</p>
      <ChartBox :option="seriesOption" height="300px" />
    </div>

    <div class="row-2">
      <!-- 日志聚类 -->
      <div class="card">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">{{ $t('log.cluster.title') }}</h3>
            <p class="card-sub">{{ $t('log.cluster.sub') }}</p>
          </div>
          <span class="cluster-total muted">{{ $t('log.cluster.total', { count: logClusters.length }) }}</span>
        </div>
        <table class="table">
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
            <tr v-for="c in logClusters" :key="c.id">
              <td class="mono">{{ c.id }}</td>
              <td class="mono pattern-cell" :title="c.pattern">{{ c.pattern }}</td>
              <td class="mono">{{ c.count.toLocaleString() }}</td>
              <td><LevelTag :level="c.level" /></td>
              <td><span v-for="s in c.services" :key="s" class="mono svc-badge">{{ s }}</span></td>
              <td><span class="trend-tag" :class="c.trend">{{ trendText[c.trend] }}</span></td>
              <td class="muted">{{ c.firstSeen }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Drain 模板提取 -->
      <div class="card">
        <h3 class="card-title">{{ $t('log.template.title') }}</h3>
        <p class="card-sub">{{ $t('log.template.sub') }}</p>
        <div class="tpl-list">
          <div v-for="(t, i) in logTemplates" :key="i" class="tpl-item">
            <div class="tpl-row">
              <span class="tpl-label">{{ $t('log.template.raw') }}</span>
              <span class="mono tpl-raw">{{ t.raw }}</span>
            </div>
            <div class="tpl-row">
              <span class="tpl-label">{{ $t('log.template.tpl') }}</span>
              <span class="mono tpl-tpl">
                <template v-for="(seg, j) in splitTemplate(t.template)" :key="j">
                  <em v-if="seg.ph" class="tpl-ph">&lt;*&gt;</em>
                  <span v-else>{{ seg.text }}</span>
                </template>
              </span>
            </div>
            <div class="tpl-row">
              <span class="tpl-label">{{ $t('log.template.params') }}</span>
              <span class="tpl-params">
                <span v-for="(p, j) in t.params" :key="j" class="mono param-badge">{{ p }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { logPipeline, logClusters, logTemplates, logSeries } from '../mock/data'

const { t } = useI18n({ useScope: 'global' })

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

// 将模板字符串按 <*> 占位符切分，用于高亮渲染
const splitTemplate = (tpl) => {
  const parts = tpl.split('<*>')
  const out = []
  parts.forEach((text, i) => {
    if (text) out.push({ text })
    if (i < parts.length - 1) out.push({ ph: true })
  })
  return out
}

const seriesOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: [t('log.chart.total'), t('log.chart.error')], top: 0, textStyle: { color: '#5c6b68' } },
  grid: { left: 60, right: 16, top: 36, bottom: 24 },
  xAxis: { type: 'category', data: logSeries.labels, axisLine: { lineStyle: { color: '#e4e9e7' } }, axisLabel: { color: '#93a19d' } },
  yAxis: { type: 'value', splitLine: { lineStyle: { color: '#eef2f0' } }, axisLabel: { color: '#93a19d' } },
  series: [
    {
      name: t('log.chart.total'), type: 'line', smooth: true, data: logSeries.total, showSymbol: false,
      lineStyle: { color: '#0e7c72', width: 2.5 }, itemStyle: { color: '#0e7c72' },
      areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(14,124,114,0.14)' }, { offset: 1, color: 'rgba(14,124,114,0)' }] } }
    },
    { name: t('log.chart.error'), type: 'line', smooth: true, data: logSeries.error, showSymbol: false, lineStyle: { color: '#c93b3b', width: 2 }, itemStyle: { color: '#c93b3b' } }
  ]
}))
</script>

<style scoped>
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
.row-2 { display: grid; grid-template-columns: 1.5fr 1fr; gap: 16px; }
.card-head-flex { display: flex; justify-content: space-between; align-items: flex-start; }
.cluster-total { margin-top: 4px; }
.pattern-cell { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.svc-badge { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); background: var(--c-primary-tint); color: var(--c-primary-dark); }
.trend-tag { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); font-size: 12px; font-weight: 600; white-space: nowrap; }
.trend-tag.spike { color: var(--c-p0); background: var(--c-p0-bg); }
.trend-tag.rising { color: var(--c-p1); background: var(--c-p1-bg); }
.trend-tag.flat { color: var(--c-text-3); background: var(--c-p4-bg); }
.tpl-list { display: flex; flex-direction: column; gap: 12px; }
.tpl-item { border: 1px solid var(--c-border); border-radius: var(--radius-card); padding: 12px 14px; display: flex; flex-direction: column; gap: 8px; background: var(--c-primary-soft); }
.tpl-row { display: flex; align-items: center; gap: 10px; }
.tpl-label { flex-shrink: 0; width: 56px; font-size: 12px; color: var(--c-text-3); }
.tpl-raw { color: var(--c-text-2); }
.tpl-tpl { color: var(--c-text); }
.tpl-ph { font-style: normal; color: var(--c-primary); font-weight: 700; background: var(--c-primary-tint); padding: 0 3px; border-radius: 4px; }
.tpl-params { display: flex; flex-wrap: wrap; gap: 6px; }
.param-badge { display: inline-block; padding: 1px 8px; border-radius: var(--radius-tag); background: var(--c-surface); border: 1px solid var(--c-border); color: var(--c-primary-dark); }
@media (max-width: 1200px) {
  .row-2 { grid-template-columns: 1fr; }
  .pipeline { flex-wrap: wrap; }
  .pipe-step { flex: 1 1 28%; }
  .pipe-arrow { display: none; }
}
</style>
