<template>
  <router-view v-if="isFullPage" />
  <AppLayout v-else />
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from './layout/AppLayout.vue'

const route = useRoute()
const isFullPage = computed(() => route.path === '/login' || route.path === '/no-permission')
</script>