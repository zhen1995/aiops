export default {
  header: {
    title: '数据源接入',
    desc: '管理 Prometheus / ElasticSearch / Pyroscope 等数据源接入配置',
    add: '新增数据源'
  },
  list: {
    title: '数据源列表',
    sub: '已配置的数据源及运行状态',
    refresh: '刷新',
    loading: '加载中...',
    empty: '暂无数据，请点击"新增数据源"添加配置',
    columns: {
      name: '名称',
      type: '类型',
      url: '接入地址',
      timeout: '超时(ms)',
      status: '状态',
      actions: '操作'
    }
  },
  status: {
    enabled: '已启用',
    disabled: '已禁用'
  },
  actions: {
    edit: '编辑',
    enable: '启用',
    disable: '禁用',
    delete: '删除'
  },
  modal: {
    createTitle: '新增数据源',
    editTitle: '编辑数据源',
    cancel: '取消',
    save: '保存',
    saving: '保存中...',
    form: {
      name: '名称',
      namePlaceholder: '请输入数据源名称',
      type: '类型',
      typePlaceholder: '请选择类型',
      url: '接入地址',
      urlPlaceholder: '请输入 HTTP 地址，例如：http://localhost:9090',
      timeout: '超时(ms)',
      timeoutPlaceholder: '默认 5000',
      skipSsl: '跳过 SSL 验证',
      username: '用户名',
      password: '密码',
      remark: '备注',
      enabled: '启用',
      optional: '选填',
      show: '显示',
      hide: '隐藏'
    }
  },
  messages: {
    loadFailed: '加载数据源列表失败：{message}',
    nameRequired: '请输入名称',
    typeRequired: '请选择类型',
    urlRequired: '请输入接入地址',
    createFailed: '新建失败：{message}',
    updateFailed: '更新失败：{message}',
    toggleFailed: '切换状态失败：{message}',
    deleteConfirm: '确定要删除数据源"{name}"吗？',
    deleteFailed: '删除失败：{message}'
  }
}
