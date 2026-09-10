<template>
  <div>
    <PageHeader :title="$t('notify.record.pageTitle')" :desc="$t('notify.record.pageDesc')" />

    <div class="card">
      <!-- 筛选栏 -->
      <div class="filters">
        <div class="filter-group">
          <label>{{ $t('notify.record.filters.status') }}</label>
          <select v-model="status" @change="search">
            <option value="">{{ $t('notify.record.filters.all') }}</option>
            <option value="success">{{ $t('notify.record.status.success') }}</option>
            <option value="failed">{{ $t('notify.record.status.failed') }}</option>
            <option value="intercepted">{{ $t('notify.record.status.intercepted') }}</option>
            <option value="skipped">{{ $t('notify.record.status.skipped') }}</option>
          </select>
        </div>
        <div class="filter-group">
          <label>{{ $t('notify.record.filters.strategy') }}</label>
          <select v-model="strategy" @change="search">
            <option value="">{{ $t('notify.record.filters.all') }}</option>
            <option value="window_aggregation">{{ $t('notify.record.strategy.windowAggregation') }}</option>
            <option value="topology_suppression">{{ $t('notify.record.strategy.topologySuppression') }}</option>
          </select>
        </div>
        <div class="search-box">
          <input
            v-model="keyword"
            type="text"
            :placeholder="$t('notify.record.filters.keywordPlaceholder')"
            @keyup.enter="search"
          />
          <button v-if="keyword" class="clear-btn" @click="clearKeyword">×</button>
        </div>
        <div class="filter-group">
          <label>{{ $t('notify.record.filters.startTime') }}</label>
          <input v-model="start" type="datetime-local" class="time-input" />
        </div>
        <div class="filter-group">
          <label>{{ $t('notify.record.filters.endTime') }}</label>
          <input v-model="end" type="datetime-local" class="time-input" />
        </div>
        <button class="btn btn-sm btn-primary" @click="search" :disabled="loading">{{ $t('notify.record.filters.search') }}</button>
        <button class="btn btn-sm" @click="reset" :disabled="loading">{{ $t('notify.record.filters.reset') }}</button>
      </div>

      <!-- 记录列表 -->
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('notify.record.table.time') }}</th>
            <th>{{ $t('notify.record.table.ruleName') }}</th>
            <th>{{ $t('notify.record.table.eventType') }}</th>
            <th>{{ $t('notify.record.table.target') }}</th>
            <th>{{ $t('notify.record.table.triggerValue') }}</th>
            <th>{{ $t('notify.record.table.status') }}</th>
            <th>{{ $t('notify.record.table.media') }}</th>
            <th>{{ $t('notify.record.table.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rec in records" :key="rec.id">
            <td class="muted">{{ fmtTime(rec.created_at) }}</td>
            <td>
              <b>{{ rec.rule_name }}</b>
              <LevelTag :level="severityMap(rec.severity)" class="sev-tag">
                {{ severityText[rec.severity] || 'P' + rec.severity }}
              </LevelTag>
            </td>
            <td>{{ rec.event_type === 'recovery' ? $t('notify.record.eventType.recovery') : $t('notify.record.eventType.alert') }}</td>
            <td class="mono">{{ rec.target_ident || '-' }}</td>
            <td class="mono">{{ rec.trigger_value || '-' }}</td>
            <td>
              <span class="status-tag" :style="statusStyle(rec.status)">
                {{ statusText(rec.status) }}
              </span>
            </td>
            <td>{{ rec.media_name || '-' }}</td>
            <td>
              <button class="btn btn-sm" @click="openDetail(rec)" :disabled="detailLoading === rec.id">
                {{ detailLoading === rec.id ? $t('notify.record.table.loading') : $t('notify.record.table.detail') }}
              </button>
            </td>
          </tr>
          <tr v-if="!loading && records.length === 0">
            <td colspan="8" class="empty-row">{{ $t('notify.record.table.empty') }}</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">{{ $t('notify.record.table.loading') }}</td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div class="pagination" v-if="total > 0">
        <span class="muted">{{ $t('notify.record.pagination.total', { count: total }) }}</span>
        <div class="page-ops">
          <select v-model.number="pageSize" class="page-size" @change="changePageSize">
            <option :value="10">{{ $t('notify.record.pagination.perPage', { count: 10 }) }}</option>
            <option :value="20">{{ $t('notify.record.pagination.perPage', { count: 20 }) }}</option>
            <option :value="50">{{ $t('notify.record.pagination.perPage', { count: 50 }) }}</option>
            <option :value="100">{{ $t('notify.record.pagination.perPage', { count: 100 }) }}</option>
          </select>
          <button class="btn btn-sm" :disabled="page === 1 || loading" @click="changePage(page - 1)">{{ $t('notify.record.pagination.prev') }}</button>
          <span class="page-info">{{ $t('notify.record.pagination.pageInfo', { page, pages: totalPages }) }}</span>
          <button class="btn btn-sm" :disabled="page >= totalPages || loading" @click="changePage(page + 1)">{{ $t('notify.record.pagination.next') }}</button>
        </div>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <div v-if="detail" class="modal-mask" @click.self="closeDetail">
      <div class="modal modal-lg">
        <div class="modal-head">
          <h3>{{ $t('notify.record.detail.title') }}</h3>
          <button class="modal-close" @click="closeDetail">×</button>
        </div>
        <div class="modal-body">
          <!-- 状态与原因 -->
          <div class="detail-block reason-block" :class="detail.status">
            <div class="reason-head">
              <span class="status-tag" :style="statusStyle(detail.status)">{{ statusText(detail.status) }}</span>
              <template v-if="detail.status === 'intercepted'">
                <span class="strategy-tag">{{ strategyText(detail.strategy) }}</span>
              </template>
            </div>
            <p v-if="detail.reason" class="reason-text">{{ detail.reason }}</p>
          </div>

          <!-- 基本信息 -->
          <div class="detail-block">
            <h4 class="block-title">{{ $t('notify.record.detail.basicInfo') }}</h4>
            <div class="info-grid">
              <div class="info-item"><label>{{ $t('notify.record.detail.ruleName') }}</label><span>{{ detail.rule_name || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.severity') }}</label><span>{{ severityText[detail.severity] || 'P' + detail.severity || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.eventType') }}</label><span>{{ detail.event_type === 'recovery' ? $t('notify.record.eventType.recovery') : $t('notify.record.eventType.alert') }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.target') }}</label><span class="mono">{{ detail.target_ident || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.triggerValue') }}</label><span class="mono">{{ detail.trigger_value || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.triggerTime') }}</label><span>{{ fmtTime(detail.trigger_time) }}</span></div>
              <div class="info-item info-wide"><label>{{ $t('notify.record.detail.tags') }}</label><span class="mono">{{ detail.tags || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.notifyTime') }}</label><span>{{ fmtTime(detail.created_at) }}</span></div>
            </div>
          </div>

          <!-- 通知链路 -->
          <div class="detail-block">
            <h4 class="block-title">{{ $t('notify.record.detail.notifyChain') }}</h4>
            <div class="info-grid">
              <div class="info-item"><label>{{ $t('notify.record.detail.notifyRule') }}</label><span>{{ detail.notify_rule_name || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.media') }}</label><span>{{ detail.media_name || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.mediaType') }}</label><span>{{ detail.media_type || '-' }}</span></div>
              <div class="info-item"><label>{{ $t('notify.record.detail.template') }}</label><span>{{ detail.template_name || '-' }}</span></div>
            </div>
          </div>

          <!-- 消息正文 -->
          <div class="detail-block">
            <h4 class="block-title">{{ $t('notify.record.detail.messageContent') }}</h4>
            <pre v-if="detail.content" class="content-pre">{{ detail.content }}</pre>
            <p v-else class="muted content-empty">{{ $t('notify.record.detail.noContent') }}</p>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn" @click="closeDetail">{{ $t('notify.record.detail.close') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import LevelTag from '../../components/LevelTag.vue'
import { notifyRecordApi } from '../../api/notifyRecord.js'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute()

const status = ref('')
const strategy = ref('')
const keyword = ref('')
const start = ref('')
const end = ref('')
const page = ref(1)
const pageSize = ref(20)

const records = ref([])
const total = ref(0)
const loading = ref(false)

const detail = ref(null)
const detailLoading = ref('')

const severityText = { 1: 'P1', 2: 'P2', 3: 'P3' }

const severityMap = (s) => {
  if (s === 1) return 'p1'
  if (s === 2) return 'p2'
  return 'p3'
}

const STATUS_COLORS = {
  success: ['var(--c-success)', 'var(--c-success-bg)'],
  failed: ['var(--c-danger)', 'var(--c-p0-bg)'],
  intercepted: ['var(--c-p1)', 'var(--c-p1-bg)'],
  skipped: ['var(--c-p4)', 'var(--c-p4-bg)']
}

const statusText = (s) => ({
  success: t('notify.record.status.success'),
  failed: t('notify.record.status.failed'),
  intercepted: t('notify.record.status.intercepted'),
  skipped: t('notify.record.status.skipped')
}[s] || s || '-')
const statusStyle = (s) => {
  const conf = STATUS_COLORS[s] || STATUS_COLORS.skipped
  return { color: conf[0], background: conf[1] }
}

const strategyText = (s) => ({
  window_aggregation: t('notify.record.strategy.windowAggregation'),
  topology_suppression: t('notify.record.strategy.topologySuppression')
}[s] || s || '-')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

// 兼容数字（unix 秒时间戳）与字符串两种时间格式
function fmtTime(ts) {
  if (!ts) return '-'
  if (typeof ts === 'number') return new Date(ts * 1000).toLocaleString('zh-CN')
  const d = new Date(ts)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

async function loadData() {
  loading.value = true
  try {
    const result = await notifyRecordApi.list({
      status: status.value,
      strategy: strategy.value,
      keyword: keyword.value.trim(),
      start: start.value,
      end: end.value,
      page: page.value,
      page_size: pageSize.value
    })
    records.value = result?.list || []
    total.value = result?.total || 0
  } catch (err) {
    alert(t('notify.record.error.loadFailed') + err.message)
    records.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  loadData()
}

function reset() {
  status.value = ''
  strategy.value = ''
  keyword.value = ''
  start.value = ''
  end.value = ''
  page.value = 1
  loadData()
}

function clearKeyword() {
  keyword.value = ''
  page.value = 1
  loadData()
}

function changePage(next) {
  page.value = next
  loadData()
}

function changePageSize() {
  page.value = 1
  loadData()
}

async function openDetail(rec) {
  detailLoading.value = rec.id
  try {
    detail.value = await notifyRecordApi.get(rec.id)
  } catch (err) {
    alert(t('notify.record.error.loadDetailFailed') + err.message)
  } finally {
    detailLoading.value = ''
  }
}

function closeDetail() {
  detail.value = null
}

onMounted(() => {
  // 支持外部跳转带筛选条件（如告警降噪页「今日已拦截」跳转）：
  // /notify/records?status=intercepted&strategy=window_aggregation&start=...&end=...
  const q = route.query
  if (q.status) status.value = String(q.status)
  if (q.strategy) strategy.value = String(q.strategy)
  if (q.start) start.value = String(q.start)
  if (q.end) end.value = String(q.end)
  loadData()
})
</script>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
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

.filter-group select,
.time-input {
  padding: 6px 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  outline: none;
}

.time-input { font-family: inherit; }

.search-box {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 6px 12px;
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

.search-box .clear-btn:hover { color: var(--c-text); }

.sev-tag { margin-left: 6px; }

.status-tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
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

.page-size {
  padding: 4px 6px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 12px;
  outline: none;
}

.page-info {
  font-size: 13px;
  color: var(--c-text-2);
}

/* 详情弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  background: var(--c-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--c-surface);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-modal);
}

.modal-lg { width: 720px; max-height: 88vh; }

.modal-head {
  padding: 14px 20px;
  border-bottom: 1px solid var(--c-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-head h3 { margin: 0; font-size: 15px; }

.modal-close {
  border: none;
  background: none;
  font-size: 20px;
  line-height: 1;
  color: var(--c-text-3);
  cursor: pointer;
}

.modal-close:hover { color: var(--c-text); }

.modal-body {
  padding: 16px 20px;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-foot {
  padding: 12px 20px;
  border-top: 1px solid var(--c-border);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.detail-block {
  border: 1px solid var(--c-border);
  border-radius: var(--radius-card);
  padding: 12px 14px;
}

.block-title {
  margin: 0 0 10px;
  font-size: 13.5px;
  font-weight: 600;
  color: var(--c-text);
}

.reason-block { background: var(--c-primary-soft); }
.reason-block.failed { background: var(--c-p0-bg); }
.reason-block.intercepted { background: var(--c-p1-bg); }

.reason-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.strategy-tag {
  font-size: 12px;
  font-weight: 600;
  color: var(--c-p1);
}

.reason-text {
  margin: 8px 0 0;
  font-size: 12.5px;
  color: var(--c-text-2);
  line-height: 1.6;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 20px;
}

.info-item {
  display: flex;
  align-items: baseline;
  gap: 10px;
  font-size: 12.5px;
}

.info-item label {
  color: var(--c-text-3);
  white-space: nowrap;
  flex-shrink: 0;
}

.info-item span {
  color: var(--c-text);
  word-break: break-all;
}

.info-wide { grid-column: 1 / -1; }

.content-pre {
  margin: 0;
  padding: 10px 12px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  font-size: 12.5px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow-y: auto;
}

.content-empty {
  margin: 0;
  padding: 12px 0;
  text-align: center;
}

@media (max-width: 900px) {
  .info-grid { grid-template-columns: 1fr; }
  .modal-lg { width: 92vw; }
}
</style>
