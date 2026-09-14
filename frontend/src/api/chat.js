import { getAuthHeader, getToken } from '../utils/auth.js'

const BASE = '/api/chat'

async function request(url, options = {}) {
  const resp = await fetch(BASE + url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader(),
      ...(options.headers || {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined
  })

  if (resp.status === 401) {
    localStorage.removeItem('aiops_token')
    localStorage.removeItem('aiops_user')
    window.location.hash = '#/login'
    throw new Error('登录已过期，请重新登录')
  }

  const data = await resp.json()

  if (data.code !== 0) {
    throw new Error(data.message || '请求失败')
  }

  return data.data
}

export function getSessions() {
  return request('/sessions')
}

export function createSession(title = '') {
  return request('/sessions', { method: 'POST', body: { title } })
}

export function deleteSession(id) {
  return request(`/sessions/${id}`, { method: 'DELETE' })
}

export function getMessages(id) {
  return request(`/sessions/${id}/messages`)
}

export function streamChat(sessionId, content, { panel = true, onChunk, onDone, onError, onPanelStart, onSpecialist } = {}) {
  const encoded = encodeURIComponent(content)
  const token = getToken()
  let url = `${BASE}/sessions/${sessionId}/stream?content=${encoded}&panel=${panel ? 1 : 0}`
  if (token) {
    url += `&token=${encodeURIComponent('Bearer ' + token)}`
  }
  const eventSource = new EventSource(url)
  let finished = false

  eventSource.addEventListener('message', (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.chunk) onChunk(data.chunk)
    } catch (e) {
      onError('解析流数据失败')
    }
  })

  // 专家会诊：会诊开始（专家名单）
  eventSource.addEventListener('panel_start', (event) => {
    try {
      const data = JSON.parse(event.data)
      if (onPanelStart) onPanelStart(data)
    } catch (e) {}
  })

  // 专家会诊：单个专家状态更新（running/completed/failed + conclusion）
  eventSource.addEventListener('specialist', (event) => {
    try {
      const data = JSON.parse(event.data)
      if (onSpecialist) onSpecialist(data)
    } catch (e) {}
  })

  eventSource.addEventListener('done', (event) => {
    finished = true
    try {
      const data = JSON.parse(event.data)
      eventSource.close()
      onDone(data)
    } catch (e) {
      eventSource.close()
      onError('解析完成事件失败')
    }
  })

  eventSource.addEventListener('error', (event) => {
    if (finished) return
    let message = '流式连接异常'
    try {
      const data = JSON.parse(event.data)
      message = data.message || message
    } catch (e) {}
    finished = true
    eventSource.close()
    onError(message)
  })

  eventSource.onerror = () => {
    if (finished) return
    finished = true
    eventSource.close()
    onError('SSE 连接中断')
  }

  return eventSource
}
