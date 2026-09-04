<template>
  <div>
    <PageHeader :title="report.title || '加载中...'" :desc="reportDesc">
      <RouterLink to="/inspection/reports" class="btn">返回列表</RouterLink>
    </PageHeader>

    <div v-if="report.id" class="score-banner" :class="scoreClass(report.score)">
      <div>
        <b>综合评分</b>
        <span class="score-value">{{ report.score }}</span>
      </div>
      <p>{{ report.summary }}</p>
    </div>

    <div v-if="report.id" class="card report-body">
      <MarkdownContent :content="report.content" />
    </div>

    <div v-if="report.error" class="card error-box">
      <b style="color: var(--c-danger);">生成失败</b>
      <p>{{ report.error }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import PageHeader from '../components/PageHeader.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import { getReport } from '../api/inspection.js'

const route = useRoute()
const report = ref({})

onMounted(async () => {
  try {
    report.value = await getReport(route.params.id)
  } catch (e) {
    alert('加载报告失败: ' + e.message)
  }
})

const reportDesc = computed(() => {
  if (!report.value.id) return ''
  const date = report.value.created_at
    ? new Date(report.value.created_at).toLocaleString('zh-CN')
    : '-'
  return `来源：${report.value.task_name || '-'} · 生成时间：${date} · 综合评分：${report.value.score} 分`
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

.report-body {
  padding: 24px 28px;
  line-height: 1.8;
}

.error-box {
  padding: 18px 22px;
  background: var(--c-p0-bg);
}
</style>
