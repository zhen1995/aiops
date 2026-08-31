import { createRequest } from './request.js'

const request = createRequest('/api/alert-engines')

export const alertEngineApi = {
  list: () => request(''),
  get: (id) => request(`/${id}`),
  create: (payload) => request('', { method: 'POST', body: payload }),
  update: (id, payload) => request(`/${id}`, { method: 'PUT', body: payload }),
  remove: (id) => request(`/${id}`, { method: 'DELETE' }),
  toggle: (id) => request(`/${id}/toggle`, { method: 'PATCH' }),
  test: (id) => request(`/${id}/test`)
}
