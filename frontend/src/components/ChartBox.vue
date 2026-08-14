<template>
  <div ref="el" :style="{ height, width: '100%' }"></div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  option: { type: Object, required: true },
  height: { type: String, default: '300px' }
})

const el = ref(null)
let chart = null
let ro = null

onMounted(() => {
  chart = echarts.init(el.value)
  chart.setOption(props.option)
  ro = new ResizeObserver(() => chart && chart.resize())
  ro.observe(el.value)
})

watch(() => props.option, (opt) => { chart && chart.setOption(opt, true) }, { deep: true })

onBeforeUnmount(() => {
  ro && ro.disconnect()
  chart && chart.dispose()
})
</script>
