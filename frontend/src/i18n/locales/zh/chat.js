export default {
  session: {
    newButton: '+ 新会话',
    noMessage: '暂无消息',
    defaultTitle: '新会话'
  },
  header: {
    title: 'AI 运维助手',
    subtitle: '基于大模型的智能运维问答、告警解读与根因分析',
    currentModel: '当前模型',
    defaultModel: '默认模型',
    noDefaultModel: '未配置默认模型'
  },
  message: {
    userAvatar: '我',
    assistantName: 'AIOPS 助手',
    userName: '运维管理员',
    thinking: '正在思考中',
    copy: '复制',
    copied: '已复制',
    error: '\n[错误：{err}]',
    aborted: '_[已中止]_'
  },
  toolbar: {
    p0Alerts: '今日 P0 告警',
    rootCause: '根因分析',
    inspectionSummary: '巡检摘要'
  },
  input: {
    placeholder: '输入问题，例如：今天系统状态如何？',
    send: '发送',
    abort: '中止'
  },
  rca: {
    invalidParam: '根因分析参数无效',
    statusRecovered: '已恢复',
    statusFiring: '告警中',
    severityP1: 'P1-紧急',
    severityP2: 'P2-警告',
    severityP3: 'P3-提醒',
    prompt: `请对以下告警事件进行根因分析：
- 规则名称：{ruleName}
- 告警对象：{targetIdent}
- 触发时间：{time}
- 级别：{severity}
- 状态：{status}
- 标签：{tags}
- 触发值：{triggerValue}

请查询相关 Prometheus 指标和 Elasticsearch 日志，排查宕机原因，并给出根因、证据链和修复建议。`
  }
}
