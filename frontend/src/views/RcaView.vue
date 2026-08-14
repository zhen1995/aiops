<template>
  <div>
    <PageHeader title="根因分析" desc="基于拓扑 + 证据链的自动根因排序 · 附修复建议">
      <button class="btn">分析历史</button>
      <button class="btn btn-primary">重新分析</button>
    </PageHeader>

    <!-- 事件选择 -->
    <div class="case-grid">
      <div
        v-for="c in rcaCases" :key="c.incidentId"
        class="card case-card" :class="{ on: c.incidentId === activeId }"
        @click="activeId = c.incidentId"
      >
        <div class="case-head">
          <span class="mono muted">{{ c.incidentId }}</span>
          <LevelTag :level="c.status" />
        </div>
        <div class="case-title">{{ c.title }}</div>
        <div class="case-foot muted">
          <span>发生时间 {{ fmtTime(c.time) }}</span>
          <span>{{ c.rootCauses.length }} 个候选根因</span>
        </div>
      </div>
    </div>

    <div class="row-main" v-if="activeCase">
      <!-- 根因排名 -->
      <div class="rc-list">
        <div v-for="rc in activeCase.rootCauses" :key="rc.rank" class="card rc-card">
          <div class="rc-head">
            <span class="rank-badge" :class="{ top: rc.rank === 1 }">Rank {{ rc.rank }}</span>
            <div class="rc-entity">
              <div class="rc-name mono">{{ rc.entity.name }}</div>
              <div class="muted">类型 {{ entityTypeText(rc.entity.type) }} · 命名空间 {{ rc.entity.namespace }}</div>
            </div>
            <div class="rc-conf">
              <div class="conf-num">{{ Math.round(rc.confidence * 100) }}<small>%</small></div>
              <div class="muted">置信度</div>
            </div>
          </div>
          <div class="progress rc-progress">
            <i :style="{ width: rc.confidence * 100 + '%', background: rc.rank === 1 ? 'var(--c-primary)' : 'var(--c-p4)' }"></i>
          </div>

          <div class="rc-body">
            <div class="rc-section">
              <div class="sec-title">证据链</div>
              <ul class="evidence-list">
                <li v-for="(e, i) in rc.evidence" :key="i">
                  <span class="ev-index">{{ i + 1 }}</span>{{ e }}
                </li>
              </ul>
            </div>
            <div class="rc-section">
              <div class="sec-title">影响范围</div>
              <div class="impact">
                <span v-for="s in rc.impact.services" :key="s" class="svc-tag mono">{{ s }}</span>
              </div>
              <div class="impact-foot">
                <span class="muted">影响用户：{{ rc.impact.users }}</span>
                <LevelTag :level="rc.impact.severity" />
              </div>
            </div>
          </div>

          <div class="rc-section">
            <div class="sec-title">修复建议</div>
            <ol class="action-list">
              <li v-for="(act, i) in rc.actions" :key="i">{{ act }}</li>
            </ol>
          </div>
        </div>
      </div>

      <!-- 服务拓扑 -->
      <div class="card topo-card">
        <h3 class="card-title">服务拓扑</h3>
        <p class="card-sub">红 = 严重异常 · 橙 = 警告 · 青 = 正常</p>
        <ChartBox :option="topoOption" height="460px" />
        <div class="topo-legend muted">
          <span><i class="dot" style="background: #0e7c72"></i>正常</span>
          <span><i class="dot" style="background: #d97b29"></i>警告</span>
          <span><i class="dot" style="background: #c93b3b"></i>严重</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import ChartBox from '../components/ChartBox.vue'
import LevelTag from '../components/LevelTag.vue'
import { rcaCases, topologyNodes, topologyEdges, fmtTime } from '../mock/data'

const activeId = ref(rcaCases[0].incidentId)
const activeCase = computed(() => rcaCases.find((c) => c.incidentId === activeId.value))

const entityTypeText = (t) => ({ database: '数据库', service: '服务', host: '主机', middleware: '中间件' }[t] || t)

const statusColor = { online: '#0e7c72', warning: '#d97b29', critical: '#c93b3b' }

const topoOption = computed(() => ({
  tooltip: { formatter: (p) => (p.dataType === 'node' ? p.name : '') },
  series: [{
    type: 'graph',
    layout: 'none',
    roam: false,
    symbolSize: 46,
    data: topologyNodes.map((n) => ({
      name: n.name, x: n.x, y: -n.y,
      itemStyle: {
        color: statusColor[n.status] || '#8a9693',
        borderColor: '#fff', borderWidth: 2,
        shadowColor: 'rgba(32,48,45,0.18)', shadowBlur: 6
      }
    })),
    links: topologyEdges.map(([s, t]) => ({ source: s, target: t })),
    lineStyle: { color: '#cdd6d3', width: 1.5, curveness: 0.12 },
    label: {
      show: true, position: 'bottom', distance: 6,
      fontSize: 11.5, color: '#5c6b68', fontFamily: 'Menlo, Consolas, monospace'
    },
    emphasis: { lineStyle: { width: 2.5 } }
  }]
}))
</script>

<style scoped>
.case-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.case-card { cursor: pointer; transition: all 0.15s ease; border: 1px solid var(--c-border); }
.case-card:hover { border-color: var(--c-primary); }
.case-card.on { border-color: var(--c-primary); box-shadow: 0 0 0 2px var(--c-primary-tint), var(--shadow-card); }
.case-head { display: flex; justify-content: space-between; align-items: center; }
.case-title { font-size: 15px; font-weight: 600; margin: 8px 0 6px; }
.case-foot { display: flex; justify-content: space-between; }

.row-main { display: grid; grid-template-columns: 1.5fr 1fr; gap: 16px; align-items: start; }

.rc-list { display: flex; flex-direction: column; gap: 16px; }
.rc-card { display: flex; flex-direction: column; gap: 12px; }
.rc-head { display: flex; align-items: center; gap: 14px; }
.rank-badge {
  flex: none; padding: 4px 12px; border-radius: var(--radius-tag);
  background: var(--c-p4-bg); color: var(--c-text-2); font-weight: 700; font-size: 13px;
}
.rank-badge.top { background: var(--c-primary); color: #fff; }
.rc-entity { flex: 1; min-width: 0; }
.rc-name { font-size: 15px; font-weight: 700; }
.rc-conf { text-align: right; }
.conf-num { font-size: 24px; font-weight: 700; line-height: 1.1; }
.conf-num small { font-size: 13px; font-weight: 400; color: var(--c-text-3); margin-left: 2px; }
.rc-progress { height: 8px; }

.rc-body { display: grid; grid-template-columns: 1.3fr 1fr; gap: 18px; }
.sec-title { font-size: 12.5px; font-weight: 600; color: var(--c-text-2); margin-bottom: 6px; }

.evidence-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 5px; }
.evidence-list li { display: flex; align-items: flex-start; gap: 8px; font-size: 12.5px; line-height: 1.55; }
.ev-index {
  flex: none; width: 18px; height: 18px; border-radius: 50%; margin-top: 1px;
  background: var(--c-primary-tint); color: var(--c-primary);
  font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center;
}

.impact { display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 8px; }
.svc-tag {
  padding: 1px 8px; border-radius: var(--radius-tag);
  background: var(--c-primary-soft); border: 1px solid var(--c-border); color: var(--c-text);
}
.impact-foot { display: flex; justify-content: space-between; align-items: center; gap: 8px; }

.action-list { margin: 0; padding-left: 20px; display: flex; flex-direction: column; gap: 4px; font-size: 12.5px; }
.action-list li::marker { color: var(--c-primary); font-weight: 700; }

.topo-card { position: sticky; top: 16px; }
.topo-legend { display: flex; gap: 18px; justify-content: center; padding-top: 4px; }

@media (max-width: 1200px) {
  .case-grid, .row-main, .rc-body { grid-template-columns: 1fr; }
  .topo-card { position: static; }
}
</style>
