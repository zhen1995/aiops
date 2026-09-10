export default {
  severity: {
    p1: 'P1-紧急',
    p2: 'P2-警告',
    p3: 'P3-提醒'
  },
  events: {
    title: '告警事件',
    desc: '由系统按告警规则自动探测生成的活跃与历史告警事件',
    tabs: {
      active: '活跃告警',
      history: '历史告警'
    },
    filters: {
      timeWindow: '时间窗口',
      last24h: '最近 24 小时',
      last7d: '最近 7 天',
      last30d: '最近 30 天',
      severity: '级别',
      all: '全部',
      searchPlaceholder: '搜索规则名称或告警对象',
      query: '查询',
      refresh: '刷新',
      loading: '加载中...'
    },
    batch: {
      selectAll: '全选本页',
      selected: '已选 {count} 条',
      cancelSelection: '取消选择',
      batchDelete: '批量删除',
      deleting: '删除中...'
    },
    table: {
      ruleName: '规则名称',
      severity: '级别',
      target: '告警对象',
      triggerTime: '触发时间',
      tags: '标签',
      triggerValue: '触发值',
      actions: '操作',
      rca: '根因分析',
      delete: '删除',
      empty: '暂无告警事件数据',
      loading: '加载中...',
      total: '共 {total} 条',
      prev: '上一页',
      next: '下一页',
      page: '第 {page} 页'
    },
    msg: {
      rcaFailed: '触发根因分析失败：{msg}',
      loadFailed: '加载告警事件失败：{msg}',
      deleteFailed: '删除告警事件失败：{msg}',
      batchDeleteFailed: '批量删除失败：{msg}',
      confirmDelete: '确认删除告警事件「{label}」？',
      confirmBatchDelete: '确认删除选中的 {count} 条告警事件？'
    }
  },
  denoise: {
    title: '告警降噪',
    desc: '多级降噪策略链，重复与级联告警自动压缩，仅有效告警触达值班人员',
    loading: '加载中...',
    retry: '重试',
    funnelTitle: '降噪漏斗',
    funnelSub: '原始告警经窗口聚合、拓扑抑制逐级压缩，最终仅有效通知触达值班人员',
    funnelEmpty: '暂无统计数据',
    funnelTooltip: '{name}：{value} 条',
    policyTitle: '降噪策略',
    policySub: '策略按序执行，可随时启停；开关变更实时生效',
    switchEnable: '点击启用',
    switchDisable: '点击停用',
    suppressedToday: '今日已拦截',
    viewRecords: '查看今日该策略拦截的通知记录',
    kpi: {
      rawTotal: '原始告警总量',
      rawHint: '今日进入降噪管道',
      effective: '有效告警',
      effectiveHint: '实际触达值班',
      compressionRate: '压缩率',
      compressionHint: '降噪整体效果',
      suppressed: '已拦截告警',
      suppressedHint: '降噪策略合计拦截'
    },
    effect: {
      windowAggregation: '减少重复告警',
      topologySuppression: '减少级联告警'
    },
    msg: {
      loadFailed: '加载降噪数据失败：{msg}',
      toggleFailed: '切换策略失败：{msg}'
    }
  },
  rules: {
    title: '告警规则',
    desc: '创建和维护系统内的告警规则，基于 PromQL 表达式触发告警',
    addRule: '新增规则',
    listTitle: '规则列表',
    listSub: '已配置的告警规则及启用状态',
    keyword: '关键字：{kw}',
    refresh: '刷新',
    loading: '加载中...',
    colStatus: '状态',
    colRuleName: '规则名称',
    colSeverity: '告警级别',
    colEvalInterval: '执行频率',
    colDuration: '持续时间（秒）',
    colNotifyRule: '通知规则',
    colEnabled: '启用状态',
    colCreatedAt: '创建时间',
    colActions: '操作',
    everySeconds: '每 {n} 秒',
    durationImmediate: '0（立即）',
    enabledOn: '已启用',
    enabledOff: '已停用',
    edit: '编辑',
    enable: '启用',
    disable: '停用',
    delete: '删除',
    activeBadge: '告警中 {count} 条',
    emptyNoMatch: '没有匹配关键字的告警规则',
    empty: '暂无告警规则，请点击"新增规则"创建',
    statusTitle: {
      disabled: '规则已停用',
      alerting: '告警中（{count} 条活跃告警），点击查看',
      noAlert: '无告警，点击查看最近告警'
    },
    eventsActive: '· 活跃告警 {count} 条',
    eventsNone: '· 暂无活跃告警',
    eventsEmpty: '该规则下当前没有活跃的告警事件',
    triggerValue: '触发值：',
    triggerType: {
      all: '触发+恢复',
      firing: '仅触发',
      recovered: '仅恢复'
    },
    modal: {
      editTitle: '编辑告警规则',
      createTitle: '新增告警规则'
    },
    form: {
      nameLabel: '规则名称',
      namePlaceholder: '例如：CPU 使用率超过 90%',
      promqlPlaceholder: '例如：100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 90',
      evalIntervalLabel: '执行频率',
      durationLabel: '持续时间（秒）',
      durationPlaceholder: '例如：300',
      durationHint: '0 表示只要有一次查询满足告警条件即触发',
      severityLabel: '告警级别',
      notifyRuleLabel: '通知规则',
      notifyOptional: '选填，告警触发/恢复时自动发送通知',
      noNotify: '不发送通知',
      notifyConfigTitle: '通知配置',
      repeatLabel: '重复通知间隔（分钟）',
      repeatPlaceholder: '例如：60',
      repeatHint: '如果告警持续未恢复，间隔 xx 分钟之后重复提醒；填 0 表示不重复',
      maxSendLabel: '最大发送次数',
      maxSendPlaceholder: '例如：3',
      maxSendHint: '如果值为 0，则不做最大发送次数的限制',
      enabledLabel: '启用状态',
      cancel: '取消',
      save: '保存',
      saving: '保存中...'
    },
    error: {
      nameRequired: '请输入规则名称',
      promqlRequired: '请输入 PromQL 表达式',
      evalRequired: '请选择执行频率',
      durationNegative: '持续时间不能为负数'
    },
    msg: {
      loadFailed: '加载告警规则失败：{msg}',
      updateFailed: '更新失败：{msg}',
      createFailed: '新建失败：{msg}',
      toggleFailed: '切换状态失败：{msg}',
      confirmDelete: '确定要删除告警规则"{name}"吗？',
      deleteFailed: '删除失败：{msg}'
    }
  }
}
