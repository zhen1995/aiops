import { getAuthHeader } from '../utils/auth.js'

const BASE_URL = '/api/auth'

async function request(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
    ...(options.headers || {})
  }

  const resp = await fetch(BASE_URL + url, {
    ...options,
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  })

  // 401 未授权
  if (resp.status === 401) {
    localStorage.removeItem('aiops_token')
    localStorage.removeItem('aiops_user')
    window.location.hash = '#/login'
    throw new Error('登录已过期，请重新登录')
  }

  const data = await resp.json()

  if (data.code !== 0) {
    throw new Error(data.message || '请求失败')
  }

  return data.data
}

export const authApi = {
  login(username, password) {
    return request('/login', {
      method: 'POST',
      body: { username, password }
    })
  },

  logout() {
    return request('/logout', {
      method: 'POST'
    })
  },

  info() {
    return request('/info')
  }
}