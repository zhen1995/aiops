<template>
  <div>
    <PageHeader title="运维知识库" desc="管理运维文档、Runbook、RCA 报告等知识库文件，支持 Word、PDF、Markdown、TXT 上传与维护">
      <button class="btn btn-primary" @click="triggerUpload">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        上传文档
      </button>
      <input
        ref="fileInput"
        type="file"
        accept=".doc,.docx,.pdf,.md,.txt"
        multiple
        style="display: none"
        @change="handleFiles"
      />
    </PageHeader>

    <div class="stat-row">
      <StatCard label="知识库文档" :value="files.length" unit="个" icon="file">
        <template #icon>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>
          </svg>
        </template>
      </StatCard>
      <StatCard label="已索引" :value="indexedCount" unit="个" icon="check">
        <template #icon>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </template>
      </StatCard>
      <StatCard label="待处理" :value="pendingCount" unit="个" icon="clock">
        <template #icon>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
        </template>
      </StatCard>
    </div>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">文档列表</h3>
          <p class="card-sub">已上传的运维知识库文件，向量化后即可被 AI 助手检索引用</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>文件名</th>
            <th>类型</th>
            <th>大小</th>
            <th>状态</th>
            <th>上传人</th>
            <th>上传时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="f in files" :key="f.id">
            <td>
              <div class="file-name">
                <svg class="file-icon" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>
                </svg>
                <b>{{ f.name }}</b>
              </div>
            </td>
            <td><span class="type-tag">{{ (f.type || '').toUpperCase() }}</span></td>
            <td>{{ formatSize(f.size) }}</td>
            <td>
              <span class="status-dot" :class="f.status"></span>
              {{ statusText(f.status) }}
            </td>
            <td>{{ f.uploader || '-' }}</td>
            <td class="muted">{{ formatTime(f.created_at) }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="reindex(f)" v-if="f.status === 'indexed' || f.status === 'failed'">重新索引</button>
                <button class="btn btn-sm" @click="removeFile(f.id)">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="files.length === 0">
            <td colspan="7" class="empty">暂无文档，点击右上角“上传文档”按钮添加</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import { knowledgeApi } from '../api/knowledge.js'

const files = ref([])
const fileInput = ref(null)
let pollTimer = null

const indexedCount = computed(() => files.value.filter((f) => f.status === 'indexed').length)
const pendingCount = computed(() =>
  files.value.filter((f) => f.status === 'pending' || f.status === 'indexing').length
)

// 列表中存在待处理/索引中的文件时轮询刷新
const hasProcessing = computed(() =>
  files.value.some((f) => f.status === 'indexing' || f.status === 'pending')
)

watch(hasProcessing, (processing) => {
  if (processing && !pollTimer) {
    pollTimer = setInterval(loadFiles, 2000)
  } else if (!processing && pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})

const statusText = (status) => {
  const map = { indexed: '已索引', pending: '待索引', indexing: '索引中', failed: '索引失败' }
  return map[status] || status
}

const formatSize = (bytes) => {
  const n = Number(bytes) || 0
  return n < 1024 * 1024
    ? `${Math.max(1, Math.ceil(n / 1024))} KB`
    : `${(n / 1024 / 1024).toFixed(1)} MB`
}

const formatTime = (t) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN', { hour12: false })
}

async function loadFiles() {
  try {
    files.value = await knowledgeApi.list()
  } catch (e) {
    console.error('加载知识库列表失败', e)
  }
}

onMounted(loadFiles)
onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleFiles = async (e) => {
  const selected = e.target.files
  if (!selected || selected.length === 0) return
  try {
    const result = await knowledgeApi.upload(selected)
    await loadFiles()
    const failed = result?.failed || []
    if (failed.length > 0) {
      alert('部分文件上传失败：\n' + failed.map((f) => `${f.name}：${f.error}`).join('\n'))
    }
  } catch (err) {
    alert(err.message || '上传失败')
  }
  e.target.value = ''
}

const reindex = async (f) => {
  try {
    await knowledgeApi.reindex(f.id)
    await loadFiles()
  } catch (err) {
    alert(err.message || '重新索引失败')
  }
}

const removeFile = async (id) => {
  if (!confirm('确定删除该文档吗？删除后不可恢复。')) return
  try {
    await knowledgeApi.remove(id)
    await loadFiles()
  } catch (err) {
    alert(err.message || '删除失败')
  }
}
</script>

<style scoped>
.stat-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 18px;
}

.card-head-flex {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 14px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  color: var(--c-primary);
  flex-shrink: 0;
}

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  background: var(--c-primary-soft);
  color: var(--c-primary);
  font-size: 11px;
  font-weight: 600;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}

.status-dot.indexed { background: var(--c-success); }
.status-dot.pending { background: var(--c-p2); }
.status-dot.indexing { background: var(--c-primary); }
.status-dot.failed { background: var(--c-danger); }

.ops { display: flex; gap: 8px; }

.empty {
  text-align: center;
  color: var(--c-text-3);
  padding: 32px 12px;
}
</style>
