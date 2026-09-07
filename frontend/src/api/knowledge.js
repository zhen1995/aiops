import { createRequest } from './request.js'
import { getAuthHeader } from '../utils/auth.js'

const request = createRequest('/api/knowledge-base')

// multipart 上传：createRequest 会将 body JSON 序列化，不适合 FormData，这里单独实现
async function uploadRequest(files) {
  const form = new FormData()
  Array.from(files).forEach((f) => form.append('files', f))
  const resp = await fetch('/api/knowledge-base', {
    method: 'POST',
    headers: getAuthHeader(),
    body: form
  })
  if (resp.status === 401) {
    localStorage.removeItem('aiops_token')
    localStorage.removeItem('aiops_user')
    window.location.hash = '#/login'
    throw new Error('登录已过期，请重新登录')
  }
  const data = await resp.json()
  if (data.code !== 0) {
    throw new Error(data.message || '上传失败')
  }
  return data.data
}

export const knowledgeApi = {
  list: () => request(''),

  upload: (files) => uploadRequest(files),

  remove: (id) => request(`/${id}`, { method: 'DELETE' }),

  reindex: (id) => request(`/${id}/reindex`, { method: 'POST' }),

  retrieve: (query, topK = 5) => request('/retrieve', { method: 'POST', body: { query, top_k: topK } })
}
