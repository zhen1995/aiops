import { createRequest } from './request.js'

const request = createRequest('/api/alert-rules')

export const alertRuleApi = {
  // 分页查询：返回 { list, total }；query 匹配规则名称/PromQL，group_id 按业务分组过滤
  list: ({ page = 1, limit = 20, query = '', groupId = '' } = {}) => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) })
    if (query) params.set('query', query)
    if (groupId) params.set('group_id', groupId)
    return request(`?${params.toString()}`)
  },
  get: (id) => request(`/${id}`),
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' }),
  toggle: (id) => request(`/${id}/toggle`, { method: 'PATCH' })
}
