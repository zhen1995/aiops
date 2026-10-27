<template>
  <div>
    <PageHeader :title="$t('knowledge.header.title')" :desc="$t('knowledge.header.desc')" />

    <!-- 页签 -->
    <div class="tabs">
      <div class="tab" :class="{ active: tab === 'wiki' }" @click="switchTab('wiki')">
        {{ $t('knowledge.wiki.tabs.wiki') }}
      </div>
      <div class="tab" :class="{ active: tab === 'docs' }" @click="switchTab('docs')">
        {{ $t('knowledge.wiki.tabs.docs') }}
      </div>
      <div class="tab" :class="{ active: tab === 'review' }" @click="switchTab('review')">
        {{ $t('knowledge.wiki.tabs.review') }}
        <span v-if="draftCount > 0" class="tab-badge">{{ draftCount }}</span>
      </div>
    </div>

    <!-- ============ Wiki 浏览 ============ -->
    <div v-if="tab === 'wiki'" class="page-split">
      <aside class="card group-panel">
        <div class="gp-head">
          <span class="gp-title">{{ $t('knowledge.wiki.category.title') }}</span>
          <button class="gp-add" :title="$t('knowledge.wiki.category.add')" @click="openCategoryModal(null)">+</button>
        </div>
        <div class="gp-tree">
          <div class="gp-node" :class="{ active: !filterCategoryId }" @click="selectCategory('')">
            <span class="gp-label">{{ $t('knowledge.wiki.entry.colCategory') }}</span>
          </div>
          <template v-for="c in categoryFlat" :key="c.id">
            <div
              class="gp-node"
              :class="{ active: filterCategoryId === c.id }"
              :style="{ paddingLeft: 8 + c.depth * 16 + 'px' }"
              @click="selectCategory(c.id)"
            >
              <span
                v-if="c.hasChildren"
                class="gp-arrow"
                @click.stop="toggleExpand(c.id)"
              >{{ isExpanded(c.id) ? '▾' : '▸' }}</span>
              <span v-else class="gp-arrow gp-arrow-spacer" />
              <input
                v-if="editingCategoryId === c.id"
                v-model="editingCategoryName"
                class="gp-rename-input"
                maxlength="128"
                @click.stop
                @keyup.enter="commitCategoryRename"
                @keyup.esc="cancelCategoryRename"
                @blur="commitCategoryRename"
                v-focus
              />
              <span v-else class="gp-label" :title="c.name">{{ c.name }}</span>
              <span v-if="editingCategoryId !== c.id" class="gp-count" v-show="c.entry_count > 0">{{ c.entry_count }}</span>
              <span v-if="editingCategoryId !== c.id" class="gp-ops">
                <button class="gp-op" :title="$t('knowledge.wiki.category.edit')" @click.stop="startCategoryRename(c)">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                    <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04a1 1 0 0 0 0-1.41l-2.34-2.34a1 1 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                  </svg>
                </button>
                <button class="gp-op gp-op-danger" :title="$t('alert.rules.delete')" @click.stop="deleteCategory(c)">
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                    <path d="M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                  </svg>
                </button>
              </span>
            </div>
          </template>
          <div v-if="categories.length === 0" class="gp-hint">{{ $t('knowledge.wiki.category.empty') }}</div>
        </div>
      </aside>

      <div class="card wiki-main">
        <div class="card-head-flex">
          <div>
            <h3 class="card-title">{{ $t('knowledge.wiki.tabs.wiki') }}</h3>
            <p class="card-sub">{{ $t('knowledge.header.desc') }}</p>
          </div>
          <div class="toolbar-right">
            <input
              v-model="entryKeyword"
              class="gp-search-input entry-search"
              :placeholder="$t('knowledge.wiki.entry.searchPlaceholder')"
              @keyup.enter="loadEntries"
            />
            <button class="btn btn-primary btn-sm" @click="openEntryModal(null)">
              + {{ $t('knowledge.wiki.entry.add') }}
            </button>
          </div>
        </div>
        <div class="table-scroll">
          <table class="table">
            <thead>
              <tr>
                <th>{{ $t('knowledge.wiki.entry.colTitle') }}</th>
                <th>{{ $t('knowledge.wiki.entry.colStatus') }}</th>
                <th>{{ $t('knowledge.wiki.entry.colCategory') }}</th>
                <th>{{ $t('knowledge.wiki.entry.colSummary') }}</th>
                <th>{{ $t('knowledge.wiki.entry.colUpdatedAt') }}</th>
                <th>{{ $t('knowledge.wiki.entry.colActions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="e in entries" :key="e.id">
                <td>
                  <a class="entry-title-link" @click="openEntryDetail(e.id)">{{ e.title }}</a>
                </td>
                <td><span class="entry-status" :class="e.status">{{ statusText(e.status) }}</span></td>
                <td>
                  <span v-if="categoryName(e.category_id)" class="type-tag">{{ categoryName(e.category_id) }}</span>
                  <span v-else class="muted">{{ $t('knowledge.wiki.entry.uncategorized') }}</span>
                </td>
                <td class="summary-cell muted">{{ e.summary || '-' }}</td>
                <td class="muted col-time">{{ formatTime(e.updated_at) }}</td>
                <td>
                  <div class="ops">
                    <button class="btn btn-sm" @click="openEntryModal(e)">{{ $t('alert.rules.edit') }}</button>
                    <button v-if="e.status === 'draft'" class="btn btn-sm btn-primary" @click="changeStatus(e, 'publish')">{{ $t('knowledge.wiki.entry.publish') }}</button>
                    <button v-if="e.status === 'published'" class="btn btn-sm" @click="changeStatus(e, 'archive')">{{ $t('knowledge.wiki.entry.archive') }}</button>
                    <button v-if="e.status === 'archived'" class="btn btn-sm" @click="changeStatus(e, 'withdraw')">{{ $t('knowledge.wiki.entry.withdraw') }}</button>
                    <button class="btn btn-sm btn-danger" @click="removeEntry(e)">{{ $t('alert.rules.delete') }}</button>
                  </div>
                </td>
              </tr>
              <tr v-if="entries.length === 0">
                <td colspan="6" class="empty">{{ $t('knowledge.wiki.entry.empty') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ============ 文档源 ============ -->
    <div v-if="tab === 'docs'">
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
          <button class="btn btn-primary" @click="triggerUpload">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
            </svg>
            {{ $t('knowledge.header.upload') }}
          </button>
          <input
            ref="fileInput"
            type="file"
            accept=".doc,.docx,.pdf,.md,.txt,.log,.csv,.json,.xml,.yaml,.yml"
            multiple
            style="display: none"
            @change="handleFiles"
          />
        </div>
        <table class="table">
          <thead>
            <tr>
              <th>{{ $t('knowledge.table.fileName') }}</th>
              <th>{{ $t('knowledge.table.type') }}</th>
              <th>{{ $t('knowledge.table.size') }}</th>
              <th>{{ $t('knowledge.table.status') }}</th>
              <th>{{ $t('knowledge.table.failReason') }}</th>
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
              <td>
                <button
                  v-if="f.status === 'failed'"
                  class="link-btn"
                  @click="showFailReason(f)"
                >{{ $t('knowledge.table.viewReason') }}</button>
                <span v-else class="muted">-</span>
              </td>
              <td>{{ f.uploader || '-' }}</td>
              <td class="muted">{{ formatTime(f.created_at) }}</td>
              <td>
                <div class="ops">
                  <button
                    v-if="isTextDoc(f) && f.status === 'indexed'"
                    class="btn btn-sm btn-primary"
                    :disabled="generatingDocId === f.id"
                    @click="generateEntries(f)"
                  >{{ generatingDocId === f.id ? $t('knowledge.wiki.entry.generating') : $t('knowledge.wiki.entry.generate') }}</button>
                  <button class="btn btn-sm" @click="reindex(f)" v-if="f.status === 'indexed' || f.status === 'failed'">{{ $t('knowledge.table.reindex') }}</button>
                  <button class="btn btn-sm" @click="removeFile(f.id)">{{ $t('knowledge.table.delete') }}</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredFiles.length === 0">
              <td colspan="8" class="empty">{{ filterKw ? $t('knowledge.table.emptyNoMatch') : $t('knowledge.table.emptyNoDocs') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ============ 待审核 ============ -->
    <div v-if="tab === 'review'" class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('knowledge.wiki.tabs.review') }}</h3>
          <p class="card-sub">{{ $t('knowledge.wiki.entry.generateDone', { count: draftCount }) }}</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="table">
          <thead>
            <tr>
              <th>{{ $t('knowledge.wiki.entry.colTitle') }}</th>
              <th>{{ $t('knowledge.wiki.entry.colCategory') }}</th>
              <th>{{ $t('knowledge.wiki.entry.colSummary') }}</th>
              <th>{{ $t('knowledge.wiki.entry.source') }}</th>
              <th>{{ $t('knowledge.wiki.entry.colUpdatedAt') }}</th>
              <th>{{ $t('knowledge.wiki.entry.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="e in draftEntries" :key="e.id">
              <td><a class="entry-title-link" @click="openEntryDetail(e.id)">{{ e.title }}</a></td>
              <td>
                <span v-if="categoryName(e.category_id)" class="type-tag">{{ categoryName(e.category_id) }}</span>
                <span v-else class="muted">{{ $t('knowledge.wiki.entry.uncategorized') }}</span>
              </td>
              <td class="summary-cell muted">{{ e.summary || '-' }}</td>
              <td class="muted">{{ e.source_name || '-' }}</td>
              <td class="muted col-time">{{ formatTime(e.updated_at) }}</td>
              <td>
                <div class="ops">
                  <button class="btn btn-sm" @click="openEntryModal(e)">{{ $t('alert.rules.edit') }}</button>
                  <button class="btn btn-sm btn-primary" @click="changeStatus(e, 'publish', true)">{{ $t('knowledge.wiki.entry.publish') }}</button>
                  <button class="btn btn-sm btn-danger" @click="removeEntry(e, true)">{{ $t('alert.rules.delete') }}</button>
                </div>
              </td>
            </tr>
            <tr v-if="draftEntries.length === 0">
              <td colspan="6" class="empty">{{ $t('knowledge.wiki.entry.empty') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ============ 词条详情弹窗 ============ -->
    <div v-if="detailEntry" class="modal-mask" @click.self="detailEntry = null">
      <div class="modal detail-modal">
        <div class="modal-header">
          <h3>
            {{ detailEntry.title }}
            <span class="entry-status" :class="detailEntry.status">{{ statusText(detailEntry.status) }}</span>
          </h3>
          <span class="modal-close" @click="detailEntry = null">×</span>
        </div>
        <div class="modal-body detail-body">
          <div class="detail-meta">
            <span v-if="categoryName(detailEntry.category_id)" class="type-tag">{{ categoryName(detailEntry.category_id) }}</span>
            <span v-for="tag in parseTags(detailEntry.tags)" :key="tag" class="tag-chip">{{ tag }}</span>
            <span v-if="detailEntry.source_name" class="muted detail-source">
              {{ $t('knowledge.wiki.entry.source') }}：{{ detailEntry.source_name }}
            </span>
          </div>
          <MarkdownContent :content="detailEntry.content" class="detail-content" />
        </div>
        <div class="modal-footer">
          <button v-if="detailEntry.status === 'draft'" class="btn btn-primary" @click="changeStatus(detailEntry, 'publish'); detailEntry = null">
            {{ $t('knowledge.wiki.entry.publish') }}
          </button>
          <button class="btn" @click="detailEntry = null">{{ $t('knowledge.failDialog.close') }}</button>
        </div>
      </div>
    </div>

    <!-- ============ 词条编辑弹窗 ============ -->
    <div v-if="entryModal.visible" class="modal-mask" @click.self="closeEntryModal">
      <div class="modal entry-modal">
        <div class="modal-header">
          <h3>{{ entryModal.isEdit ? $t('knowledge.wiki.entry.edit') : $t('knowledge.wiki.entry.add') }}</h3>
          <span class="modal-close" @click="closeEntryModal">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('knowledge.wiki.entry.titleLabel') }}</label>
            <input v-model="entryForm.title" class="form-input" :placeholder="$t('knowledge.wiki.entry.titlePlaceholder')" maxlength="200" />
            <span v-if="entryErrors.title" class="form-error">{{ entryErrors.title }}</span>
          </div>
          <div class="form-item">
            <label class="form-label">{{ $t('knowledge.wiki.entry.categoryLabel') }}</label>
            <div class="sel-wrap">
              <button type="button" class="sel-trigger" @click="entryCatOpen = !entryCatOpen">
                <span class="sel-trigger-label">{{ entryCategoryLabel }}</span>
                <span class="sel-arrow">▾</span>
              </button>
              <div v-if="entryCatOpen" class="more-mask" @click="entryCatOpen = false"></div>
              <div v-if="entryCatOpen" class="sel-menu">
                <div class="sel-option" :class="{ active: entryForm.category_id === '' }" @click="pickEntryCategory('')">
                  {{ $t('knowledge.wiki.entry.uncategorized') }}
                </div>
                <div
                  v-for="c in categoryOptions"
                  :key="c.id"
                  class="sel-option"
                  :class="{ active: entryForm.category_id === c.id }"
                  @click="pickEntryCategory(c.id)"
                >{{ c.label }}</div>
              </div>
            </div>
          </div>
          <div class="form-item">
            <label class="form-label">{{ $t('knowledge.wiki.entry.tagsLabel') }}</label>
            <input v-model="entryForm.tagsInput" class="form-input" :placeholder="$t('knowledge.wiki.entry.tagsPlaceholder')" />
          </div>
          <div class="form-item">
            <label class="form-label">{{ $t('knowledge.wiki.entry.summaryLabel') }}</label>
            <input v-model="entryForm.summary" class="form-input" :placeholder="$t('knowledge.wiki.entry.summaryPlaceholder')" maxlength="500" />
          </div>
          <div class="form-item">
            <label class="form-label required">{{ $t('knowledge.wiki.entry.contentLabel') }}</label>
            <textarea
              v-model="entryForm.content"
              class="form-input form-textarea content-textarea"
              rows="10"
              :placeholder="$t('knowledge.wiki.entry.contentPlaceholder')"
            />
            <span v-if="entryErrors.content" class="form-error">{{ entryErrors.content }}</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="closeEntryModal" :disabled="entryModal.saving">{{ $t('knowledge.failDialog.close') }}</button>
          <button class="btn" @click="saveEntry(false)" :disabled="entryModal.saving">
            {{ entryModal.saving ? $t('system.actions.saving') : $t('system.actions.save') }}
          </button>
          <button class="btn btn-primary" @click="saveEntry(true)" :disabled="entryModal.saving">
            {{ entryModal.saving ? $t('system.actions.saving') : $t('knowledge.wiki.entry.publish') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ============ 分类编辑弹窗 ============ -->
    <div v-if="categoryModal.visible" class="modal-mask" @click.self="categoryModal.visible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ $t('knowledge.wiki.category.add') }}</h3>
          <span class="modal-close" @click="categoryModal.visible = false">×</span>
        </div>
        <div class="modal-body">
          <div class="form-item">
            <label class="form-label required">{{ $t('knowledge.wiki.category.name') }}</label>
            <input
              v-model="categoryForm.name"
              class="form-input"
              maxlength="128"
              @keyup.enter="saveCategory"
            />
            <span v-if="categoryError" class="form-error">{{ categoryError }}</span>
          </div>
          <div class="form-item">
            <label class="form-label">{{ $t('knowledge.wiki.category.parent') }}</label>
            <div class="sel-wrap">
              <button type="button" class="sel-trigger" @click="catModalOpen = !catModalOpen">
                <span class="sel-trigger-label">{{ catModalParentLabel }}</span>
                <span class="sel-arrow">▾</span>
              </button>
              <div v-if="catModalOpen" class="more-mask" @click="catModalOpen = false"></div>
              <div v-if="catModalOpen" class="sel-menu">
                <div class="sel-option" :class="{ active: categoryForm.parent_id === '' }" @click="pickCatModalParent('')">
                  {{ $t('knowledge.wiki.category.none') }}
                </div>
                <div
                  v-for="c in categoryOptions"
                  :key="c.id"
                  class="sel-option"
                  :class="{ active: categoryForm.parent_id === c.id }"
                  @click="pickCatModalParent(c.id)"
                >{{ c.label }}</div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="categoryModal.visible = false" :disabled="categoryModal.saving">{{ $t('knowledge.failDialog.close') }}</button>
          <button class="btn btn-primary" @click="saveCategory" :disabled="categoryModal.saving">
            {{ categoryModal.saving ? $t('system.actions.saving') : $t('system.actions.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 索引失败原因弹窗 -->
    <div v-if="failDetail" class="modal-mask" @click.self="failDetail = null">
      <div class="modal fail-modal" role="dialog" :aria-label="$t('knowledge.failDialog.title')">
        <div class="modal-header">
          <h3>{{ $t('knowledge.failDialog.title') }}</h3>
          <span class="modal-close" role="button" tabindex="0" :aria-label="$t('knowledge.failDialog.close')"
                @click="failDetail = null" @keydown.enter="failDetail = null">&times;</span>
        </div>
        <div class="modal-body">
          <p class="fail-modal-file">{{ failDetail.name }}</p>
          <pre class="fail-modal-error">{{ failDetail.error_msg || $t('knowledge.failDialog.empty') }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import { knowledgeApi } from '../api/knowledge.js'
import { wikiApi } from '../api/knowledgeWiki.js'

const { t } = useI18n({ useScope: 'global' })

const route = useRoute()
const router = useRouter()

const vFocus = {
  mounted: (el) => { el.focus(); el.select && el.select() }
}

// ---------- 页签 ----------
const tab = ref('wiki')

function switchTab(next) {
  tab.value = next
  if (next === 'review') loadDrafts()
}

// ---------- 文档源（原有逻辑） ----------
const files = ref([])
const fileInput = ref(null)
let pollTimer = null

const failDetail = ref(null)
const showFailReason = (doc) => { failDetail.value = doc }

const onKeydown = (e) => {
  if (e.key === 'Escape') { failDetail.value = null; detailEntry.value = null }
}
window.addEventListener('keydown', onKeydown)

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
    failed: t('knowledge.status.failed'),
    draft: t('knowledge.wiki.status.draft'),
    published: t('knowledge.wiki.status.published'),
    archived: t('knowledge.wiki.status.archived')
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

// 支持 LLM 提炼词条的文本类文档
const textExts = ['txt', 'md', 'log', 'csv', 'json', 'xml', 'yaml', 'yml', 'conf', 'ini', 'sh']
const isTextDoc = (f) => textExts.includes((f.type || '').toLowerCase())

const generatingDocId = ref('')

async function generateEntries(f) {
  generatingDocId.value = f.id
  try {
    const res = await wikiApi.generateFromDocument(f.id)
    alert(t('knowledge.wiki.entry.generateDone', { count: res?.count || 0 }))
    await loadDraftCount()
  } catch (err) {
    alert(t('knowledge.wiki.msg.generateFailed', { msg: err.message }))
  } finally {
    generatingDocId.value = ''
  }
}

// ---------- Wiki 分类树 ----------
const categories = ref([])
const filterCategoryId = ref('')
const expanded = ref(new Set())
const editingCategoryId = ref(null)
const editingCategoryName = ref('')
let catRenameSnapshot = ''

const childrenMap = computed(() => {
  const m = {}
  for (const c of categories.value) {
    const pid = c.parent_id || ''
    if (!m[pid]) m[pid] = []
    m[pid].push(c)
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

const categoryFlat = computed(() => {
  const out = []
  const walk = (nodes, depth) => {
    for (const n of nodes) {
      const children = childrenMap.value[n.id] || []
      out.push({ ...n, depth, hasChildren: children.length > 0 })
      if (children.length > 0 && isExpanded(n.id)) walk(children, depth + 1)
    }
  }
  walk(childrenMap.value[''] || [], 0)
  return out
})

// 层级缩进选项（编辑词条/分类弹窗用）
const categoryOptions = computed(() => {
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

const categoryName = (id) => (categories.value.find(c => c.id === id) || {}).name || ''

async function loadCategories() {
  try {
    categories.value = await wikiApi.categories() || []
    const s = new Set(expanded.value)
    for (const c of categories.value) {
      if ((childrenMap.value[c.id] || []).length > 0) s.add(c.id)
    }
    expanded.value = s
  } catch (err) {
    alert(t('knowledge.wiki.msg.loadCategoriesFailed', { msg: err.message }))
  }
}

function selectCategory(id) {
  if (editingCategoryId.value) return
  filterCategoryId.value = id
  loadEntries()
}

function startCategoryRename(c) {
  editingCategoryId.value = c.id
  editingCategoryName.value = c.name
  catRenameSnapshot = c.name
}

async function commitCategoryRename() {
  const id = editingCategoryId.value
  if (!id) return
  const name = editingCategoryName.value.trim()
  editingCategoryId.value = null
  if (!name || name === catRenameSnapshot) return
  try {
    await wikiApi.updateCategory(id, { name })
    await loadCategories()
  } catch (err) {
    alert(t('knowledge.wiki.msg.saveCategoryFailed', { msg: err.message }))
    await loadCategories()
  }
}

function cancelCategoryRename() {
  editingCategoryId.value = null
}

async function deleteCategory(c) {
  if (!confirm(t('knowledge.wiki.category.confirmDelete', { name: c.name }))) return
  try {
    await wikiApi.removeCategory(c.id)
    if (filterCategoryId.value === c.id) filterCategoryId.value = ''
    await loadCategories()
    await loadEntries()
  } catch (err) {
    alert(t('knowledge.wiki.msg.deleteCategoryFailed', { msg: err.message }))
  }
}

// ---------- 分类弹窗 ----------
const categoryModal = reactive({ visible: false, saving: false })
const categoryForm = reactive({ name: '', parent_id: '' })
const categoryError = ref('')
const catModalOpen = ref(false)

const catModalParentLabel = computed(() => {
  if (!categoryForm.parent_id) return t('knowledge.wiki.category.none')
  return categoryName(categoryForm.parent_id)
})

function pickCatModalParent(id) {
  categoryForm.parent_id = id
  catModalOpen.value = false
}

function openCategoryModal() {
  categoryForm.name = ''
  categoryForm.parent_id = ''
  categoryError.value = ''
  categoryModal.visible = true
}

async function saveCategory() {
  categoryError.value = ''
  if (!categoryForm.name.trim()) {
    categoryError.value = t('knowledge.wiki.error.categoryNameRequired')
    return
  }
  categoryModal.saving = true
  try {
    await wikiApi.createCategory({ name: categoryForm.name, parent_id: categoryForm.parent_id })
    categoryModal.saving = false
    categoryModal.visible = false
    await loadCategories()
  } catch (err) {
    alert(t('knowledge.wiki.msg.saveCategoryFailed', { msg: err.message }))
    categoryModal.saving = false
  }
}

// ---------- 词条 ----------
const entries = ref([])
const draftEntries = ref([])
const draftCount = ref(0)
const entryKeyword = ref('')

const entryForm = reactive({
  title: '', category_id: '', tagsInput: '', summary: '', content: ''
})
const entryErrors = reactive({ title: '', content: '' })
const entryModal = reactive({ visible: false, isEdit: false, editingId: null, saving: false })
const entryCatOpen = ref(false)
const detailEntry = ref(null)

const entryCategoryLabel = computed(() => {
  if (!entryForm.category_id) return t('knowledge.wiki.entry.uncategorized')
  return categoryName(entryForm.category_id)
})

function pickEntryCategory(id) {
  entryForm.category_id = id
  entryCatOpen.value = false
}

async function loadEntries() {
  try {
    const data = await wikiApi.entries({
      category_id: filterCategoryId.value,
      q: entryKeyword.value.trim(),
      limit: 200
    })
    entries.value = data?.list || []
  } catch (err) {
    alert(t('knowledge.wiki.msg.loadEntriesFailed', { msg: err.message }))
  }
}

async function loadDrafts() {
  try {
    const data = await wikiApi.entries({ status: 'review', limit: 200 })
    draftEntries.value = data?.list || []
  } catch (err) {
    alert(t('knowledge.wiki.msg.loadEntriesFailed', { msg: err.message }))
  }
}

async function loadDraftCount() {
  try {
    const data = await wikiApi.entries({ status: 'review', limit: 1 })
    draftCount.value = data?.total || 0
  } catch { /* 忽略角标加载失败 */ }
}

function openEntryModal(entry) {
  entryModal.isEdit = !!entry
  entryModal.editingId = entry ? entry.id : null
  entryForm.title = entry ? entry.title : ''
  entryForm.category_id = entry ? (entry.category_id || '') : (filterCategoryId.value || '')
  entryForm.tagsInput = entry ? parseTags(entry.tags).join(', ') : ''
  entryForm.summary = entry ? (entry.summary || '') : ''
  entryForm.content = entry ? entry.content : ''
  entryErrors.title = ''
  entryErrors.content = ''
  entryCatOpen.value = false
  entryModal.visible = true
}

function closeEntryModal() {
  if (entryModal.saving) return
  entryModal.visible = false
}

function parseTags(raw) {
  if (!raw) return []
  try {
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

async function saveEntry(publish) {
  entryErrors.title = ''
  entryErrors.content = ''
  if (!entryForm.title.trim()) {
    entryErrors.title = t('knowledge.wiki.error.titleRequired')
    return
  }
  if (!entryForm.content.trim()) {
    entryErrors.content = t('knowledge.wiki.error.contentRequired')
    return
  }
  const tags = entryForm.tagsInput.split(/[,，]/).map(s => s.trim()).filter(Boolean)
  const payload = {
    title: entryForm.title,
    category_id: entryForm.category_id,
    tags: JSON.stringify(tags),
    summary: entryForm.summary,
    content: entryForm.content
  }
  entryModal.saving = true
  try {
    if (entryModal.isEdit) {
      await wikiApi.updateEntry(entryModal.editingId, payload)
    } else {
      await wikiApi.createEntry(payload, publish)
    }
    entryModal.saving = false
    entryModal.visible = false
    await loadEntries()
    await loadDraftCount()
    await loadCategories()
  } catch (err) {
    alert(t('knowledge.wiki.msg.saveEntryFailed', { msg: err.message }))
    entryModal.saving = false
  }
}

async function openEntryDetail(id) {
  try {
    detailEntry.value = await wikiApi.getEntry(id)
  } catch (err) {
    alert(t('knowledge.wiki.msg.loadEntriesFailed', { msg: err.message }))
  }
}

async function changeStatus(entry, action, isDraftTab = false) {
  try {
    if (action === 'publish') await wikiApi.publishEntry(entry.id)
    else if (action === 'archive') await wikiApi.archiveEntry(entry.id)
    else if (action === 'withdraw') await wikiApi.withdrawEntry(entry.id)
    await loadEntries()
    await loadDraftCount()
    await loadCategories()
    if (isDraftTab) await loadDrafts()
  } catch (err) {
    alert(t('knowledge.wiki.msg.statusFailed', { msg: err.message }))
  }
}

async function removeEntry(entry, isDraftTab = false) {
  if (!confirm(t('knowledge.wiki.entry.confirmDelete', { title: entry.title }))) return
  try {
    await wikiApi.removeEntry(entry.id)
    await loadEntries()
    await loadDraftCount()
    await loadCategories()
    if (isDraftTab) await loadDrafts()
  } catch (err) {
    alert(t('knowledge.wiki.msg.deleteEntryFailed', { msg: err.message }))
  }
}

// ---------- 上传 / 文档操作（原有逻辑） ----------
const triggerUpload = () => { fileInput.value?.click() }

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

// ---------- 生命周期 ----------
onMounted(async () => {
  await Promise.all([loadFiles(), loadCategories(), loadEntries(), loadDraftCount()])
  handleSearchParam()
})

function handleSearchParam() {
  const { doc, q } = route.query
  if (!doc && !q) return
  if (q) filterKw.value = String(q)
  if (doc) highlightDocId.value = String(doc)
  tab.value = 'docs'
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
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
/* ---------- 页签 ---------- */
.tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--c-border);
}

.tab {
  padding: 9px 18px;
  font-size: 13.5px;
  color: var(--c-text-2);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: all 0.15s;
  user-select: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.tab:hover { color: var(--c-primary); }

.tab.active {
  color: var(--c-primary);
  font-weight: 600;
  border-bottom-color: var(--c-primary);
}

.tab-badge {
  font-size: 11px;
  font-weight: 600;
  background: var(--c-danger);
  color: #fff;
  border-radius: 9px;
  padding: 0 6px;
  line-height: 16px;
}

/* ---------- Wiki 两栏 ---------- */
.page-split {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.group-panel {
  width: 240px;
  flex-shrink: 0;
  padding: 14px;
  box-sizing: border-box;
}

.wiki-main {
  flex: 1;
  min-width: 0;
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

.gp-tree {
  max-height: calc(100vh - 300px);
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

.entry-search { width: 200px; }

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

/* ---------- 词条表格 ---------- */
.table-scroll { overflow-x: auto; }

.entry-title-link {
  color: var(--c-primary);
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.entry-title-link:hover {
  opacity: 0.8;
  text-decoration: underline;
}

.entry-status {
  font-size: 11px;
  font-weight: 600;
  border-radius: 10px;
  padding: 2px 9px;
  white-space: nowrap;
}

.entry-status.draft { background: #fdf6dc; color: #d9a400; }
.entry-status.published { background: var(--c-primary-tint); color: var(--c-primary); }
.entry-status.archived { background: var(--c-bg); color: var(--c-text-3); border: 1px solid var(--c-border); }

.summary-cell {
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-time { white-space: nowrap; }

.tag-chip {
  font-size: 11px;
  background: var(--c-primary-soft, rgba(22, 119, 255, 0.08));
  color: var(--c-primary);
  padding: 1px 8px;
  border-radius: 10px;
  white-space: nowrap;
}

/* ---------- 弹窗 ---------- */
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
  box-shadow: var(--shadow-modal);
  width: 580px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 18px 24px;
  border-bottom: 1px solid var(--c-border);
}
.modal-header h3 { margin: 0; font-size: 16px; font-weight: 600; display: flex; align-items: center; gap: 10px; }

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

.detail-modal { width: 760px; }

.detail-body { padding: 16px 24px 20px; }

.detail-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

.detail-source { font-size: 12px; }

.detail-content {
  font-size: 13.5px;
  line-height: 1.7;
  color: var(--c-text);
  word-break: break-word;
}

.entry-modal { width: 720px; }

.form-item { margin-bottom: 14px; }

.form-label {
  display: block;
  font-size: 13px;
  color: var(--c-text-2);
  margin-bottom: 6px;
  font-weight: 500;
}
.form-label.required::before { content: '*'; color: var(--c-danger); margin-right: 3px; }

.form-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 13px;
  transition: border-color 0.15s;
  outline: none;
  font-family: inherit;
  box-sizing: border-box;
}
.form-input:focus { border-color: var(--c-primary); }
.form-input::placeholder { color: var(--c-text-3); }
.form-textarea { resize: vertical; min-height: 70px; }
.content-textarea { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12.5px; }

.form-error {
  display: block;
  color: var(--c-danger);
  font-size: 12px;
  margin-top: 4px;
}

/* 自定义下拉 */
.sel-wrap { position: relative; }

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
.sel-trigger:hover { border-color: var(--c-primary); }
.sel-trigger-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sel-arrow { color: var(--c-text-3); font-size: 11px; flex-shrink: 0; }

.more-mask {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 90;
}

.sel-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 220px;
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
.sel-option:hover { background: var(--c-primary-soft, rgba(22, 119, 255, 0.08)); color: var(--c-primary); }
.sel-option.active { background: var(--c-primary-tint); color: var(--c-primary); font-weight: 600; }

/* ---------- 文档源（原有样式） ---------- */
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

.link-btn {
  background: none;
  border: none;
  padding: 0;
  font-size: inherit;
  color: var(--c-danger);
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 3px;
}
.link-btn:hover { opacity: 0.8; }

.fail-modal { width: min(560px, calc(100vw - 48px)); }
.fail-modal-file {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--c-text-2);
  word-break: break-all;
}
.fail-modal-error {
  margin: 0;
  padding: 12px 14px;
  background: var(--c-p0-bg);
  border: 1px solid rgba(201, 59, 59, 0.2);
  border-radius: 8px;
  color: var(--c-p0);
  font-size: 12.5px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: inherit;
}
</style>
