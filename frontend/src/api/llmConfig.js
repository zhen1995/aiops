// 通用请求封装
const BASE_URL = '/api/llm-config'

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

// LLM 配置 API
export const llmConfigApi = {
  // 获取列表
  list() {
    return request('')
  },

  // 获取单个
  get(id) {
    return request(`/${id}`)
  },

  // 新建
  create(payload) {
    return request('', { method: 'POST', body: payload })
  },

  // 更新
  update(id, payload) {
    return request(`/${id}`, { method: 'PUT', body: payload })
  },

  // 删除
  remove(id) {
    return request(`/${id}`, { method: 'DELETE' })
  },

  // 切换启用状态
  toggle(id) {
    return request(`/${id}/toggle`, { method: 'PATCH' })
  },

  // 获取供应商类型列表
  suppliers() {
    return request('/suppliers')
  }
}