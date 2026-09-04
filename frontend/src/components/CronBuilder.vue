<template>
  <div class="cron-builder">
    <!-- 秒字段开关 -->
    <div class="cb-toggle" v-if="allowSeconds">
      <label class="switch-sm">
        <input type="checkbox" v-model="state.useSeconds" />
        <span class="slider"></span>
      </label>
      <span class="cb-toggle-label">显示秒字段（表达式为 6 段）</span>
    </div>

    <!-- 字段 Tab -->
    <div class="cb-tabs">
      <button
        v-for="f in visibleFields"
        :key="f.key"
        :class="['cb-tab', { active: state.activeField === f.key }]"
        @click="state.activeField = f.key"
      >{{ f.label }}</button>
    </div>

    <!-- 当前字段的配置面板 -->
    <div class="cb-panel">
      <template v-if="currentField">
        <div class="cb-section">
          <label class="cb-radio">
            <input type="radio" :name="'mode_' + currentField.key" value="every" v-model="currentState.mode" />
            <span>每 {{ currentField.label }} ({{ currentField.wildcard }})</span>
          </label>
        </div>

        <div class="cb-section">
          <label class="cb-radio">
            <input type="radio" :name="'mode_' + currentField.key" value="interval" v-model="currentState.mode" />
            <span>每隔</span>
          </label>
          <div v-if="currentState.mode === 'interval'" class="cb-inline">
            <input
              type="number" min="1" :max="currentField.max"
              v-model.number="currentState.intervalN"
              class="num-input"
            />
            <span class="cb-unit">{{ currentField.unitLabel }}</span>
          </div>
        </div>

        <div class="cb-section">
          <label class="cb-radio">
            <input type="radio" :name="'mode_' + currentField.key" value="range" v-model="currentState.mode" />
            <span>范围</span>
          </label>
          <div v-if="currentState.mode === 'range'" class="cb-inline">
            <input
              type="number" :min="currentField.min" :max="currentField.max"
              v-model.number="currentState.rangeStart"
              class="num-input"
            />
            <span class="cb-sep">—</span>
            <input
              type="number" :min="currentField.min" :max="currentField.max"
              v-model.number="currentState.rangeEnd"
              class="num-input"
            />
          </div>
        </div>

        <div class="cb-section">
          <label class="cb-radio">
            <input type="radio" :name="'mode_' + currentField.key" value="specific" v-model="currentState.mode" />
            <span>指定 {{ currentField.unitLabel }} <span class="cb-hint">（多选）</span></span>
          </label>
          <div v-if="currentState.mode === 'specific'" class="cb-chips">
            <button
              v-for="v in currentField.values" :key="v.value"
              :class="['cb-chip', { active: currentState.values.includes(v.value) }]"
              @click="toggleValue(currentField.key, v.value)"
            >{{ v.label }}</button>
          </div>
        </div>

        <!-- 日 vs 周 互斥提示 -->
        <div v-if="currentField.key === 'day' || currentField.key === 'week'" class="cb-note">
          <span v-if="currentField.key === 'day'">配置日后，周字段将自动设为 <code>?</code></span>
          <span v-else>配置周后，日字段将自动设为 <code>?</code></span>
        </div>
      </template>
    </div>

    <!-- 底部预览 + 按钮 -->
    <div class="cb-footer">
      <div class="cb-expr">
        <code class="cb-expr-text">{{ cronExpression }}</code>
      </div>
      <div class="cb-summary">{{ summaryText }}</div>
      <div class="cb-actions">
        <button class="btn btn-sm" @click="$emit('cancel')">取消</button>
        <button class="btn btn-sm btn-primary" @click="confirm">确定</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, watch, onMounted } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  allowSeconds: { type: Boolean, default: true }
})
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

// 字段定义（5 段 + 可选秒）
const allFields = {
  second: { key: 'second', label: '秒', wildcard: '*', min: 0, max: 59, unitLabel: '秒', values: [] },
  minute: { key: 'minute', label: '分钟', wildcard: '*', min: 0, max: 59, unitLabel: '分钟', values: [] },
  hour: { key: 'hour', label: '小时', wildcard: '*', min: 0, max: 23, unitLabel: '小时', values: [] },
  day: { key: 'day', label: '日', wildcard: '*', min: 1, max: 31, unitLabel: '日', values: [] },
  month: { key: 'month', label: '月', wildcard: '*', min: 1, max: 12, unitLabel: '月', values: [] },
  week: { key: 'week', label: '周', wildcard: '?', min: 1, max: 7, unitLabel: '周', values: [] }
}

// 为每个字段预先生成可选值列表（避免响应式深拷贝问题）
const weekLabels = ['', '周一', '周二', '周三', '周四', '周五', '周六', '周日']
const monthLabels = ['', '1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']
Object.keys(allFields).forEach(k => {
  const f = allFields[k]
  f.values = []
  for (let i = f.min; i <= f.max; i++) {
    let label
    if (k === 'week') label = weekLabels[i]
    else if (k === 'month') label = monthLabels[i]
    else label = String(i).padStart(2, '0')
    f.values.push({ value: i, label })
  }
})

function makeFieldState(mode = 'every') {
  return { mode, intervalN: 1, rangeStart: 0, rangeEnd: 59, values: [] }
}

function initAllFields() {
  state.fields = {}
  state.fields.second = makeFieldState('every')
  state.fields.minute = makeFieldState('every')
  state.fields.hour = makeFieldState('every')
  state.fields.day = makeFieldState('every')
  state.fields.month = makeFieldState('every')
  state.fields.week = makeFieldState('every') // every → '?'
}

// state
const state = reactive({
  useSeconds: true,
  activeField: 'minute',
  fields: {}
})

// 初始化 state
initAllFields()

// 反向解析 cron 表达式到 UI state
function parseField(raw, fieldDef) {
  const s = makeFieldState('every')
  s.rangeStart = fieldDef.min
  s.rangeEnd = fieldDef.max
  if (raw === '*' || raw === '?') {
    s.mode = 'every'
    return s
  }
  const m = raw.match(/^\*\/(\d+)$/)
  if (m) { s.mode = 'interval'; s.intervalN = parseInt(m[1]); return s }
  const r = raw.match(/^(\d+)-(\d+)\/(\d+)$/)
  if (r) {
    s.mode = 'range'
    s.rangeStart = parseInt(r[1]); s.rangeEnd = parseInt(r[2])
    s.intervalN = parseInt(r[3]); return s
  }
  const r2 = raw.match(/^(\d+)-(\d+)$/)
  if (r2) {
    s.mode = 'range'
    s.rangeStart = parseInt(r2[1]); s.rangeEnd = parseInt(r2[2])
    return s
  }
  if (/^[\d,]+$/.test(raw)) {
    s.mode = 'specific'
    s.values = raw.split(',').map(x => parseInt(x))
    return s
  }
  return s
}

function setupFromCron(expr) {
  initAllFields()
  if (!expr) return
  const parts = expr.trim().split(/\s+/)
  if (parts.length === 6) {
    state.useSeconds = !!props.allowSeconds
    const keys = ['second', 'minute', 'hour', 'day', 'month', 'week']
    keys.forEach((k, i) => { state.fields[k] = parseField(parts[i], allFields[k]) })
  } else if (parts.length === 5) {
    state.useSeconds = false
    const keys = ['minute', 'hour', 'day', 'month', 'week']
    keys.forEach((k, i) => { state.fields[k] = parseField(parts[i], allFields[k]) })
  }
}

// 序列化单个字段
function serializeField(key) {
  const fd = allFields[key]
  const s = state.fields[key] || makeFieldState('every')
  const W = '?'

  // 日/周 互斥
  if (key === 'day' && (state.fields.week || {}).mode !== 'every') return W
  if (key === 'week' && (state.fields.day || {}).mode !== 'every') return W

  switch (s.mode) {
    case 'every': return key === 'week' ? W : '*'
    case 'interval': return `*/${s.intervalN || 1}`
    case 'range': {
      if (s.rangeStart >= s.rangeEnd) return String(s.rangeStart)
      let e = `${s.rangeStart}-${s.rangeEnd}`
      if (s.intervalN > 1) e += `/${s.intervalN}`
      return e
    }
    case 'specific': {
      if (!s.values || !s.values.length) return key === 'week' ? W : '*'
      if (s.values.length === fd.max - fd.min + 1) return key === 'week' ? W : '*'
      return [...s.values].sort((a, b) => a - b).join(',')
    }
  }
  return key === 'week' ? W : '*'
}

function toggleValue(key, v) {
  const arr = state.fields[key].values
  const i = arr.indexOf(v)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(v)
}

// 确保 activeField 在可见字段内
const visibleFieldKeys = computed(() => {
  const base = ['minute', 'hour', 'day', 'month', 'week']
  if (state.useSeconds && props.allowSeconds) return ['second', ...base]
  return base
})

// 根据 keys 获取完整 field 定义对象
const visibleFields = computed(() => visibleFieldKeys.value.map(k => allFields[k]))

// 确保 activeField 有效
watch(visibleFieldKeys, (keys) => {
  if (!keys.includes(state.activeField)) {
    state.activeField = keys[0]
  }
}, { immediate: true })

// 当前字段定义对象
const currentField = computed(() => allFields[state.activeField])
// 当前字段的 state
const currentState = computed(() => state.fields[state.activeField] || makeFieldState('every'))

// 最终 cron 表达式
const cronExpression = computed(() => visibleFieldKeys.value.map(k => serializeField(k)).join(' '))

// 自然语言摘要
const summaryText = computed(() => {
  const dayS = serializeField('day')
  const weekS = serializeField('week')
  const hourS = serializeField('hour')
  const minS = serializeField('minute')

  const parts = []
  if (minS !== '*' && !minS.startsWith('*/') && minS !== '0') parts.push(`在 ${minS} 分`)

  if (weekS !== '?' && weekS !== '*') {
    const labels = ['', '周一', '周二', '周三', '周四', '周五', '周六', '周日']
    if (weekS.includes('-')) {
      const [a, b] = weekS.split('-').map(Number)
      parts.push(`每周 ${labels[a]}~${labels[b]}`)
    } else if (weekS.includes(',')) {
      parts.push(`每周 ${weekS.split(',').map(x => labels[Number(x)]).join('、')}`)
    } else {
      parts.push(`每周 ${labels[Number(weekS)]}`)
    }
  } else if (dayS !== '*') {
    parts.push(`每月 ${dayS} 日`)
  } else {
    parts.push('每天')
  }

  if (hourS !== '*' && !hourS.startsWith('*/')) parts.push(`${hourS} 点`)
  else if (hourS.startsWith('*/')) parts.push(`每 ${hourS.slice(2)} 小时`)

  return parts.join(' · ')
})

function confirm() {
  emit('update:modelValue', cronExpression.value)
  emit('confirm', cronExpression.value)
}

// mount 后 + modelValue 变化时解析
onMounted(() => {
  setupFromCron(props.modelValue || defaultCron())
})
watch(() => props.modelValue, (v) => {
  if (v) setupFromCron(v)
})

function defaultCron() {
  return '0 0 9 * * ?'
}
</script>

<style scoped>
.cron-builder {
  margin-top: 12px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  background: var(--c-bg);
  overflow: hidden;
}

.cb-toggle {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--c-border);
  background: var(--c-surface);
}
.cb-toggle-label { font-size: 12.5px; color: var(--c-text-2); }

.switch-sm { position: relative; display: inline-block; width: 30px; height: 16px; cursor: pointer; }
.switch-sm input { display: none; }
.switch-sm .slider {
  position: absolute; inset: 0; background: #ccc; border-radius: 16px; transition: 0.2s;
}
.switch-sm .slider::before {
  content: ''; position: absolute; width: 12px; height: 12px; left: 2px; top: 2px;
  background: #fff; border-radius: 50%; transition: 0.2s;
}
.switch-sm input:checked + .slider { background: var(--c-primary); }
.switch-sm input:checked + .slider::before { transform: translateX(14px); }

.cb-tabs {
  display: flex;
  border-bottom: 1px solid var(--c-border);
  background: var(--c-surface);
}
.cb-tab {
  flex: 1;
  padding: 10px 0;
  font-size: 13px;
  border: none;
  background: transparent;
  color: var(--c-text-2);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: 0.15s;
}
.cb-tab:hover { color: var(--c-text); }
.cb-tab.active {
  color: var(--c-primary);
  border-bottom-color: var(--c-primary);
  background: var(--c-primary-soft, rgba(22,119,255,0.06));
  font-weight: 600;
}

.cb-panel { padding: 14px 16px; }

.cb-section { margin-bottom: 12px; }
.cb-radio {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; color: var(--c-text); cursor: pointer;
  padding: 4px 0;
}
.cb-radio input[type="radio"] { accent-color: var(--c-primary); }

.cb-inline {
  display: flex; align-items: center; gap: 8px; padding-left: 22px;
  margin-top: 6px;
}
.num-input {
  width: 72px;
  padding: 5px 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  font-size: 13px;
  background: var(--c-surface);
  color: var(--c-text);
  outline: none;
}
.num-input:focus { border-color: var(--c-primary); }
.cb-sep { color: var(--c-text-3); }
.cb-unit { font-size: 12px; color: var(--c-text-3); }

.cb-hint { color: var(--c-text-3); font-size: 11px; }

.cb-chips {
  display: flex; flex-wrap: wrap; gap: 5px;
  padding-left: 22px;
  margin-top: 6px;
  max-height: 140px; overflow-y: auto;
}
.cb-chip {
  padding: 3px 9px;
  border-radius: 4px;
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  font-size: 12px;
  color: var(--c-text-2);
  cursor: pointer;
  transition: 0.12s;
}
.cb-chip:hover { color: var(--c-text); }
.cb-chip.active {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
}

.cb-note {
  margin-top: 10px;
  padding: 6px 10px;
  font-size: 11.5px;
  background: var(--c-primary-soft, rgba(22,119,255,0.08));
  border-radius: 6px;
  color: var(--c-text-2);
}
.cb-note code {
  background: var(--c-surface);
  padding: 0 4px;
  border-radius: 3px;
  font-size: 11px;
}

.cb-footer {
  border-top: 1px solid var(--c-border);
  padding: 12px 16px;
  background: var(--c-surface);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.cb-expr-text {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 14px;
  color: var(--c-primary);
  background: var(--c-bg);
  padding: 6px 12px;
  border-radius: 6px;
  letter-spacing: 1px;
}
.cb-summary {
  font-size: 12px;
  color: var(--c-text-2);
}
.cb-actions {
  display: flex; gap: 8px; justify-content: flex-end;
}

.btn-sm {
  padding: 5px 14px;
  font-size: 12.5px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  color: var(--c-text);
}
.btn-sm:hover { border-color: var(--c-primary); }
.btn-sm.btn-primary {
  background: var(--c-primary);
  border-color: var(--c-primary);
  color: #fff;
}
.btn-sm.btn-primary:hover { opacity: 0.9; }
</style>
