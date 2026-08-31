<template>
  <div>
    <PageHeader title="告警规则" desc="查看 Nightingale 告警规则列表">
      <div class="header-actions">
        <div class="filter-box">
          <label>业务组ID</label>
          <input
            v-model="gids"
            type="text"
            placeholder="逗号分隔，留空使用引擎配置"
            @keyup.enter="loadData"
          />
        </div>
        <button class="btn btn-sm" @click="loadData" :disabled="loading">
          {{ loading ? '加载中...' : '刷新' }}
        </button>
      </div>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">规则列表</h3>
          <p class="card-sub">Nightingale 告警规则同步结果</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>规则名称</th>
            <th>业务组ID</th>
            <th>告警级别</th>
            <th>启用状态</th>
            <th>评估间隔（秒）</th>
            <th>持续时间（秒）</th>
            <th>当前事件数</th>
            <th>创建时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in rules" :key="rule.id">
            <td><b>{{ rule.name }}</b></td>
            <td>{{ rule.group_id }}</td>
            <td>
              <LevelTag :level="severityLevel(rule.severity)">
                {{ severityText[rule.severity] || rule.severity }}
              </LevelTag>
            </td>
            <td>
              <LevelTag :level="rule.disabled === 1 ? 'info' : 'running'">
                {{ statusText[rule.disabled] || rule.disabled }}
              </LevelTag>
            </td>
            <td>{{ rule.prom_eval_interval }}</td>
            <td>{{ rule.prom_for_duration }}</td>
            <td>{{ rule.cur_event_count }}</td>
            <td class="muted">{{ fmtTime(rule.create_at) }}</td>
          </tr>
          <tr v-if="!loading && rules.length === 0">
            <td colspan="8" class="empty-row">暂无告警规则数据</td>
          </tr>
          <tr v-if="loading">
            <td colspan="8" class="empty-row">加载中...</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import LevelTag from '../components/LevelTag.vue'
import { alertRuleApi } from '../api/alertRule.js'

const severityText = { 1: 'P1-紧急', 2: 'P2-警告', 3: 'P3-提醒' }
const statusText = { 0: '已启用', 1: '已禁用' }

const fmtTime = (ts) => ts ? new Date(ts * 1000).toLocaleString() : '-'

const severityLevel = (s) => {
  if (s === 1) return 'critical'
  if (s === 2) return 'warning'
  return 'info'
}

const rules = ref([])
const loading = ref(false)
const gids = ref('')

async function loadData() {
  loading.value = true
  try {
    const data = await alertRuleApi.list(gids.value.trim())
    rules.value = data || []
  } catch (err) {
    alert('加载告警规则失败：' + err.message)
    rules.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 6px 12px;
}

.filter-box label {
  font-size: 13px;
  color: var(--c-text-2);
  white-space: nowrap;
}

.filter-box input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  width: 200px;
  color: var(--c-text);
}

.empty-row {
  text-align: center;
  color: var(--c-text-3);
  padding: 40px 0 !important;
}
</style>
