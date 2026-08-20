import { getAuthHeader } from '../utils/auth.js'

const BASE_URL = '/api/users'

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

export const userApi = {
  list(keyword) {
    const url = keyword ? `?keyword=${encodeURIComponent(keyword)}` : ''
    return request(url)
  },

  get(id) {
    return request(`/${id}`)
  },

  create(payload) {
    return request('', { method: 'POST', body: payload })
  },

  update(id, payload) {
    return request(`/${id}`, { method: 'PUT', body: payload })
  },

  remove(id) {
    return request(`/${id}`, { method: 'DELETE' })
  },

  resetPassword(id, password) {
    return request(`/${id}/reset-password`, {
      method: 'PATCH',
      body: { password }
    })
  }
}
