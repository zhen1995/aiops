<template>
  <div class="page-header">
    <div>
      <h2>{{ title }}</h2>
      <p v-if="desc">{{ desc }}</p>
    </div>
    <div v-if="$slots.default" class="page-header-actions">
      <slot />
    </div>
  </div>
</template>

<script setup>
defineProps({ title: String, desc: String })
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}
h2 { margin: 0; font-size: 18px; font-weight: 600; }
p { margin: 3px 0 0; font-size: 13px; color: var(--c-text-2); }
.page-header-actions { display: flex; gap: 10px; }
</style>
