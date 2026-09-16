export default {
  header: {
    title: '系统配置',
    desc: '配置运维知识库 Qdrant 向量存储与前端访问地址等全局配置'
  },
  card: {
    title: '知识库向量存储',
    sub: 'Qdrant 向量库服务地址配置'
  },
  frontend: {
    title: '前端访问地址',
    sub: '用于巡检报告通知中的"完整报告"链接与告警通知模板的 $.domain 变量',
    urlLabel: '前端地址',
    urlPlaceholder: '例如：http://localhost:5173'
  },
  form: {
    qdrantUrl: 'Qdrant 地址',
    qdrantUrlPlaceholder: '例如：http://localhost:6333'
  },
  actions: {
    save: '保存',
    saving: '保存中...',
    testConnection: '测试连接',
    testing: '测试中...'
  },
  tips: {
    title: '配置说明',
    item1: 'Qdrant 地址为运维知识库使用的向量存储服务地址，保存后即时生效。',
    item2: '更换 Qdrant 地址后，需重新上传文档并重建向量索引，旧向量数据不会自动迁移。',
    item3: '前端访问地址保存后即时生效，巡检报告通知中的"完整报告"链接与告警通知模板的 $.domain 变量均使用该地址。',
    item4: '生产环境部署后，请将该地址配置为可公开访问的前端地址。'
  },
  result: {
    success: '连接成功',
    failed: '连接失败'
  },
  messages: {
    loadFailed: '加载系统配置失败：{message}',
    saveSuccess: '保存成功',
    saveFailed: '保存失败：{message}',
    testFailed: '测试失败：{message}',
    urlRequired: '请输入 Qdrant 地址',
    frontendUrlRequired: '请输入前端访问地址',
    frontendUrlInvalid: '前端访问地址必须以 http:// 或 https:// 开头'
  }
}
