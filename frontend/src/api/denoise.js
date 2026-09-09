import { createRequest } from './request.js'

const request = createRequest('/api/denoise')

export const denoiseApi = {
  // 降噪策略列表，返回 { policies: [...] }
  getPolicies() {
    return request('/policies')
  },

  // 启停指定策略，body: { enabled }
  updatePolicy(strategy, enabled) {
    return request(`/policies/${strategy}`, { method: 'PUT', body: { enabled } })
  },

  // 降噪统计（总量 / 压缩率 / 漏斗），返回 { raw_total, effective, compression_rate, suppressed_total, funnel }
  getStats() {
    return request('/stats')
  }
}
