import { createRequest } from './request.js'

const request = createRequest('/api/alert-rules')

export const alertRuleApi = {
  list: () => request(''),
  get: (id) => request(`/${id}`),
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' }),
  toggle: (id) => request(`/${id}/toggle`, { method: 'PATCH' })
}
