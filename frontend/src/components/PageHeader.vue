<template>
  <!-- 页面标题已由顶栏菜单展示，此处不再渲染标题，保留描述文字与操作按钮区 -->
  <div v-if="desc || $slots.default" class="page-header">
    <p v-if="desc">{{ desc }}</p>
    <div v-if="$slots.default" class="page-header-actions">
      <slot />
    </div>
  </div>
</template>

<script setup>
// title 保留声明以兼容既有调用（不再渲染，避免与顶栏菜单标题重复）
defineProps({ title: String, desc: String })
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
p { margin: 0; font-size: 13px; color: var(--c-text-2); }
.page-header-actions { display: flex; gap: 10px; flex-shrink: 0; }
</style>
