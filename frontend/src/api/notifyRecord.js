import { createRequest } from './request.js'

const request = createRequest('/api/notify-records')

export const notifyRecordApi = {
  // 通知记录分页列表，返回 { total, list }
  // params: { status, event_type, strategy, keyword, start, end, page, page_size }
  list({ status = '', event_type = '', strategy = '', keyword = '', start = '', end = '', page = 1, page_size = 20 } = {}) {
    const params = new URLSearchParams({ page: String(page), page_size: String(page_size) })
    if (status) params.set('status', status)
    if (event_type) params.set('event_type', event_type)
    if (strategy) params.set('strategy', strategy)
    if (keyword) params.set('keyword', keyword)
    if (start) params.set('start', start)
    if (end) params.set('end', end)
    return request(`?${params.toString()}`)
  },

  // 单条通知记录详情
  get(id) {
    return request(`/${id}`)
  }
}
