<template>
  <span class="level-tag" :style="style"><slot>{{ text }}</slot></span>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n({ useScope: 'global' })

const props = defineProps({ level: { type: String, required: true } })

// [字面文案, i18n key（字面文案为空时按 key 翻译）, 文字颜色, 背景色]
const map = {
  p0: ['P0', null, 'var(--c-p0)', 'var(--c-p0-bg)'],
  p1: ['P1', null, 'var(--c-p0)', 'var(--c-p0-bg)'],
  p2: ['P2', null, 'var(--c-p1)', 'var(--c-p1-bg)'],
  p3: ['P3', null, 'var(--c-p2)', 'var(--c-p2-bg)'],
  p4: ['P4', null, 'var(--c-p4)', 'var(--c-p4-bg)'],
  critical: [null, 'critical', 'var(--c-p0)', 'var(--c-p0-bg)'],
  major: [null, 'major', 'var(--c-p1)', 'var(--c-p1-bg)'],
  minor: [null, 'minor', 'var(--c-p2)', 'var(--c-p2-bg)'],
  warning: [null, 'warning', 'var(--c-p2)', 'var(--c-p2-bg)'],
  info: [null, 'info', 'var(--c-p4)', 'var(--c-p4-bg)'],
  active: [null, 'active', 'var(--c-p0)', 'var(--c-p0-bg)'],
  acked: [null, 'acked', 'var(--c-p1)', 'var(--c-p1-bg)'],
  resolved: [null, 'resolved', 'var(--c-success)', 'var(--c-success-bg)'],
  running: [null, 'running', 'var(--c-primary)', 'var(--c-primary-tint)'],
  online: [null, 'online', 'var(--c-success)', 'var(--c-success-bg)'],
  error: ['ERROR', null, 'var(--c-p0)', 'var(--c-p0-bg)'],
  warn: ['WARN', null, 'var(--c-p1)', 'var(--c-p1-bg)']
}

const conf = computed(() => map[props.level?.toLowerCase()] || map.info)
const text = computed(() => (conf.value[0] ? conf.value[0] : t(`common.levelTag.${conf.value[1]}`)))
const style = computed(() => ({ color: conf.value[2], background: conf.value[3] }))
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
