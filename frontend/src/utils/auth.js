// 认证相关工具函数
const TOKEN_KEY = 'aiops_token'
const USER_KEY = 'aiops_user'

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

export function isLoggedIn() {
  return !!getToken()
}

export function setAuth(token, user) {
  setToken(token)
  setUser(user)
}

export function clearAuth() {
  removeToken()
  removeUser()
}

// 获取 Authorization 请求头
export function getAuthHeader() {
  const token = getToken()
  if (token) {
    return { 'Authorization': 'Bearer ' + token }
  }
  return {}
}