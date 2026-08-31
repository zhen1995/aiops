import { createRequest } from './request.js'

const request = createRequest('/api/alert-events')

export const alertEventApi = {
  list({ scope = 'active', hours = 24, page = 1, limit = 20, query = '', severity = '' } = {}) {
    const params = new URLSearchParams({ scope, hours: String(hours), page: String(page), limit: String(limit) })
    if (query) params.set('query', query)
    if (severity) params.set('severity', String(severity))
    return request(`?${params.toString()}`)
  }
}
