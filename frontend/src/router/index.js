import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn } from '../utils/auth.js'

const routes = [
  { path: '/login', name: 'login', component: () => import('../views/LoginView.vue'), meta: { title: '登录', public: true } },
  { path: '/', redirect: '/chat' },
  { path: '/chat', name: 'chat', component: () => import('../views/ChatView.vue'), meta: { title: 'AI 对话' } },
  { path: '/dashboard', name: 'dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '总览大盘' } },
  { path: '/alerts/console', name: 'alerts-console', component: () => import('../views/AlertsView.vue'), meta: { title: '告警控制台' } },
  { path: '/alerts/denoise', name: 'alerts-denoise', component: () => import('../views/DenoiseView.vue'), meta: { title: '告警降噪' } },
  { path: '/anomaly', name: 'anomaly', component: () => import('../views/AnomalyView.vue'), meta: { title: '异常检测' } },
  { path: '/rca', name: 'rca', component: () => import('../views/RcaView.vue'), meta: { title: '根因分析' } },
  { path: '/logs', name: 'logs', component: () => import('../views/LogAnalysisView.vue'), meta: { title: '日志分析' } },
  { path: '/ai-config', redirect: '/ai-config/llm' },
  { path: '/ai-config/llm', name: 'ai-config-llm', component: () => import('../views/ai-config/LLMManagementView.vue'), meta: { title: 'LLM 管理' } },
  { path: '/ai-config/skill', name: 'ai-config-skill', component: () => import('../views/ai-config/SkillManagementView.vue'), meta: { title: 'Skill 管理' } },
  { path: '/knowledge-base', name: 'knowledge-base', component: () => import('../views/KnowledgeBaseView.vue'), meta: { title: '运维知识库' } },
  { path: '/inspection', name: 'inspection', component: () => import('../views/InspectionReportView.vue'), meta: { title: '巡检报告管理' } },
  { path: '/inspection/:id', name: 'inspection-detail', component: () => import('../views/InspectionReportDetailView.vue'), meta: { title: '巡检报告详情' } },
  { path: '/notification', redirect: '/notification/policy' },
  { path: '/notification/policy', name: 'notification-policy', component: () => import('../views/notification/PolicyView.vue'), meta: { title: '通知策略' } },
  { path: '/notification/medium', name: 'notification-medium', component: () => import('../views/notification/MediumView.vue'), meta: { title: '通知媒介' } },
  { path: '/org', redirect: '/org/user' },
  { path: '/org/user', name: 'org-user', component: () => import('../views/org/UserManagementView.vue'), meta: { title: '用户管理' } },
  { path: '/org/role', name: 'org-role', component: () => import('../views/org/RoleManagementView.vue'), meta: { title: '角色管理' } },
  { path: '/datasource', name: 'datasource', component: () => import('../views/DatasourceView.vue'), meta: { title: '数据源接入' } },
  { path: '/:pathMatch(.*)*', redirect: '/chat' }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 全局路由守卫
router.beforeEach((to, from, next) => {
  const loggedIn = isLoggedIn()

  // 已登录访问登录页，跳转到首页
  if (to.path === '/login' && loggedIn) {
    next({ path: '/chat' })
    return
  }

  // 未登录访问非公开页面，跳转到登录页
  if (!to.meta.public && !loggedIn) {
    next({ path: '/login' })
    return
  }

  next()
})

export default router