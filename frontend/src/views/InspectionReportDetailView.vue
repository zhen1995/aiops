<template>
  <div>
    <PageHeader :title="report.title" :desc="`巡检日期：${report.date} · 综合评分：${report.score} 分`">
      <RouterLink to="/inspection" class="btn">返回列表</RouterLink>
      <button class="btn btn-primary">推送报告</button>
    </PageHeader>

    <div class="score-banner" :class="scoreClass(report.score)">
      <div>
        <b>综合评分</b>
        <span class="score-value">{{ report.score }}</span>
      </div>
      <p>{{ report.summary }}</p>
    </div>

    <div class="report-sections">
      <div v-for="(section, idx) in report.sections" :key="idx" class="card report-section">
        <h3 class="card-title">{{ section.title }}</h3>
        <ul>
          <li v-for="(item, i) in section.content" :key="i">{{ item }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import { inspectionReports, inspectionReportDetails } from '../mock/data'

const route = useRoute()
const reportId = route.params.id

const report = computed(() => {
  const detail = inspectionReportDetails[reportId]
  if (detail) return detail
  const summary = inspectionReports.find((r) => r.id === reportId)
  return summary || { title: '报告不存在', date: '-', score: 0, summary: '', sections: [] }
})

const scoreClass = (score) => {
  if (score >= 90) return 'good'
  if (score >= 80) return 'warning'
  return 'danger'
}
</script>

<style scoped>
.score-banner {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 18px 22px;
  border-radius: var(--radius-card);
  margin-bottom: 20px;
  border: 1px solid var(--c-border);
}

.score-banner.good { background: var(--c-success-bg); border-color: #cee6d8; }
.score-banner.warning { background: var(--c-p2-bg); border-color: #ece4c8; }
.score-banner.danger { background: var(--c-p0-bg); border-color: #f0d0d0; }

.score-banner b { display: block; font-size: 13px; margin-bottom: 4px; }
.score-value { font-size: 32px; font-weight: 700; line-height: 1; }
.score-banner p { margin: 0; font-size: 14px; flex: 1; }

.report-sections { display: flex; flex-direction: column; gap: 16px; }

.report-section ul {
  margin: 10px 0 0;
  padding-left: 18px;
  color: var(--c-text-2);
  font-size: 13.5px;
  line-height: 1.9;
}
</style>
