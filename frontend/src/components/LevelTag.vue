<template>
  <span class="level-tag" :style="style">{{ text }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ level: { type: String, required: true } })

const map = {
  p0: ['P0', 'var(--c-p0)', 'var(--c-p0-bg)'],
  p1: ['P1', 'var(--c-p1)', 'var(--c-p1-bg)'],
  p2: ['P2', 'var(--c-p2)', 'var(--c-p2-bg)'],
  p3: ['P3', 'var(--c-p3)', 'var(--c-p3-bg)'],
  p4: ['P4', 'var(--c-p4)', 'var(--c-p4-bg)'],
  critical: ['严重', 'var(--c-p0)', 'var(--c-p0-bg)'],
  major: ['重要', 'var(--c-p1)', 'var(--c-p1-bg)'],
  minor: ['次要', 'var(--c-p2)', 'var(--c-p2-bg)'],
  warning: ['警告', 'var(--c-p2)', 'var(--c-p2-bg)'],
  info: ['提示', 'var(--c-p4)', 'var(--c-p4-bg)'],
  active: ['活动', 'var(--c-p0)', 'var(--c-p0-bg)'],
  acked: ['已确认', 'var(--c-p1)', 'var(--c-p1-bg)'],
  resolved: ['已解决', 'var(--c-success)', 'var(--c-success-bg)'],
  running: ['运行中', 'var(--c-primary)', 'var(--c-primary-tint)'],
  online: ['在线', 'var(--c-success)', 'var(--c-success-bg)'],
  error: ['ERROR', 'var(--c-p0)', 'var(--c-p0-bg)'],
  warn: ['WARN', 'var(--c-p1)', 'var(--c-p1-bg)']
}

const conf = computed(() => map[props.level?.toLowerCase()] || map.info)
const text = computed(() => conf.value[0])
const style = computed(() => ({ color: conf.value[1], background: conf.value[2] }))
</script>

<style scoped>
.level-tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
</style>
