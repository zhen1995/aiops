import { createRequest } from './request.js'

const request = createRequest('/api/root-cause-analyses')
const apiRequest = createRequest('/api')

export const rcaApi = {
  // 触发（或重新触发）某告警事件的根因分析，返回 {id, status}
  trigger(eventId) {
    return apiRequest(`/alert-events/${eventId}/root-cause`, { method: 'POST' })
  },
  // 查询分析历史，可按告警事件过滤
  list({ alert_event_id = '', page = 1, limit = 50 } = {}) {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (alert_event_id) params.set('alert_event_id', String(alert_event_id))
    return request(`?${params.toString()}`)
  },
  // 单条分析记录详情
  detail(id) {
    return request(`/${id}`)
  }
}
