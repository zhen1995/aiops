import { createRequest } from './request.js'

const request = createRequest('/api/shield-rules')

export const shieldRuleApi = {
  list: (params = {}) => {
    const sp = new URLSearchParams()
    if (params.q) sp.set('q', params.q)
    if (params.includeExpired) sp.set('include_expired', '1')
    const qs = sp.toString()
    return request(qs ? `?${qs}` : '')
  },
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  toggle: (id) => request(`/${id}/toggle`, { method: 'PATCH' }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' })
}
