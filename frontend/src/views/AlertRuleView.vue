<template>
  <div>
    <PageHeader :title="$t('alert.rules.title')" :desc="$t('alert.rules.desc')" />

    <div class="page-split">
      <!-- 左侧：业务分组树面板 -->
      <aside class="card group-panel">
        <div class="gp-head">
          <span class="gp-title">{{ $t('group.title') }}</span>
          <button class="gp-add" :title="$t('group.modal.createTitle')" @click="openGroupModal">+</button>
        </div>
        <div class="gp-search">
          <input
            v-model="groupKeyword"
            class="gp-search-input"
            :placeholder="$t('group.panel.searchPlaceholder')"
          />
        </div>
        <div class="gp-tree">
          <div class="gp-node" :class="{ active: !filterGroupId }" @click="selectGroup('')">
            <span class="gp-arrow gp-arrow-spacer" />
            <span class="gp-label">{{ $t('group.panel.allRules') }}</span>
            <span v-if="total > 0" class="gp-count">{{ total }}</span>
          </div>

          <template v-for="g in treeFlat" :key="g.id">
            <div
              class="gp-node"
              :class="{ active: filterGroupId === g.id }"
              :style="{ paddingLeft: 8 + g.depth * 16 + 'px' }"
              @click="selectGroup(g.id)"
            >
              <span
                v-if="g.hasChildren"
                class="gp-arrow"
                @click.stop="toggleExpand(g.id)"
              >{{ isExpanded(g.id) ? '▾' : '▸' }}</span>
              <span v-else class="gp-arrow gp-arrow-spacer" />

              <input
                v-if="editingGroupId === g.id"
                v-model="editingName"
                class="gp-rename-input"
                maxlength="128"
                @click.stop
                @keyup.enter="commitRename"
                @keyup.esc="cancelRename"
                @blur="commitRename"
                v-focus
              />
              <span v-else class="gp-label" :title="g.name">{{ g.name }}</span>

              <span v-if="editingGroupId !== g.id" class="gp-count" v-show="g.rule_count > 0">{{ g.rule_count }}</span>

              <span v-if="editingGroupId !== g.id" class="gp-ops">
                <button class="gp-op" :title="$t('group.edit')" @click.stop="startRename(g)">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                    <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04a1 1 0 0 0 0-1.41l-2.34-2.34a1 1 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                  </svg>
                </button>
                <button class="gp-op gp-op-danger" :title="$t('group.delete')" @click.stop="deleteGroup(g)">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                    <path d="M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                  </svg>
                </button>
              </span>
            </div>
          </template>

          <div v-if="!groupLoading && groups.length === 0" class="gp-hint">{{ $t('group.panel.empty') }}</div>
          <div v-else-if="groupKeyword && treeFlat.length === 0" class="gp-hint">{{ $t('group.panel.noMatch') }}</div>
          <div v-if="groupLoading" class="gp-hint">{{ $t('group.loading') }}</div>
        </div>
      </aside>

      <!-- 右侧：规则列表 -->
      <div class="card rules-card">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">{{ $t('alert.rules.listTitle') }}</h3>
            <p class="card-sub">{{ $t('alert.rules.listSub') }}</p>
          </div>
          <div class="toolbar-right">
            <span v-if="currentGroup" class="filter-tag">
              {{ $t('group.panel.filtering', { name: currentGroup.name }) }}
              <em @click="selectGroup('')">×</em>
            </span>
            <span v-if="filterKw" class="filter-tag">
              {{ $t('alert.rules.keyword', { kw: filterKw }) }}
              <em @click="clearKeyword">×</em>
            </span>
            <button class="btn btn-sm" @click="loadData" :disabled="loading">
              {{ loading ? $t('alert.rules.loading') : $t('alert.rules.refresh') }}
            </button>
            <button class="btn btn-primary btn-sm" @click="openCreateModal" :disabled="loading">
              + {{ $t('alert.rules.addRule') }}
            </button>
            <div class="more-wrap">
              <button class="btn btn-sm" @click="moreOpen = !moreOpen">
                {{ $t('alert.rules.more') }} ▾
              </button>
              <div v-if="moreOpen" class="more-mask" @click="moreOpen = false"></div>
              <div v-if="moreOpen" class="more-menu">
                <div
                  class="more-item"
                  :class="{ disabled: selectedCount === 0 }"
                  @click="onBatchDelete"
                >{{ $t('alert.rules.batchDelete') }}</div>
                <div
                  class="more-item"
                  :class="{ disabled: selectedCount === 0 }"
                  @click="onOpenCloneToGroup"
                >{{ $t('alert.rules.cloneToGroup') }}</div>
              </div>
            </div>
          </div>
        </div>
        <div class="table-scroll">
        <table class="table rules-table">
          <thead>
            <tr>
              <th class="col-check"><input type="checkbox" :checked="allChecked" @change="toggleSelectAll" /></th>
              <th class="col-status">{{ $t('alert.rules.colStatus') }}</th>
              <th>{{ $t('alert.rules.colRuleName') }}</th>
              <th>{{ $t('alert.rules.colGroup') }}</th>
              <th>{{ $t('alert.rules.colSeverity') }}</th>
              <th>PromQL</th>
              <th>{{ $t('alert.rules.colNotifyRule') }}</th>
              <th>{{ $t('alert.rules.colEnabled') }}</th>
              <th>{{ $t('alert.rules.colCreatedAt') }}</th>
              <th>{{ $t('alert.rules.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id">
              <td class="col-check"><input type="checkbox" :checked="selectedIds.has(rule.id)" @change="toggleSelect(rule)" /></td>
              <td class="col-status">
                <span
                  class="status-icon"
                  :class="ruleStatusClass(rule)"
                  :title="ruleStatusTitle(rule)"
                  @click="openEventsModal(rule)"
                >
                  <template v-if="activeCount(rule.id) > 0">
                    <span class="status-dot status-alert">
                      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                        <path d="M12 2L1 21h22L12 2zm0 6l7.53 13H4.47L12 8zm-1 5v4h2v-4h-2zm0 5v2h2v-2h-2z"/>
                      </svg>
                    </span>
                    <span class="status-badge" :title="$t('alert.rules.activeBadge', { count: activeCount(rule.id) })">{{ activeCount(rule.id) }}</span>
                  </template>
                  <template v-else>
                    <span class="status-dot status-ok">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
                      </svg>
                    </span>
                  </template>
                </span>
              </td>
              <td>
                <a class="rule-name-link" :title="$t('alert.rules.edit')" @click="openEditModal(rule)">{{ rule.name }}</a>
              </td>
              <td>
                <span v-if="rule.group_name" class="group-chip">{{ rule.group_name }}</span>
                <span v-else class="muted">—</span>
              </td>
              <td>
                <LevelTag :level="severityLevel(rule.severity)">
                  {{ severityText(rule.severity) }}
                </LevelTag>
              </td>
              <td class="mono" :title="rule.prom_ql">{{ rule.prom_ql || '-' }}</td>
              <td>
                <template v-if="rule.notify_rule_id">
                  <span class="notify-link">{{ (notifyRules.find(nr => nr.id === rule.notify_rule_id) || {}).name || rule.notify_rule_id }}</span>
                </template>
                <span v-else class="muted">—</span>
              </td>
              <td>
                <label class="switch" :title="rule.is_enabled === 1 ? $t('alert.rules.disable') : $t('alert.rules.enable')">
                  <input type="checkbox" :checked="rule.is_enabled === 1" @change="toggleEnabled(rule)" />
                  <span class="slider"></span>
                </label>
              </td>
              <td class="muted col-time">{{ fmtTime(rule.created_at) }}</td>
              <td>
                <div class="ops">
                  <button class="gp-op row-del-op" :title="$t('alert.rules.clone')" @click="openCloneModal(rule)">
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                      <path d="M16 1H4a2 2 0 0 0-2 2v14h2V3h12V1zm3 4H8a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                  </button>
                  <button class="gp-op row-del-op" :title="$t('alert.rules.delete')" @click="deleteRule(rule)">
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                      <path d="M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!loading && rules.length === 0">
              <td colspan="10" class="empty-row">{{ filterKw ? $t('alert.rules.emptyNoMatch') : $t('alert.rules.empty') }}</td>
            </tr>
            <tr v-if="loading">
              <td colspan="10" class="empty-row">{{ $t('alert.rules.loading') }}</td>
            </tr>
          </tbody>
        </table>
        </div>

        <!-- 分页栏 -->
        <div class="pagination" v-if="total > 0">
          <span class="pager-total">{{ $t('alert.rules.pager.total', { total }) }}</span>
          <div class="pager-pages">
            <button class="pager-btn" :disabled="page === 1 || loading" @click="changePage(page - 1)">‹</button>
            <button
              v-for="(p, idx) in pageNumbers"
              :key="idx"
              class="pager-btn"
              :class="{ active: p === page }"
              :disabled="p === '…'"
              @click="typeof p === 'number' && changePage(p)"
            >{{ p }}</button>
            <button class="pager-btn" :disabled="page >= totalPages || loading" @click="changePage(page + 1)">›</button>
          </div>
          <select v-model.number="limit" class="pager-size" @change="changeLimit">
            <option v-for="n in [10, 20, 50, 100]" :key="n" :value="n">{{ $t('alert.rules.pager.perPage', { n }) }}</option>
          </select>
          <span class="pager-jump">
            {{ $t('alert.rules.pager.jumpTo') }}
            <input v-model.number="jumpTo" type="number" min="1" :max="totalPages" @keyup.enter="jumpToPage" />
            {{ $t('alert.rules.pager.pageUnit') }}
          </span>
        </div>
      </div>
    </div>

    <!-- 规则活跃事件弹窗 -->
    <div v-if="eventsModal.visible" class="modal-mask" @click.self="eventsModal.visible = false">
      <div class="modal modal-lg">
        <div class="modal-header">
          <h3>
            {{ eventsModal.rule?.name }}
            <span v-if="eventsModal.list.length > 0" class="title-sub">{{ $t('alert.rules.eventsActive', { count: eventsModal.list.length }) }}</span>
            <span v-else class="title-sub title-ok">{{ $t('alert.rules.eventsNone') }}</span>
          </h3>
          <span class="modal-close" @click="eventsModal.visible = false">×</span>
        </div>
        <div class="modal-body events-modal-body">
          <div v-if="eventsModal.loading" class="empty-inner">{{ $t('alert.rules.loading') }}</div>
          <div v-else-if="eventsModal.list.length === 0" class="empty-inner">
            <div class="empty-icon">✓</div>
            <div>{{ $t('alert.rules.eventsEmpty') }}</div>
          </div>
          <div v-else class="events-list">
            <div v-for="ev in eventsModal.list" :key="ev.id" class="event-card">
              <div class="event-top">
                <span class="event-sev" :class="'sev-' + ev.severity">{{ severityText(ev.severity) }}</span>
                <span class="event-target" :title="ev.target_ident">{{ ev.target_ident || '-' }}</span>
                <span class="event-time">{{ fmtTime(ev.trigger_time) }}</span>
              </div>
              <div class="event-meta">
                <span class="meta-item">{{ $t('alert.rules.triggerValue') }}<b>{{ ev.trigger_value || '-' }}</b></span>
              </div>
              <div v-if="ev.tags" class="event-tags" :title="ev.tags">
                <span v-for="(tag, idx) in parseTags(ev.tags)" :key="idx" class="tag-chip">{{ tag }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建/编辑 规则弹窗 -->
    <div v-if="modalVisible" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ isEdit ? $t('alert.rules.modal.editTitle') : $t('alert.rules.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('alert.rules.form.nameLabel') }}</label>
            <input v-model="form.name" class="form-input" :placeholder="$t('alert.rules.form.namePlaceholder')" maxlength="128" />
            <span v-if="errors.name" class="form-error">{{ errors.name }}</span>
          </div>

          <div class="form-item">
            <label class="form-label required">PromQL</label>
            <textarea
              v-model="form.prom_ql"
              class="form-input form-textarea mono-textarea"
              rows="3"
              :placeholder="$t('alert.rules.form.promqlPlaceholder')"
            />
            <span v-if="errors.prom_ql" class="form-error">{{ errors.prom_ql }}</span>
          </div>

          <div class="form-row">
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('alert.rules.form.evalIntervalLabel') }}</label>
              <select v-model.number="form.eval_interval" class="form-input">
                <option :value="15">{{ $t('alert.rules.everySeconds', { n: 15 }) }}</option>
                <option :value="30">{{ $t('alert.rules.everySeconds', { n: 30 }) }}</option>
                <option :value="45">{{ $t('alert.rules.everySeconds', { n: 45 }) }}</option>
                <option :value="60">{{ $t('alert.rules.everySeconds', { n: 60 }) }}</option>
                <option :value="120">{{ $t('alert.rules.everySeconds', { n: 120 }) }}</option>
                <option :value="180">{{ $t('alert.rules.everySeconds', { n: 180 }) }}</option>
                <option :value="300">{{ $t('alert.rules.everySeconds', { n: 300 }) }}</option>
              </select>
              <span v-if="errors.eval_interval" class="form-error">{{ errors.eval_interval }}</span>
            </div>
            <div class="form-item form-item-half">
              <label class="form-label required">{{ $t('alert.rules.form.durationLabel') }}</label>
              <input v-model.number="form.duration" type="number" min="0" class="form-input" :placeholder="$t('alert.rules.form.durationPlaceholder')" />
              <span class="form-hint">{{ $t('alert.rules.form.durationHint') }}</span>
              <span v-if="errors.duration" class="form-error">{{ errors.duration }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label required">{{ $t('alert.rules.form.severityLabel') }}</label>
            <select v-model.number="form.severity" class="form-input">
              <option :value="1">{{ $t('alert.severity.p1') }}</option>
              <option :value="2">{{ $t('alert.severity.p2') }}</option>
              <option :value="3">{{ $t('alert.severity.p3') }}</option>
            </select>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('alert.rules.form.groupLabel') }}</label>
            <select v-model="form.group_id" class="form-input">
              <option value="">{{ $t('alert.rules.form.noGroup') }}</option>
              <option v-for="g in groupOptions" :key="g.id" :value="g.id">{{ g.label }}</option>
            </select>
          </div>

          <div class="form-item">
            <label class="form-label">{{ $t('alert.rules.form.notifyRuleLabel') }} <span class="form-hint-inline">{{ $t('alert.rules.form.notifyOptional') }}</span></label>
            <select v-model="form.notify_rule_id" class="form-input">
              <option value="">{{ $t('alert.rules.form.noNotify') }}</option>
              <option v-for="nr in notifyRules" :key="nr.id" :value="nr.id">{{ nr.name }}（{{ triggerLabel(nr.trigger_types) }}）</option>
            </select>
          </div>

          <!-- 通知配置 -->
          <div class="form-section">
            <div class="form-section-title">{{ $t('alert.rules.form.notifyConfigTitle') }}</div>
            <div class="form-row">
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('alert.rules.form.repeatLabel') }}</label>
                <input v-model.number="form.repeat_interval_minutes" type="number" min="0" class="form-input" :placeholder="$t('alert.rules.form.repeatPlaceholder')" />
                <span class="form-hint">{{ $t('alert.rules.form.repeatHint') }}</span>
              </div>
              <div class="form-item form-item-half">
                <label class="form-label">{{ $t('alert.rules.form.maxSendLabel') }}</label>
                <input v-model.number="form.max_send_count" type="number" min="0" class="form-input" :placeholder="$t('alert.rules.form.maxSendPlaceholder')" />
                <span class="form-hint">{{ $t('alert.rules.form.maxSendHint') }}</span>
              </div>
            </div>
          </div>

          <div class="form-item form-item-toggle">
            <label class="form-label">{{ $t('alert.rules.form.enabledLabel') }}</label>
            <label class="switch">
              <input type="checkbox" v-model="form.is_enabled" :true-value="1" :false-value="0" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeModal" :disabled="saving">{{ $t('alert.rules.form.cancel') }}</button>
          <button class="btn btn-primary" @click="saveRule" :disabled="saving">
            {{ saving ? $t('alert.rules.form.saving') : $t('alert.rules.form.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 新建/编辑 业务分组弹窗 -->
    <div v-if="groupModalVisible" class="modal-mask" @click.self="closeGroupModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('group.modal.createTitle') }}</h3>
          <span class="modal-close" @click="closeGroupModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('group.form.nameLabel') }}</label>
            <input
              v-model="groupForm.name"
              class="form-input"
              :placeholder="$t('group.form.namePlaceholder')"
              maxlength="128"
              @keyup.enter="saveGroup"
            />
            <span v-if="groupError" class="form-error">{{ groupError }}</span>
          </div>
          <div class="form-item">
            <label class="form-label">{{ $t('group.form.parentLabel') }}</label>
            <select v-model="groupForm.parent_id" class="form-input">
              <option value="">{{ $t('group.form.parentNone') }}</option>
              <option v-for="g in groupOptions" :key="g.id" :value="g.id">{{ g.label }}</option>
            </select>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeGroupModal" :disabled="groupSaving">{{ $t('group.form.cancel') }}</button>
          <button class="btn btn-primary" @click="saveGroup" :disabled="groupSaving">
            {{ groupSaving ? $t('group.form.saving') : $t('group.form.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 克隆到其他业务分组弹窗 -->
    <div v-if="cloneModal.visible" class="modal-mask" @click.self="cloneModal.visible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('alert.rules.cloneToGroup') }}</h3>
          <span class="modal-close" @click="cloneModal.visible = false">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label">{{ $t('alert.rules.cloneTarget') }}</label>
            <div class="sel-wrap">
              <button type="button" class="sel-trigger" @click="cloneGroupOpen = !cloneGroupOpen">
                <span class="sel-trigger-label">{{ cloneTargetLabel }}</span>
                <span class="sel-arrow">▾</span>
              </button>
              <div v-if="cloneGroupOpen" class="more-mask" @click="cloneGroupOpen = false"></div>
              <div v-if="cloneGroupOpen" class="sel-menu">
                <div
                  v-for="g in groupOptions"
                  :key="g.id"
                  class="sel-option"
                  :class="{ active: cloneModal.targetGroupId === g.id }"
                  @click="pickCloneTarget(g)"
                >{{ g.label }}</div>
              </div>
            </div>
            <span class="form-hint">{{ $t('alert.rules.cloneHint', { count: selectedCount }) }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="cloneModal.visible = false" :disabled="cloneModal.saving">{{ $t('group.form.cancel') }}</button>
          <button class="btn btn-primary" @click="submitClone" :disabled="cloneModal.saving">
            {{ cloneModal.saving ? $t('group.form.saving') : $t('group.form.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { alertRuleApi } from '../api/alertRule.js'
import { notifyApi } from '../api/notify.js'
import { alertEventApi } from '../api/alertEvent.js'
import { businessGroupApi } from '../api/businessGroup.js'

const { t } = useI18n({ useScope: 'global' })
const route = useRoute()

const vFocus = {
  mounted: (el) => { el.focus(); el.select && el.select() }
}

const severityText = (s) => ({ 1: t('alert.severity.p1'), 2: t('alert.severity.p2'), 3: t('alert.severity.p3') }[s] || `P${s}`)

const severityLevel = (s) => {
  if (s === 1) return 'critical'
  if (s === 2) return 'major'
  return 'warning'
}

const fmtTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  return isNaN(d.getTime()) ? '-' : d.toLocaleString('zh-CN')
}

// -------- 规则状态相关 --------
// activeFiringEvents: 一次拉全量活跃告警事件，前端按 rule_id 分组
const activeFiringEvents = ref([])

async function loadActiveEvents() {
  try {
    const result = await alertEventApi.list({ scope: 'active', limit: 500 })
    activeFiringEvents.value = result?.list || []
  } catch {
    activeFiringEvents.value = []
  }
}

// 按 rule_id 的活跃告警数
const activeCountMap = computed(() => {
  const m = {}
  for (const ev of activeFiringEvents.value) {
    m[ev.rule_id] = (m[ev.rule_id] || 0) + 1
  }
  return m
})

function activeCount(ruleId) {
  return activeCountMap.value[ruleId] || 0
}

function ruleStatusClass(rule) {
  const count = activeCount(rule.id)
  if (count > 0) return 'has-alert'
  if (rule.is_enabled !== 1) return 'disabled'
  return 'no-alert'
}

function ruleStatusTitle(rule) {
  if (rule.is_enabled !== 1) return t('alert.rules.statusTitle.disabled')
  const count = activeCount(rule.id)
  if (count > 0) return t('alert.rules.statusTitle.alerting', { count })
  return t('alert.rules.statusTitle.noAlert')
}

// -------- 活跃事件弹窗 --------
const eventsModal = reactive({
  visible: false,
  rule: null,
  list: [],
  loading: false
})

async function openEventsModal(rule) {
  eventsModal.rule = rule
  eventsModal.visible = true
  eventsModal.loading = true
  // 先从已经缓存的全量活跃事件里过滤
  let list = activeFiringEvents.value.filter(e => e.rule_id === rule.id)
  eventsModal.list = list
  eventsModal.loading = false
}

function parseTags(s) {
  if (!s) return []
  return s.split(',').map(t => t.trim()).filter(Boolean)
}

// -------- 基础状态 --------
const rules = ref([])
const notifyRules = ref([])
const groups = ref([])
const loading = ref(false)
const saving = ref(false)

// 规则多选（更多菜单：批量删除 / 克隆到其他业务组）
const selectedIds = ref(new Set())
const moreOpen = ref(false)

const selectedCount = computed(() => selectedIds.value.size)
const allChecked = computed(() => rules.value.length > 0 && rules.value.every(r => selectedIds.value.has(r.id)))

function toggleSelect(rule) {
  const s = new Set(selectedIds.value)
  if (s.has(rule.id)) s.delete(rule.id)
  else s.add(rule.id)
  selectedIds.value = s
}

function toggleSelectAll() {
  if (allChecked.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(rules.value.map(r => r.id))
  }
}

// 克隆到其他业务分组弹窗
const cloneModal = reactive({ visible: false, targetGroupId: '', saving: false })
const cloneGroupOpen = ref(false)

const cloneTargetLabel = computed(() => {
  const g = groups.value.find(x => x.id === cloneModal.targetGroupId)
  return g ? g.name : '-'
})

function pickCloneTarget(g) {
  cloneModal.targetGroupId = g.id
  cloneGroupOpen.value = false
}

function onBatchDelete() {
  if (selectedCount.value === 0) return
  moreOpen.value = false
  if (!confirm(t('alert.rules.confirmBatchDelete', { count: selectedCount.value }))) return
  alertRuleApi.batchDelete([...selectedIds.value])
    .then(async () => {
      selectedIds.value = new Set()
      await loadData()
      await loadGroups()
    })
    .catch(err => alert(t('alert.rules.msg.batchDeleteFailed', { msg: err.message })))
}

function onOpenCloneToGroup() {
  if (selectedCount.value === 0) return
  moreOpen.value = false
  cloneGroupOpen.value = false
  cloneModal.targetGroupId = filterGroupId.value || (groupOptions.value[0] && groupOptions.value[0].id) || ''
  cloneModal.visible = true
}

async function submitClone() {
  if (!cloneModal.targetGroupId) return
  cloneModal.saving = true
  try {
    await alertRuleApi.clone({ ids: [...selectedIds.value], group_id: cloneModal.targetGroupId })
    cloneModal.saving = false
    cloneModal.visible = false
    selectedIds.value = new Set()
    await loadData()
    await loadGroups()
  } catch (err) {
    alert(t('alert.rules.msg.cloneFailed', { msg: err.message }))
    cloneModal.saving = false
  }
}

// 全局搜索落地：/alert-rules?q=关键字，服务端过滤规则名称与 PromQL
const filterKw = ref(route.query.q ? String(route.query.q) : '')
// 业务分组筛选（服务端过滤），由左侧分组树驱动
const filterGroupId = ref('')
// 分页
const page = ref(1)
const limit = ref(20)
const total = ref(0)
const jumpTo = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit.value)))

// 页码序列：最多 7 个，首尾固定，中间用省略号
const pageNumbers = computed(() => {
  const t = totalPages.value
  const c = page.value
  if (t <= 7) return Array.from({ length: t }, (_, i) => i + 1)
  const pages = [1]
  const lo = Math.max(2, c - 1)
  const hi = Math.min(t - 1, c + 1)
  if (lo > 2) pages.push('…')
  for (let i = lo; i <= hi; i++) pages.push(i)
  if (hi < t - 1) pages.push('…')
  pages.push(t)
  return pages
})

// -------- 业务分组树 --------
const groupKeyword = ref('')
const groupLoading = ref(false)
const expanded = ref(new Set())
const editingGroupId = ref(null)
const editingName = ref('')
let renameSnapshot = ''

// 子分组映射：parent_id -> groups
const childrenMap = computed(() => {
  const m = {}
  for (const g of groups.value) {
    const pid = g.parent_id || ''
    if (!m[pid]) m[pid] = []
    m[pid].push(g)
  }
  return m
})

const isExpanded = (id) => expanded.value.has(id)

function toggleExpand(id) {
  const s = new Set(expanded.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  expanded.value = s
}

// 树扁平化输出（含深度），供模板逐行渲染；搜索关键字时展示匹配分组（忽略展开状态，平铺展示）
const treeFlat = computed(() => {
  const kw = groupKeyword.value.trim().toLowerCase()
  const out = []
  const walk = (nodes, depth) => {
    for (const n of nodes) {
      const children = childrenMap.value[n.id] || []
      out.push({ ...n, depth, hasChildren: children.length > 0 })
      if (children.length > 0 && isExpanded(n.id)) walk(children, depth + 1)
    }
  }
  if (kw) {
    for (const g of groups.value) {
      if (g.name.toLowerCase().includes(kw)) {
        out.push({ ...g, depth: 0, hasChildren: false })
      }
    }
    return out
  }
  walk(childrenMap.value[''] || [], 0)
  return out
})

// 全部分组的层级缩进选项（用于规则表单与新建分组弹窗的父分组选择）
const groupOptions = computed(() => {
  const out = []
  const walk = (nodes, depth) => {
    for (const n of nodes) {
      out.push({ id: n.id, label: '　'.repeat(depth) + n.name })
      walk(childrenMap.value[n.id] || [], depth + 1)
    }
  }
  walk(childrenMap.value[''] || [], 0)
  return out
})

const currentGroup = computed(() => groups.value.find(g => g.id === filterGroupId.value) || null)

async function loadGroups() {
  groupLoading.value = true
  try {
    groups.value = await businessGroupApi.list() || []
    // 新加载的分组默认展开（与已有展开状态取并集）
    const s = new Set(expanded.value)
    for (const g of groups.value) {
      if ((childrenMap.value[g.id] || []).length > 0) s.add(g.id)
    }
    expanded.value = s
  } catch {
    groups.value = []
  } finally {
    groupLoading.value = false
  }
}

function selectGroup(id) {
  if (editingGroupId.value) return
  if (filterGroupId.value === id) return
  filterGroupId.value = id
  page.value = 1
  loadData()
}

function startRename(g) {
  editingGroupId.value = g.id
  editingName.value = g.name
  renameSnapshot = g.name
}

async function commitRename() {
  const id = editingGroupId.value
  if (!id) return
  const name = editingName.value.trim()
  editingGroupId.value = null
  if (!name || name === renameSnapshot) return
  try {
    await businessGroupApi.update(id, { name })
    await loadGroups()
    await loadData()
  } catch (err) {
    alert(t('group.rename.saveFailed', { msg: err.message }))
    await loadGroups()
  }
}

function cancelRename() {
  editingGroupId.value = null
  editingName.value = ''
}

async function deleteGroup(g) {
  if (!confirm(t('group.msg.confirmDelete', { name: g.name }))) return
  try {
    await businessGroupApi.remove(g.id)
    if (filterGroupId.value === g.id) {
      filterGroupId.value = ''
      page.value = 1
    }
    await loadGroups()
    await loadData()
  } catch (err) {
    alert(t('group.msg.deleteFailed', { msg: err.message }))
  }
}

// -------- 新建分组弹窗 --------
const groupModalVisible = ref(false)
const groupSaving = ref(false)
const groupForm = reactive({ name: '', parent_id: '' })
const groupError = ref('')

function openGroupModal() {
  groupForm.name = ''
  groupForm.parent_id = ''
  groupError.value = ''
  groupModalVisible.value = true
}

function closeGroupModal() {
  if (groupSaving.value) return
  groupModalVisible.value = false
}

async function saveGroup() {
  groupError.value = ''
  if (!groupForm.name.trim()) {
    groupError.value = t('group.error.nameRequired')
    return
  }
  groupSaving.value = true
  try {
    await businessGroupApi.create({ name: groupForm.name, parent_id: groupForm.parent_id })
    groupSaving.value = false
    closeGroupModal()
    await loadGroups()
  } catch (err) {
    alert(t('group.msg.createFailed', { msg: err.message }))
    groupSaving.value = false
  }
}

// -------- 规则弹窗状态 --------
const modalVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const errors = reactive({})

const form = reactive({
  name: '',
  prom_ql: '',
  eval_interval: 60,
  duration: 300,
  severity: 2,
  group_id: '',
  notify_rule_id: '',
  repeat_interval_minutes: 0,
  max_send_count: 0,
  is_enabled: 1
})

async function loadData() {
  loading.value = true
  try {
    const data = await alertRuleApi.list({
      page: page.value,
      limit: limit.value,
      query: filterKw.value.trim(),
      groupId: filterGroupId.value
    })
    rules.value = data?.list || []
    total.value = data?.total || 0
    // 数据刷新后剔除已不存在的勾选（如被删除）
    const alive = new Set(rules.value.map(r => r.id))
    selectedIds.value = new Set([...selectedIds.value].filter(id => alive.has(id)))
    // 当前页已空（如删除末页最后一条）时回退到上一页
    if (rules.value.length === 0 && page.value > 1 && total.value > 0) {
      page.value--
      loading.value = false
      return loadData()
    }
    // 同时刷新活跃事件列表（规则状态图标依赖）
    await loadActiveEvents()
  } catch (err) {
    alert(t('alert.rules.msg.loadFailed', { msg: err.message }))
    rules.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 筛选条件变化时回到第一页重新查询
function reloadFirstPage() {
  page.value = 1
  loadData()
}

function clearKeyword() {
  filterKw.value = ''
  reloadFirstPage()
}

function changePage(p) {
  if (p < 1 || p > totalPages.value || p === page.value) return
  page.value = p
  loadData()
}

function changeLimit() {
  page.value = 1
  loadData()
}

function jumpToPage() {
  const p = parseInt(jumpTo.value, 10)
  if (!isNaN(p)) changePage(Math.min(Math.max(p, 1), totalPages.value))
  jumpTo.value = ''
}

function resetForm() {
  form.name = ''
  form.prom_ql = ''
  form.eval_interval = 60
  form.duration = 300
  form.severity = 2
  form.group_id = ''
  form.notify_rule_id = ''
  form.repeat_interval_minutes = 0
  form.max_send_count = 0
  form.is_enabled = 1
  Object.keys(errors).forEach(k => delete errors[k])
}

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  resetForm()
  modalVisible.value = true
}

// 克隆：以新建模式打开弹窗并预填被克隆规则的内容（名称保持原样，直接保存将触发唯一性校验）
function openCloneModal(rule) {
  isEdit.value = false
  editingId.value = null
  resetForm()
  form.name = rule.name
  form.prom_ql = rule.prom_ql
  form.eval_interval = rule.eval_interval || 60
  form.duration = rule.duration
  form.severity = rule.severity
  form.group_id = rule.group_id || ''
  form.notify_rule_id = rule.notify_rule_id || ''
  form.repeat_interval_minutes = rule.repeat_interval_minutes || 0
  form.max_send_count = rule.max_send_count || 0
  form.is_enabled = rule.is_enabled
  modalVisible.value = true
}

function openEditModal(rule) {
  isEdit.value = true
  editingId.value = rule.id
  resetForm()
  form.name = rule.name
  form.prom_ql = rule.prom_ql
  form.eval_interval = rule.eval_interval || 60
  form.duration = rule.duration
  form.severity = rule.severity
  form.group_id = rule.group_id || ''
  form.notify_rule_id = rule.notify_rule_id || ''
  form.repeat_interval_minutes = rule.repeat_interval_minutes || 0
  form.max_send_count = rule.max_send_count || 0
  form.is_enabled = rule.is_enabled
  modalVisible.value = true
}

function closeModal() {
  if (saving.value) return
  modalVisible.value = false
}

function validateForm() {
  Object.keys(errors).forEach(k => delete errors[k])
  let valid = true

  if (!form.name.trim()) {
    errors.name = t('alert.rules.error.nameRequired')
    valid = false
  }
  if (!form.prom_ql.trim()) {
    errors.prom_ql = t('alert.rules.error.promqlRequired')
    valid = false
  }
  if (form.eval_interval === null || form.eval_interval === undefined || form.eval_interval <= 0) {
    errors.eval_interval = t('alert.rules.error.evalRequired')
    valid = false
  }
  if (form.duration === null || form.duration === undefined || form.duration < 0) {
    errors.duration = t('alert.rules.error.durationNegative')
    valid = false
  }

  return valid
}

async function saveRule() {
  if (!validateForm()) return

  saving.value = true
  try {
    const payload = {
      name: form.name,
      prom_ql: form.prom_ql,
      eval_interval: form.eval_interval,
      duration: form.duration,
      severity: form.severity,
      group_id: form.group_id || '',
      notify_rule_id: form.notify_rule_id || '',
      repeat_interval_minutes: form.repeat_interval_minutes || 0,
      max_send_count: form.max_send_count || 0,
      is_enabled: form.is_enabled
    }

    if (isEdit.value) {
      await alertRuleApi.update(editingId.value, payload)
    } else {
      await alertRuleApi.create(payload)
    }

    saving.value = false
    closeModal()
    if (!isEdit.value) page.value = 1
    await loadData()
    await loadGroups()
  } catch (err) {
    alert(t(isEdit.value ? 'alert.rules.msg.updateFailed' : 'alert.rules.msg.createFailed', { msg: err.message }))
    saving.value = false
  }
}

async function toggleEnabled(rule) {
  try {
    const result = await alertRuleApi.toggle(rule.id)
    const idx = rules.value.findIndex(r => r.id === rule.id)
    if (idx !== -1 && result) {
      // 合并而非整体替换：接口异常时保留本地冗余字段（如 group_name）
      rules.value[idx] = { ...rules.value[idx], ...result }
    }
  } catch (err) {
    alert(t('alert.rules.msg.toggleFailed', { msg: err.message }))
  }
}

async function deleteRule(rule) {
  if (!confirm(t('alert.rules.msg.confirmDelete', { name: rule.name }))) return
  try {
    await alertRuleApi.remove(rule.id)
    await loadData()
    await loadGroups()
  } catch (err) {
    alert(t('alert.rules.msg.deleteFailed', { msg: err.message }))
  }
}

const triggerLabel = (triggerTypes) => t('alert.rules.triggerType.' + triggerTypes)

async function loadNotifyRules() {
  try {
    notifyRules.value = await notifyApi.listRules() || []
  } catch {
    notifyRules.value = []
  }
}

onMounted(() => {
  loadData()
  loadNotifyRules()
  loadGroups()
})
</script>

<style scoped>
/* ========== 左右两栏布局 ========== */
.page-split {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

/* ========== 左侧分组树面板 ========== */
.group-panel {
  width: 240px;
  flex-shrink: 0;
  padding: 14px;
  box-sizing: border-box;
}

.gp-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.gp-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--c-text);
}

.gp-add {
  width: 22px;
  height: 22px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.gp-add:hover {
  border-color: var(--c-primary);
  color: var(--c-primary);
}

.gp-search {
  margin-bottom: 8px;
}

.gp-search-input {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 12.5px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.gp-search-input:focus { border-color: var(--c-primary); }
.gp-search-input::placeholder { color: var(--c-text-3); }

.gp-tree {
  max-height: calc(100vh - 280px);
  overflow-y: auto;
}

.gp-node {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--c-text-2);
  transition: background 0.15s;
  user-select: none;
}

.gp-node:hover { background: var(--c-primary-soft, rgba(22, 119, 255, 0.08)); }

.gp-node.active {
  background: var(--c-primary-tint);
  color: var(--c-primary);
  font-weight: 600;
}

.gp-arrow {
  width: 14px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: var(--c-text-3);
}

.gp-arrow-spacer { visibility: hidden; }

.gp-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gp-count {
  font-size: 11px;
  font-weight: 600;
  background: var(--c-primary-tint);
  color: var(--c-primary);
  border-radius: 9px;
  padding: 0 6px;
  flex-shrink: 0;
}

.gp-node.active .gp-count {
  background: var(--c-primary);
  color: #fff;
}

.gp-ops {
  display: inline-flex;
  gap: 2px;
  opacity: 0;
  flex-shrink: 0;
  transition: opacity 0.15s;
}

.gp-node:hover .gp-ops { opacity: 1; }

.gp-op {
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--c-text-3);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  transition: all 0.15s;
}

.gp-op:hover { background: var(--c-primary-tint); color: var(--c-primary); }
.gp-op-danger:hover { background: #ffe5e5; color: var(--c-danger); }

.gp-rename-input {
  flex: 1;
  min-width: 0;
  padding: 2px 6px;
  border: 1px solid var(--c-primary);
  border-radius: 4px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}

.gp-hint {
  padding: 14px 8px;
  font-size: 12px;
  color: var(--c-text-3);
  text-align: center;
}

/* ========== 右侧规则卡片 ========== */
.rules-card {
  flex: 1;
  min-width: 0;
}

/* 工具栏与表格间距 + 按钮尺寸 */
.rules-card .card-head-flex {
  margin-bottom: 14px;
}

.rules-card .toolbar-right .btn {
  padding: 7px 16px;
  font-size: 13px;
  height: 32px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

/* 更多操作下拉菜单 */
.more-wrap {
  position: relative;
}

.more-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 90;
}

.more-menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 160px;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  z-index: 100;
  padding: 4px;
}

.more-item {
  padding: 7px 12px;
  font-size: 12.5px;
  color: var(--c-text-2);
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}

.more-item:hover:not(.disabled) {
  background: var(--c-primary-soft, rgba(22, 119, 255, 0.08));
  color: var(--c-primary);
}

.more-item.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* 自定义下拉选择（克隆到其他业务分组） */
.sel-wrap {
  position: relative;
}

.sel-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 0.15s;
  font-family: inherit;
}

.sel-trigger:hover {
  border-color: var(--c-primary);
}

.sel-trigger-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sel-arrow {
  color: var(--c-text-3);
  font-size: 11px;
  flex-shrink: 0;
}

.sel-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 240px;
  overflow-y: auto;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
  z-index: 110;
  padding: 4px;
  box-sizing: border-box;
}

.sel-option {
  padding: 7px 12px;
  font-size: 13px;
  color: var(--c-text-2);
  border-radius: 6px;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: all 0.15s;
}

.sel-option:hover {
  background: var(--c-primary-soft, rgba(22, 119, 255, 0.08));
  color: var(--c-primary);
}

.sel-option.active {
  background: var(--c-primary-tint);
  color: var(--c-primary);
  font-weight: 600;
}

.table-scroll {
  overflow-x: auto;
}

.rules-table {
  min-width: 980px;
}

.rules-table td {
  vertical-align: middle;
}

.col-time {
  white-space: nowrap;
}

.ops {
  flex-wrap: nowrap;
  white-space: nowrap;
}
.ops .btn {
  flex-shrink: 0;
  white-space: nowrap;
}

.group-chip {
  font-size: 12px;
  background: var(--c-primary-soft, rgba(22,119,255,0.08));
  color: var(--c-primary);
  border-radius: var(--radius-tag, 10px);
  padding: 2px 9px;
  white-space: nowrap;
}

/* 规则名称：可点击，点击弹出编辑弹窗 */
.rule-name-link {
  color: var(--c-text);
  font-weight: 600;
  cursor: pointer;
  transition: color 0.15s;
}

.rule-name-link:hover {
  color: var(--c-primary);
  text-decoration: underline;
}

/* 删除图标按钮（夜莺风格，常显） */
.row-del-op {
  width: 24px;
  height: 24px;
}

/* 列表内的启停开关 */
td .switch {
  vertical-align: middle;
}

/* 分页栏 */
.pagination {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 16px;
  flex-wrap: wrap;
  font-size: 12.5px;
  color: var(--c-text-2);
}

.pager-total { white-space: nowrap; }

.pager-pages { display: flex; align-items: center; gap: 6px; }

.pager-btn {
  min-width: 28px;
  height: 28px;
  padding: 0 7px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 12.5px;
  cursor: pointer;
  transition: all 0.15s;
}
.pager-btn:hover:not(:disabled) {
  border-color: var(--c-primary);
  color: var(--c-primary);
}
.pager-btn.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
  cursor: default;
}
.pager-btn:disabled { cursor: default; opacity: 0.6; }
.pager-btn.active:disabled { opacity: 1; }

.pager-size {
  height: 28px;
  padding: 0 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 12.5px;
  outline: none;
  cursor: pointer;
}

.pager-jump {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.pager-jump input {
  width: 48px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 12.5px;
  outline: none;
  box-sizing: border-box;
}
.pager-jump input:focus { border-color: var(--c-primary); }

.filter-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--c-primary);
  background: var(--c-primary-tint);
  border-radius: var(--radius-tag);
  padding: 3px 9px;
}

.filter-tag em {
  font-style: normal;
  cursor: pointer;
  color: var(--c-text-3);
}

.filter-tag em:hover { color: var(--c-danger); }

/* ========== 状态图标列 ========== */
.col-check { width: 34px; text-align: center; }
.col-check input { cursor: pointer; accent-color: var(--c-primary); }
.col-status { width: 60px; text-align: center; }
.status-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 3px 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}
.status-icon:hover { background: var(--c-primary-soft, rgba(22,119,255,0.08)); }

.status-dot {
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: 50%;
}
.status-alert {
  color: #e74c3c;
  animation: pulse 1.8s ease-in-out infinite;
}
.status-ok { color: #27ae60; }
.status-icon.disabled .status-dot { color: #bbb; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.55; }
}

.status-badge {
  font-size: 11px;
  font-weight: 600;
  background: #e74c3c;
  color: #fff;
  border-radius: 10px;
  padding: 1px 6px;
  min-width: 16px;
  display: inline-block;
  text-align: center;
}

/* ========== 活跃事件弹窗 ========== */
.modal-lg { width: 620px; }

.events-modal-body { padding: 8px 24px 20px; }
.title-sub { font-size: 13px; font-weight: 400; color: var(--c-text-2); margin-left: 6px; }
.title-ok { color: #27ae60; }

.empty-inner { padding: 40px 20px; text-align: center; color: var(--c-text-3); }
.empty-icon { font-size: 40px; color: #27ae60; margin-bottom: 10px; }

.events-list { display: flex; flex-direction: column; gap: 10px; }

.event-card {
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 10px 14px;
  background: var(--c-bg);
}
.event-top { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; flex-wrap: wrap; }
.event-sev {
  font-size: 11px; padding: 1px 8px; border-radius: 10px; font-weight: 600;
}
.event-sev.sev-1 { background: #ffe5e5; color: #e74c3c; }
.event-sev.sev-2 { background: #ffecd8; color: #f39c12; }
.event-sev.sev-3 { background: #fdf6dc; color: #d9a400; }
.event-target { font-size: 13px; font-family: ui-monospace, Menlo, monospace; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.event-time { margin-left: auto; font-size: 11.5px; color: var(--c-text-3); }
.event-meta { font-size: 12px; color: var(--c-text-2); margin-bottom: 5px; }
.event-meta b { color: var(--c-text); font-weight: 600; }
.event-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-chip {
  font-size: 11px; background: var(--c-primary-soft, rgba(22,119,255,0.08));
  color: var(--c-primary); padding: 1px 7px; border-radius: 10px;
  max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* ========== 其余 ========== */
.ops { display: flex; gap: 8px; }
.btn-danger { color: var(--c-danger); }
.btn-danger:hover { border-color: var(--c-danger); color: var(--c-danger); }

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.modal-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--c-surface);
  border-radius: var(--radius-card);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  width: 580px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  border-bottom: 1px solid var(--c-border);
}
.modal-header h3 { margin: 0; font-size: 16px; font-weight: 600; }
.modal-close {
  cursor: pointer;
  font-size: 24px;
  color: var(--c-text-3);
  line-height: 1;
  transition: color 0.15s;
}
.modal-close:hover { color: var(--c-text); }

.modal-body {
  padding: 20px 24px;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 24px;
  border-top: 1px solid var(--c-border);
}

.form-row { display: flex; gap: 16px; margin-bottom: 14px; }
.form-item { margin-bottom: 14px; flex: 1; }
.form-item-half { flex: 1; }
.form-item-toggle { display: flex; align-items: center; gap: 12px; }
.form-label { display: block; font-size: 13px; color: var(--c-text-2); margin-bottom: 6px; font-weight: 500; }
.form-label.required::before { content: '*'; color: var(--c-danger); margin-right: 3px; }

.form-input {
  width: 100%; padding: 8px 12px; border: 1px solid var(--c-border); border-radius: 6px;
  background: var(--c-surface); color: var(--c-text); font-size: 13px;
  transition: border-color 0.15s; outline: none; font-family: inherit; box-sizing: border-box;
}
.form-input:focus { border-color: var(--c-primary); }
.form-input::placeholder { color: var(--c-text-3); }
.form-textarea { resize: vertical; min-height: 70px; }
.mono-textarea { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12.5px; }

.form-error { display: block; color: var(--c-danger); font-size: 12px; margin-top: 4px; }
.form-hint { display: block; color: var(--c-text-3); font-size: 12px; margin-top: 4px; }

.switch { position: relative; display: inline-block; width: 40px; height: 22px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--c-border); transition: .3s; border-radius: 22px;
}
.slider:before {
  position: absolute; content: ""; height: 16px; width: 16px; left: 3px; bottom: 3px;
  background-color: white; transition: .3s; border-radius: 50%;
}
input:checked + .slider { background-color: var(--c-primary); }
input:checked + .slider:before { transform: translateX(18px); }
</style>
