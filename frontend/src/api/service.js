import { createRequest } from './request.js'

const request = createRequest('/api/services')

export const serviceApi = {
  // 分页查询服务列表，支持关键字与状态过滤
  list({ keyword = '', status = '', page = 1, page_size = 20 } = {}) {
    const params = new URLSearchParams({ page: String(page), page_size: String(page_size) })
    if (keyword) params.set('keyword', keyword)
    if (status !== '') params.set('status', String(status))
    return request(`?${params.toString()}`)
  },

  get(id) {
    return request(`/${id}`)
  },

  create(data) {
    return request('', { method: 'POST', body: data })
  },

  update(id, data) {
    return request(`/${id}`, { method: 'PUT', body: data })
  },

  remove(id) {
    return request(`/${id}`, { method: 'DELETE' })
  },

  toggle(id) {
    return request(`/${id}/toggle`, { method: 'PATCH' })
  },

  // 验证服务的 ES / Prometheus 连通性，返回 { es_ok, es_message, prom_ok, prom_message, verified }
  verify(id) {
    return request(`/${id}/verify`, { method: 'POST' })
  },

  // 按标签选择器预览匹配的 Prometheus job，返回 { jobs: [...], message }
  previewJobs(data) {
    return request('/preview-jobs', { method: 'POST', body: data })
  },

  // 下拉选项（供其他模块引用服务）
  options() {
    return request('/options')
  }
}
