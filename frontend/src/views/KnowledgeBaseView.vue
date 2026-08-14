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
            <td><span class="type-tag">{{ f.type.toUpperCase() }}</span></td>
            <td>{{ f.size }}</td>
            <td>
              <span class="status-dot" :class="f.status"></span>
              {{ statusText(f.status) }}
            </td>
            <td>{{ f.uploader }}</td>
            <td class="muted">{{ f.uploadTime }}</td>
            <td>
              <div class="ops">
                <button class="btn btn-sm" @click="reindex(f)" v-if="f.status !== 'pending'">重新索引</button>
                <button class="btn btn-sm" @click="remove(f.id)">删除</button>
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
import { ref, computed } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import { knowledgeBaseFiles as rawFiles } from '../mock/data'

const files = ref(rawFiles)
const fileInput = ref(null)

const indexedCount = computed(() => files.value.filter((f) => f.status === 'indexed').length)
const pendingCount = computed(() => files.value.filter((f) => f.status === 'pending').length)

const statusText = (status) => {
  const map = { indexed: '已索引', pending: '待索引', indexing: '索引中', failed: '索引失败' }
  return map[status] || status
}

const triggerUpload = () => {
  fileInput.value?.click()
}

const handleFiles = (e) => {
  const selected = Array.from(e.target.files || [])
  selected.forEach((file) => {
    const ext = file.name.split('.').pop().toLowerCase()
    const sizeText = file.size > 1024 * 1024
      ? `${(file.size / 1024 / 1024).toFixed(1)} MB`
      : `${Math.ceil(file.size / 1024)} KB`
    files.value.unshift({
      id: `kb-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name: file.name,
      type: ext,
      size: sizeText,
      status: 'pending',
      uploadTime: new Date().toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).replace(/\//g, '-'),
      uploader: '运维管理员'
    })
  })
  e.target.value = ''
}

const reindex = (f) => {
  f.status = 'indexing'
  setTimeout(() => {
    f.status = 'indexed'
  }, 1200)
}

const remove = (id) => {
  files.value = files.value.filter((f) => f.id !== id)
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
