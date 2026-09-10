<template>
  <div ref="boxRef" class="gs-box">
    <div class="gs-input" :class="{ focused: panelOpen }">
      <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" />
      </svg>
      <input
        v-model="keyword"
        :placeholder="$t('common.search.placeholder')"
        @focus="onFocus"
        @keydown="onKeydown"
      />
      <button v-if="keyword" class="gs-clear" @click="clearKeyword">×</button>
    </div>

    <div v-if="panelOpen" class="gs-panel">
      <!-- 空输入：最近搜索 + 快捷入口 -->
      <template v-if="!keyword.trim()">
        <div v-if="recents.length" class="gs-section">
          <div class="gs-sec-head">
            <span>{{ $t('common.search.recent') }}</span>
            <button class="gs-link-btn" @click="clearRecents">{{ $t('common.search.clear') }}</button>
          </div>
          <div class="gs-recent-list">
            <span v-for="(r, i) in recents" :key="r + i" class="gs-chip" @click="applyRecent(r)">
              {{ r }}
              <em class="gs-chip-remove" @click.stop="removeRecent(i)">×</em>
            </span>
          </div>
        </div>
        <div class="gs-section">
          <div class="gs-sec-head"><span>{{ $t('common.search.quick') }}</span></div>
          <div
            v-for="item in quickEntries"
            :key="item.label"
            class="gs-item"
            @click="openQuick(item)"
          >
            <span class="gs-icon gs-icon-quick" v-html="item.icon"></span>
            <span class="gs-main">{{ item.label }}</span>
            <span class="gs-sub">{{ $t('common.search.quick') }}</span>
          </div>
        </div>
      </template>

      <!-- 搜索中 -->
      <div v-else-if="searching" class="gs-empty">{{ $t('common.search.searching') }}</div>

      <!-- 搜索结果 -->
      <template v-else-if="resultGroups.length">
        <div v-for="g in resultGroups" :key="g.type" class="gs-section">
          <div class="gs-sec-head">
            <span>{{ g.label }}</span>
            <span class="gs-count">{{ g.items.length }}</span>
          </div>
          <div
            v-for="(item, idx) in g.items"
            :key="g.type + item.id"
            class="gs-item"
            :class="{ active: flatIndex(g, idx) === activeIndex }"
            @click="openItem(item)"
            @mouseenter="activeIndex = flatIndex(g, idx)"
          >
            <span class="gs-icon" :class="'gs-icon-' + item.type" v-html="item.icon"></span>
            <span class="gs-body">
              <span class="gs-main" v-html="item.titleHtml"></span>
              <span class="gs-sub" v-if="item.subHtml" v-html="item.subHtml"></span>
              <span v-if="item.type === 'alert'" class="gs-tags">
                <LevelTag :level="item.severityLevel">{{ item.severityText }}</LevelTag>
                <LevelTag :level="item.status === 'firing' ? 'active' : 'resolved'">
                  {{ item.status === 'firing' ? $t('common.search.firing') : $t('common.search.resolved') }}
                </LevelTag>
                <span class="gs-time">{{ relTime(item.triggerTime) }}</span>
              </span>
            </span>
          </div>
          <div v-if="g.more" class="gs-more" @click="openGroupAll(g)">
            {{ $t('common.search.viewAll', { label: g.label }) }}
          </div>
        </div>
        <div class="gs-foot">{{ $t('common.search.footerHint') }}</div>
      </template>

      <!-- 无结果 -->
      <template v-else-if="searched">
        <div class="gs-empty">{{ $t('common.search.noResults') }}</div>
        <div class="gs-item gs-ai" @click="askAi">
          <span class="gs-icon gs-icon-ai">AI</span>
          <span class="gs-main">{{ $t('common.search.askAi', { keyword: keyword.trim() }) }}</span>
          <span class="gs-sub">{{ $t('common.search.goToChat') }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { searchApi } from '../api/search.js'
import { getPermissions } from '../utils/auth.js'
import LevelTag from './LevelTag.vue'

const { t } = useI18n({ useScope: 'global' })
const router = useRouter()

const RECENT_KEY = 'aiops_recent_searches'
const MAX_RECENT = 8

const boxRef = ref(null)
const keyword = ref('')
const panelOpen = ref(false)
const searching = ref(false)
const searched = ref(false)
const results = ref(null)
const activeIndex = ref(-1)
const recents = ref(loadRecents())

const quickEntries = computed(() => [
  {
    label: t('common.quick.todayP0'),
    path: '/alerts/events',
    icon: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/></svg>'
  },
  {
    label: t('common.quick.rca'),
    path: '/rca',
    icon: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><circle cx="12" cy="12" r="8"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2"/></svg>'
  },
  {
    label: t('common.quick.inspectionSummary'),
    path: '/inspection/reports',
    icon: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/></svg>'
  }
])

// 可搜索的本地页面（auth 为权限名，保持中文不翻译；title 随语言切换）
const allPages = computed(() => [
  { title: t('common.pages.aiChat'), path: '/chat', auth: '对话' },
  { title: t('common.pages.dashboard'), path: '/dashboard', auth: '总览大盘' },
  { title: t('common.pages.serviceRegistry'), path: '/services', auth: '服务注册' },
  { title: t('common.pages.alertDenoise'), path: '/alerts/denoise', auth: '告警降噪' },
  { title: t('common.pages.alertEvents'), path: '/alerts/events', auth: '告警事件' },
  { title: t('common.pages.alertRules'), path: '/alert-rules', auth: '告警规则' },
  { title: t('common.pages.messageTemplates'), path: '/notify/templates', auth: '消息模板' },
  { title: t('common.pages.notifyRules'), path: '/notify/rules', auth: '通知规则' },
  { title: t('common.pages.engineConfig'), path: '/n9e/config', auth: '夜莺引擎配置' },
  { title: t('common.pages.n9eAlertRules'), path: '/n9e/alert-rules', auth: '夜莺告警规则' },
  { title: t('common.pages.n9eAlertEvents'), path: '/n9e/alert-events', auth: '夜莺告警事件' },
  { title: t('common.pages.rca'), path: '/rca', auth: '根因分析' },
  { title: t('common.pages.logAnalysis'), path: '/logs', auth: '日志分析' },
  { title: t('common.pages.llmMgmt'), path: '/ai-config/llm', auth: 'LLM 管理' },
  { title: t('common.pages.skillMgmt'), path: '/ai-config/skill', auth: 'Skill 管理' },
  { title: t('common.pages.knowledgeBase'), path: '/knowledge-base', auth: '运维知识库' },
  { title: t('common.pages.inspectionTasks'), path: '/inspection/tasks', auth: '巡检任务' },
  { title: t('common.pages.inspectionReports'), path: '/inspection/reports', auth: '巡检报告' },
  { title: t('common.pages.notifyMedium'), path: '/notification/medium', auth: '通知媒介' },
  { title: t('common.pages.userMgmt'), path: '/org/user', auth: '用户管理' },
  { title: t('common.pages.roleMgmt'), path: '/org/role', auth: '角色管理' },
  { title: t('common.pages.datasource'), path: '/datasource', auth: '数据源接入' }
])

const severityTextMap = { 1: 'P1', 2: 'P2', 3: 'P3' }

const icons = {
  menu: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18"/></svg>',
  service: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><path d="M6 6.5h.01M6 17.5h.01"/></svg>',
  alert: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/></svg>',
  rule: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 15s1-2 4-2 5 2 8 2 4-2 4-2"/><path d="M4 9s1-2 4-2 5 2 8 2 4-2 4-2"/></svg>',
  doc: '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/></svg>'
}

// ---------- 结果组装 ----------

const resultGroups = computed(() => {
  const kw = keyword.value.trim()
  if (!kw || !results.value) return []

  const perms = getPermissions() || []
  const lower = kw.toLowerCase()
  const menuItems = allPages.value
    .filter(p => (!p.auth || perms.includes(p.auth)) && p.title.toLowerCase().includes(lower))
    .slice(0, 5)
    .map(p => ({
      type: 'menu',
      id: p.path,
      titleHtml: hl(p.title),
      subHtml: t('common.search.menuLabel'),
      path: p.path
    }))

  const groups = []
  if (menuItems.length) groups.push({ type: 'menu', label: t('common.search.menuLabel'), items: menuItems, more: false })

  const data = results.value
  if (data.services?.length) {
    groups.push({
      type: 'service',
      label: t('common.search.serviceLabel'),
      more: false,
      items: data.services.map(s => ({
        type: 'service',
        id: s.id,
        titleHtml: hl(s.matched || s.name),
        subHtml: s.code ? `${escapeHtml(s.code)} · ${t('common.search.serviceLabel')}` : t('common.search.serviceLabel'),
        raw: s
      }))
    })
  }
  if (data.alerts?.length) {
    groups.push({
      type: 'alert',
      label: t('common.search.alertLabel'),
      more: data.alerts.length >= 5,
      items: data.alerts.map(a => ({
        type: 'alert',
        id: a.id,
        titleHtml: hl(a.matched || a.rule_name),
        subHtml: '',
        severityLevel: 'p' + (a.severity || 3),
        severityText: severityTextMap[a.severity] || ('P' + a.severity),
        status: a.status,
        triggerTime: a.trigger_time,
        raw: a
      }))
    })
  }
  if (data.rules?.length) {
    groups.push({
      type: 'rule',
      label: t('common.search.ruleLabel'),
      more: data.rules.length >= 5,
      items: data.rules.map(r => ({
        type: 'rule',
        id: r.id,
        titleHtml: hl(r.matched || r.name),
        subHtml: r.promql ? escapeHtml(truncate(r.promql, 60)) : '',
        raw: r
      }))
    })
  }
  if (data.documents?.length) {
    groups.push({
      type: 'doc',
      label: t('common.search.docLabel'),
      more: data.documents.length >= 5,
      items: data.documents.map(d => ({
        type: 'doc',
        id: d.id,
        titleHtml: hl(d.matched || d.title),
        subHtml: t('common.search.knowledgeBase'),
        raw: d
      }))
    })
  }
  return groups
})

const flatItems = computed(() => resultGroups.value.flatMap(g => g.items))

function flatIndex(g, idx) {
  let base = 0
  for (const group of resultGroups.value) {
    if (group === g) return base + idx
    base += group.items.length
  }
  return -1
}

// ---------- 高亮与工具 ----------

function escapeHtml(s) {
  return String(s ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function stripHtml(s) {
  return String(s ?? '').replace(/<[^>]*>/g, '')
}

function hl(text) {
  const t = stripHtml(text)
  const kw = keyword.value.trim()
  if (!kw) return escapeHtml(t)
  const lower = t.toLowerCase()
  const k = kw.toLowerCase()
  const idx = lower.indexOf(k)
  if (idx === -1) return escapeHtml(t)
  return (
    escapeHtml(t.slice(0, idx)) +
    '<em>' + escapeHtml(t.slice(idx, idx + k.length)) + '</em>' +
    escapeHtml(t.slice(idx + k.length))
  )
}

function truncate(s, n) {
  const str = String(s ?? '')
  return str.length > n ? str.slice(0, n) + '…' : str
}

function relTime(ts) {
  if (!ts) return ''
  const t = typeof ts === 'number' ? ts * 1000 : new Date(ts).getTime()
  if (isNaN(t)) return ''
  const diff = Date.now() - t
  if (diff < 60 * 1000) return t('common.search.justNow')
  if (diff < 3600 * 1000) return t('common.search.minutesAgo', { n: Math.floor(diff / 60000) })
  if (diff < 86400 * 1000) return t('common.search.hoursAgo', { n: Math.floor(diff / 3600000) })
  return t('common.search.daysAgo', { n: Math.floor(diff / 86400000) })
}

// ---------- 最近搜索 ----------

function loadRecents() {
  try {
    const arr = JSON.parse(localStorage.getItem(RECENT_KEY) || '[]')
    return Array.isArray(arr) ? arr.filter(x => typeof x === 'string') : []
  } catch {
    return []
  }
}

function saveRecents() {
  localStorage.setItem(RECENT_KEY, JSON.stringify(recents.value))
}

function addRecent(kw) {
  const k = kw.trim()
  if (!k) return
  recents.value = [k, ...recents.value.filter(x => x !== k)].slice(0, MAX_RECENT)
  saveRecents()
}

function removeRecent(i) {
  recents.value.splice(i, 1)
  saveRecents()
}

function clearRecents() {
  recents.value = []
  saveRecents()
}

function applyRecent(r) {
  keyword.value = r
}

// ---------- 搜索 ----------

let debounceTimer = null
let seq = 0

function onKeydown(e) {
  if (e.key === 'Escape') {
    panelOpen.value = false
    activeIndex.value = -1
    e.target.blur()
    return
  }
  const items = flatItems.value
  if (!items.length) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = (activeIndex.value + 1) % items.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = (activeIndex.value - 1 + items.length) % items.length
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const item = items[activeIndex.value >= 0 ? activeIndex.value : 0]
    if (item) openItem(item)
  }
}

function onFocus() {
  panelOpen.value = true
}

function clearKeyword() {
  keyword.value = ''
  results.value = null
  searched.value = false
  searching.value = false
  activeIndex.value = -1
}

// 输入防抖 300ms，空关键字不发请求
watch(keyword, (val) => {
  clearTimeout(debounceTimer)
  activeIndex.value = -1
  const kw = val.trim()
  if (!kw) {
    results.value = null
    searched.value = false
    searching.value = false
    return
  }
  debounceTimer = setTimeout(() => doSearch(kw), 300)
})

async function doSearch(kw) {
  const mySeq = ++seq
  searching.value = true
  try {
    const data = await searchApi.query(kw, 5)
    if (mySeq !== seq) return
    results.value = data || {}
    searched.value = true
  } catch (e) {
    if (mySeq !== seq) return
    results.value = { services: [], alerts: [], rules: [], documents: [] }
    searched.value = true
  } finally {
    if (mySeq === seq) searching.value = false
  }
}

// ---------- 跳转 ----------

function closePanel() {
  panelOpen.value = false
  activeIndex.value = -1
  keyword.value = ''
  results.value = null
  searched.value = false
  searching.value = false
}

function openQuick(item) {
  closePanel()
  router.push(item.path)
}

function openItem(item) {
  const kw = keyword.value.trim()
  addRecent(kw)
  closePanel()
  switch (item.type) {
    case 'menu':
      router.push(item.path)
      break
    case 'service':
      router.push({ path: '/services', query: { highlight: item.id } })
      break
    case 'alert': {
      const q = stripHtml(item.raw?.rule_name || item.raw?.matched || kw)
      router.push({ path: '/alerts/events', query: { q } })
      break
    }
    case 'rule':
      router.push({ path: '/alert-rules', query: { q: kw } })
      break
    case 'doc':
      router.push({
        path: '/knowledge-base',
        query: { doc: item.id, q: stripHtml(item.raw?.title || kw) }
      })
      break
  }
}

function openGroupAll(g) {
  const kw = keyword.value.trim()
  addRecent(kw)
  closePanel()
  if (g.type === 'alert') router.push({ path: '/alerts/events', query: { q: kw } })
  else if (g.type === 'rule') router.push({ path: '/alert-rules', query: { q: kw } })
  else if (g.type === 'doc') router.push({ path: '/knowledge-base', query: { q: kw } })
  else router.push('/services')
}

function askAi() {
  const kw = keyword.value.trim()
  addRecent(kw)
  closePanel()
  router.push({ path: '/chat', query: { q: kw } })
}

// ---------- 面板外点击关闭 ----------

function handleDocClick(e) {
  if (e.target instanceof Element && !e.target.closest('.gs-box')) {
    panelOpen.value = false
    activeIndex.value = -1
  }
}

onMounted(() => document.addEventListener('click', handleDocClick))
onBeforeUnmount(() => document.removeEventListener('click', handleDocClick))
</script>

<style scoped>
.gs-box { position: relative; width: 240px; }

.gs-input {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 7px 12px;
  color: var(--c-text-3);
  width: 100%;
  transition: border-color 0.15s;
}

.gs-input.focused { border-color: var(--c-primary); }

.gs-input input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  width: 100%;
  color: var(--c-text);
  font-family: inherit;
}

.gs-input input::placeholder { color: var(--c-text-3); }

.gs-clear {
  border: none;
  background: none;
  color: var(--c-text-3);
  cursor: pointer;
  font-size: 15px;
  line-height: 1;
  padding: 0 2px;
}

.gs-clear:hover { color: var(--c-text); }

.gs-panel {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 380px;
  max-height: 70vh;
  overflow-y: auto;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 10px;
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.12);
  z-index: 200;
  padding: 6px;
}

.gs-section { margin-bottom: 4px; }

.gs-sec-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px 4px;
  font-size: 12px;
  color: var(--c-text-3);
}

.gs-count {
  background: var(--c-p4-bg);
  color: var(--c-text-2);
  border-radius: 9px;
  padding: 0 7px;
  line-height: 16px;
  font-size: 11px;
}

.gs-link-btn {
  border: none;
  background: none;
  color: var(--c-text-3);
  font-size: 12px;
  cursor: pointer;
  padding: 0;
}

.gs-link-btn:hover { color: var(--c-danger); }

.gs-recent-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 2px 10px 6px;
}

.gs-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: var(--c-primary-soft);
  color: var(--c-text-2);
  border-radius: var(--radius-tag);
  font-size: 12px;
  padding: 3px 9px;
  cursor: pointer;
  max-width: 100%;
}

.gs-chip:hover { color: var(--c-primary); }

.gs-chip-remove {
  font-style: normal;
  color: var(--c-text-3);
  line-height: 1;
}

.gs-chip-remove:hover { color: var(--c-danger); }

.gs-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  color: var(--c-text);
}

.gs-item:hover,
.gs-item.active { background: var(--c-primary-soft); }

.gs-icon {
  flex-shrink: 0;
  width: 26px;
  height: 26px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--c-p4-bg);
  color: var(--c-text-2);
  font-size: 11px;
  font-weight: 700;
}

.gs-icon-menu { background: var(--c-primary-tint); color: var(--c-primary); }
.gs-icon-service { background: var(--c-p3-bg); color: var(--c-p3); }
.gs-icon-alert { background: var(--c-p0-bg); color: var(--c-p0); }
.gs-icon-rule { background: var(--c-p1-bg); color: var(--c-p1); }
.gs-icon-doc { background: var(--c-p2-bg); color: var(--c-p2); }
.gs-icon-quick { background: var(--c-primary-tint); color: var(--c-primary); }
.gs-icon-ai { background: var(--c-primary-tint); color: var(--c-primary); }

.gs-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.gs-main {
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.gs-main :deep(em) {
  font-style: normal;
  color: var(--c-primary);
  font-weight: 600;
}

.gs-sub {
  font-size: 11.5px;
  color: var(--c-text-3);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.gs-tags {
  display: flex;
  align-items: center;
  gap: 6px;
}

.gs-time { font-size: 11px; color: var(--c-text-3); }

.gs-more {
  padding: 6px 10px 6px 36px;
  font-size: 12px;
  color: var(--c-primary);
  cursor: pointer;
  border-radius: 8px;
}

.gs-more:hover { background: var(--c-primary-soft); }

.gs-empty {
  padding: 26px 10px;
  text-align: center;
  color: var(--c-text-3);
  font-size: 13px;
}

.gs-foot {
  padding: 6px 10px;
  font-size: 11px;
  color: var(--c-text-3);
  text-align: center;
  border-top: 1px solid var(--c-border);
  margin-top: 4px;
}

@media (max-width: 768px) {
  .gs-box { display: none; }
}
</style>
