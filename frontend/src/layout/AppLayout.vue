<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="logo">
        <span class="logo-mark">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 12h4l3-8 4 16 3-8h4" />
          </svg>
        </span>
        <div class="logo-text">
          <b>AIOPS</b>
          <span>智能运维平台</span>
        </div>
      </div>

      <nav class="menu">
        <template v-for="m in menus" :key="m.path || m.title">
          <RouterLink
            v-if="!m.children"
            :to="m.path"
            class="menu-item"
            :class="{ active: isActive(m.path) }"
          >
            <span class="menu-icon" v-html="m.icon"></span>
            <span>{{ m.title }}</span>
            <em v-if="m.badge" class="menu-badge">{{ m.badge }}</em>
          </RouterLink>

          <div v-else class="menu-group">
            <button
              class="menu-item menu-group-title"
              :class="{ active: isGroupActive(m), expanded: expanded[m.title] }"
              @click="toggle(m.title)"
            >
              <span class="menu-icon" v-html="m.icon"></span>
              <span>{{ m.title }}</span>
              <svg class="arrow" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="6 9 12 15 18 9" />
              </svg>
            </button>
            <div v-show="expanded[m.title]" class="submenu">
              <RouterLink
                v-for="c in m.children"
                :key="c.path"
                :to="c.path"
                class="submenu-item"
                :class="{ active: isActive(c.path) }"
              >
                <span>{{ c.title }}</span>
                <em v-if="c.badge" class="menu-badge">{{ c.badge }}</em>
              </RouterLink>
            </div>
          </div>
        </template>
      </nav>

      <div class="sidebar-foot">
        <div class="ver muted">智能运维平台</div>
      </div>
    </aside>

    <div class="main">
      <header class="topbar">
        <h1 class="page-title">{{ route.meta.title || 'AIOPS' }}</h1>
        <div class="topbar-right">
          <div class="search">
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" />
            </svg>
            <input placeholder="搜索服务 / 告警 / 事件…" />
          </div>
          <div class="user" @click="toggleUserMenu">
            <span class="avatar">{{ avatarText }}</span>
            <div class="user-info">
              <b>{{ currentUser?.name || currentUser?.username || '用户' }}</b>
              <span>{{ currentUser?.username || '' }}</span>
            </div>
            <div v-if="showUserMenu" class="user-dropdown" @click.stop>
              <div class="user-dropdown-item" @click="handleLogout">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                退出登录
              </div>
            </div>
          </div>
        </div>
      </header>

      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser, clearAuth, getPermissions } from '../utils/auth.js'

const route = useRoute()
const router = useRouter()

const showUserMenu = ref(false)
const currentUser = ref(getUser())
const userPermissions = ref(getPermissions())

const avatarText = computed(() => {
  const name = currentUser.value?.name || currentUser.value?.username || 'U'
  return name.charAt(0).toUpperCase()
})

function handleLogout() {
  showUserMenu.value = false
  if (confirm('确定要退出登录吗？')) {
    clearAuth()
    router.push('/login')
  }
}

function toggleUserMenu() {
  showUserMenu.value = !showUserMenu.value
}

function handleDocClick(e) {
  const target = e.target
  if (target instanceof Element && !target.closest('.user')) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocClick)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocClick)
})

const isActive = (p) => (p === '/' ? route.path === '/' : route.path.startsWith(p))

const isGroupActive = (m) => m.children?.some((c) => isActive(c.path))

const expanded = reactive({
  '告警管理': true,
  '夜莺': false,
  'AI 配置': false,
  '人员组织': false,
  '巡检管理': false
})

const toggle = (title) => {
  expanded[title] = !expanded[title]
}

const ic = (d) =>
  `<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${d}</svg>`

// 菜单定义，auth 字段对应 sys_auth 表中的 name
const allMenus = [
  { path: '/chat', title: '对话', auth: '对话', icon: ic('<path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>') },
  { path: '/dashboard', title: '总览大盘', auth: '总览大盘', icon: ic('<rect x="3" y="3" width="7" height="9" rx="1.5"/><rect x="14" y="3" width="7" height="5" rx="1.5"/><rect x="14" y="12" width="7" height="9" rx="1.5"/><rect x="3" y="16" width="7" height="5" rx="1.5"/>') },
  { path: '/services', title: '服务注册', auth: '服务注册', icon: ic('<rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><path d="M6 6.5h.01M6 17.5h.01"/>') },
  {
    title: '告警管理',
    icon: ic('<path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/>'),
    children: [
      { path: '/alerts/denoise', title: '告警降噪', auth: '告警降噪' },
      { path: '/alerts/events', title: '告警事件', auth: '告警事件' },
      { path: '/alert-rules', title: '告警规则', auth: '告警规则' },
      { path: '/notify/templates', title: '消息模板', auth: '消息模板' },
      { path: '/notify/rules', title: '通知规则', auth: '通知规则' }
    ]
  },
  {
    title: '夜莺',
    icon: ic('<circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>'),
    children: [
      { path: '/n9e/config', title: '引擎配置', auth: '夜莺引擎配置' },
      { path: '/n9e/alert-rules', title: '告警规则', auth: '夜莺告警规则' },
      { path: '/n9e/alert-events', title: '告警事件', auth: '夜莺告警事件' }
    ]
  },
  { path: '/rca', title: '根因分析', auth: '根因分析', icon: ic('<circle cx="12" cy="12" r="3"/><circle cx="12" cy="12" r="8"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2"/>') },
  { path: '/logs', title: '日志分析', auth: '日志分析', icon: ic('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M8 13h8M8 17h5"/>') },
  { path: '/ai-config/llm', title: 'LLM配置', auth: 'LLM 管理', icon: ic('<path d="M12 2a10 10 0 1 0 10 10H12V2z"/><path d="M12 2a10 10 0 0 1 10 10"/><path d="M12 12l7-7"/>') },
  { path: '/knowledge-base', title: '运维知识库', auth: '运维知识库', icon: ic('<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>') },
  {
    title: '巡检管理',
    icon: ic('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/>'),
    children: [
      { path: '/inspection/tasks', title: '巡检任务', auth: '巡检任务' },
      { path: '/inspection/reports', title: '巡检报告', auth: '巡检报告' }
    ]
  },
  { path: '/notification/medium', title: '通知媒介', auth: '通知媒介', icon: ic('<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/><path d="M2 8c0-2.2 1.8-4 4-4h12a4 4 0 0 1 4 4"/>') },
  {
    title: '人员组织',
    icon: ic('<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>'),
    children: [
      { path: '/org/user', title: '用户管理', auth: '用户管理' },
      { path: '/org/role', title: '角色管理', auth: '角色管理' }
    ]
  },
  { path: '/datasource', title: '数据源接入', auth: '数据源接入', icon: ic('<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5"/><path d="M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>') }
]

// 根据用户权限过滤菜单
const menus = computed(() => {
  const perms = userPermissions.value

  // 无权限时不显示任何菜单
  if (!perms || perms.length === 0) {
    return []
  }

  const filterMenu = (menu) => {
    // 有 children 的菜单组：过滤子项，至少有一个子项有权限才显示
    if (menu.children) {
      const visibleChildren = menu.children.filter(c => c.auth && perms.includes(c.auth))
      if (visibleChildren.length === 0) return null
      return { ...menu, children: visibleChildren }
    }
    // 无子项的顶级菜单：检查自身权限
    if (menu.auth && !perms.includes(menu.auth)) return null
    return menu
  }

  return allMenus.map(filterMenu).filter(Boolean)
})
</script>

<style scoped>
.layout { display: flex; height: 100vh; overflow: hidden; }

.sidebar {
  width: var(--sidebar-w);
  flex-shrink: 0;
  background: var(--c-surface);
  border-right: 1px solid var(--c-border);
  display: flex;
  flex-direction: column;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 18px 16px;
  border-bottom: 1px solid var(--c-border);
}

.logo-mark {
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: var(--c-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text { display: flex; flex-direction: column; line-height: 1.25; }
.logo-text b { font-size: 16px; letter-spacing: 0.5px; }
.logo-text span { font-size: 11px; color: var(--c-text-3); }

.menu { flex: 1; padding: 12px 10px; overflow-y: auto; }

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  color: var(--c-text-2);
  font-size: 13.5px;
  margin-bottom: 2px;
  position: relative;
  transition: all 0.15s ease;
  background: transparent;
  border: none;
  width: 100%;
  cursor: pointer;
  font-family: inherit;
}

.menu-item:hover { background: var(--c-primary-soft); color: var(--c-primary); }

.menu-item.active {
  background: var(--c-primary-tint);
  color: var(--c-primary);
  font-weight: 600;
}

.menu-item.active::before {
  content: '';
  position: absolute;
  left: -10px;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--c-primary);
}

.menu-icon { display: flex; }

.menu-badge {
  margin-left: auto;
  font-style: normal;
  font-size: 11px;
  background: var(--c-p0);
  color: #fff;
  border-radius: 9px;
  padding: 0 7px;
  line-height: 17px;
}

.menu-group-title .arrow {
  margin-left: auto;
  transition: transform 0.2s ease;
  color: var(--c-text-3);
}

.menu-group-title.expanded .arrow { transform: rotate(180deg); }

.submenu {
  padding-left: 16px;
  margin-bottom: 4px;
}

.submenu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px 8px 22px;
  border-radius: 8px;
  color: var(--c-text-2);
  font-size: 13px;
  position: relative;
  transition: all 0.15s ease;
}

.submenu-item::before {
  content: '';
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--c-text-3);
}

.submenu-item:hover { background: var(--c-primary-soft); color: var(--c-primary); }

.submenu-item.active {
  background: var(--c-primary-tint);
  color: var(--c-primary);
  font-weight: 600;
}

.submenu-item.active::before { background: var(--c-primary); }

.sidebar-foot { padding: 14px; border-top: 1px solid var(--c-border); }

.sys-status {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: var(--c-primary-soft);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 10px 12px;
  margin-bottom: 10px;
}

.sys-status .dot { margin-top: 7px; }
.sys-status b { display: block; font-size: 12.5px; }
.sys-status span { font-size: 11px; color: var(--c-text-3); }

.ver { text-align: center; }

.main { flex: 1; display: flex; flex-direction: column; min-width: 0; }

.topbar {
  height: var(--topbar-h);
  flex-shrink: 0;
  background: var(--c-surface);
  border-bottom: 1px solid var(--c-border);
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 16px;
}

.page-title { font-size: 17px; font-weight: 600; margin: 0; }

.topbar-right { margin-left: auto; display: flex; align-items: center; gap: 14px; }

.search {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 7px 12px;
  color: var(--c-text-3);
  width: 240px;
}

.search input {
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  width: 100%;
  color: var(--c-text);
}

.env-tag {
  font-size: 12px;
  color: var(--c-primary);
  background: var(--c-primary-tint);
  border: 1px solid #cfe6e2;
  padding: 3px 10px;
  border-radius: 6px;
  font-family: monospace;
}

.icon-btn {
  position: relative;
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  border-radius: 8px;
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--c-text-2);
  cursor: pointer;
}

.icon-btn:hover { color: var(--c-primary); border-color: var(--c-primary); }

.bell-dot {
  position: absolute;
  top: 7px;
  right: 8px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--c-p0);
  border: 1.5px solid #fff;
}

.user { display: flex; align-items: center; gap: 9px; }

.avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: var(--c-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.user-info { display: flex; flex-direction: column; line-height: 1.3; cursor: pointer; }
.user-info b { font-size: 13px; }
.user-info span { font-size: 11px; color: var(--c-text-3); }

.user {
  position: relative;
  display: flex;
  align-items: center;
  gap: 9px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.15s;
}
.user:hover { background: var(--c-primary-soft); }

.user-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  min-width: 140px;
  padding: 6px;
  z-index: 100;
}

.user-dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  font-size: 13px;
  color: var(--c-text-2);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}

.user-dropdown-item:hover {
  background: var(--c-primary-soft);
  color: var(--c-primary);
}

.content { flex: 1; overflow-y: auto; padding: 22px 24px 32px; }
</style>
