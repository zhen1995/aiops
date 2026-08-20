<template>
  <div class="login-page">
    <div class="login-left">
      <div class="brand">
        <div class="logo-mark">
          <svg viewBox="0 0 24 24" width="32" height="32" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 12h4l3-8 4 16 3-8h4" />
          </svg>
        </div>
        <div class="brand-text">
          <h1>AIOPS</h1>
          <span>智能运维平台</span>
        </div>
      </div>
      <div class="slogan">
        <h2>面向大规模分布式系统的<br/>智能运维平台</h2>
        <p>融合多源监控数据，实现异常检测、根因分析、告警降噪与智能运维助手</p>
      </div>
      <div class="features">
        <div class="feature-item">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 12h4l3-8 4 16 3-8h4" />
            </svg>
          </div>
          <div>
            <b>异常检测</b>
            <span>基于机器学习的智能异常检测</span>
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3"/><circle cx="12" cy="12" r="8"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2"/>
            </svg>
          </div>
          <div>
            <b>根因分析</b>
            <span>自动化故障根因定位</span>
          </div>
        </div>
        <div class="feature-item">
          <div class="feature-icon">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/>
            </svg>
          </div>
          <div>
            <b>告警降噪</b>
            <span>智能压缩与聚合告警风暴</span>
          </div>
        </div>
      </div>
    </div>

    <div class="login-right">
      <div class="login-form-wrap">
        <div class="form-header">
          <h2>欢迎登录</h2>
          <p>请输入您的账号信息</p>
        </div>

        <form @submit.prevent="handleLogin" class="login-form">
          <div class="form-group">
            <label>用户名</label>
            <div class="input-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
              </svg>
              <input
                v-model="username"
                type="text"
                placeholder="请输入用户名"
                autocomplete="username"
                :disabled="loading"
                @keyup.enter="handleLogin"
              />
            </div>
          </div>

          <div class="form-group">
            <label>密码</label>
            <div class="input-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="请输入密码"
                autocomplete="current-password"
                :disabled="loading"
                @keyup.enter="handleLogin"
              />
              <span class="input-action" @click="showPassword = !showPassword">
                {{ showPassword ? '隐藏' : '显示' }}
              </span>
            </div>
          </div>

          <div class="form-options">
            <label class="remember">
              <input type="checkbox" v-model="rememberMe" />
              <span>记住我</span>
            </label>
          </div>

          <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

          <button type="submit" class="submit-btn" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>

        <div class="form-footer">
          <span>AIOPS 智能运维平台 © 2026</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../api/auth.js'
import { setAuth } from '../utils/auth.js'

const router = useRouter()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const rememberMe = ref(true)
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  errorMsg.value = ''

  if (!username.value.trim()) {
    errorMsg.value = '请输入用户名'
    return
  }
  if (!password.value.trim()) {
    errorMsg.value = '请输入密码'
    return
  }

  loading.value = true
  try {
    const result = await authApi.login(username.value, password.value)
    setAuth(result.token, result.user)
    router.push('/chat')
  } catch (err) {
    errorMsg.value = err.message || '登录失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

/* 左侧品牌区域 */
.login-left {
  flex: 1;
  background: linear-gradient(135deg, #0a5f57 0%, #0e7c72 60%, #1a9b8e 100%);
  color: #fff;
  padding: 60px 64px;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
}

.login-left::before {
  content: '';
  position: absolute;
  top: -150px;
  right: -150px;
  width: 400px;
  height: 400px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 50%;
}

.login-left::after {
  content: '';
  position: absolute;
  bottom: -100px;
  left: -100px;
  width: 300px;
  height: 300px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 50%;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  position: relative;
  z-index: 1;
}

.logo-mark {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.brand-text h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 1px;
}

.brand-text span {
  font-size: 12px;
  opacity: 0.85;
}

.slogan {
  margin-top: 100px;
  position: relative;
  z-index: 1;
}

.slogan h2 {
  font-size: 30px;
  font-weight: 600;
  line-height: 1.4;
  margin: 0 0 16px;
}

.slogan p {
  font-size: 14px;
  opacity: 0.8;
  line-height: 1.7;
  max-width: 400px;
}

.features {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 18px;
  position: relative;
  z-index: 1;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 14px;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 14px 18px;
}

.feature-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.feature-item b {
  display: block;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 2px;
}

.feature-item span {
  font-size: 12px;
  opacity: 0.8;
}

/* 右侧登录表单 */
.login-right {
  flex: 0 0 480px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.login-form-wrap {
  width: 100%;
  max-width: 360px;
}

.form-header {
  text-align: center;
  margin-bottom: 36px;
}

.form-header h2 {
  font-size: 24px;
  font-weight: 600;
  color: var(--c-text);
  margin: 0 0 8px;
}

.form-header p {
  font-size: 14px;
  color: var(--c-text-3);
  margin: 0;
}

.login-form .form-group {
  margin-bottom: 20px;
}

.login-form .form-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--c-text-2);
  margin-bottom: 8px;
}

.input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.input-wrap .input-icon {
  position: absolute;
  left: 12px;
  color: var(--c-text-3);
  pointer-events: none;
}

.input-wrap input {
  width: 100%;
  padding: 10px 40px 10px 38px;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  font-size: 14px;
  background: var(--c-bg);
  color: var(--c-text);
  outline: none;
  transition: border-color 0.15s, background 0.15s;
}

.input-wrap input:focus {
  border-color: var(--c-primary);
  background: #fff;
}

.input-wrap input::placeholder {
  color: var(--c-text-3);
}

.input-action {
  position: absolute;
  right: 12px;
  cursor: pointer;
  font-size: 12px;
  color: var(--c-primary);
  user-select: none;
  padding: 2px 4px;
}

.input-action:hover {
  color: var(--c-primary-dark);
}

.form-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.remember {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--c-text-2);
  cursor: pointer;
}

.remember input {
  accent-color: var(--c-primary);
}

.error-msg {
  background: var(--c-p0-bg);
  color: var(--c-p0);
  font-size: 13px;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 16px;
  border: 1px solid rgba(201, 59, 59, 0.2);
}

.submit-btn {
  width: 100%;
  padding: 11px;
  background: var(--c-primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.submit-btn:hover:not(:disabled) {
  background: var(--c-primary-dark);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.form-footer {
  margin-top: 48px;
  text-align: center;
  font-size: 12px;
  color: var(--c-text-3);
}

/* 响应式 */
@media (max-width: 960px) {
  .login-left {
    display: none;
  }
  .login-right {
    flex: 1;
  }
}
</style>