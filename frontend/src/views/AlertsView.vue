<template>
  <div>
    <PageHeader title="告警控制台" desc="智能降噪后的告警全生命周期管理 · 确认 / 解决闭环">
      <button class="btn" @click="resetFilters">重置筛选</button>
      <button class="btn btn-primary">导出告警</button>
    </PageHeader>

    <div class="kpi-grid">
      <StatCard v-for="k in statCards" :key="k.label" v-bind="k" />
    </div>

    <!-- 筛选栏 -->
    <div class="card filter-bar">
      <div class="filter-group">
        <span class="filter-label">级别</span>
        <div class="seg">
          <button
            v-for="opt in levelOptions" :key="opt.value"
            class="seg-btn" :class="{ on: filters.level === opt.value }"
            @click="filters.level = opt.value"
          >{{ opt.label }}</button>
        </div>
      </div>
      <div class="filter-group">
        <span class="filter-label">状态</span>
        <div class="seg">
          <button
            v-for="opt in statusOptions" :key="opt.value"
            class="seg-btn" :class="{ on: filters.status === opt.value }"
            @click="filters.status = opt.value"
          >{{ opt.label }}</button>
        </div>
      </div>
      <div class="filter-group">
        <span class="filter-label">服务</span>
        <select v-model="filters.service" class="select">
          <option value="">全部服务</option>
          <option v-for="s in serviceOptions" :key="s" :value="s">{{ s }}</option>
        </select>
      </div>
      <div class="filter-group filter-search">
        <input v-model.trim="filters.keyword" class="input" type="text" placeholder="搜索告警标题 / ID / 来源…" />
      </div>
      <span class="muted filter-count">匹配 {{ filteredAlerts.length }} 条</span>
    </div>

    <!-- 告警表格 -->
    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">告警列表</h3>
          <p class="card-sub">降噪后有效告警 · 点击操作可确认 / 解决</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>级别</th><th>告警内容</th><th>服务</th><th>来源</th>
            <th>次数</th><th>状态</th><th>时间</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in filteredAlerts" :key="a.id">
            <td><LevelTag :level="a.level" /></td>
            <td class="alert-title">
              <div class="alert-name">{{ a.title }}</div>
              <div class="muted mono">{{ a.id }}</div>
            </td>
            <td class="mono muted">{{ a.service }}</td>
            <td class="muted">{{ a.source }}</td>
            <td><span class="count-badge">{{ a.count }}</span></td>
            <td><LevelTag :level="a.status" /></td>
            <td class="muted">{{ fmtTime(a.time) }}</td>
            <td>
              <div class="ops">
                <button v-if="a.status === 'active'" class="btn btn-sm" @click="ack(a)">确认</button>
                <button v-if="a.status !== 'resolved'" class="btn btn-sm btn-primary" @click="resolve(a)">解决</button>
                <span v-else class="muted">已闭环</span>
              </div>
            </td>
          </tr>
          <tr v-if="!filteredAlerts.length">
            <td colspan="8" class="empty muted">没有匹配当前筛选条件的告警</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 告警生命周期 -->
    <div class="card">
      <h3 class="card-title">告警生命周期 · 降噪流水线</h3>
      <p class="card-sub">今日原始告警 2,316 条，经 7 个阶段处理，累计拦截 {{ totalReduced }} 条，最终通知 156 条</p>
      <div class="lifecycle">
        <template v-for="(s, i) in alertLifecycle" :key="s.stage">
          <div class="stage" :class="{ terminal: s.reduced === 0 }">
            <div class="stage-index">{{ i + 1 }}</div>
            <div class="stage-name">{{ s.stage }}</div>
            <div class="stage-desc">{{ s.desc }}</div>
            <div class="stage-reduced" :class="{ zero: s.reduced === 0 }">
              {{ s.reduced > 0 ? `拦截 ${s.reduced} 条` : '不拦截' }}
            </div>
          </div>
          <div v-if="i < alertLifecycle.length - 1" class="stage-arrow">→</div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import LevelTag from '../components/LevelTag.vue'
import { alerts, alertStats, alertLifecycle, fmtTime } from '../mock/data'

const alertList = ref(alerts.map((a) => ({ ...a })))

const statCards = [
  { label: '活动告警', value: alertStats.active, delta: '-64%', deltaType: 'up', hint: '较昨日 33 条' },
  { label: '今日新增', value: alertStats.todayNew, delta: '+8', deltaType: 'down', hint: '降噪后有效告警' },
  { label: '已确认', value: alertStats.acked, deltaType: 'flat', hint: '处理中工单' },
  { label: '已解决', value: alertStats.resolved, delta: '+12', deltaType: 'up', hint: '今日闭环' },
  { label: '告警压缩率', value: alertStats.compressRate, unit: '%', delta: '+4.2%', deltaType: 'up', hint: '目标 ≥80%' }
]

const totalReduced = alertLifecycle.reduce((sum, s) => sum + s.reduced, 0)

const filters = reactive({ level: '', status: '', service: '', keyword: '' })

const levelOptions = [
  { label: '全部', value: '' },
  { label: 'P0', value: 'p0' }, { label: 'P1', value: 'p1' }, { label: 'P2', value: 'p2' },
  { label: 'P3', value: 'p3' }, { label: 'P4', value: 'p4' }
]
const statusOptions = [
  { label: '全部', value: '' },
  { label: '活动', value: 'active' }, { label: '已确认', value: 'acked' }, { label: '已解决', value: 'resolved' }
]
const serviceOptions = [...new Set(alertList.value.map((a) => a.service))]

const filteredAlerts = computed(() => {
  const kw = filters.keyword.toLowerCase()
  return alertList.value.filter((a) => {
    if (filters.level && a.level !== filters.level) return false
    if (filters.status && a.status !== filters.status) return false
    if (filters.service && a.service !== filters.service) return false
    if (kw && !(a.title + a.id + a.source).toLowerCase().includes(kw)) return false
    return true
  })
})

function ack(a) { a.status = 'acked' }
function resolve(a) { a.status = 'resolved' }
function resetFilters() { Object.assign(filters, { level: '', status: '', service: '', keyword: '' }) }
</script>

<style scoped>
.kpi-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 16px; margin-bottom: 16px; }

.filter-bar {
  display: flex; align-items: center; flex-wrap: wrap; gap: 14px 22px;
  padding: 14px 20px; margin-bottom: 16px;
}
.filter-group { display: flex; align-items: center; gap: 8px; }
.filter-label { font-size: 13px; color: var(--c-text-2); white-space: nowrap; }
.seg { display: inline-flex; background: var(--c-bg); border: 1px solid var(--c-border); border-radius: 8px; padding: 2px; }
.seg-btn {
  border: none; background: transparent; padding: 4px 12px; border-radius: 6px;
  font-size: 12.5px; color: var(--c-text-2); cursor: pointer; transition: all 0.15s ease;
}
.seg-btn.on { background: var(--c-surface); color: var(--c-primary); font-weight: 600; box-shadow: var(--shadow-card); }
.select, .input {
  border: 1px solid var(--c-border); border-radius: 8px; background: var(--c-surface);
  padding: 6px 10px; font-size: 13px; color: var(--c-text); outline: none;
}
.select:focus, .input:focus { border-color: var(--c-primary); }
.filter-search { flex: 1; min-width: 200px; }
.input { width: 100%; }
.filter-count { margin-left: auto; white-space: nowrap; }

.card { margin-bottom: 16px; }
.card:last-child { margin-bottom: 0; }
.card-head-flex { display: flex; justify-content: space-between; align-items: flex-start; }
.alert-title { max-width: 420px; }
.alert-name { font-size: 13px; }
.count-badge {
  display: inline-block; min-width: 26px; text-align: center;
  background: var(--c-primary-tint); color: var(--c-primary);
  border-radius: var(--radius-tag); padding: 1px 8px; font-weight: 600; font-size: 12px;
}
.ops { display: flex; gap: 6px; align-items: center; }
.empty { text-align: center; padding: 28px 0; }

.lifecycle { display: flex; align-items: stretch; gap: 4px; }
.stage {
  flex: 1; background: var(--c-primary-soft); border: 1px solid var(--c-border);
  border-radius: var(--radius-card); padding: 14px 12px; text-align: center;
  display: flex; flex-direction: column; gap: 4px;
}
.stage.terminal { background: var(--c-p4-bg); }
.stage-index {
  width: 24px; height: 24px; margin: 0 auto; border-radius: 50%;
  background: var(--c-primary); color: #fff; font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.stage.terminal .stage-index { background: var(--c-p4); }
.stage-name { font-size: 13px; font-weight: 600; }
.stage-desc { font-size: 11.5px; color: var(--c-text-3); }
.stage-reduced { font-size: 12.5px; font-weight: 700; color: var(--c-p0); margin-top: 2px; }
.stage-reduced.zero { color: var(--c-text-3); font-weight: 500; }
.stage-arrow { align-self: center; color: var(--c-text-3); font-size: 15px; padding: 0 2px; }

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .lifecycle { flex-direction: column; }
  .stage-arrow { transform: rotate(90deg); }
}
</style>
