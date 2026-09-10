<template>
  <div>
    <PageHeader :title="$t('n9e.rules.title')" :desc="$t('n9e.rules.desc')">
      <button class="btn" @click="load" :disabled="loading">{{ $t('n9e.rules.refresh') }}</button>
    </PageHeader>

    <div v-if="message" class="banner" :class="messageType">{{ message }}</div>

    <div class="card" v-if="rules.length > 0">
      <table class="table">
        <thead>
          <tr>
            <th style="width:50px">#</th>
            <th>{{ $t('n9e.rules.colName') }}</th>
            <th style="width:120px">{{ $t('n9e.rules.colGroup') }}</th>
            <th style="width:90px">{{ $t('n9e.rules.colLevel') }}</th>
            <th style="width:100px">{{ $t('n9e.rules.colDuration') }}</th>
            <th>{{ $t('n9e.rules.colPromql') }}</th>
            <th style="width:80px">{{ $t('n9e.rules.colEnabled') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(r, i) in rules" :key="keyOf(r, i)">
            <td class="muted">{{ i + 1 }}</td>
            <td><b>{{ r.name || r.cate_name || '-' }}</b></td>
            <td class="muted">{{ groupOf(r) }}</td>
            <td><span class="level">{{ severityOf(r) }}</span></td>
            <td class="muted">{{ durationOf(r) }}</td>
            <td class="mono" :title="promqlOf(r)">{{ truncate(promqlOf(r), 60) }}</td>
            <td>
              <span class="tag" :class="enabledOf(r) ? 'ok' : 'off'">
                {{ $t(enabledOf(r) ? 'n9e.rules.yes' : 'n9e.rules.no') }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else-if="!loading" class="empty-card">
      <p class="empty-title">{{ $t('n9e.rules.emptyTitle') }}</p>
      <p class="empty-sub">{{ $t('n9e.rules.emptySub') }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '../../components/PageHeader.vue'
import { n9eApi } from '../../api/n9e.js'

const { t } = useI18n({ useScope: 'global' })

const loading = ref(false)
const rules = ref([])
const message = ref('')
const messageType = ref('info')

// 业务组 id → name 映射
const groupMap = reactive({})

onMounted(async () => {
  await Promise.all([loadGroups(), load()])
})

async function loadGroups() {
  try {
    const raw = await n9eApi.getBusiGroups()
    const list = extractList(raw)
    for (const g of list) {
      if (g.id !== undefined && g.name !== undefined) {
        groupMap[g.id] = g.name
      }
    }
  } catch (e) {
    // 业务组加载失败不阻塞规则展示
    console.warn('加载业务组失败:', e.message)
  }
}

async function load() {
  loading.value = true
  message.value = ''
  try {
    const raw = await n9eApi.getAlertRules()
    rules.value = extractList(raw)
  } catch (e) {
    messageType.value = 'error'
    message.value = e.message
    rules.value = []
  } finally {
    loading.value = false
  }
}

function extractList(raw) {
  if (Array.isArray(raw)) return raw
  if (raw && typeof raw === 'object') {
    for (const topKey of ['data', 'dat', 'list', 'items']) {
      const v = raw[topKey]
      if (Array.isArray(v)) return v
      if (v && typeof v === 'object') {
        for (const subKey of ['list', 'data', 'items', 'records']) {
          const sv = v[subKey]
          if (Array.isArray(sv)) return sv
        }
      }
    }
    for (const v of Object.values(raw)) {
      if (Array.isArray(v)) return v
      if (v && typeof v === 'object') {
        for (const sv of Object.values(v)) {
          if (Array.isArray(sv)) return sv
        }
      }
    }
  }
  return []
}

function keyOf(r, i) { return r.id ?? r.rule_id ?? r.ruleId ?? r.ident ?? i }

function groupOf(r) {
  if (r.group_id !== undefined && groupMap[r.group_id]) {
    return groupMap[r.group_id]
  }
  if (r.busi_group_name) return r.busi_group_name
  if (r.busi_group) return r.busi_group
  if (r.group) return r.group
  return '-'
}

function severityOf(r) {
  // 夜莺 severities 数组是每个 query 的级别：1=P3, 2=P2, 3=P1
  if (Array.isArray(r.severities) && r.severities.length > 0) {
    const lv = r.severities[0]
    return { 1: t('n9e.rules.severityP3'), 2: t('n9e.rules.severityP2'), 3: t('n9e.rules.severityP1') }[lv] ?? `P${lv}`
  }
  const lv = r.severity
  if (lv === undefined || lv === null) return '-'
  if (typeof lv === 'number') {
    return { 0: t('n9e.rules.severityP3'), 1: t('n9e.rules.severityP2'), 2: t('n9e.rules.severityP1') }[lv] ?? `P${lv}`
  }
  return String(lv)
}

function promqlOf(r) {
  // 夜莺 v6+ 把 PromQL 放在 rule_config.queries 里
  if (r.rule_config && Array.isArray(r.rule_config.queries) && r.rule_config.queries.length > 0) {
    const q = r.rule_config.queries[0]
    if (q.prom_ql) return q.prom_ql
    if (q.promQL) return q.promQL
  }
  // 回退到顶层字段（兼容旧版本）
  return r.prom_ql ?? r.promQL ?? r.promql ?? r.rule ?? r.expr ?? ''
}

function durationOf(r) {
  // rule_config.queries[0].recover_config 不是持续时长
  if (r.prom_for_duration !== undefined) return r.prom_for_duration
  return r.duration ?? '-'
}

function enabledOf(r) {
  if (r.enabled !== undefined) return !!r.enabled
  if (r.is_enabled !== undefined) return !!r.is_enabled
  if (r.disable !== undefined) return !r.disable
  return true
}

function truncate(s, n) {
  if (!s) return '-'
  return s.length > n ? s.slice(0, n) + '…' : s
}
</script>

<style scoped>
.banner {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 12px;
  border: 1px solid transparent;
}
.banner.info { background: var(--c-primary-soft, rgba(22,119,255,0.08)); color: var(--c-primary); border-color: #bde2ff; }
.banner.error { background: var(--c-p0-bg, #ffe5e5); color: var(--c-danger, #c93b3b); border-color: #f0d0d0; }

.mono { font-family: ui-monospace, Menlo, Monaco, Consolas, monospace; font-size: 12px; max-width: 360px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.muted { color: var(--c-text-3); }

.level {
  padding: 2px 10px; border-radius: 10px; font-size: 11.5px; font-weight: 500;
  background: var(--c-primary-soft); color: var(--c-primary);
}

.tag { padding: 2px 8px; border-radius: 10px; font-size: 11.5px; }
.tag.ok { background: var(--c-success-bg, #e8f5ee); color: #1f7a45; }
.tag.off { background: #f0f0f0; color: var(--c-text-3); }

.empty-card {
  padding: 60px 20px; text-align: center; color: var(--c-text-3);
  border: 1px dashed var(--c-border); border-radius: 10px;
}
.empty-title { font-size: 15px; font-weight: 500; color: var(--c-text-2); margin-bottom: 6px; }
.empty-sub { font-size: 12.5px; }

.table th { font-size: 12px; color: var(--c-text-3); font-weight: 500; text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--c-border); }
.table td { padding: 10px 12px; font-size: 13px; border-bottom: 1px solid var(--c-border); }
.table tr:last-child td { border-bottom: none; }
</style>
