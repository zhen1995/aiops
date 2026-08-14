<template>
  <div>
    <PageHeader title="LLM 管理" desc="管理大语言模型接入配置，支持多厂商、多模型切换">
      <button class="btn btn-primary">新增模型</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">模型列表</h3>
          <p class="card-sub">已接入的 LLM 服务与运行状态</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>厂商</th>
            <th>模型</th>
            <th>接入端点</th>
            <th>API Key</th>
            <th>温度 / 最大 Token</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in llmConfigs" :key="m.id">
            <td><b>{{ m.name }}</b></td>
            <td>{{ m.provider }}</td>
            <td class="mono">{{ m.model }}</td>
            <td class="muted mono" style="max-width: 220px; overflow: hidden; text-overflow: ellipsis">{{ m.endpoint }}</td>
            <td class="mono">{{ m.apiKeyMasked }}</td>
            <td>{{ m.temperature }} / {{ m.maxTokens }}</td>
            <td><LevelTag :level="m.status === 'active' ? 'running' : 'info'" /></td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm" :class="m.status === 'active' ? '' : 'btn-primary'" @click="toggle(m)">
                  {{ m.status === 'active' ? '停用' : '启用' }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import PageHeader from '../../components/PageHeader.vue'
import LevelTag from '../../components/LevelTag.vue'
import { llmConfigs as rawConfigs } from '../../mock/data'

const llmConfigs = ref(rawConfigs)

const toggle = (m) => {
  m.status = m.status === 'active' ? 'paused' : 'active'
}
</script>

<style scoped>
.ops { display: flex; gap: 8px; }
</style>
