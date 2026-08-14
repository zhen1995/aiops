<template>
  <div>
    <PageHeader title="Skill 管理" desc="管理 AI 助手可调用的能力（Skill），包括告警解读、根因分析、巡检报告等">
      <button class="btn btn-primary">新增 Skill</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">Skill 列表</h3>
          <p class="card-sub">已注册的智能运维能力</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>Skill 名称</th>
            <th>描述</th>
            <th>版本</th>
            <th>触发方式</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in skills" :key="s.id">
            <td><b>{{ s.name }}</b></td>
            <td class="muted" style="max-width: 360px">{{ s.description }}</td>
            <td class="mono">{{ s.version }}</td>
            <td>{{ s.trigger }}</td>
            <td><LevelTag :level="s.status === 'running' ? 'running' : 'info'" /></td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm" :class="s.status === 'running' ? '' : 'btn-primary'" @click="toggle(s)">
                  {{ s.status === 'running' ? '停用' : '启用' }}
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
import { skills as rawSkills } from '../../mock/data'

const skills = ref(rawSkills)

const toggle = (s) => {
  s.status = s.status === 'running' ? 'paused' : 'running'
}
</script>

<style scoped>
.ops { display: flex; gap: 8px; }
</style>
