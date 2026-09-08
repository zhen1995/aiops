import { getAuthHeader } from '../utils/auth.js'

const BASE = '/api/n9e'

async function request(url, options = {}) {
  const resp = await fetch(BASE + url, {
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

export const n9eApi = {
  // 引擎配置
  listConfigs() { return request('/configs') },
  saveConfig(cfg) { return request('/configs', { method: 'POST', body: cfg }) },
  deleteConfig(id) { return request(`/configs/${id}`, { method: 'DELETE' }) },
  testConfig(cfg) { return request('/configs/test', { method: 'POST', body: cfg }) },

  // 代理到夜莺
  getBusiGroups() { return request('/busi-groups') },
  getAlertRules() { return request('/alert-rules') },
  getCurEvents(params = {}) {
    const qs = new URLSearchParams()
    if (params.p) qs.set('p', params.p)
    if (params.limit) qs.set('limit', params.limit)
    if (params.my_groups !== undefined) qs.set('my_groups', params.my_groups)
    const q = qs.toString()
    return request(`/alert-cur-events${q ? '?' + q : ''}`)
  },
  getHisEvents(params = {}) {
    const qs = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') qs.set(k, v)
    })
    const q = qs.toString()
    return request(`/alert-his-events${q ? '?' + q : ''}`)
  }
}
