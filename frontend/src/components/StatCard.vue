<template>
  <div class="card stat-card">
    <div class="stat-head">
      <span class="stat-label">
        {{ label }}
        <span v-if="help" class="help-tip" tabindex="0">
          <span class="help-tip-icon">?</span>
          <span class="help-tip-text" role="tooltip">{{ help }}</span>
        </span>
      </span>
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
  hint: String,
  help: String // 指标说明，悬停问号图标显示悬浮窗
})
</script>

<style scoped>
.stat-card { display: flex; flex-direction: column; gap: 6px; }
.stat-head { display: flex; align-items: center; justify-content: space-between; }
.stat-label { font-size: 13px; color: var(--c-text-2); }
.help-tip { position: relative; display: inline-flex; margin-left: 4px; vertical-align: -1px; }
.help-tip-icon {
  width: 14px; height: 14px; border-radius: 50%;
  border: 1px solid var(--c-text-3); color: var(--c-text-3);
  font-size: 10px; line-height: 12px; text-align: center; cursor: help;
}
.help-tip:hover .help-tip-icon, .help-tip:focus .help-tip-icon { border-color: var(--c-primary); color: var(--c-primary); }
.help-tip-text {
  position: absolute; bottom: calc(100% + 8px); left: 50%; transform: translateX(-50%);
  width: max-content; max-width: 240px; padding: 8px 12px;
  background: #2f3635; color: #fff; font-size: 12px; line-height: 1.6;
  border-radius: 6px; white-space: normal; text-align: left; font-weight: 400;
  visibility: hidden; opacity: 0; transition: opacity .15s ease; z-index: 20;
  pointer-events: none; box-shadow: 0 4px 12px rgba(0, 0, 0, .18);
}
.help-tip-text::after {
  content: ""; position: absolute; top: 100%; left: 50%; transform: translateX(-50%);
  border: 5px solid transparent; border-top-color: #2f3635;
}
.help-tip:hover .help-tip-text, .help-tip:focus .help-tip-text { visibility: visible; opacity: 1; }
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
