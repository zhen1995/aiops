<template>
  <div class="no-permission-page">
    <div class="no-permission-card">
      <div class="icon-wrap">
        <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          <path d="M9 12l2 2 4-4"/>
        </svg>
      </div>
      <h2>暂无访问权限</h2>
      <p>您的账号尚未分配任何菜单权限，请联系管理员为您配置角色权限。</p>
      <div class="actions">
        <button class="btn-logout" @click="handleLogout">退出登录</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { clearAuth } from '../utils/auth.js'

const router = useRouter()

function handleLogout() {
  clearAuth()
  router.push('/login')
}
</script>

<style scoped>
.no-permission-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--c-bg);
}

.no-permission-card {
  text-align: center;
  padding: 60px 80px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  max-width: 420px;
}

.icon-wrap {
  width: 88px;
  height: 88px;
  margin: 0 auto 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f87171 0%, #ef4444 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.no-permission-card h2 {
  margin: 0 0 12px;
  font-size: 22px;
  font-weight: 600;
  color: var(--c-text);
}

.no-permission-card p {
  margin: 0 0 28px;
  font-size: 14px;
  color: var(--c-text-3);
  line-height: 1.7;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}

.btn-logout {
  padding: 10px 28px;
  background: var(--c-primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-logout:hover {
  background: var(--c-primary-dark);
}
</style>
