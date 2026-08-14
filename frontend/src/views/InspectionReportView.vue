<template>
  <div>
    <PageHeader title="巡检报告管理" desc="查看历史推送的日巡检报告，点击标题可查看详情">
      <button class="btn">导出全部</button>
      <button class="btn btn-primary">手动生成</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">历史巡检报告</h3>
          <p class="card-sub">共 {{ reports.length }} 份报告，按日期倒序排列</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th style="width: 60px">评分</th>
            <th>报告标题</th>
            <th>日期</th>
            <th>摘要</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in reports" :key="r.id">
            <td>
              <span class="score" :class="scoreClass(r.score)">{{ r.score }}</span>
            </td>
            <td>
              <RouterLink :to="`/inspection/${r.id}`" class="report-title">{{ r.title }}</RouterLink>
            </td>
            <td class="muted">{{ r.date }}</td>
            <td class="muted" style="max-width: 420px">{{ r.summary }}</td>
            <td><LevelTag level="running" /></td>
            <td>
              <RouterLink :to="`/inspection/${r.id}`" class="btn btn-sm">查看详情</RouterLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { inspectionReports } from '../mock/data'

const reports = inspectionReports

const scoreClass = (score) => {
  if (score >= 90) return 'good'
  if (score >= 80) return 'warning'
  return 'danger'
}
</script>

<style scoped>
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
</style>
