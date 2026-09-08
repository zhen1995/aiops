import { getAuthHeader } from '../utils/auth.js'

async function request(url, options = {}) {
  const resp = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader(),
      ...(options.headers || {})
    },
    body: options.body !== undefined ? (typeof options.body === 'string' ? options.body : JSON.stringify(options.body)) : undefined
  })
  if (resp.status === 401) {
    localStorage.removeItem('aiops_token')
    localStorage.removeItem('aiops_user')
    window.location.hash = '#/login'
    throw new Error('登录已过期')
  }
  const data = await resp.json()
  if (data.code !== 0) throw new Error(data.message || '请求失败')
  return data.data
}

export const notifyApi = {
  // 通知媒介（CRUD 已由 notify_media 路由暴露，这里只是便捷封装）
  listMedia() { return request('/api/notify-media') },

  // 消息模板
  listTemplates() { return request('/api/notify-templates') },
  createTemplate(tpl) { return request('/api/notify-templates', { method: 'POST', body: tpl }) },
  updateTemplate(id, tpl) { return request(`/api/notify-templates/${id}`, { method: 'PUT', body: tpl }) },
  deleteTemplate(id) { return request(`/api/notify-templates/${id}`, { method: 'DELETE' }) },
  validateTemplate(content) { return request('/api/notify-templates/validate', { method: 'POST', body: { content } }) },
  previewTemplate(content, evt) { return request('/api/notify-templates/preview', { method: 'POST', body: { content, event: evt } }) },
  sampleEvent() { return request('/api/notify-templates/sample-event') },

  // 通知规则
  listRules() { return request('/api/notify-rules') },
  createRule(rule) { return request('/api/notify-rules', { method: 'POST', body: rule }) },
  updateRule(id, rule) { return request(`/api/notify-rules/${id}`, { method: 'PUT', body: rule }) },
  deleteRule(id) { return request(`/api/notify-rules/${id}`, { method: 'DELETE' }) },
  toggleRule(id) { return request(`/api/notify-rules/${id}/toggle`, { method: 'PATCH' }) }
}
