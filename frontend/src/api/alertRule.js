import { getAuthHeader } from '../utils/auth.js'

const BASE_URL = '/api/alert-rules'

async function request(url, options = {}) {
  const resp = await fetch(BASE_URL + url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader(),
      ...(options.headers || {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined
  })
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

export const alertRuleApi = {
  list(gids = '') {
    const qs = gids ? `?gids=${encodeURIComponent(gids)}` : ''
    return request(qs)
  }
}
