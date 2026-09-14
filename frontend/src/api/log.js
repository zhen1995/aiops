import { createRequest } from './request.js'

const request = createRequest('/api/logs')

function buildQuery(params = {}) {
  const q = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      q.set(key, String(value))
    }
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export const logApi = {
  // 日志量趋势：{ service, hours }
  getTrend(params = {}) {
    return request(`/trend${buildQuery(params)}`)
  },
  // 日志聚类：{ service, trend, level, hours, page, page_size }
  getClusters(params = {}) {
    return request(`/clusters${buildQuery(params)}`)
  },
  // Drain 模板提取：{ service, limit }
  getTemplates(params = {}) {
    return request(`/templates${buildQuery(params)}`)
  }
}
