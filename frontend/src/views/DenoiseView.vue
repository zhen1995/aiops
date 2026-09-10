<template>
  <div>
    <PageHeader :title="$t('alert.denoise.title')" :desc="$t('alert.denoise.desc')" />

    <!-- 统计卡片 -->
    <div v-if="loading" class="kpi-grid">
      <div class="card stat-placeholder">{{ $t('alert.denoise.loading') }}</div>
    </div>
    <div v-else-if="loadError" class="card error-card">
      <span>{{ loadError }}</span>
      <button class="btn btn-sm btn-primary" @click="loadAll">{{ $t('alert.denoise.retry') }}</button>
    </div>
    <div v-else class="kpi-grid">
      <StatCard v-for="k in kpiCards" :key="k.label" v-bind="k" />
    </div>

    <!-- 降噪漏斗 -->
    <div class="card funnel-card">
      <h3 class="card-title">{{ $t('alert.denoise.funnelTitle') }}</h3>
      <p class="card-sub">{{ $t('alert.denoise.funnelSub') }}</p>
      <ChartBox v-if="funnelData.length" :option="funnelOption" height="340px" />
      <p v-else class="empty-funnel muted">{{ $t('alert.denoise.funnelEmpty') }}</p>
    </div>

    <!-- 降噪策略 -->
    <div class="card">
      <h3 class="card-title">{{ $t('alert.denoise.policyTitle') }}</h3>
      <p class="card-sub">{{ $t('alert.denoise.policySub') }}</p>
      <div class="policy-grid">
        <div v-for="p in policies" :key="p.strategy" class="policy-card" :class="{ off: !p.enabled }">
          <div class="policy-head">
            <span class="policy-name">{{ p.name }}</span>
            <label class="switch" :title="p.enabled ? $t('alert.denoise.switchDisable') : $t('alert.denoise.switchEnable')">
              <input
                type="checkbox"
                :checked="p.enabled"
                :disabled="toggling === p.strategy"
                @change="togglePolicy(p, $event.target.checked)"
              />
              <span class="slider"></span>
            </label>
          </div>
          <p class="policy-desc">{{ p.description }}</p>
          <p class="policy-effect">{{ effectText(p.strategy) }}</p>
          <div class="policy-foot">
            <span class="muted">{{ $t('alert.denoise.suppressedToday') }}</span>
            <button class="policy-reduced mono" :title="$t('alert.denoise.viewRecords')" @click="goRecords(p)">
              {{ Number(p.suppressed_today || 0).toLocaleString() }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import { denoiseApi } from '../api/denoise.js'

const { t } = useI18n({ useScope: 'global' })
const router = useRouter()

const loading = ref(false)
const loadError = ref('')
const stats = ref(null)
const policies = ref([])
const toggling = ref('')

const num = (v) => (typeof v === 'number' ? v : Number(v) || 0)

// 兼容后端 snake_case 与历史 camelCase 字段
function pickStats(s) {
  return {
    raw_total: num(s?.raw_total ?? s?.rawTotal),
    effective: num(s?.effective ?? s?.final),
    compression_rate: num(s?.compression_rate ?? s?.compressionRate),
    suppressed_total: num(s?.suppressed_total ?? s?.suppressedTotal),
    funnel: Array.isArray(s?.funnel) ? s.funnel : []
  }
}

const kpiCards = computed(() => {
  const s = pickStats(stats.value)
  return [
    { label: t('alert.denoise.kpi.rawTotal'), value: s.raw_total.toLocaleString(), hint: t('alert.denoise.kpi.rawHint') },
    { label: t('alert.denoise.kpi.effective'), value: s.effective.toLocaleString(), hint: t('alert.denoise.kpi.effectiveHint') },
    { label: t('alert.denoise.kpi.compressionRate'), value: s.compression_rate.toFixed(2), unit: '%', hint: t('alert.denoise.kpi.compressionHint') },
    { label: t('alert.denoise.kpi.suppressed'), value: s.suppressed_total.toLocaleString(), hint: t('alert.denoise.kpi.suppressedHint') }
  ]
})

// 浅青 → 深青 层次配色（4 级漏斗）
const funnelColors = ['#c4e2de', '#9bd0ca', '#43a298', '#0a5f57']

const funnelData = computed(() => pickStats(stats.value).funnel)

const funnelOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    formatter: (params) => t('alert.denoise.funnelTooltip', { name: params.name, value: params.value })
  },
  series: [{
    type: 'funnel',
    left: '12%', right: '12%', top: 8, bottom: 8,
    minSize: '18%', maxSize: '100%',
    sort: 'descending',
    gap: 4,
    label: { show: true, position: 'inside', formatter: '{b}  {c}', fontSize: 12, color: '#20302d' },
    itemStyle: { borderColor: '#fff', borderWidth: 1 },
    emphasis: { label: { fontSize: 13, fontWeight: 600 } },
    data: funnelData.value.map((f, i) => ({
      ...f,
      itemStyle: { color: funnelColors[i % funnelColors.length] },
      label: { color: i >= 2 ? '#fff' : '#20302d' }
    }))
  }]
}))

// 策略效果文案（本地补充，优先展示接口返回的名称/描述）
function effectText(strategy) {
  if (strategy === 'window_aggregation') return t('alert.denoise.effect.windowAggregation')
  if (strategy === 'topology_suppression') return t('alert.denoise.effect.topologySuppression')
  return ''
}

// datetime-local 格式（yyyy-MM-ddTHH:mm），与通知记录页时间筛选输入框格式一致
function fmtLocal(d) {
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}

// 跳转通知记录：查看今日被该策略拦截的记录
function goRecords(policy) {
  const now = new Date()
  const dayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  router.push({
    path: '/notify/records',
    query: { status: 'intercepted', strategy: policy.strategy, start: fmtLocal(dayStart), end: fmtLocal(now) }
  })
}

async function loadAll() {
  loading.value = true
  loadError.value = ''
  try {
    const [statsData, policyData] = await Promise.all([
      denoiseApi.getStats(),
      denoiseApi.getPolicies()
    ])
    stats.value = statsData
    policies.value = (policyData?.policies || []).map((p) => ({ ...p }))
  } catch (err) {
    loadError.value = t('alert.denoise.msg.loadFailed', { msg: err.message })
  } finally {
    loading.value = false
  }
}

async function togglePolicy(policy, enabled) {
  toggling.value = policy.strategy
  try {
    await denoiseApi.updatePolicy(policy.strategy, enabled)
    await loadAll()
  } catch (err) {
    alert(t('alert.denoise.msg.toggleFailed', { msg: err.message }))
  } finally {
    toggling.value = ''
  }
}

onMounted(loadAll)
</script>

<style scoped>
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.stat-placeholder { grid-column: 1 / -1; text-align: center; color: var(--c-text-3); padding: 32px 0; }
.error-card { display: flex; align-items: center; justify-content: center; gap: 12px; color: var(--c-danger); margin-bottom: 16px; padding: 20px; }
.funnel-card { margin-bottom: 16px; }
.empty-funnel { text-align: center; padding: 40px 0; }
.policy-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 14px; }
.policy-card {
  border: 1px solid var(--c-border);
  border-radius: var(--radius-card);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: opacity 0.2s ease, filter 0.2s ease;
}
.policy-card.off { opacity: 0.55; filter: grayscale(0.6); }
.policy-head { display: flex; justify-content: space-between; align-items: center; }
.policy-name { font-size: 14px; font-weight: 600; }
.policy-desc { margin: 0; font-size: 12.5px; color: var(--c-text-2); line-height: 1.5; }
.policy-effect { margin: 0; font-size: 12.5px; font-weight: 600; color: var(--c-primary); }
.policy-foot { margin-top: auto; display: flex; justify-content: space-between; align-items: center; border-top: 1px dashed var(--c-border); padding-top: 8px; }
.policy-reduced {
  font-size: 16px; font-weight: 700; color: var(--c-primary-dark);
  border: none; background: none; padding: 0; cursor: pointer; font-family: inherit;
}
.policy-reduced:hover { color: var(--c-primary); text-decoration: underline; }

/* 开关 */
.switch { position: relative; width: 38px; height: 20px; flex-shrink: 0; cursor: pointer; }
.switch input { opacity: 0; width: 0; height: 0; }
.switch input:disabled + .slider { cursor: not-allowed; opacity: 0.7; }
.slider { position: absolute; inset: 0; border-radius: 20px; background: var(--c-p4); transition: background 0.2s; }
.slider::before {
  content: ''; position: absolute; left: 3px; top: 3px;
  width: 14px; height: 14px; border-radius: 50%;
  background: #fff; transition: transform 0.2s;
}
.switch input:checked + .slider { background: var(--c-primary); }
.switch input:checked + .slider::before { transform: translateX(18px); }

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
