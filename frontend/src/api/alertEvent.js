import { createRequest } from './request.js'

const request = createRequest('/api/alert-events')

export const alertEventApi = {
  list({ scope = 'active', hours = '', page = 1, limit = 20, query = '', severity = '' } = {}) {
    const params = new URLSearchParams({ scope, page: String(page), limit: String(limit) })
    // 活跃告警不传 hours（后端返回全部未恢复事件）；历史告警默认 24h
    if (hours !== '' && hours !== undefined && hours !== null) params.set('hours', String(hours))
    if (query) params.set('query', query)
    if (severity) params.set('severity', String(severity))
    return request(`?${params.toString()}`)
  },
  remove(id) {
    return request(`/${id}`, { method: 'DELETE' })
  },
  removeBatch(ids) {
    return request('/batch', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids })
    })
  }
}
