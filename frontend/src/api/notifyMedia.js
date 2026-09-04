import { createRequest } from './request.js'

const request = createRequest('/api/notify-media')

export const notifyMediaApi = {
  list: () => request(''),
  get: (id) => request(`/${id}`),
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' }),
  toggle: (id) => request(`/${id}/toggle`, { method: 'PATCH' }),
  types: () => request('/types'),
  // 测试失败时后端返回 code=500，createRequest 会抛出带 message 的错误
  // payload 可传 { content } 自定义测试内容
  test: (id, payload) => request(`/${id}/test`, { method: 'POST', body: payload })
}
