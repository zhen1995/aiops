<template>
  <div>
    <PageHeader title="通知媒介" desc="维护钉钉机器人、Webhook 回调等通知通道配置">
      <button class="btn btn-primary">新增媒介</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">媒介列表</h3>
          <p class="card-sub">通知策略将关联以下媒介进行告警分发</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>媒介名称</th>
            <th>类型</th>
            <th>Webhook / 配置</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in media" :key="m.id">
            <td><b>{{ m.name }}</b></td>
            <td>
              <span class="type-tag" :class="m.type">{{ typeText(m.type) }}</span>
            </td>
            <td class="mono muted" style="max-width: 360px; overflow: hidden; text-overflow: ellipsis">
              {{ m.config.webhook }}
            </td>
            <td><LevelTag :level="m.status === 'active' ? 'running' : 'info'" /></td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm" @click="test(m)">测试</button>
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
import { notificationMedia as rawMedia } from '../../mock/data'

const media = ref(rawMedia)

const typeText = (type) => {
  const map = { dingtalk: '钉钉机器人', webhook: 'Webhook 回调' }
  return map[type] || type
}

const toggle = (m) => {
  m.status = m.status === 'active' ? 'paused' : 'active'
}

const test = (m) => {
  alert(`已向「${m.name}」发送测试消息（原型演示）`)
}
</script>

<style scoped>
.ops { display: flex; gap: 8px; }

.type-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  font-weight: 600;
}

.type-tag.dingtalk { background: #e6f4ff; color: #1677ff; }
.type-tag.webhook { background: var(--c-primary-soft); color: var(--c-primary); }
</style>
