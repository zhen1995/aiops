export default {
  task: {
    title: '巡检任务',
    desc: '配置定时巡检任务，服务会按 Cron 表达式自动执行并生成报告',
    newBtn: '+ 新建任务',
    table: {
      enabled: '启用',
      name: '任务名称',
      cron: 'Cron 表达式',
      prompt: '任务提示词',
      lastRun: '上次执行',
      nextRun: '预计下次',
      actions: '操作',
      edit: '编辑',
      runNow: '立即执行',
      remove: '删除',
      empty: '暂无巡检任务，点击右上角新建'
    },
    modal: {
      editTitle: '编辑巡检任务',
      createTitle: '新建巡检任务',
      name: '任务名称',
      namePlaceholder: '例如：每日系统巡检',
      cronExpr: 'Cron 表达式',
      cronHint: '支持 5 段（分 时 日 月 周）或 6 段（秒 分 时 日 月 周）',
      cronPlaceholder: '点击右侧「配置」生成表达式',
      collapse: '收起',
      expand: '配置',
      preview: '预览',
      nextRuns: '预计下次执行时间：',
      promptLabel: '任务提示词',
      promptPlaceholder: '描述你希望巡检任务分析的内容，例如：检查过去 24 小时内的异常告警、错误日志和服务健康状态...',
      media: '通知媒介',
      mediaHint: '报告生成后自动推送，可多选',
      mediaEmpty: '暂无启用的通知媒介，请先在「通知媒介」菜单中创建',
      dingtalk: '钉钉',
      enabledSwitch: '启用任务',
      cancel: '取消',
      saving: '保存中...',
      save: '保存'
    },
    messages: {
      loadFailed: '加载任务失败: ',
      nameRequired: '请填写任务名称',
      cronRequired: '请填写 Cron 表达式',
      promptRequired: '请填写任务提示词',
      saveFailed: '保存失败: ',
      toggleFailed: '操作失败: ',
      runNowConfirm: '立即执行巡检任务 "{name}"？（调用 LLM 可能需要几十秒）',
      runSucceeded: '执行完成，已生成巡检报告',
      runFailed: '执行失败: ',
      deleteConfirm: '确认删除巡检任务 "{name}"？',
      deleteFailed: '删除失败: '
    }
  },
  report: {
    title: '巡检报告',
    desc: '查看定时巡检任务生成的历史报告',
    filters: {
      task: '来源任务',
      allTasks: '全部任务',
      status: '状态',
      allStatus: '全部状态',
      query: '查询',
      reset: '重置'
    },
    table: {
      score: '评分',
      reportTitle: '报告标题',
      task: '来源任务',
      generatedAt: '生成时间',
      summary: '摘要',
      status: '状态',
      actions: '操作',
      view: '查看',
      remove: '删除',
      emptyFiltered: '没有匹配当前查询条件的巡检报告',
      empty: '暂无巡检报告，请先配置巡检任务'
    },
    status: {
      completed: '已完成',
      generating: '生成中',
      failed: '失败'
    },
    messages: {
      loadFailed: '加载报告失败: ',
      deleteConfirm: '确认删除巡检报告 "{title}"？',
      deleteFailed: '删除失败: '
    }
  },
  detail: {
    loading: '加载中...',
    back: '返回列表',
    overallScore: '综合评分',
    generateFailed: '生成失败',
    loadFailed: '加载报告失败: ',
    desc: '来源：{task} · 生成时间：{date} · 综合评分：{score} 分'
  },
  cron: {
    showSeconds: '显示秒字段（表达式为 6 段）',
    everyLabel: '每 {label} ({wildcard})',
    interval: '每隔',
    range: '范围',
    specific: '指定 {unit}',
    multiple: '（多选）',
    dayNote: '配置日后，周字段将自动设为',
    weekNote: '配置周后，日字段将自动设为',
    cancel: '取消',
    confirm: '确定',
    fields: {
      second: '秒',
      minute: '分钟',
      hour: '小时',
      day: '日',
      month: '月',
      week: '周'
    },
    units: {
      second: '秒',
      minute: '分钟',
      hour: '小时',
      day: '日',
      month: '月',
      week: '周'
    },
    weekDays: {
      1: '周一',
      2: '周二',
      3: '周三',
      4: '周四',
      5: '周五',
      6: '周六',
      7: '周日'
    },
    months: {
      1: '1月',
      2: '2月',
      3: '3月',
      4: '4月',
      5: '5月',
      6: '6月',
      7: '7月',
      8: '8月',
      9: '9月',
      10: '10月',
      11: '11月',
      12: '12月'
    },
    summary: {
      atMinute: '在 {m} 分',
      weekRange: '每周 {from}~{to}',
      weekList: '每周 {days}',
      weekSingle: '每周 {day}',
      monthDay: '每月 {d} 日',
      daily: '每天',
      atHour: '{h} 点',
      everyNHours: '每 {n} 小时',
      daySeparator: '、'
    }
  }
}
