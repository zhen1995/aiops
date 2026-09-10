export default {
  record: {
    pageTitle: '通知记录',
    pageDesc: '每一次告警通知的发送轨迹：成功、失败、被降噪拦截均可追溯',
    filters: {
      status: '状态',
      strategy: '降噪策略',
      all: '全部',
      keywordPlaceholder: '搜索规则名称或告警对象',
      startTime: '开始时间',
      endTime: '结束时间',
      search: '查询',
      reset: '重置'
    },
    status: {
      success: '已发送',
      failed: '发送失败',
      intercepted: '已拦截',
      skipped: '已跳过'
    },
    strategy: {
      windowAggregation: '时间窗口聚合',
      topologySuppression: '拓扑抑制'
    },
    eventType: {
      alert: '告警',
      recovery: '恢复'
    },
    table: {
      time: '通知时间',
      ruleName: '规则名称',
      eventType: '事件类型',
      target: '告警对象',
      triggerValue: '触发值',
      status: '状态',
      media: '通知媒介',
      actions: '操作',
      detail: '详情',
      loading: '加载中...',
      empty: '暂无通知记录'
    },
    pagination: {
      total: '共 {count} 条',
      perPage: '{count} 条/页',
      prev: '上一页',
      next: '下一页',
      pageInfo: '第 {page} / {pages} 页'
    },
    detail: {
      title: '通知记录详情',
      basicInfo: '基本信息',
      notifyChain: '通知链路',
      messageContent: '消息正文',
      ruleName: '规则名称',
      severity: '级别',
      eventType: '事件类型',
      target: '告警对象',
      triggerValue: '触发值',
      triggerTime: '触发时间',
      tags: '标签',
      notifyTime: '通知时间',
      notifyRule: '通知规则',
      media: '通知媒介',
      mediaType: '媒介类型',
      template: '消息模板',
      noContent: '无（未进入发送）',
      close: '关闭'
    },
    error: {
      loadFailed: '加载通知记录失败：',
      loadDetailFailed: '加载通知记录详情失败：'
    }
  },
  rule: {
    pageTitle: '通知规则',
    pageDesc: '配置通知媒介与消息模板的多对多关联，可绑定夜莺告警规则',
    create: '+ 新建通知规则',
    modal: {
      editTitle: '编辑通知规则',
      createTitle: '新建通知规则',
      name: '规则名称',
      namePlaceholder: '例如：告警触发-钉钉',
      remark: '备注',
      remarkPlaceholder: '选填，说明这条规则的用途',
      media: '通知媒介',
      pickOne: '选一个',
      mediaEmpty: '暂无通知媒介，请到「通知媒介」菜单新增',
      selectMedia: '请选择通知媒介',
      template: '消息模板',
      templateEmpty: '暂无消息模板，请到「消息模板」菜单新增',
      selectTemplate: '请选择消息模板',
      enabled: '启用状态',
      cancel: '取消',
      save: '保存',
      saving: '保存中...'
    },
    list: {
      total: '共 {count} 条规则',
      name: '规则名称',
      triggerScene: '触发场景',
      media: '通知媒介',
      template: '消息模板',
      status: '状态',
      actions: '操作',
      edit: '编辑',
      delete: '删除',
      empty: '暂无通知规则',
      loading: '加载中...'
    },
    mediaType: {
      dingtalk: '钉钉',
      webhook: 'Webhook',
      email: '邮件',
      wecom: '企微'
    },
    trigger: {
      all: '触发+恢复',
      firing: '仅触发',
      recovered: '仅恢复'
    },
    error: {
      nameRequired: '请填写规则名称',
      saveFailed: '保存失败: ',
      deleteFailed: '删除失败: ',
      toggleFailed: '切换失败: ',
      confirmDelete: '确认删除通知规则「{name}」？'
    }
  },
  template: {
    pageTitle: '消息模板',
    pageDesc: '定义告警通知的内容模板，使用 Go template 语法，支持告警事件变量',
    create: '+ 新建模板',
    modal: {
      editTitle: '编辑消息模板',
      createTitle: '新建消息模板',
      name: '模板名称',
      namePlaceholder: '例如：钉钉-通用告警模板',
      scene: '适用场景',
      sceneAll: '通用（触发 + 恢复）',
      sceneFiring: '仅告警触发',
      sceneRecovered: '仅告警恢复',
      mediaType: '媒介类型',
      mediaDingtalk: '钉钉',
      mediaWebhook: 'Webhook',
      mediaEmail: '邮件',
      mediaWecom: '企业微信',
      description: '描述',
      descriptionPlaceholder: '模板说明，帮助识别用途',
      contentLabel: '模板内容 (Go template)',
      contentPlaceholder: '{{$event.RuleName}} 告警&#10;触发值: {{$event.TriggerValue}}&#10;站点: {{$.domain}}',
      preview: '实时预览',
      previewHint: '（模板合法后将显示渲染结果）',
      emptyPreview: '无内容',
      refTitle: '变量参考',
      refHint: '（点击可插入模板）',
      tabVars: '告警事件变量',
      tabFuncs: '内置函数',
      colVar: '变量',
      colFunc: '函数',
      colDesc: '说明',
      clickInsert: '点击插入: ',
      cancel: '取消',
      save: '保存模板',
      saving: '保存中...'
    },
    list: {
      edit: '编辑',
      delete: '删除',
      empty: '暂无模板，点击右上角新建',
      loading: '加载中...'
    },
    type: {
      all: '通用',
      firing: '触发',
      recovered: '恢复'
    },
    mediaLabel: {
      dingtalk: '钉钉',
      webhook: 'Webhook',
      email: '邮件',
      wecom: '企微'
    },
    vars: {
      event: '整个告警事件对象，可用于调试查看全部字段',
      labels: '事件标签 map，等价于 $event.TagsMap',
      value: '触发值，等价于 $event.TriggerValue',
      domain: '站点地址，用于拼接详情链接',
      ruleName: '告警规则名称',
      ruleNote: '告警规则备注/描述',
      id: '告警事件唯一 ID',
      severityLabel: '告警级别中文（P1-紧急 / P2-警告 / P3-提醒）',
      triggerValue: '触发值（PromQL 表达式计算结果）',
      busiGroupName: '夜莺业务组名称',
      cluster: '告警集群标识',
      triggerTime: '触发时间',
      lastEvalTime: '最近一次 PromQL 命中时间',
      firstTrigger: '首次触发时间',
      isRecovered: '是否已恢复（布尔）',
      tagsMap: '事件标签 map，支持按 key 取值',
      tagsJson: '事件标签 JSON 字符串',
      annotations: '附加 annotations 字段'
    },
    funcs: {
      timeformat: '格式化时间，第二个参数为 Go layout',
      timeformatCN: '简化版，输出 yyyy-MM-dd HH:mm:ss',
      timestamp: '当前时间字符串（常用于"发送时间"）',
      now: '当前秒级时间戳（int64），可用来算持续时长',
      sub: '减法，常用于计算告警持续秒数',
      add: '加法',
      mul: '乘法',
      humanizeDuration: '秒数 → 人类可读（3分 / 2时 / 1.5天）',
      humanizeDurationIfc: 'humanizeDuration 别名，兼容夜莺模板',
      durationHuman: '简化版（旧模板用，已废弃建议用 humanizeDuration）'
    },
    error: {
      loadFailed: '加载失败: ',
      nameContentRequired: '请填写模板名称和内容',
      saveFailed: '保存失败: ',
      deleteFailed: '删除失败: ',
      confirmDelete: '确认删除模板「{name}」？'
    }
  },
  medium: {
    pageTitle: '通知媒介',
    pageDesc: '维护钉钉、Webhook 回调等通知通道配置',
    create: '+ 新增媒介',
    list: {
      title: '媒介列表',
      subtitle: '告警通知将通过以下媒介进行分发',
      refresh: '刷新',
      loading: '加载中...',
      name: '媒介名称',
      type: '类型',
      summary: '配置摘要',
      status: '状态',
      actions: '操作',
      running: '运行中',
      disabled: '已停用',
      edit: '编辑',
      testing: '测试中...',
      test: '测试',
      disable: '停用',
      enable: '启用',
      delete: '删除',
      empty: '暂无数据，请点击"新增媒介"添加配置'
    },
    type: {
      dingtalk: '钉钉',
      webhook: 'Webhook 回调'
    },
    modal: {
      editTitle: '编辑媒介',
      createTitle: '新增媒介',
      name: '媒介名称',
      namePlaceholder: '例如：SRE 值班钉钉群',
      type: '媒介类型',
      selectType: '请选择类型',
      webhookUrl: '回调地址',
      webhookUrlPlaceholder: '请输入回调 URL，例如：https://example.com/api/notify',
      method: '请求方法',
      timeout: '超时(ms)',
      timeoutPlaceholder: '默认 5000',
      dingWebhook: '机器人 Webhook',
      secret: '加签 Secret',
      secretPlaceholder: '选填，开启加签时填写',
      show: '显示',
      hide: '隐藏',
      remark: '备注',
      remarkPlaceholder: '选填',
      enabled: '启用',
      cancel: '取消',
      save: '保存',
      saving: '保存中...'
    },
    test: {
      title: '发送测试消息',
      titleWithName: '发送测试消息 — {name}',
      content: '消息内容',
      contentPlaceholder: '请输入测试消息内容',
      sending: '发送中...',
      send: '发送测试'
    },
    error: {
      loadFailed: '加载媒介列表失败：',
      nameRequired: '请输入媒介名称',
      typeRequired: '请选择媒介类型',
      webhookUrlRequired: '请输入回调地址',
      dingWebhookRequired: '请输入机器人 Webhook 地址',
      updateFailed: '更新失败：',
      createFailed: '新建失败：',
      testContentRequired: '请输入测试消息内容',
      testSent: '测试消息已发送至「{name}」',
      testFailed: '「{name}」测试失败：{message}',
      toggleFailed: '切换状态失败：',
      deleteFailed: '删除失败：',
      confirmDelete: '确定要删除媒介"{name}"吗？'
    },
    signed: '（已加签）',
    defaultTestMessage: '【AIOPS】通知媒介「测试」消息，时间：{time}'
  }
}
