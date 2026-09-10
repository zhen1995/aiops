<template>
  <div>
    <PageHeader :title="$t('llm.skill.title')" :desc="$t('llm.skill.desc')">
      <button class="btn btn-primary">{{ $t('llm.skill.add') }}</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">{{ $t('llm.skill.listTitle') }}</h3>
          <p class="card-sub">{{ $t('llm.skill.listSub') }}</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>{{ $t('llm.skill.colName') }}</th>
            <th>{{ $t('llm.skill.colDesc') }}</th>
            <th>{{ $t('llm.skill.colVersion') }}</th>
            <th>{{ $t('llm.skill.colTrigger') }}</th>
            <th>{{ $t('llm.skill.colStatus') }}</th>
            <th>{{ $t('llm.skill.colActions') }}</th>
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
                <button class="btn btn-sm">{{ $t('llm.skill.edit') }}</button>
                <button class="btn btn-sm" :class="s.status === 'running' ? '' : 'btn-primary'" @click="toggle(s)">
                  {{ s.status === 'running' ? $t('llm.skill.disable') : $t('llm.skill.enable') }}
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
