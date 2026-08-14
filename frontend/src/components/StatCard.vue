<template>
  <div class="card stat-card">
    <div class="stat-head">
      <span class="stat-label">{{ label }}</span>
      <span v-if="$slots.icon" class="stat-icon"><slot name="icon" /></span>
    </div>
    <div class="stat-value">{{ value }}<small v-if="unit">{{ unit }}</small></div>
    <div class="stat-foot">
      <span v-if="delta" class="stat-delta" :class="deltaType">{{ delta }}</span>
      <span v-if="hint" class="muted">{{ hint }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  label: String,
  value: [String, Number],
  unit: String,
  delta: String,
  deltaType: { type: String, default: 'up' }, // up(青) / down(红) / flat
  hint: String
})
</script>

<style scoped>
.stat-card { display: flex; flex-direction: column; gap: 6px; }
.stat-head { display: flex; align-items: center; justify-content: space-between; }
.stat-label { font-size: 13px; color: var(--c-text-2); }
.stat-icon {
  width: 34px; height: 34px; border-radius: 8px;
  background: var(--c-primary-tint); color: var(--c-primary);
  display: flex; align-items: center; justify-content: center;
}
.stat-value { font-size: 26px; font-weight: 700; line-height: 1.2; letter-spacing: -0.5px; }
.stat-value small { font-size: 13px; font-weight: 400; color: var(--c-text-3); margin-left: 4px; }
.stat-foot { display: flex; align-items: center; gap: 8px; font-size: 12px; }
.stat-delta { padding: 1px 7px; border-radius: var(--radius-tag); font-weight: 600; }
.stat-delta.up { color: var(--c-primary); background: var(--c-primary-tint); }
.stat-delta.down { color: var(--c-p0); background: var(--c-p0-bg); }
.stat-delta.flat { color: var(--c-text-3); background: var(--c-p4-bg); }
</style>
