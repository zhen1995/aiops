<template>
  <div>
    <PageHeader title="告警事件" desc="由系统按告警规则自动探测生成的活跃与历史告警事件" />

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
              @keyup.enter="search"
            />
            <button v-if="query" class="clear-btn" @click="clearQuery">×</button>
          </div>
          <button class="btn btn-sm btn-primary" @click="search" :disabled="loading">查询</button>
          <button class="btn btn-sm" @click="loadData" :disabled="loading">
            {{ loading ? '加载中...' : '刷新' }}
          </button>
        </div>
      </div>

      <!-- 批量操作条 -->
      <div v-if="selectedIds.size > 0" class="batch-bar">
        <label class="batch-check">
          <input type="checkbox" :checked="isAllSelected" :indeterminate="isIndeterminate" @change="toggleSelectAll" />
          <span>全选本页</span>
        </label>
        <span class="batch-info">已选 <b>{{ selectedIds.size }}</b> 条</span>
        <button class="btn btn-sm" @click="clearSelection">取消选择</button>
        <button class="btn btn-sm btn-danger" @click="batchRemove" :disabled="batchLoading">
          {{ batchLoading ? '删除中...' : '批量删除' }}
        </button>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th class="col-check">
              <input type="checkbox" :checked="isAllSelected" :indeterminate="isIndeterminate" @change="toggleSelectAll" />
            </th>
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
          <tr v-for="ev in events" :key="ev.id" :class="{ selected: selectedIds.has(ev.id) }">
            <td class="col-check">
              <input type="checkbox" :checked="selectedIds.has(ev.id)" @change="toggleOne(ev.id)" />
            </td>
            <td><b>{{ ev.rule_name }}</b></td>
            <td>
              <LevelTag :level="severityMap(ev.severity)">
                {{ severityText[ev.severity] || ev.severity }}
              </LevelTag>
            </td>
            <td>
              <LevelTag :level="statusMap(ev)">
                {{ statusText(ev) }}
              </LevelTag>
            </td>
            <td class="mono">{{ ev.target_ident || '-' }}</td>
            <td class="muted">{{ fmtTime(ev.trigger_time) }}</td>
            <td class="muted mono tags-cell" :title="ev.tags">{{ ev.tags || '-' }}</td>
            <td class="mono">{{ ev.trigger_value || '-' }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="goToRca(ev)">根因分析</button>
                <button class="btn btn-sm btn-danger" @click="removeEvent(ev)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && events.length === 0">
            <td colspan="9" class="empty-row">暂无告警事件数据</td>
          </tr>
          <tr v-if="loading">
            <td colspan="9" class="empty-row">加载中...</td>
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import LevelTag from '../components/LevelTag.vue'
import { alertEventApi } from '../api/alertEvent.js'
import { rcaApi } from '../api/rca.js'

const route = useRoute()
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
const batchLoading = ref(false)

// 选中的告警事件 ID 集合
const selectedIds = ref(new Set())

const severityText = { 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }

const severityMap = (s) => {
  if (s === 1) return 'critical'
  if (s === 2) return 'warning'
  return 'info'
}

const statusMap = (ev) => ev.status === 'firing' ? 'active' : 'resolved'

// 状态文案：告警中 / 已恢复（恢复事件单独标注）
const statusText = (ev) => {
  if (ev.status === 'firing') return '告警中'
  return ev.type === 'recovery' ? '已恢复(恢复事件)' : '已恢复'
}

// 兼容数字（旧 unix 秒时间戳）与字符串两种时间格式
const fmtTime = (ts) => {
  if (!ts) return '-'
  if (typeof ts === 'number') return new Date(ts * 1000).toLocaleString()
  const d = new Date(ts)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

async function goToRca(event) {
  try {
    await rcaApi.trigger(event.id)
    router.push({ path: '/rca', query: { event: event.id } })
  } catch (err) {
    alert('触发根因分析失败：' + err.message)
  }
}

const activeCount = computed(() => scope.value === 'active' ? events.value.length : 0)
const historyCount = computed(() => scope.value === 'history' ? events.value.length : 0)
const p1Count = computed(() => events.value.filter(e => e.severity === 1).length)
const p2Count = computed(() => events.value.filter(e => e.severity === 2).length)

// 全选状态
const isAllSelected = computed(() => events.value.length > 0 && events.value.every(ev => selectedIds.value.has(ev.id)))
const isIndeterminate = computed(() => {
  const sel = events.value.filter(ev => selectedIds.value.has(ev.id)).length
  return sel > 0 && sel < events.value.length
})

// 切换单条选中
function toggleOne(id) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

// 切换全选
function toggleSelectAll(e) {
  const next = new Set(selectedIds.value)
  if (e.target.checked) {
    events.value.forEach(ev => next.add(ev.id))
  } else {
    events.value.forEach(ev => next.delete(ev.id))
  }
  selectedIds.value = next
}

function clearSelection() {
  selectedIds.value = new Set()
}

function switchScope(next) {
  if (scope.value === next) return
  scope.value = next
  page.value = 1
  clearSelection()
  loadData()
}

function clearQuery() {
  query.value = ''
  page.value = 1
  clearSelection()
  loadData()
}

function changePage(next) {
  page.value = next
  clearSelection()
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
  // 全局搜索落地：/alerts/events?q=关键字，作为列表搜索关键字
  if (route.query.q) {
    query.value = String(route.query.q)
    router.replace({ path: '/alerts/events', query: {} })
  }
  loadData()
})

// 查询：先回到第一页再加载
function search() {
  page.value = 1
  clearSelection()
  loadData()
}

// 删除告警事件（单条）
async function removeEvent(ev) {
  const label = `${ev.rule_name}${ev.target_ident ? '（' + ev.target_ident + '）' : ''}`
  if (!window.confirm(`确认删除告警事件「${label}」？`)) return
  try {
    await alertEventApi.remove(ev.id)
    clearSelection()
    loadData()
  } catch (err) {
    alert('删除告警事件失败：' + err.message)
  }
}

// 批量删除
async function batchRemove() {
  const ids = Array.from(selectedIds.value)
  if (ids.length === 0) return
  if (!window.confirm(`确认删除选中的 ${ids.length} 条告警事件？`)) return

  batchLoading.value = true
  try {
    await alertEventApi.removeBatch(ids)
    clearSelection()
    loadData()
  } catch (err) {
    alert('批量删除失败：' + err.message)
  } finally {
    batchLoading.value = false
  }
}

// 当列表变化（加载完成）时，移除不存在的选中项（防止翻页后残留）
watch(events, (list) => {
  if (selectedIds.value.size === 0) return
  const valid = new Set(list.map(e => e.id))
  const next = new Set()
  for (const id of selectedIds.value) {
    if (valid.has(id)) next.add(id)
  }
  selectedIds.value = next
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

/* 批量操作条 */
.batch-bar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  margin-bottom: 12px;
  border-radius: 8px;
  background: var(--c-primary-soft, rgba(22,119,255,0.06));
  border: 1px solid var(--c-primary, #1677ff);
}
.batch-check {
  display: flex; align-items: center; gap: 6px;
  font-size: 12.5px; color: var(--c-text-2); cursor: pointer;
  user-select: none;
}
.batch-check input { width: 15px; height: 15px; cursor: pointer; }
.batch-info { font-size: 13px; color: var(--c-text); }
.batch-info b { color: var(--c-primary); font-size: 14px; }

.col-check { width: 42px; text-align: center; }
.col-check input { width: 15px; height: 15px; cursor: pointer; }

.table tr.selected td { background: var(--c-primary-soft, rgba(22,119,255,0.05)); }

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

.ops {
  display: flex;
  gap: 8px;
}

.btn-danger {
  color: var(--c-danger);
}

.btn-danger:hover {
  border-color: var(--c-danger);
  color: var(--c-danger);
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
