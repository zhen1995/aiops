export default {
  header: {
    title: '运维总览大盘',
    desc: '多源监控数据融合 · 异常检测 / 根因分析 / 告警降噪 实时概览',
    range24h: '近 24 小时',
    rangeCustom: '自定义范围',
    to: '至',
    confirm: '确定',
    retry: '重试',
    noData: '暂无数据'
  },
  kpi: {
    activeAlerts: {
      label: '活动告警',
      hint: '较昨日变化',
      help: '当前仍处于未恢复状态（firing）的告警总数，不限触发时间。下方百分比为与 24 小时前未恢复告警数的对比。'
    },
    anomalyToday: {
      label: '今日异常检测',
      hint: '命中异常点',
      help: '今日 0 点起新触发的告警事件数，「+N」为与昨日同一时段相比的差值。'
    },
    compressionRate: {
      label: '告警压缩率',
      hint: '目标 ≥80%',
      help: '降噪拦截量占（告警事件 + 降噪拦截）的比例，反映智能降噪的压缩效果，目标 ≥80%。'
    },
    mttr: {
      label: '平均 MTTR',
      hint: '目标 <12min',
      help: '范围内已恢复告警从触发到恢复的平均耗时（分钟），目标 <12min。'
    }
  },
  trend: {
    title: '告警趋势（原始 vs 降噪后）',
    sub: '智能降噪实时生效，压缩率 {rate}%',
    legendRaw: '原始告警',
    legendDenoised: '降噪后告警'
  },
  severity: {
    title: '告警级别分布',
    sub: '范围内全部告警按级别分级'
  },
  health: {
    title: '服务健康度',
    sub: '基于多指标关联检测综合评分',
    abnormalMetrics: '异常指标 {count} 个',
    trendWorse: '↓ 恶化',
    trendBetter: '↑ 好转',
    trendStable: '→ 平稳'
  },
  alerts: {
    title: '最新高危告警',
    sub: '范围内最新告警事件（按级别排序）',
    view: '查看告警事件',
    colLevel: '级别',
    colContent: '告警内容',
    colService: '服务',
    colTime: '时间',
    p1: 'P1-紧急',
    p2: 'P2-警告',
    p3: 'P3-提醒'
  },
  error: {
    loadFailed: '加载总览数据失败'
  }
}
