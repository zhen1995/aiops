import { createRequest } from './request.js'

const request = createRequest('/api/alert-rules')

export const alertRuleApi = {
  list(gids = '') {
    const qs = gids ? `?gids=${encodeURIComponent(gids)}` : ''
    return request(qs)
  }
}
