<template>
  <div>
    <PageHeader title="巡检报告" desc="查看定时巡检任务生成的历史报告">
    </PageHeader>

    <!-- 查询栏 -->
    <div class="card filter-bar">
      <div class="filter-group">
        <span class="filter-label">来源任务</span>
        <select v-model="filters.task_id" class="select">
          <option value="">全部任务</option>
          <option v-for="t in taskOptions" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
      </div>
      <div class="filter-group">
        <span class="filter-label">状态</span>
        <select v-model="filters.status" class="select">
          <option value="">全部状态</option>
          <option value="completed">已完成</option>
          <option value="generating">生成中</option>
          <option value="failed">失败</option>
        </select>
      </div>
      <button class="btn btn-primary" @click="load">查询</button>
      <button class="btn" @click="resetFilters">重置</button>
    </div>

    <div class="card">
      <table class="table">
        <thead>
          <tr>
            <th style="width: 60px">评分</th>
            <th>报告标题</th>
            <th style="width: 130px">来源任务</th>
            <th style="width: 170px">生成时间</th>
            <th>摘要</th>
            <th style="width: 90px">状态</th>
            <th style="width: 120px">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in reports" :key="r.id">
            <td>
              <span class="score" :class="scoreClass(r.score)">{{ r.score }}</span>
            </td>
            <td>
              <RouterLink :to="`/inspection/reports/${r.id}`" class="report-title">{{ r.title }}</RouterLink>
            </td>
            <td class="muted">{{ r.task_name }}</td>
            <td class="muted">{{ fmtTime(r.created_at) }}</td>
            <td class="muted" style="max-width: 380px">{{ r.summary }}</td>
            <td>
              <span v-if="r.status === 'completed'" class="status-ok">已完成</span>
              <span v-else-if="r.status === 'generating'" class="status-run">生成中</span>
              <span v-else class="status-err">失败</span>
            </td>
            <td>
              <RouterLink v-if="r.status === 'completed'" :to="`/inspection/reports/${r.id}`" class="btn btn-sm">查看</RouterLink>
              <button class="btn btn-sm btn-danger-link" @click="remove(r)">删除</button>
            </td>
          </tr>
          <tr v-if="reports.length === 0">
            <td colspan="7" class="empty">{{ hasFilter ? '没有匹配当前查询条件的巡检报告' : '暂无巡检报告，请先配置巡检任务' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import { listReports, listTasks, deleteReport } from '../api/inspection.js'

const reports = ref([])
const taskOptions = ref([])
const filters = reactive({ task_id: '', status: '' })

const hasFilter = computed(() => !!(filters.task_id || filters.status))

onMounted(() => {
  load()
  loadTaskOptions()
})

async function loadTaskOptions() {
  try {
    taskOptions.value = await listTasks()
  } catch (e) {
    taskOptions.value = []
  }
}

async function load() {
  try {
    reports.value = await listReports({ task_id: filters.task_id, status: filters.status })
  } catch (e) {
    alert('加载报告失败: ' + e.message)
  }
}

function resetFilters() {
  filters.task_id = ''
  filters.status = ''
  load()
}

async function remove(report) {
  if (!confirm(`确认删除巡检报告 "${report.title}"？`)) return
  try {
    await deleteReport(report.id)
    await load()
  } catch (e) {
    alert('删除失败: ' + e.message)
  }
}

const scoreClass = (score) => {
  if (score >= 90) return 'good'
  if (score >= 80) return 'warning'
  return 'danger'
}

function fmtTime(v) {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return v
  return d.toLocaleString('zh-CN')
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px 18px;
  padding: 14px 20px;
  margin-bottom: 16px;
}
.filter-group { display: flex; align-items: center; gap: 8px; }
.filter-label { font-size: 13px; color: var(--c-text-2); white-space: nowrap; }
.select {
  border: 1px solid var(--c-border);
  border-radius: 8px;
  background: var(--c-surface);
  padding: 6px 10px;
  font-size: 13px;
  color: var(--c-text);
  outline: none;
}
.select:focus { border-color: var(--c-primary); }

.btn-danger-link {
  background: transparent;
  color: var(--c-danger, #c93b3b);
  border: none;
  padding: 4px 8px;
  font-size: 12px;
  cursor: pointer;
}
.btn-danger-link:hover { text-decoration: underline; }

.score {
  display: inline-block;
  width: 34px;
  height: 34px;
  line-height: 34px;
  text-align: center;
  border-radius: 50%;
  font-size: 13px;
  font-weight: 700;
}
.score.good { background: var(--c-success-bg); color: var(--c-success); }
.score.warning { background: var(--c-p2-bg); color: var(--c-p2); }
.score.danger { background: var(--c-p0-bg); color: var(--c-p0); }

.report-title {
  font-weight: 600;
  color: var(--c-primary);
}
.report-title:hover { text-decoration: underline; }

.status-ok {
  display: inline-block;
  background: var(--c-success-bg);
  color: var(--c-success);
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
}
.status-run {
  display: inline-block;
  background: var(--c-primary-tint);
  color: var(--c-primary);
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
}
.status-err {
  display: inline-block;
  background: var(--c-p0-bg);
  color: var(--c-danger, #c93b3b);
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
}

.table .empty {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 20px;
}
</style>
