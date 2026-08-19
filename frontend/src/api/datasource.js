// 数据源 API
const BASE_URL = '/api/datasources'

async function request(url, options = {}) {
  const resp = await fetch(BASE_URL + url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
    body: options.body ? JSON.stringify(options.body) : undefined
  })

  const data = await resp.json()

  if (data.code !== 0) {
    throw new Error(data.message || '请求失败')
  }

  return data.data
}

export const datasourceApi = {
  list() {
    return request('')
  },

  get(id) {
    return request(`/${id}`)
  },

  create(payload) {
    return request('', { method: 'POST', body: payload })
  },

  update(id, payload) {
    return request(`/${id}`, { method: 'PUT', body: payload })
  },

  remove(id) {
    return request(`/${id}`, { method: 'DELETE' })
  },

  toggle(id) {
    return request(`/${id}/toggle`, { method: 'PATCH' })
  },

  types() {
    return request('/types')
  }
}