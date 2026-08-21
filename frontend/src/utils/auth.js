// 认证相关工具函数
const TOKEN_KEY = 'aiops_token'
const USER_KEY = 'aiops_user'
const PERMISSIONS_KEY = 'aiops_permissions'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getUser() {
  const user = localStorage.getItem(USER_KEY)
  if (!user) return null
  try {
    return JSON.parse(user)
  } catch {
    localStorage.removeItem(USER_KEY)
    return null
  }
}

export function setUser(user) {
  if (user != null) {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  } else {
    localStorage.removeItem(USER_KEY)
  }
}

export function removeUser() {
  localStorage.removeItem(USER_KEY)
}

export function getPermissions() {
  const p = localStorage.getItem(PERMISSIONS_KEY)
  if (!p) return []
  try {
    return JSON.parse(p)
  } catch {
    localStorage.removeItem(PERMISSIONS_KEY)
    return []
  }
}

export function setPermissions(permissions) {
  if (permissions && Array.isArray(permissions)) {
    localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(permissions))
  } else {
    localStorage.removeItem(PERMISSIONS_KEY)
  }
}

export function removePermissions() {
  localStorage.removeItem(PERMISSIONS_KEY)
}

export function hasPermission(name) {
  const perms = getPermissions()
  return perms.includes(name)
}

export function isLoggedIn() {
  return !!getToken()
}

export function setAuth(token, user) {
  setToken(token)
  setUser(user)
  if (user && user.permissions) {
    setPermissions(user.permissions)
  } else {
    // 即使 permissions 为空或不存在，也要清除旧权限
    setPermissions([])
  }
}

export function clearAuth() {
  removeToken()
  removeUser()
  removePermissions()
}

// 获取 Authorization 请求头
export function getAuthHeader() {
  const token = getToken()
  if (token) {
    return { 'Authorization': 'Bearer ' + token }
  }
  return {}
}
