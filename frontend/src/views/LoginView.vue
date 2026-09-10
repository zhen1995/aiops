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
          <span>{{ $t('login.brand') }}</span>
        </div>
      </div>
      <div class="hero-illustration">
        <svg viewBox="0 0 560 420" width="100%" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
          <!-- 网格底纹 -->
          <defs>
            <pattern id="grid" width="28" height="28" patternUnits="userSpaceOnUse">
              <path d="M28 0H0v28" stroke="rgba(255,255,255,0.06)" stroke-width="1"/>
            </pattern>
            <linearGradient id="areaFill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="rgba(255,255,255,0.28)"/>
              <stop offset="100%" stop-color="rgba(255,255,255,0)"/>
            </linearGradient>
          </defs>
          <rect x="0" y="0" width="560" height="420" fill="url(#grid)" rx="16"/>

          <!-- 监控大屏面板 -->
          <rect x="70" y="60" width="420" height="250" rx="14" fill="rgba(255,255,255,0.08)" stroke="rgba(255,255,255,0.22)" stroke-width="1.5"/>
          <!-- 面板标题栏 -->
          <rect x="70" y="60" width="420" height="36" rx="14" fill="rgba(255,255,255,0.08)"/>
          <rect x="70" y="82" width="420" height="14" fill="rgba(255,255,255,0.08)"/>
          <circle cx="92" cy="78" r="4" fill="rgba(255,255,255,0.55)"/>
          <circle cx="108" cy="78" r="4" fill="rgba(255,255,255,0.35)"/>
          <circle cx="124" cy="78" r="4" fill="rgba(255,255,255,0.2)"/>
          <rect x="330" y="72" width="66" height="12" rx="6" fill="rgba(255,255,255,0.18)"/>
          <circle cx="462" cy="78" r="5" fill="#5eead4"/>
          <circle cx="462" cy="78" r="9" stroke="rgba(94,234,212,0.4)" stroke-width="1.5"/>

          <!-- 折线图区域 -->
          <rect x="90" y="110" width="250" height="140" rx="8" fill="rgba(255,255,255,0.05)" stroke="rgba(255,255,255,0.12)"/>
          <line x1="90" y1="150" x2="340" y2="150" stroke="rgba(255,255,255,0.08)" stroke-dasharray="4 4"/>
          <line x1="90" y1="190" x2="340" y2="190" stroke="rgba(255,255,255,0.08)" stroke-dasharray="4 4"/>
          <line x1="90" y1="230" x2="340" y2="230" stroke="rgba(255,255,255,0.08)" stroke-dasharray="4 4"/>
          <path d="M96 200 L130 190 L160 205 L190 175 L220 185 L250 150 L280 160 L310 120 L336 130" stroke="rgba(255,255,255,0.85)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M96 200 L130 190 L160 205 L190 175 L220 185 L250 150 L280 160 L310 120 L336 130 L336 245 L96 245 Z" fill="url(#areaFill)"/>
          <!-- 异常尖峰 -->
          <path d="M190 175 L220 185 L250 150 L280 160 L310 120 L336 130" stroke="#fda4af" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" stroke-dasharray="1 0"/>
          <circle cx="310" cy="120" r="5" fill="#fb7185" stroke="#fff" stroke-width="2"/>
          <circle cx="310" cy="120" r="11" stroke="rgba(251,113,133,0.45)" stroke-width="1.5"/>
          <!-- 数据点 -->
          <circle cx="190" cy="175" r="3.5" fill="#fff"/>
          <circle cx="250" cy="150" r="3.5" fill="#fff"/>
          <circle cx="336" cy="130" r="3.5" fill="#fff"/>

          <!-- 环形仪表盘 -->
          <circle cx="420" cy="160" r="40" stroke="rgba(255,255,255,0.15)" stroke-width="9"/>
          <path d="M420 120 A40 40 0 1 1 384 178" stroke="#5eead4" stroke-width="9" stroke-linecap="round"/>
          <line x1="420" y1="160" x2="443" y2="138" stroke="#fff" stroke-width="3" stroke-linecap="round"/>
          <circle cx="420" cy="160" r="5" fill="#fff"/>
          <text x="420" y="224" text-anchor="middle" fill="rgba(255,255,255,0.75)" font-size="12" font-family="inherit">CPU 62%</text>
          <circle cx="380" cy="252" r="4" fill="#5eead4"/>
          <text x="392" y="256" fill="rgba(255,255,255,0.65)" font-size="11" font-family="inherit">{{ $t('login.chartNormal') }}</text>
          <circle cx="444" cy="252" r="4" fill="#fb7185"/>
          <text x="456" y="256" fill="rgba(255,255,255,0.65)" font-size="11" font-family="inherit">{{ $t('login.chartAlert') }}</text>

          <!-- 柱状图 -->
          <rect x="90" y="272" width="16" height="24" rx="3" fill="rgba(255,255,255,0.25)"/>
          <rect x="114" y="266" width="16" height="30" rx="3" fill="rgba(255,255,255,0.35)"/>
          <rect x="138" y="276" width="16" height="20" rx="3" fill="rgba(255,255,255,0.25)"/>
          <rect x="162" y="258" width="16" height="38" rx="3" fill="rgba(255,255,255,0.45)"/>
          <rect x="186" y="270" width="16" height="26" rx="3" fill="rgba(255,255,255,0.3)"/>
          <rect x="210" y="250" width="16" height="46" rx="3" fill="#5eead4" opacity="0.85"/>
          <rect x="234" y="264" width="16" height="32" rx="3" fill="rgba(255,255,255,0.35)"/>
          <rect x="258" y="274" width="16" height="22" rx="3" fill="rgba(255,255,255,0.25)"/>
          <rect x="282" y="256" width="16" height="40" rx="3" fill="rgba(255,255,255,0.4)"/>
          <rect x="306" y="268" width="16" height="28" rx="3" fill="rgba(255,255,255,0.3)"/>

          <!-- 告警铃铛 -->
          <g transform="translate(392 280)">
            <path d="M16 2a8 8 0 0 0-8 8c0 9-4 12-4 12h24s-4-3-4-12a8 8 0 0 0-8-8" stroke="rgba(255,255,255,0.85)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M13.5 26a2.5 2.5 0 0 0 5 0" stroke="rgba(255,255,255,0.85)" stroke-width="2.5" stroke-linecap="round"/>
            <circle cx="27" cy="4" r="5" fill="#fb7185"/>
          </g>

          <!-- 服务器节点拓扑 -->
          <g stroke="rgba(255,255,255,0.35)" stroke-width="1.5" stroke-dasharray="3 4">
            <line x1="140" y1="340" x2="250" y2="356"/>
            <line x1="250" y1="356" x2="360" y2="340"/>
            <line x1="140" y1="340" x2="360" y2="340" stroke-opacity="0.4"/>
          </g>
          <g>
            <rect x="118" y="326" width="44" height="28" rx="6" fill="rgba(255,255,255,0.1)" stroke="rgba(255,255,255,0.35)" stroke-width="1.5"/>
            <circle cx="130" cy="340" r="3" fill="#5eead4"/>
            <rect x="138" y="337" width="18" height="6" rx="3" fill="rgba(255,255,255,0.3)"/>
            <rect x="228" y="342" width="44" height="28" rx="6" fill="rgba(255,255,255,0.14)" stroke="rgba(255,255,255,0.45)" stroke-width="1.5"/>
            <circle cx="240" cy="356" r="3" fill="#5eead4"/>
            <rect x="248" y="353" width="18" height="6" rx="3" fill="rgba(255,255,255,0.35)"/>
            <rect x="338" y="326" width="44" height="28" rx="6" fill="rgba(255,255,255,0.1)" stroke="rgba(255,255,255,0.35)" stroke-width="1.5"/>
            <circle cx="350" cy="340" r="3" fill="#fb7185"/>
            <rect x="358" y="337" width="18" height="6" rx="3" fill="rgba(255,255,255,0.3)"/>
          </g>
        </svg>
      </div>
    </div>

    <div class="login-right">
      <div class="login-form-wrap">
        <div class="form-header">
          <h2>{{ $t('login.welcome') }}</h2>
          <p>{{ $t('login.welcomeDesc') }}</p>
        </div>

        <form @submit.prevent="handleLogin" class="login-form">
          <div class="form-group">
            <label>{{ $t('login.username') }}</label>
            <div class="input-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
              </svg>
              <input
                v-model="username"
                type="text"
                :placeholder="$t('login.usernamePlaceholder')"
                autocomplete="username"
                :disabled="loading"
                @keyup.enter="handleLogin"
              />
            </div>
          </div>

          <div class="form-group">
            <label>{{ $t('login.password') }}</label>
            <div class="input-wrap">
              <svg class="input-icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                :placeholder="$t('login.passwordPlaceholder')"
                autocomplete="current-password"
                :disabled="loading"
                @keyup.enter="handleLogin"
              />
              <span class="input-action" @click="showPassword = !showPassword">
                {{ $t(showPassword ? 'login.hide' : 'login.show') }}
              </span>
            </div>
          </div>

          <div class="form-options">
            <label class="remember">
              <input type="checkbox" v-model="rememberMe" />
              <span>{{ $t('login.rememberMe') }}</span>
            </label>
          </div>

          <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

          <button type="submit" class="submit-btn" :disabled="loading">
            {{ loading ? $t('login.loggingIn') : $t('login.submit') }}
          </button>
        </form>

        <div class="lang-row">
          <LangSwitch />
        </div>

        <div class="form-footer">
          <span>{{ $t('login.footer', { year: 2026 }) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authApi } from '../api/auth.js'
import { setAuth } from '../utils/auth.js'
import LangSwitch from '../components/LangSwitch.vue'

const router = useRouter()
const { t } = useI18n({ useScope: 'global' })

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const rememberMe = ref(true)
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  errorMsg.value = ''

  if (!username.value.trim()) {
    errorMsg.value = t('login.usernameRequired')
    return
  }
  if (!password.value.trim()) {
    errorMsg.value = t('login.passwordRequired')
    return
  }

  loading.value = true
  try {
    const result = await authApi.login(username.value, password.value)
    setAuth(result.token, result.user)
    router.push('/chat')
  } catch (err) {
    errorMsg.value = err.message || t('login.failed')
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

.hero-illustration {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
  padding: 32px 0;
}

.hero-illustration svg {
  max-width: 720px;
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

.lang-row {
  display: flex;
  justify-content: center;
  margin-top: 14px;
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