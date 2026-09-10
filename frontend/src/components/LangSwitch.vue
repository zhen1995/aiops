<template>
  <select class="lang-switch" :value="locale" :title="$t('common.language')" @change="onChange">
    <option v-for="l in SUPPORTED_LOCALES" :key="l.value" :value="l.value">{{ l.label }}</option>
  </select>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES, setLocale } from '../i18n'

const { locale } = useI18n({ useScope: 'global' })

function onChange(e) {
  setLocale(e.target.value)
}
</script>

<style scoped>
.lang-switch {
  padding: 5px 8px;
  border: 1px solid var(--c-border);
  border-radius: 6px;
  background: var(--c-surface);
  color: var(--c-text-2);
  font-size: 12.5px;
  outline: none;
  cursor: pointer;
  font-family: inherit;
}
.lang-switch:hover { border-color: var(--c-primary); color: var(--c-primary); }
</style>
