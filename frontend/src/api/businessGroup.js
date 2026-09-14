import { createRequest } from './request.js'

const request = createRequest('/api/business-groups')

export const businessGroupApi = {
  list: () => request(''),
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' })
}
