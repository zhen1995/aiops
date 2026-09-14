export default {
  header: {
    title: '日志分析',
    desc: '非结构化日志智能解析 · Drain 模板提取 · 日志聚类与异常检测',
    last1h: '近 1 小时',
    last6h: '近 6 小时',
    last12h: '近 12 小时',
    last24h: '近 24 小时',
    query: '查询',
    querying: '查询中…',
    serviceAll: '全部服务',
    serviceSelected: '已选 {count} 个服务',
    selectAll: '全选',
    clear: '清空'
  },
  pipeline: {
    title: '日志处理流水线',
    sub: '原始日志经六环节实时处理，最终输出异常事件与告警',
    step1: '日志解析',
    step2: '模板提取 (Drain)',
    step3: '参数分离',
    step4: '向量化',
    step5: '聚类分析',
    step6: '异常检测',
    desc1: '格式识别与字段抽取',
    desc2: '在线聚类生成日志模板',
    desc3: '常量模板与变量参数拆分',
    desc4: '日志序列转为特征向量',
    desc5: '相似模式聚合为簇',
    desc6: '频次/新模板异常识别'
  },
  trend: {
    title: '日志量趋势',
    sub: '日志总量与 ERROR 日志量（条/小时）'
  },
  cluster: {
    title: '日志聚类',
    sub: '相似日志自动聚合为模式簇，按出现频次与趋势排序',
    total: '共 {count} 个活跃聚类',
    colId: '聚类 ID',
    colPattern: '日志模式',
    colCount: '数量',
    colLevel: '级别',
    colServices: '关联服务',
    colTrend: '趋势',
    colFirstSeen: '首次出现',
    trendSpike: '激增',
    trendRising: '上升',
    trendFlat: '平稳',
    filterAnomaly: '异常级别（ERROR/WARN）',
    filterError: 'ERROR',
    filterWarn: 'WARN',
    filterInfo: 'INFO',
    filterAll: '全部级别'
  },
  template: {
    title: 'Drain 模板提取',
    sub: '原始日志 → 常量模板 + 变量参数，在线学习实时更新',
    raw: '原始日志',
    tpl: '模板',
    params: '参数',
    count: '共 {count} 条'
  },
  chart: {
    total: '日志总量',
    error: 'ERROR 日志'
  },
  common: {
    empty: '暂无数据，日志分析任务运行后展示',
    loadFailed: '加载失败，请稍后重试',
    copy: '复制',
    copied: '已复制'
  }
}
