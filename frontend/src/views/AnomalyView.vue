<template>
  <div>
    <PageHeader title="告警事件" desc="接入 Nightingale 的活跃与历史告警事件查询" />

    <div class="kpi-grid">
      <StatCard label="活跃告警数" :value="activeCount" deltaType="flat" hint="当前列表" />
      <StatCard label="历史告警数（今日）" :value="historyCount" deltaType="flat" hint="当前列表" />
      <StatCard label="P1 告警数" :value="p1Count" deltaType="down" hint="当前列表" />
      <StatCard label="P2 告警数" :value="p2Count" deltaType="up" hint="当前列表" />
    </div>

    <div class="card">
      <div class="card-head-flex">
        <div class="tabs">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            class="tab-btn"
            :class="{ active: scope === tab.value }"
            @click="switchScope(tab.value)"
          >
            {{ tab.label }}
          </button>
        </div>
        <div class="filters">
          <div class="filter-group">
            <label>时间窗口</label>
            <select v-model.number="hours" @change="loadData">
              <option :value="24">最近 24 小时</option>
              <option :value="168">最近 7 天</option>
              <option :value="720">最近 30 天</option>
            </select>
          </div>
          <div class="filter-group">
            <label>级别</label>
            <select v-model="severity" @change="loadData">
              <option value="">全部</option>
              <option value="1">P1-紧急</option>
              <option value="2">P2-警告</option>
              <option value="3">P3-提醒</option>
            </select>
          </div>
          <div class="search-box">
            <input
              v-model="query"
              type="text"
              placeholder="搜索规则名称或告警对象"
              @keyup.enter="loadData"
            />
            <button v-if="query" class="clear-btn" @click="clearQuery">×</button>
          </div>
          <button class="btn btn-sm" @click="loadData" :disabled="loading">
            {{ loading ? '加载中...' : '刷新' }}
          </button>
        </div>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>规则名称</th>
            <th>级别</th>
            <th>状态</th>
            <th>告警对象</th>
            <th>触发时间</th>
            <th>标签</th>
            <th>触发值</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ev in events" :key="ev.id">
            <td><b>{{ ev.rule_name }}</b></td>
            <td>
              <LevelTag :level="severityMap(ev.severity)">
                {{ severityText[ev.severity] || ev.severity }}
              </LevelTag>
            </td>
            <td>
              <LevelTag :level="statusMap(ev)">
                {{ ev.is_recovered ? '已恢复' : '告警中' }}
              </LevelTag>
            </td>
            <td class="mono">{{ ev.target_ident || '-' }}</td>
            <td class="muted">{{ fmtTime(ev.trigger_time) }}</td>
            <td class="muted mono tags-cell" :title="ev.tags">{{ ev.tags || '-' }}</td>
            <td class="mono">{{ ev.trigger_value || '-' }}</td>
            <td>
              <button class="btn btn-sm" @click="goToRca(ev)">根因分析</button>
            </td>
          </tr>
          <tr v-if="!loading && events.length === 0">
            <td colspan="8" class="empty-row">暂无告警事件数据</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>

      <div class="pagination" v-if="total > 0">
        <span class="muted">共 {{ total }} 条</span>
        <div class="page-ops">
          <button class="btn btn-sm" :disabled="page === 1 || loading" @click="changePage(page - 1)">上一页</button>
          <span class="page-info">第 {{ page }} 页</span>
          <button class="btn btn-sm" :disabled="page * limit >= total || loading" @click="changePage(page + 1)">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import LevelTag from '../components/LevelTag.vue'
import { alertEventApi } from '../api/alertEvent.js'

const router = useRouter()

const tabs = [
  { label: '活跃告警', value: 'active' },
  { label: '历史告警', value: 'history' }
]

const scope = ref('active')
const hours = ref(24)
const query = ref('')
const severity = ref('')
const page = ref(1)
const limit = ref(20)

const events = ref([])
const total = ref(0)
const loading = ref(false)

const severityText = { 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }

const severityMap = (s) => {
  if (s === 1) return 'critical'
  if (s === 2) return 'warning'
  return 'info'
}

const statusMap = (ev) => ev.is_recovered ? 'resolved' : 'active'

const fmtTime = (ts) => ts ? new Date(ts * 1000).toLocaleString() : '-'

function encodeEvent(event) {
  const payload = {
    rule_name: event.rule_name,
    target_ident: event.target_ident,
    tags: event.tags,
    trigger_time: event.trigger_time,
    trigger_value: event.trigger_value,
    severity: event.severity,
    is_recovered: event.is_recovered
  }
  return btoa(encodeURIComponent(JSON.stringify(payload)))
}

function goToRca(event) {
  router.push({
    path: '/chat',
    query: { rca: '1', event: encodeEvent(event) }
  })
}

const activeCount = computed(() => scope.value === 'active' ? events.value.length : 0)
const historyCount = computed(() => scope.value === 'history' ? events.value.length : 0)
const p1Count = computed(() => events.value.filter(e => e.severity === 1).length)
const p2Count = computed(() => events.value.filter(e => e.severity === 2).length)

function switchScope(next) {
  if (scope.value === next) return
  scope.value = next
  page.value = 1
  loadData()
}

function clearQuery() {
  query.value = ''
  loadData()
}

function changePage(next) {
  page.value = next
  loadData()
}

async function loadData() {
  loading.value = true
  try {
    const result = await alertEventApi.list({
      scope: scope.value,
      hours: hours.value,
      page: page.value,
      limit: limit.value,
      query: query.value.trim(),
      severity: severity.value === '' ? '' : String(severity.value)
    })
    events.value = result?.list || []
    total.value = result?.total || 0
  } catch (err) {
    alert('加载告警事件失败：' + err.message)
    events.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 18px;
}

.card-head-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 16px;
}

.tabs {
  display: flex;
  gap: 8px;
}

.tab-btn {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.tab-btn:hover {
  border-color: var(--c-primary);
  color: var(--c-primary);
}

.tab-btn.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: var(--c-surface);
}

.filters {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-group label {
  font-size: 12px;
  color: var(--c-text-2);
  white-space: nowrap;
}

.filter-group select {
  padding: 6px 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  outline: none;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 6px 12px;
  color: var(--c-text-3);
  width: 220px;
  position: relative;
}

.search-box input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  width: 100%;
  color: var(--c-text);
}

.search-box .clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  border: none;
  background: none;
  color: var(--c-text-3);
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 0 4px;
}

.search-box .clear-btn:hover {
  color: var(--c-text);
}

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}

.tags-cell {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--c-border);
}

.page-ops {
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-info {
  font-size: 13px;
  color: var(--c-text-2);
}

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .card-head-flex { flex-direction: column; align-items: flex-start; }
}
</style>
