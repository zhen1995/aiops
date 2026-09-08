// AIOPS 前端原型 Mock 数据 —— 全部静态数据，无任何后端请求

// 确定性伪随机序列生成器（保证每次刷新图形一致）
export function seededSeries(seed, n, base, amp, noise = 0.15) {
  let s = seed
  const rnd = () => {
    s = (s * 9301 + 49297) % 233280
    return s / 233280
  }
  const out = []
  for (let i = 0; i < n; i++) {
    const wave = Math.sin((i / n) * Math.PI * 2) * amp * 0.6
    const jitter = (rnd() - 0.5) * amp * noise * 2
    out.push(Math.round((base + wave + jitter) * 10) / 10)
  }
  return out
}

export function timeLabels(n, stepMin = 5) {
  const out = []
  const now = new Date('2026-08-12T14:00:00')
  for (let i = n - 1; i >= 0; i--) {
    const t = new Date(now.getTime() - i * stepMin * 60000)
    out.push(`${String(t.getHours()).padStart(2, '0')}:${String(t.getMinutes()).padStart(2, '0')}`)
  }
  return out
}

export const fmtTime = (iso) => {
  const d = new Date(iso)
  return `${d.getMonth() + 1}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

// ---------------- 总览大盘 ----------------
export const kpiStats = [
  { label: '活动告警', value: 12, delta: '-64%', deltaType: 'up', hint: '较昨日 33 条' },
  { label: '今日异常检测', value: 47, delta: '+8', deltaType: 'down', hint: '命中异常点' },
  { label: '告警压缩率', value: '86.4', unit: '%', delta: '+4.2%', deltaType: 'up', hint: '目标 ≥80%' },
  { label: '平均 MTTR', value: '11.5', unit: 'min', delta: '-62%', deltaType: 'up', hint: '目标 <12min' }
]

export const alertTrend = {
  labels: timeLabels(24, 60),
  raw: seededSeries(7, 24, 210, 90, 0.3).map((v) => Math.round(v)),
  effective: seededSeries(13, 24, 32, 14, 0.35).map((v) => Math.round(v))
}

export const severityDist = [
  { name: 'P0', value: 2, itemStyle: { color: '#c93b3b' } },
  { name: 'P1', value: 4, itemStyle: { color: '#d97b29' } },
  { name: 'P2', value: 6, itemStyle: { color: '#b8961f' } },
  { name: 'P3', value: 18, itemStyle: { color: '#0e7c72' } },
  { name: 'P4', value: 26, itemStyle: { color: '#8a9693' } }
]

export const serviceHealth = [
  { service: 'order-service', health: 72, status: 'warning', anomaly: 3, trend: 'down' },
  { service: 'payment-service', health: 81, status: 'warning', anomaly: 2, trend: 'flat' },
  { service: 'user-service', health: 96, status: 'online', anomaly: 0, trend: 'up' },
  { service: 'inventory-service', health: 93, status: 'online', anomaly: 1, trend: 'up' },
  { service: 'gateway-service', health: 98, status: 'online', anomaly: 0, trend: 'up' },
  { service: 'search-service', health: 89, status: 'online', anomaly: 1, trend: 'flat' }
]

// ---------------- 告警 ----------------
export const alerts = [
  { id: 'ALT-260812-091', title: '数据库连接超时激增：order-db-primary 连接超时 150 条/分钟', service: 'order-service', level: 'p0', status: 'active', source: '异常检测 · Prophet', count: 156, time: '2026-08-12T13:42:00' },
  { id: 'ALT-260812-088', title: 'order-db-primary CPU 使用率 98%，超出动态阈值上界 3σ', service: 'order-service', level: 'p0', status: 'active', source: '动态阈值', count: 42, time: '2026-08-12T13:38:00' },
  { id: 'ALT-260812-085', title: 'payment-service P99 延迟 2.8s，超出基线 240%', service: 'payment-service', level: 'p1', status: 'active', source: '异常检测 · IsolationForest', count: 28, time: '2026-08-12T13:21:00' },
  { id: 'ALT-260812-082', title: '慢查询数量突增：order_db 慢 SQL 较基线增长 500%', service: 'order-service', level: 'p1', status: 'acked', source: '日志异常 · Drain', count: 89, time: '2026-08-12T13:15:00' },
  { id: 'ALT-260812-079', title: 'inventory-service 库存扣减失败率 4.2%', service: 'inventory-service', level: 'p2', status: 'acked', source: '指标异常 · GMM', count: 17, time: '2026-08-12T12:56:00' },
  { id: 'ALT-260812-076', title: 'Kafka 消费滞后：topic logs-raw lag 超过 50k', service: 'aiops-collector', level: 'p2', status: 'active', source: '静态规则', count: 9, time: '2026-08-12T12:40:00' },
  { id: 'ALT-260812-071', title: 'node-07 内存可用量低于 10%，预测 26 小时后耗尽', service: 'node-07', level: 'p2', status: 'active', source: '时序预测 · Prophet', count: 5, time: '2026-08-12T12:02:00' },
  { id: 'ALT-260812-068', title: 'payment-service 调用 order-service 错误率上升至 8.6%', service: 'payment-service', level: 'p1', status: 'acked', source: '关联分析 · VAE', count: 34, time: '2026-08-12T11:47:00' },
  { id: 'ALT-260812-063', title: 'ES 集群写入拒绝率 0.8%，bulk 队列接近饱和', service: 'elasticsearch', level: 'p3', status: 'resolved', source: '静态规则', count: 12, time: '2026-08-12T11:20:00' },
  { id: 'ALT-260812-058', title: 'gateway-service 5xx 比例短时波动（已自动恢复）', service: 'gateway-service', level: 'p3', status: 'resolved', source: '动态阈值', count: 6, time: '2026-08-12T10:55:00' },
  { id: 'ALT-260812-054', title: 'search-service 索引重建耗时超出预期 35%', service: 'search-service', level: 'p4', status: 'resolved', source: '时序预测', count: 3, time: '2026-08-12T10:12:00' },
  { id: 'ALT-260812-049', title: '证书剩余有效期不足 30 天：api.internal.example', service: 'gateway-service', level: 'p4', status: 'acked', source: '巡检任务', count: 1, time: '2026-08-12T09:30:00' }
]

export const alertRules = [
  { name: 'CPU 使用率动态阈值', type: '动态阈值', target: 'node_*', status: 'running' },
  { name: 'HTTP 5xx 比例 > 1%', type: '静态阈值', target: 'gateway-service', status: 'running' },
  { name: '数据库连接池耗尽预警', type: '时序预测', target: '*-db-*', status: 'running' },
  { name: '日志 ERROR 频率突增', type: '日志异常', target: '全部服务', status: 'running' }
]

// ---------------- 异常检测 ----------------
export const anomalyAlgorithms = [
  { scene: '单指标异常', algo: 'Isolation Forest', type: '无监督', advantage: '无需标注数据，适合高维', data: 'CPU、内存、QPS', status: 'running' },
  { scene: '单指标异常', algo: 'Prophet', type: '时序分解', advantage: '可解释性强，自动处理趋势/季节', data: '有周期性规律的指标', status: 'running' },
  { scene: '单指标异常', algo: 'LSTM Autoencoder', type: '深度学习', advantage: '捕捉复杂时序模式', data: '高频指标（秒级）', status: 'running' },
  { scene: '多指标关联', algo: 'VAE', type: '深度学习', advantage: '学习正常模式流形', data: '多维度指标向量', status: 'running' },
  { scene: '多指标关联', algo: 'GMM', type: '概率模型', advantage: '计算快，可解释', data: '中等维度（<50）', status: 'running' },
  { scene: '日志异常', algo: 'Drain', type: '模板提取', advantage: '高效，在线学习', data: '非结构化日志', status: 'running' },
  { scene: '日志异常', algo: 'LogBERT', type: 'NLP', advantage: '语义理解能力强', data: '含语义信息的日志', status: 'paused' },
  { scene: '日志异常', algo: 'DeepLog', type: 'LSTM', advantage: '序列模式学习', data: '系统调用序列', status: 'running' }
]

export const anomalyResults = [
  { metric: 'node_cpu_seconds_total', service: 'order-db-primary', algorithm: 'Prophet', score: 0.89, confidence: 0.92, value: 95.2, expected: 45.0, deviation: 50.2, severity: 'critical', time: '2026-08-12T13:38:00',
    explanation: { method: '基于 Prophet 时序分解', trend: '整体呈上升趋势', seasonality: '检测到日周期模式', reason: '实际值超出预测区间 3 个标准差' } },
  { metric: 'http_requests_duration_seconds', service: 'payment-service', algorithm: 'Isolation Forest', score: 0.81, confidence: 0.87, value: 2800, expected: 820, deviation: 1980, severity: 'major', time: '2026-08-12T13:21:00',
    explanation: { method: '孤立森林异常评分', trend: 'P99 延迟阶跃式上升', seasonality: '无明显周期', reason: '延迟分布与历史正常样本显著偏离' } },
  { metric: 'node_memory_MemAvailable_bytes', service: 'node-07', algorithm: 'Prophet', score: 0.74, confidence: 0.83, value: 1.2, expected: 4.8, deviation: -3.6, severity: 'major', time: '2026-08-12T12:02:00',
    explanation: { method: '基于 Prophet 时序分解', trend: '可用内存持续下降', seasonality: '检测到日周期模式', reason: '按当前斜率 26 小时后耗尽' } },
  { metric: 'kafka_consumer_lag', service: 'aiops-collector', algorithm: 'GMM', score: 0.66, confidence: 0.78, value: 52000, expected: 3000, deviation: 49000, severity: 'minor', time: '2026-08-12T12:40:00',
    explanation: { method: '高斯混合模型聚类偏离', trend: '滞后量阶梯式累积', seasonality: '无明显周期', reason: '消费速率低于生产速率持续 40 分钟' } }
]

export const metricSeries = {
  labels: timeLabels(48, 30),
  actual: seededSeries(29, 48, 46, 18, 0.25).map((v, i) => (i >= 40 ? v + (i - 39) * 6 : v)),
  expected: seededSeries(29, 48, 46, 18, 0.25),
  anomalyIndex: [41, 42, 44, 46, 47]
}

// ---------------- 根因分析 ----------------
export const rcaCases = [
  {
    incidentId: 'inc_2026081201', title: '订单服务大面积超时', status: 'active', time: '2026-08-12T13:38:00',
    rootCauses: [
      { rank: 1, confidence: 0.88, entity: { type: 'database', name: 'order-db-primary', namespace: 'production' },
        evidence: ['数据库连接超时日志激增（150 条/分钟）', '数据库 CPU 使用率从 30% 飙升至 98%', '慢查询数量在 13:35 突增 500%'],
        impact: { services: ['order-service', 'payment-service', 'inventory-service'], users: '约 12,000 用户', severity: 'p0' },
        actions: ['检查数据库慢查询日志，定位耗时 SQL', '评估是否需要扩容数据库连接池', '考虑启用读写分离或数据库主从切换'] },
      { rank: 2, confidence: 0.61, entity: { type: 'service', name: 'order-service', namespace: 'production' },
        evidence: ['13:32 有一次新版本发布（v2.18.0）', '发布后 GC 停顿时间增加 3 倍', '线程池活跃线程数接近上限'],
        impact: { services: ['order-service'], users: '约 4,000 用户', severity: 'p1' },
        actions: ['回滚至 v2.17.2 并观察指标', '检查新版本中的连接泄漏问题'] },
      { rank: 3, confidence: 0.34, entity: { type: 'host', name: 'node-07', namespace: 'production' },
        evidence: ['宿主机内存可用量持续下降', '磁盘 IO await 高于基线 2 倍'],
        impact: { services: ['多个中间件 Pod'], users: '间接影响', severity: 'p2' },
        actions: ['驱逐节点上的非关键 Pod', '排查内存泄漏进程'] }
    ]
  },
  {
    incidentId: 'inc_2026081103', title: '支付服务延迟抖动', status: 'resolved', time: '2026-08-11T22:14:00',
    rootCauses: [
      { rank: 1, confidence: 0.79, entity: { type: 'middleware', name: 'redis-cluster-3', namespace: 'production' },
        evidence: ['Redis 慢日志突增', '某个大 Key 过期引发缓存击穿', 'payment-service 缓存命中率从 97% 降至 61%'],
        impact: { services: ['payment-service'], users: '约 2,300 用户', severity: 'p1' },
        actions: ['对大 Key 设置随机过期时间', '增加热点 Key 本地缓存'] }
    ]
  }
]

export const topologyNodes = [
  { name: 'gateway-service', x: 300, y: 40, status: 'online' },
  { name: 'order-service', x: 160, y: 140, status: 'critical' },
  { name: 'payment-service', x: 300, y: 140, status: 'warning' },
  { name: 'user-service', x: 440, y: 140, status: 'online' },
  { name: 'inventory-service', x: 90, y: 250, status: 'warning' },
  { name: 'order-db-primary', x: 230, y: 250, status: 'critical' },
  { name: 'redis-cluster', x: 370, y: 250, status: 'warning' },
  { name: 'user-db', x: 500, y: 250, status: 'online' }
]

export const topologyEdges = [
  ['gateway-service', 'order-service'], ['gateway-service', 'payment-service'], ['gateway-service', 'user-service'],
  ['order-service', 'inventory-service'], ['order-service', 'order-db-primary'],
  ['payment-service', 'order-db-primary'], ['payment-service', 'redis-cluster'], ['user-service', 'user-db']
]

// ---------------- 日志分析 ----------------
export const logPipeline = ['日志解析', '模板提取 (Drain)', '参数分离', '向量化', '聚类分析', '异常检测']

export const logClusters = [
  { id: 'C-018', pattern: 'Connection to database <*> failed after <*>ms', count: 1240, level: 'error', services: ['order-service'], trend: 'spike', firstSeen: '13:32' },
  { id: 'C-017', pattern: 'Slow query took <*>ms: SELECT * FROM orders WHERE <*>', count: 486, level: 'warn', services: ['order-service'], trend: 'spike', firstSeen: '13:35' },
  { id: 'C-014', pattern: 'Payment callback timeout, orderId=<*>', count: 96, level: 'error', services: ['payment-service'], trend: 'rising', firstSeen: '13:10' },
  { id: 'C-011', pattern: 'Cache miss for key <*>, fallback to db', count: 2043, level: 'warn', services: ['payment-service'], trend: 'flat', firstSeen: '09:00' },
  { id: 'C-009', pattern: 'Request completed in <*>ms', count: 184203, level: 'info', services: ['gateway-service'], trend: 'flat', firstSeen: '00:00' },
  { id: 'C-006', pattern: 'GC pause (G1 Evacuation Pause) <*>ms', count: 58, level: 'warn', services: ['order-service'], trend: 'rising', firstSeen: '13:33' }
]

export const logTemplates = [
  { raw: 'Connection to database order-db failed after 5000ms', template: 'Connection to database <*> failed after <*>ms', params: ['order-db', '5000'] },
  { raw: 'Connection to database user-db failed after 3000ms', template: 'Connection to database <*> failed after <*>ms', params: ['user-db', '3000'] },
  { raw: 'Slow query took 2341ms: SELECT * FROM orders WHERE user_id=88213', template: 'Slow query took <*>ms: SELECT * FROM orders WHERE <*>', params: ['2341', 'user_id=88213'] }
]

export const logSeries = {
  labels: timeLabels(24, 60),
  total: seededSeries(41, 24, 8200, 2400, 0.2).map((v) => Math.round(v)),
  error: seededSeries(43, 24, 60, 25, 0.4).map((v, i) => Math.round(i >= 20 ? v * 4 : v))
}

// ---------------- 告警降噪 ----------------
export const denoisePolicies = [
  { name: '时间窗口聚合', desc: '5 分钟内相同告警合并为一条', effect: '减少 40% 重复告警', enabled: true, reduced: 860 },
  { name: '相似度合并', desc: '基于文本相似度合并同类告警', effect: '减少 30% 相似告警', enabled: true, reduced: 640 },
  { name: '拓扑抑制', desc: '父节点故障抑制子节点告警', effect: '减少 50% 级联告警', enabled: true, reduced: 210 },
  { name: '动态阈值', desc: '基于历史数据自适应调整阈值', effect: '减少 60% 阈值误报', enabled: true, reduced: 130 },
  { name: '智能分级', desc: '基于影响面和紧急度自动定级', effect: '提升关键告警识别率', enabled: true, reduced: 0 }
]

export const denoiseStats = {
  rawTotal: 2316, afterFilter: 1996, afterWindow: 1136, afterSimilarity: 496, afterTopology: 286, afterThreshold: 156, final: 156,
  funnel: [
    { name: '原始告警', value: 2316 },
    { name: '规则过滤后', value: 1996 },
    { name: '窗口聚合后', value: 1136 },
    { name: '相似度去重后', value: 496 },
    { name: '拓扑抑制后', value: 286 },
    { name: '动态阈值后', value: 156 },
    { name: '有效通知', value: 156 }
  ]
}

// ---------------- 数据源接入 ----------------
export const dataSources = [
  { name: 'Prometheus', type: '指标数据', status: 'online', method: 'Remote Write / Federation', endpoint: 'http://prometheus:9090', desc: '指标采集：Counter / Gauge / Histogram / Summary 全类型覆盖', metricsRate: '42k samples/s' },
  { name: 'ELK (Elasticsearch)', type: '日志数据', status: 'online', method: 'API 查询 / DSL', endpoint: 'http://elasticsearch:9200', desc: '日志接入：结构化解析、时间窗口拉取、写入 Kafka logs-raw', metricsRate: '8.6k docs/s' }
]

export const collectorStats = {
  labels: timeLabels(24, 60),
  metrics: seededSeries(51, 24, 42, 8, 0.15).map((v) => Math.round(v * 10) / 10),
  logs: seededSeries(53, 24, 8.6, 2.2, 0.2).map((v) => Math.round(v * 10) / 10)
}

export const logFields = [
  { field: '@timestamp', type: 'date', desc: '日志时间', example: '2026-08-12T08:00:00.000Z' },
  { field: 'service', type: 'keyword', desc: '服务名', example: 'order-service' },
  { field: 'level', type: 'keyword', desc: '日志级别', example: 'ERROR, WARN, INFO' },
  { field: 'trace_id', type: 'keyword', desc: '链路追踪 ID', example: 'abc123def456' },
  { field: 'host', type: 'keyword', desc: '主机名', example: 'node-01' },
  { field: 'message', type: 'text', desc: '原始日志内容', example: 'Connection timeout...' },
  { field: 'error_type', type: 'keyword', desc: '错误类型', example: 'TimeoutException' },
  { field: 'response_time', type: 'long', desc: '响应时间 (ms)', example: '5000' }
]

export const metricTypes = [
  { type: 'Counter', desc: '累计值，只增不减', scene: '请求总量、错误总量突变检测' },
  { type: 'Gauge', desc: '可增可减的瞬时值', scene: 'CPU、内存、磁盘使用率异常' },
  { type: 'Histogram', desc: '分桶统计', scene: '请求延迟分布异常' },
  { type: 'Summary', desc: '分位数统计', scene: 'P99/P95 延迟突增检测' }
]

// ---------------- AI 对话 ----------------
export const chatMessages = [
  { role: 'assistant', content: '你好，我是 AIOPS 智能运维助手。你可以问我关于告警、异常检测、根因分析或巡检报告的问题。', time: '2026-08-12T09:00:00' },
  { role: 'user', content: '今天有哪些 P0 告警？', time: '2026-08-12T09:01:12' },
  { role: 'assistant', content: '今日共有 2 条 P0 告警：\n1. order-db-primary 数据库连接超时激增（156 条/分钟）\n2. order-db-primary CPU 使用率 98%，超出动态阈值上界 3σ\n\n两条告警均指向 order-service 链路，建议优先排查数据库侧。', time: '2026-08-12T09:01:15' },
  { role: 'user', content: '帮我分析一下根因', time: '2026-08-12T09:02:30' },
  { role: 'assistant', content: '基于当前告警和拓扑关联，最高置信度根因是 **order-db-primary 数据库实例**：\n- 13:35 慢查询数量突增 500%\n- 13:38 CPU 飙升至 98%\n- 13:42 连接超时日志达到 150 条/分钟\n\n次要根因（置信度 0.61）为 order-service v2.18.0 发布后 GC 停顿增加。建议先检查数据库慢查询，再评估是否需要回滚应用版本。', time: '2026-08-12T09:02:35' }
]

// ---------------- AI 配置 / LLM 管理 ----------------
export const llmConfigs = [
  { id: 'llm-001', name: '生产 GPT-4o', provider: 'OpenAI', model: 'gpt-4o', endpoint: 'https://api.openai.com/v1', apiKeyMasked: 'sk-****4f2a', status: 'active', temperature: 0.3, maxTokens: 4096 },
  { id: 'llm-002', name: '国产 Qwen-Max', provider: '阿里云', model: 'qwen-max', endpoint: 'https://dashscope.aliyuncs.com/api/v1', apiKeyMasked: 'sk-****9c1b', status: 'active', temperature: 0.5, maxTokens: 2048 },
  { id: 'llm-003', name: '内部 DeepSeek', provider: '私有化', model: 'deepseek-coder', endpoint: 'http://llm.internal:8000/v1', apiKeyMasked: '-', status: 'paused', temperature: 0.2, maxTokens: 8192 }
]

// ---------------- AI 配置 / Skill 管理 ----------------
export const skills = [
  { id: 'skill-001', name: '告警解读', description: '自动解读告警内容、影响面和可能原因，生成自然语言摘要。', version: 'v1.2.0', status: 'running', trigger: '告警产生 / 用户询问' },
  { id: 'skill-002', name: '根因分析', description: '结合拓扑、指标、日志进行多维度关联，输出 Top-N 根因假设。', version: 'v2.0.1', status: 'running', trigger: 'P0/P1 告警聚合 / 手动触发' },
  { id: 'skill-003', name: '降噪建议', description: '基于历史处置记录推荐降噪策略与阈值调整建议。', version: 'v1.0.4', status: 'running', trigger: '降噪效果评估任务' },
  { id: 'skill-004', name: '巡检报告生成', description: '每日定时汇总系统健康度、告警闭环、容量趋势，生成巡检报告。', version: 'v1.1.0', status: 'running', trigger: '每日 09:00 定时任务' },
  { id: 'skill-005', name: '处置方案推荐', description: '根据告警类型和历史 SOP 推荐标准处置步骤。', version: 'v0.9.0', status: 'paused', trigger: '告警确认后' }
]

// ---------------- 巡检报告管理 ----------------
export const inspectionReports = [
  { id: 'RPT-20260812', title: '2026-08-12 日巡检报告', date: '2026-08-12', score: 87, summary: '今日系统整体健康，P0 告警 2 条已闭环，容量暂无风险。', status: 'pushed' },
  { id: 'RPT-20260811', title: '2026-08-11 日巡检报告', date: '2026-08-11', score: 92, summary: '昨日无 P0 告警，平均 MTTR 11.5 分钟，异常检测命中率平稳。', status: 'pushed' },
  { id: 'RPT-20260810', title: '2026-08-10 日巡检报告', date: '2026-08-10', score: 78, summary: '支付服务延迟抖动导致 P1 告警，已定位 Redis 大 Key 问题。', status: 'pushed' },
  { id: 'RPT-20260809', title: '2026-08-09 日巡检报告', date: '2026-08-09', score: 95, summary: '全天运行平稳，无高危告警，容量余量充足。', status: 'pushed' },
  { id: 'RPT-20260808', title: '2026-08-08 日巡检报告', date: '2026-08-08', score: 83, summary: '网关 5xx 比例短时波动，已自动恢复，建议关注证书有效期。', status: 'pushed' }
]

export const inspectionReportDetails = {
  'RPT-20260812': {
    id: 'RPT-20260812',
    title: '2026-08-12 日巡检报告',
    date: '2026-08-12',
    score: 87,
    sections: [
      {
        title: '一、总体健康度',
        content: [
          '系统综合健康度 87 分（较昨日 -5 分）。',
          '核心服务在线率 99.97%，中间件无异常。',
          '活动告警 12 条，其中 P0 2 条、P1 4 条，均已确认或解决。'
        ]
      },
      {
        title: '二、告警闭环',
        content: [
          '今日新增告警 56 条，降噪后有效 156 条，压缩率 86.4%。',
          '平均确认时长 4.3 分钟，平均解决时长 11.5 分钟。',
          '主要事件：order-db-primary 数据库 CPU 与连接超时告警，已定位慢查询。'
        ]
      },
      {
        title: '三、异常检测',
        content: [
          '今日异常检测命中 47 个点，误报率估算 6.2%。',
          '关键异常：payment-service P99 延迟 2.8s、node-07 内存可用量低于 10%。',
          'Prophet / IsolationForest / VAE 模型运行正常。'
        ]
      },
      {
        title: '四、容量与风险',
        content: [
          '数据库连接池峰值 78%，余量充足。',
          'node-07 内存按当前斜率预计 26 小时后耗尽，建议持续关注。',
          '证书：api.internal.example 剩余 28 天，建议下周安排续期。'
        ]
      },
      {
        title: '五、今日建议',
        content: [
          '1. 优化 order_db 慢查询（SELECT * FROM orders WHERE user_id=?）。',
          '2. 评估 payment-service 热点 Key 本地缓存方案。',
          '3. 跟进 node-07 内存趋势，必要时驱逐非关键 Pod。'
        ]
      }
    ]
  },
  'RPT-20260811': {
    id: 'RPT-20260811',
    title: '2026-08-11 日巡检报告',
    date: '2026-08-11',
    score: 92,
    sections: [
      { title: '一、总体健康度', content: ['系统综合健康度 92 分，运行平稳。', '无 P0 告警，P1 告警 1 条已闭环。'] },
      { title: '二、告警闭环', content: ['今日新增告警 33 条，降噪后有效 89 条。', '平均 MTTR 11.5 分钟。'] },
      { title: '三、异常检测', content: ['异常检测命中 39 个点，模型运行正常。'] },
      { title: '四、容量与风险', content: ['各节点资源使用率处于安全水位。'] },
      { title: '五、今日建议', content: ['继续保持当前阈值策略。'] }
    ]
  }
}

// ---------------- 人员组织 / 用户管理 ----------------
export const users = [
  { id: 'u-001', username: 'admin', realName: '运维管理员', email: 'admin@aiops.local', phone: '138****0001', department: 'SRE 团队', roles: ['系统管理员', '值班经理'], status: 'active' },
  { id: 'u-002', username: 'zhangsan', realName: '张三', email: 'zhangsan@aiops.local', phone: '138****0002', department: 'SRE 团队', roles: ['值班工程师'], status: 'active' },
  { id: 'u-003', username: 'lisi', realName: '李四', email: 'lisi@aiops.local', phone: '138****0003', department: '平台研发', roles: ['开发工程师'], status: 'active' },
  { id: 'u-004', username: 'wangwu', realName: '王五', email: 'wangwu@aiops.local', phone: '138****0004', department: '安全团队', roles: ['安全审计员'], status: 'paused' },
  { id: 'u-005', username: 'zhaoliu', realName: '赵六', email: 'zhaoliu@aiops.local', phone: '138****0005', department: 'SRE 团队', roles: ['值班工程师', '告警处理人'], status: 'active' }
]

// ---------------- 人员组织 / 角色管理 ----------------
export const roles = [
  { id: 'role-001', name: '系统管理员', code: 'admin', description: '拥有系统全部权限，可管理用户、角色与全局配置。', userCount: 1, permissions: ['用户管理', '角色管理', '系统配置', '数据查看'] },
  { id: 'role-002', name: '值班经理', code: 'duty_manager', description: '负责值班排班、告警升级策略与重大事件协调。', userCount: 1, permissions: ['告警管理', '通知策略', '巡检报告'] },
  { id: 'role-003', name: '值班工程师', code: 'duty_engineer', description: '日常告警处理、异常确认与基础运维操作。', userCount: 2, permissions: ['告警处理', '异常查看', '日志查询'] },
  { id: 'role-004', name: '开发工程师', code: 'developer', description: '查看服务指标、日志与根因分析结果，辅助定位问题。', userCount: 1, permissions: ['指标查看', '日志查看', '根因分析'] },
  { id: 'role-005', name: '安全审计员', code: 'auditor', description: '审计系统操作日志、访问记录与合规报告。', userCount: 1, permissions: ['审计日志', '用户查看', '报告导出'] },
  { id: 'role-006', name: '告警处理人', code: 'alert_handler', description: '接收并处理分配的告警，可执行确认、解决等操作。', userCount: 1, permissions: ['告警处理', '工单查看'] }
]

// ---------------- 运维知识库 ----------------
export const knowledgeBaseFiles = [
  { id: 'kb-001', name: 'MySQL 连接超时排查手册.docx', type: 'docx', size: '1.8 MB', status: 'indexed', uploadTime: '2026-08-10 14:32', uploader: '张三' },
  { id: 'kb-002', name: 'order-service 发布变更规范.pdf', type: 'pdf', size: '856 KB', status: 'indexed', uploadTime: '2026-08-11 09:15', uploader: '李四' },
  { id: 'kb-003', name: 'Redis 集群故障应急 Runbook.md', type: 'md', size: '42 KB', status: 'indexed', uploadTime: '2026-08-11 16:48', uploader: '王五' },
  { id: 'kb-004', name: 'Kafka 消费延迟处理建议.txt', type: 'txt', size: '8 KB', status: 'pending', uploadTime: '2026-08-12 11:20', uploader: '赵六' },
  { id: 'kb-005', name: 'P0 告警升级流程.doc', type: 'doc', size: '2.1 MB', status: 'indexed', uploadTime: '2026-08-12 13:55', uploader: '运维管理员' }
]

