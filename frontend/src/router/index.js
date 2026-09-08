import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn, getPermissions } from '../utils/auth.js'

// 每个路由的 meta.auth 对应 sys_auth 表中的 name
const routes = [
  { path: '/login', name: 'login', component: () => import('../views/LoginView.vue'), meta: { title: '登录', public: true } },
  { path: '/no-permission', name: 'no-permission', component: () => import('../views/NoPermissionView.vue'), meta: { title: '无权限', public: true } },
  { path: '/', redirect: '/chat' },
  { path: '/chat', name: 'chat', component: () => import('../views/ChatView.vue'), meta: { title: 'AI 对话', auth: '对话' } },
  { path: '/dashboard', name: 'dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '总览大盘', auth: '总览大盘' } },
  { path: '/alerts/denoise', name: 'alerts-denoise', component: () => import('../views/DenoiseView.vue'), meta: { title: '告警降噪', auth: '告警降噪' } },
  { path: '/alerts/events', name: 'alert-events', component: () => import('../views/AnomalyView.vue'), meta: { title: '告警事件', auth: '告警事件' } },
  { path: '/alert-rules', name: 'alert-rules', component: () => import('../views/AlertRuleView.vue'), meta: { title: '告警规则', auth: '告警规则' } },
  // 通知（消息模板 / 通知规则）
  { path: '/notify/templates', name: 'notify-templates', component: () => import('../views/notify/NotifyTemplateView.vue'), meta: { title: '消息模板', auth: '消息模板' } },
  { path: '/notify/rules', name: 'notify-rules', component: () => import('../views/notify/NotifyRuleView.vue'), meta: { title: '通知规则', auth: '通知规则' } },
  { path: '/notify', redirect: '/notify/templates' },
  // 夜莺
  { path: '/n9e/config', name: 'n9e-config', component: () => import('../views/n9e/EngineConfigView.vue'), meta: { title: '夜莺引擎配置', auth: '夜莺引擎配置' } },
  { path: '/n9e/alert-rules', name: 'n9e-alert-rules', component: () => import('../views/n9e/AlertRulesView.vue'), meta: { title: '夜莺告警规则', auth: '夜莺告警规则' } },
  { path: '/n9e/alert-events', name: 'n9e-alert-events', component: () => import('../views/n9e/AlertEventsView.vue'), meta: { title: '夜莺告警事件', auth: '夜莺告警事件' } },
  { path: '/n9e', redirect: '/n9e/config' },
  { path: '/rca', name: 'rca', component: () => import('../views/RcaView.vue'), meta: { title: '根因分析', auth: '根因分析' } },
  { path: '/logs', name: 'logs', component: () => import('../views/LogAnalysisView.vue'), meta: { title: '日志分析', auth: '日志分析' } },
  { path: '/ai-config', redirect: '/ai-config/llm' },
  { path: '/ai-config/llm', name: 'ai-config-llm', component: () => import('../views/ai-config/LLMManagementView.vue'), meta: { title: 'LLM 管理', auth: 'LLM 管理' } },
  { path: '/ai-config/skill', name: 'ai-config-skill', component: () => import('../views/ai-config/SkillManagementView.vue'), meta: { title: 'Skill 管理', auth: 'Skill 管理' } },
  { path: '/knowledge-base', name: 'knowledge-base', component: () => import('../views/KnowledgeBaseView.vue'), meta: { title: '运维知识库', auth: '运维知识库' } },
  { path: '/inspection/tasks', name: 'inspection-tasks', component: () => import('../views/InspectionTaskView.vue'), meta: { title: '巡检任务', auth: '巡检任务' } },
  { path: '/inspection/reports', name: 'inspection-reports', component: () => import('../views/InspectionReportView.vue'), meta: { title: '巡检报告', auth: '巡检报告' } },
  { path: '/inspection/reports/:id', name: 'inspection-detail', component: () => import('../views/InspectionReportDetailView.vue'), meta: { title: '巡检报告详情', auth: '巡检报告' } },
  { path: '/inspection', redirect: '/inspection/tasks' },
  { path: '/notification', redirect: '/notification/medium' },
  { path: '/notification/medium', name: 'notification-medium', component: () => import('../views/notification/MediumView.vue'), meta: { title: '通知媒介', auth: '通知媒介' } },
  { path: '/org', redirect: '/org/user' },
  { path: '/org/user', name: 'org-user', component: () => import('../views/org/UserManagementView.vue'), meta: { title: '用户管理', auth: '用户管理' } },
  { path: '/org/role', name: 'org-role', component: () => import('../views/org/RoleManagementView.vue'), meta: { title: '角色管理', auth: '角色管理' } },
  { path: '/datasource', name: 'datasource', component: () => import('../views/DatasourceView.vue'), meta: { title: '数据源接入', auth: '数据源接入' } },
  { path: '/:pathMatch(.*)*', redirect: '/chat' }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
  const loggedIn = isLoggedIn()

  // 已登录访问登录页，跳转到无权限页或首页
  if (to.path === '/login' && loggedIn) {
    const perms = getPermissions()
    next({ path: perms.length === 0 ? '/no-permission' : '/chat' })
    return
  }

  // 未登录访问非公开页面，跳转到登录页
  if (!to.meta.public && !loggedIn) {
    next({ path: '/login' })
    return
  }

  // 已登录但无权限访问该页面
  if (loggedIn && to.meta.auth) {
    const perms = getPermissions()
    // 无权限时跳转到无权限提示页
    if (perms.length === 0) {
      next({ path: '/no-permission' })
      return
    }
    // 有权限但不包含当前路由所需权限，跳转到首页
    if (!perms.includes(to.meta.auth)) {
      next({ path: '/chat' })
      return
    }
  }

  next()
})

export default router
