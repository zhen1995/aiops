import { createRequest } from './request.js'

const request = createRequest('/api/system-config')

export const systemConfigApi = {
  // 获取系统配置，返回 { qdrant_url, frontend_base_url }
  getSystemConfig() {
    return request('')
  },

  // 更新系统配置，payload: { qdrant_url?, frontend_base_url? }（只更新显式传入的字段）
  updateSystemConfig(payload) {
    return request('', { method: 'PUT', body: payload })
  },

  // 测试 Qdrant 连通性，payload: { qdrant_url }，返回 { ok, message }
  testQdrant(payload) {
    return request('/qdrant/test', { method: 'POST', body: payload })
  }
}
