import { createRequest } from './request.js'

const request = createRequest('/api/dashboard')

export const dashboardApi = {
  // 近 N 小时：hours；自定义范围：start/end（ISO 字符串）
  overview({ hours = 24, start = '', end = '' } = {}) {
    const params = new URLSearchParams()
    if (start && end) {
      params.set('start', start)
      params.set('end', end)
    } else {
      params.set('hours', String(hours))
    }
    return request(`/overview?${params.toString()}`)
  }
}
