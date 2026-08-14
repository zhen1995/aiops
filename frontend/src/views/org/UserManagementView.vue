<template>
  <div>
    <PageHeader title="用户管理" desc="维护平台用户账号、所属部门与角色绑定">
      <button class="btn btn-primary">新增用户</button>
    </PageHeader>

    <div class="card">
      <div class="card-head-flex">
        <div>
          <h3 class="card-title">用户列表</h3>
          <p class="card-sub">平台全部账号及其角色分配情况</p>
        </div>
      </div>
      <table class="table">
        <thead>
          <tr>
            <th>用户名</th>
            <th>姓名</th>
            <th>邮箱</th>
            <th>手机号</th>
            <th>部门</th>
            <th>角色</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in userList" :key="u.id">
            <td><b>{{ u.username }}</b></td>
            <td>{{ u.realName }}</td>
            <td class="muted">{{ u.email }}</td>
            <td class="muted">{{ u.phone }}</td>
            <td>{{ u.department }}</td>
            <td>
              <span v-for="role in u.roles" :key="role" class="role-tag">{{ role }}</span>
            </td>
            <td><LevelTag :level="u.status === 'active' ? 'running' : 'info'" /></td>
            <td>
              <div class="ops">
                <button class="btn btn-sm">编辑</button>
                <button class="btn btn-sm" @click="resetPwd(u)">重置密码</button>
                <button class="btn btn-sm" :class="u.status === 'active' ? '' : 'btn-primary'" @click="toggle(u)">
                  {{ u.status === 'active' ? '停用' : '启用' }}
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
import { users as rawUsers } from '../../mock/data'

const userList = ref(rawUsers)

const toggle = (u) => {
  u.status = u.status === 'active' ? 'paused' : 'active'
}

const resetPwd = (u) => {
  alert(`已为用户「${u.realName}」发送密码重置邮件（原型演示）`)
}
</script>

<style scoped>
.ops { display: flex; gap: 8px; }

.role-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-tag);
  font-size: 12px;
  background: var(--c-primary-soft);
  color: var(--c-primary);
  margin-right: 6px;
}
</style>
