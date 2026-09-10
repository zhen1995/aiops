import { createI18n } from 'vue-i18n'

import zhCommon from './locales/zh/common'
import zhLayout from './locales/zh/layout'
import zhLogin from './locales/zh/login'
import zhDashboard from './locales/zh/dashboard'
import zhChat from './locales/zh/chat'
import zhService from './locales/zh/service'
import zhAlert from './locales/zh/alert'
import zhNotify from './locales/zh/notify'
import zhRca from './locales/zh/rca'
import zhLog from './locales/zh/log'
import zhLlm from './locales/zh/llm'
import zhKnowledge from './locales/zh/knowledge'
import zhInspection from './locales/zh/inspection'
import zhOrg from './locales/zh/org'
import zhDatasource from './locales/zh/datasource'
import zhN9e from './locales/zh/n9e'

import enCommon from './locales/en/common'
import enLayout from './locales/en/layout'
import enLogin from './locales/en/login'
import enDashboard from './locales/en/dashboard'
import enChat from './locales/en/chat'
import enService from './locales/en/service'
import enAlert from './locales/en/alert'
import enNotify from './locales/en/notify'
import enRca from './locales/en/rca'
import enLog from './locales/en/log'
import enLlm from './locales/en/llm'
import enKnowledge from './locales/en/knowledge'
import enInspection from './locales/en/inspection'
import enOrg from './locales/en/org'
import enDatasource from './locales/en/datasource'
import enN9e from './locales/en/n9e'

export const SUPPORTED_LOCALES = [
  { value: 'zh', label: '简体中文' },
  { value: 'en', label: 'English' }
]

const STORAGE_KEY = 'aiops_locale'

export function getLocale() {
  return localStorage.getItem(STORAGE_KEY) || 'zh'
}

const i18n = createI18n({
  legacy: false,
  locale: getLocale(),
  fallbackLocale: 'zh',
  globalInjection: true,
  messages: {
    zh: {
      common: zhCommon,
      layout: zhLayout,
      login: zhLogin,
      dashboard: zhDashboard,
      chat: zhChat,
      service: zhService,
      alert: zhAlert,
      notify: zhNotify,
      rca: zhRca,
      log: zhLog,
      llm: zhLlm,
      knowledge: zhKnowledge,
      inspection: zhInspection,
      org: zhOrg,
      datasource: zhDatasource,
      n9e: zhN9e
    },
    en: {
      common: enCommon,
      layout: enLayout,
      login: enLogin,
      dashboard: enDashboard,
      chat: enChat,
      service: enService,
      alert: enAlert,
      notify: enNotify,
      rca: enRca,
      log: enLog,
      llm: enLlm,
      knowledge: enKnowledge,
      inspection: enInspection,
      org: enOrg,
      datasource: enDatasource,
      n9e: enN9e
    }
  }
})

// 切换系统语言：更新 i18n、写入 localStorage、同步 <html lang>
export function setLocale(locale) {
  i18n.global.locale.value = locale
  localStorage.setItem(STORAGE_KEY, locale)
  document.documentElement.setAttribute('lang', locale === 'zh' ? 'zh-CN' : 'en')
}

export default i18n
