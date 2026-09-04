<template>
  <div>
    <PageHeader title="根因分析" desc="基于告警事件的自动根因推理 · 附证据链、影响范围与修复建议">
      <button class="btn btn-primary" :disabled="!selected || reanalyzing" @click="reanalyze">
        {{ reanalyzing ? '触发中...' : '重新分析' }}
      </button>
    </PageHeader>

    <!-- 分析历史 -->
    <div class="case-grid" v-if="list.length">
      <div
        v-for="item in list"
        :key="item.id"
        class="card case-card"
        :class="{ on: selected && item.id === selected.id }"
        @click="selectAnalysis(item)"
      >
        <div class="case-head">
          <b class="case-rule">{{ item.rule_name || '-' }}</b>
          <LevelTag :level="statusLevel(item.status)">{{ statusText(item.status) }}</LevelTag>
        </div>
        <div class="case-target mono muted">{{ item.target_ident || '-' }}</div>
        <div class="case-foot muted">
          <span>触发时间 {{ fmtTime(item.trigger_time) }}</span>
          <span>{{ candidateCount(item) }} 个候选根因</span>
        </div>
      </div>
    </div>
    <div class="card empty-card" v-if="!list.length && !loading">
      暂无根因分析记录，请先在「告警事件」页对告警发起根因分析
    </div>

    <!-- 详情区 -->
    <template v-if="selected">
      <!-- 分析中 -->
      <div class="card running-card" v-if="selected.status === 'running'">
        <div class="running-bar"><i></i></div>
        <p class="muted">根因分析进行中，正在聚合指标、日志与性能剖析证据，请稍候...</p>
      </div>

      <!-- 失败 -->
      <div class="card failed-card" v-else-if="selected.status === 'failed'">
        <div class="failed-title">根因分析失败</div>
        <p class="failed-error mono">{{ selected.error || '未知错误' }}</p>
      </div>

      <!-- 成功且有结构化结果 -->
      <template v-else-if="parsedResult">
        <div class="card summary-card" v-if="parsedResult.summary">
          <div class="sec-title">分析结论</div>
          <p class="summary-text">{{ parsedResult.summary }}</p>
        </div>

        <div class="rc-list">
          <div v-for="rc in parsedResult.root_causes || []" :key="rc.rank" class="card rc-card">
            <div class="rc-head">
              <span class="rank-badge" :class="{ top: rc.rank === 1 }">Rank {{ rc.rank }}</span>
              <div class="rc-entity">
                <div class="rc-name mono">{{ rc.entity?.name || '-' }}</div>
                <div class="muted">类型 {{ entityTypeText(rc.entity?.type) }} · 命名空间 {{ rc.entity?.namespace || '-' }}</div>
              </div>
              <div class="rc-conf">
                <div class="conf-num">{{ Math.round((rc.confidence || 0) * 100) }}<small>%</small></div>
                <div class="muted">可信度</div>
              </div>
            </div>
            <div class="progress rc-progress">
              <i :style="{ width: (rc.confidence || 0) * 100 + '%', background: rc.rank === 1 ? 'var(--c-primary)' : 'var(--c-p4)' }"></i>
            </div>

            <div class="rc-body">
              <div class="rc-section">
                <div class="sec-title">证据链</div>
                <ul class="evidence-list">
                  <li v-for="(e, i) in rc.evidence || []" :key="i">
                    <span class="ev-index">{{ i + 1 }}</span>{{ e }}
                  </li>
                </ul>
              </div>
              <div class="rc-section">
                <div class="sec-title">影响范围</div>
                <div class="impact">
                  <span v-for="s in rc.impact?.services || []" :key="s" class="svc-tag mono">{{ s }}</span>
                  <span class="muted" v-if="!(rc.impact?.services || []).length">-</span>
                </div>
                <div class="impact-foot">
                  <span class="muted">影响用户：{{ rc.impact?.users || '-' }}</span>
                  <LevelTag :level="impactLevel(rc.impact?.severity)" />
                </div>
              </div>
            </div>

            <div class="rc-section" v-if="(rc.actions || []).length">
              <div class="sec-title">修复建议</div>
              <ol class="action-list">
                <li v-for="(act, i) in rc.actions" :key="i">{{ act }}</li>
              </ol>
            </div>
          </div>
        </div>
      </template>

      <!-- 成功但无结构化结果：回退展示 Markdown 报告 -->
      <div class="card" v-else>
        <div class="sec-title">分析报告</div>
        <MarkdownContent v-if="selected.content" :content="selected.content" />
        <p class="muted" v-else>分析已完成，但未生成结果内容</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import { rcaApi } from '../api/rca.js'

const route = useRoute()

const list = ref([])
const selected = ref(null)
const loading = ref(false)
const reanalyzing = ref(false)

// 兼容数字（旧 unix 秒时间戳）与字符串两种时间格式
const fmtTime = (ts) => {
  if (!ts) return '-'
  if (typeof ts === 'number') return new Date(ts * 1000).toLocaleString()
  const d = new Date(ts)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

const entityTypeText = (t) => ({ database: '数据库', service: '服务', host: '主机', middleware: '中间件' }[t] || t || '-')

const statusLevel = (s) => ({ running: 'running', completed: 'resolved', failed: 'error' }[s] || 'info')
const statusText = (s) => ({ running: '分析中', completed: '已完成', failed: '失败' }[s] || s || '-')

// 影响级别映射：p0~p2 直出，none/空 归为提示
const impactLevel = (s) => {
  const v = (s || '').toLowerCase()
  if (!v || v === 'none') return 'info'
  return v
}

// result 为结构化 JSON 字符串，可能为空或解析失败
function parseResult(item) {
  if (!item?.result) return null
  try {
    return JSON.parse(item.result)
  } catch {
    return null
  }
}

const parsedResult = computed(() => parseResult(selected.value))

const candidateCount = (item) => {
  const r = parseResult(item)
  return r?.root_causes?.length ?? '-'
}

// ---- running 状态轮询 ----
let pollTimer = null
let pollCount = 0
const POLL_INTERVAL = 3000
const POLL_MAX = 200

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function refreshDetail(id) {
  const data = await rcaApi.detail(id)
  selected.value = data
  // 同步历史卡片上的状态与候选根因数
  const idx = list.value.findIndex((x) => x.id === id)
  if (idx >= 0) list.value[idx] = { ...list.value[idx], ...data }
  return data
}

function startPolling(id) {
  stopPolling()
  pollCount = 0
  pollTimer = setInterval(async () => {
    pollCount += 1
    if (pollCount >= POLL_MAX) {
      stopPolling()
      return
    }
    try {
      const data = await refreshDetail(id)
      if (data.status !== 'running') stopPolling()
    } catch (err) {
      stopPolling()
      alert('查询根因分析进度失败：' + err.message)
    }
  }, POLL_INTERVAL)
}

async function selectAnalysis(item) {
  stopPolling()
  try {
    const data = await refreshDetail(item.id)
    if (data.status === 'running') startPolling(data.id)
  } catch (err) {
    alert('加载根因分析详情失败：' + err.message)
  }
}

async function loadList(alertEventId = '') {
  loading.value = true
  try {
    const result = await rcaApi.list({ alert_event_id: alertEventId, page: 1, limit: 50 })
    list.value = result?.list || []
  } catch (err) {
    alert('加载根因分析历史失败：' + err.message)
    list.value = []
  } finally {
    loading.value = false
  }
}

async function reanalyze() {
  if (!selected.value || reanalyzing.value) return
  reanalyzing.value = true
  stopPolling()
  try {
    const ret = await rcaApi.trigger(selected.value.alert_event_id)
    // 重新拉取该事件的分析列表并选中新触发的分析
    await loadList(selected.value.alert_event_id)
    const target = list.value.find((x) => x.id === ret?.id)
      || list.value.find((x) => x.alert_event_id === selected.value.alert_event_id)
    if (target) selectAnalysis(target)
  } catch (err) {
    alert('触发根因分析失败：' + err.message)
  } finally {
    reanalyzing.value = false
  }
}

onMounted(async () => {
  await loadList()
  const eventId = route.query.event
  if (eventId) {
    // 从告警事件页带参进入：优先在现有列表中定位，否则按事件过滤拉取
    let target = list.value.find((x) => String(x.alert_event_id) === String(eventId))
    if (!target) {
      await loadList(eventId)
      target = list.value.find((x) => String(x.alert_event_id) === String(eventId))
    }
    if (target) selectAnalysis(target)
  } else if (list.value.length) {
    selectAnalysis(list.value[0])
  }
})

onUnmounted(stopPolling)
</script>

<style scoped>
.case-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.case-card { cursor: pointer; transition: all 0.15s ease; border: 1px solid var(--c-border); }
.case-card:hover { border-color: var(--c-primary); }
.case-card.on { border-color: var(--c-primary); box-shadow: 0 0 0 2px var(--c-primary-tint), var(--shadow-card); }
.case-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.case-rule { font-size: 14px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.case-target { margin: 6px 0; font-size: 12.5px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.case-foot { display: flex; justify-content: space-between; gap: 8px; font-size: 12px; }

.empty-card { text-align: center; color: var(--c-text-3); padding: 40px 0; }

.running-card { display: flex; flex-direction: column; align-items: center; gap: 14px; padding: 32px 20px; }
.running-bar {
  width: 60%; max-width: 420px; height: 6px; border-radius: 3px;
  background: var(--c-bg); overflow: hidden; position: relative;
}
.running-bar i {
  position: absolute; top: 0; bottom: 0; width: 30%; border-radius: 3px;
  background: var(--c-primary); animation: running-slide 1.2s ease-in-out infinite;
}
@keyframes running-slide {
  0% { left: -30%; }
  100% { left: 100%; }
}

.failed-card { padding: 20px; }
.failed-title { font-weight: 600; color: var(--c-p0); margin-bottom: 8px; }
.failed-error { font-size: 12.5px; color: var(--c-text-2); white-space: pre-wrap; }

.summary-card { margin-bottom: 16px; }
.summary-text { margin: 0; font-size: 13.5px; line-height: 1.7; }

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

@media (max-width: 1200px) {
  .case-grid, .rc-body { grid-template-columns: 1fr; }
}
</style>
