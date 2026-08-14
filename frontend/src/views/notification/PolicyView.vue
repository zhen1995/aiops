<template>
  <div>
    <PageHeader title="通知策略" desc="创建通知策略并关联通知媒介，实现告警分级分发">
      <button class="btn btn-primary">新增策略</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">策略列表</h3>
          <p class="card-sub">按告警级别、服务、时间等条件匹配通知媒介</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>策略名称</th>
            <th>匹配级别</th>
            <th>触发条件</th>
            <th>延迟</th>
            <th>关联媒介</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in policies" :key="p.id">
            <td><b>{{ p.name }}</b></td>
            <td>
              <span v-for="lv in p.level" :key="lv" class="level-dot" :class="lv">{{ lv.toUpperCase() }}</span>
            </td>
            <td class="muted">{{ p.conditions }}</td>
            <td>{{ p.delayMinutes }} 分钟</td>
            <td>
              <div class="channels">
                <span v-for="ch in resolveChannels(p.channels)" :key="ch.id" class="channel-tag">{{ ch.name }}</span>
              </div>
            </td>
            <td><LevelTag :level="p.status === 'active' ? 'running' : 'info'" /></td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm" :class="p.status === 'active' ? '' : 'btn-primary'" @click="toggle(p)">
                  {{ p.status === 'active' ? '停用' : '启用' }}
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
import { notificationPolicies as rawPolicies, notificationMedia } from '../../mock/data'

const policies = ref(rawPolicies)

const resolveChannels = (ids) => {
  return notificationMedia.filter((m) => ids.includes(m.id))
}

const toggle = (p) => {
  p.status = p.status === 'active' ? 'paused' : 'active'
}
</script>

<style scoped>
.ops { display: flex; gap: 8px; }

.level-dot {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  margin-right: 4px;
}

.level-dot.p0 { background: var(--c-p0-bg); color: var(--c-p0); }
.level-dot.p1 { background: var(--c-p1-bg); color: var(--c-p1); }
.level-dot.p2 { background: var(--c-p2-bg); color: var(--c-p2); }
.level-dot.p3 { background: var(--c-p3-bg); color: var(--c-p3); }
.level-dot.p4 { background: var(--c-p4-bg); color: var(--c-p4); }

.channels { display: flex; flex-wrap: wrap; gap: 6px; }

.channel-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  background: var(--c-primary-soft);
  color: var(--c-primary);
}
</style>
