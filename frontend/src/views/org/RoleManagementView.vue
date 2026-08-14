<template>
  <div>
    <PageHeader title="角色管理" desc="维护平台角色、权限范围与角色成员">
      <button class="btn btn-primary">新增角色</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">角色列表</h3>
          <p class="card-sub">基于 RBAC 的角色权限配置</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>角色名称</th>
            <th>角色编码</th>
            <th>描述</th>
            <th>成员数</th>
            <th>权限范围</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in roleList" :key="r.id">
            <td><b>{{ r.name }}</b></td>
            <td class="mono muted">{{ r.code }}</td>
            <td class="muted" style="max-width: 280px; overflow: hidden; text-overflow: ellipsis">{{ r.description }}</td>
            <td>{{ r.userCount }} 人</td>
            <td>
              <span v-for="perm in r.permissions" :key="perm" class="perm-tag">{{ perm }}</span>
            </td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm">权限</button>
                <button class="btn btn-sm">删除</button>
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
import { roles as rawRoles } from '../../mock/data'

const roleList = ref(rawRoles)
</script>

<style scoped>
.ops { display: flex; gap: 8px; }

.perm-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  background: var(--c-primary-soft);
  color: var(--c-primary);
  margin-right: 6px;
  margin-bottom: 4px;
}
</style>
