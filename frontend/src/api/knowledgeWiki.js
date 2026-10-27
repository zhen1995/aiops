import { createRequest } from './request.js'

const request = createRequest('/api/knowledge-wiki')

export const wikiApi = {
  // 分类
  categories: () => request('/categories'),
  createCategory: (payload) => request('/categories', { method: 'POST', body: payload }),
  updateCategory: (id, payload) => request(`/categories/${id}`, { method: 'PUT', body: payload }),
  removeCategory: (id) => request(`/categories/${id}`, { method: 'DELETE' }),

  // 词条
  entries: (params = {}) => {
    const sp = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) {
      if (v !== '' && v !== null && v !== undefined) sp.set(k, String(v))
    }
    const qs = sp.toString()
    return request(qs ? `/entries?${qs}` : '/entries')
  },
  createEntry: (payload, publish = false) => request(publish ? '/entries?publish=1' : '/entries', { method: 'POST', body: payload }),
  getEntry: (id) => request(`/entries/${id}`),
  updateEntry: (id, payload) => request(`/entries/${id}`, { method: 'PUT', body: payload }),
  removeEntry: (id) => request(`/entries/${id}`, { method: 'DELETE' }),
  publishEntry: (id) => request(`/entries/${id}/publish`, { method: 'POST', body: {} }),
  withdrawEntry: (id) => request(`/entries/${id}/withdraw`, { method: 'POST', body: {} }),
  archiveEntry: (id) => request(`/entries/${id}/archive`, { method: 'POST', body: {} }),
  generateFromDocument: (documentId) => request('/entries/generate-from-document', { method: 'POST', body: { document_id: documentId } }),

  // 混合检索
  search: (q, limit = 5) => request(`/search?q=${encodeURIComponent(q)}&limit=${limit}`)
}
