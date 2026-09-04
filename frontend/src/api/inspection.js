import { createRequest } from './request.js'

const request = createRequest('/api/inspection')

// ---------- 巡检任务 ----------

export function listTasks() {
  return request('/tasks')
}

export function getTask(id) {
  return request(`/tasks/${id}`)
}

export function createTask(data) {
  return request('/tasks', { method: 'POST', body: data })
}

export function updateTask(id, data) {
  return request(`/tasks/${id}`, { method: 'PUT', body: data })
}

export function deleteTask(id) {
  return request(`/tasks/${id}`, { method: 'DELETE' })
}

export function toggleTask(id) {
  return request(`/tasks/${id}/toggle`, { method: 'PATCH' })
}

export function triggerTask(id) {
  return request(`/tasks/${id}/trigger`, { method: 'POST' })
}

export function previewCron(cronExpr, count = 5) {
  return request('/cron-preview', {
    method: 'POST',
    body: { cron_expr: cronExpr, count }
  })
}

// ---------- 巡检报告 ----------

export function listReports(params = {}) {
  const qs = new URLSearchParams()
  if (params.task_id) qs.set('task_id', params.task_id)
  if (params.status) qs.set('status', params.status)
  const q = qs.toString()
  return request('/reports' + (q ? `?${q}` : ''))
}

export function getReport(id) {
  return request(`/reports/${id}`)
}

export function deleteReport(id) {
  return request(`/reports/${id}`, { method: 'DELETE' })
}
