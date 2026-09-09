import { createRequest } from './request.js'

const request = createRequest('/api')

export const searchApi = {
  query: (keyword, limit = 5) =>
    request(`/search?q=${encodeURIComponent(keyword)}&limit=${limit}`)
}
