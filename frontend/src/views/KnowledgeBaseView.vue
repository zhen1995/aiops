<template>
  <div>
    <PageHeader :title="$t('knowledge.header.title')" :desc="$t('knowledge.header.desc')">
      <button class="btn btn-primary" @click="triggerUpload">
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        {{ $t('knowledge.header.upload') }}
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
      <StatCard :label="$t('knowledge.stats.total')" :value="files.length" :unit="$t('knowledge.stats.unit')" icon="file">
        <template #icon>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>
          </svg>
        </template>
      </StatCard>
      <StatCard :label="$t('knowledge.stats.indexed')" :value="indexedCount" :unit="$t('knowledge.stats.unit')" icon="check">
        <template #icon>
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
        </template>
      </StatCard>
      <StatCard :label="$t('knowledge.stats.pending')" :value="pendingCount" :unit="$t('knowledge.stats.unit')" icon="clock">
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
          <h3 class="card-title">{{ $t('knowledge.list.title') }}</h3>
          <p class="card-sub">{{ $t('knowledge.list.sub') }}</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('knowledge.table.fileName') }}</th>
            <th>{{ $t('knowledge.table.type') }}</th>
            <th>{{ $t('knowledge.table.size') }}</th>
            <th>{{ $t('knowledge.table.status') }}</th>
            <th>{{ $t('knowledge.table.uploader') }}</th>
            <th>{{ $t('knowledge.table.uploadTime') }}</th>
            <th>{{ $t('knowledge.table.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="f in filteredFiles"
            :key="f.id"
            :class="{ 'row-highlight': f.id === highlightDocId }"
            :data-doc-id="f.id"
          >
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
                <button class="btn btn-sm" @click="reindex(f)" v-if="f.status === 'indexed' || f.status === 'failed'">{{ $t('knowledge.table.reindex') }}</button>
                <button class="btn btn-sm" @click="removeFile(f.id)">{{ $t('knowledge.table.delete') }}</button>
              </div>
            </td>
          </tr>
          <tr v-if="filteredFiles.length === 0">
            <td colspan="7" class="empty">{{ filterKw ? $t('knowledge.table.emptyNoMatch') : $t('knowledge.table.emptyNoDocs') }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import { knowledgeApi } from '../api/knowledge.js'

const { t } = useI18n({ useScope: 'global' })

const route = useRoute()
const router = useRouter()

const files = ref([])
const fileInput = ref(null)
let pollTimer = null

// 全局搜索落地：/knowledge-base?doc=<文档ID>&q=标题，按关键字过滤并高亮该行
const highlightDocId = ref('')
const filterKw = ref('')

const filteredFiles = computed(() => {
  const kw = filterKw.value.trim().toLowerCase()
  if (!kw) return files.value
  return files.value.filter(f => (f.name || '').toLowerCase().includes(kw))
})

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
  const map = {
    indexed: t('knowledge.status.indexed'),
    pending: t('knowledge.status.pending'),
    indexing: t('knowledge.status.indexing'),
    failed: t('knowledge.status.failed')
  }
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

onMounted(async () => {
  await loadFiles()
  handleSearchParam()
})

// 全局搜索落地：/knowledge-base?doc=<文档ID>&q=标题
// 知识库页面为文件列表（无文档详情视图），退化为：按标题关键字过滤 + 滚动高亮目标行
function handleSearchParam() {
  const { doc, q } = route.query
  if (!doc && !q) return
  if (q) filterKw.value = String(q)
  if (doc) highlightDocId.value = String(doc)
  router.replace({ path: '/knowledge-base', query: {} })
  if (doc) {
    nextTick(() => {
      const el = document.querySelector(`tr[data-doc-id="${CSS.escape(String(doc))}"]`)
      if (el) el.scrollIntoView({ block: 'center', behavior: 'smooth' })
    })
    setTimeout(() => { highlightDocId.value = '' }, 3000)
  }
}
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
      alert(t('knowledge.alert.partialUploadFailed') + '\n' + failed.map((f) => `${f.name}：${f.error}`).join('\n'))
    }
  } catch (err) {
    alert(err.message || t('knowledge.alert.uploadFailed'))
  }
  e.target.value = ''
}

const reindex = async (f) => {
  try {
    await knowledgeApi.reindex(f.id)
    await loadFiles()
  } catch (err) {
    alert(err.message || t('knowledge.alert.reindexFailed'))
  }
}

const removeFile = async (id) => {
  if (!confirm(t('knowledge.alert.deleteConfirm'))) return
  try {
    await knowledgeApi.remove(id)
    await loadFiles()
  } catch (err) {
    alert(err.message || t('knowledge.alert.deleteFailed'))
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

tr.row-highlight td {
  background: var(--c-primary-tint);
  animation: row-flash 1s ease-in-out 2;
}

@keyframes row-flash {
  0%, 100% { background: var(--c-primary-tint); }
  50% { background: var(--c-primary-soft); }
}

.empty {
  text-align: center;
  color: var(--c-text-3);
  padding: 32px 12px;
}
</style>
