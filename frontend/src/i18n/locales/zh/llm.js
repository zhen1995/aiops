export default {
  header: {
    title: 'LLM 管理',
    desc: '管理大语言模型接入配置，支持多厂商、多模型切换',
    add: '新增模型'
  },
  list: {
    title: '模型列表',
    sub: '已接入的 LLM 服务与运行状态',
    refresh: '刷新',
    loading: '加载中...'
  },
  table: {
    name: '名称',
    modelType: '模型类型',
    supplierCategory: '提供商类型',
    model: '模型',
    endpoint: '接入端点',
    apiKey: 'API Key',
    status: '状态',
    actions: '操作',
    default: '默认',
    statusRunning: '运行中',
    statusStopped: '已停用',
    edit: '编辑',
    disable: '停用',
    enable: '启用',
    delete: '删除',
    empty: '暂无数据，请点击"新增模型"添加配置'
  },
  modal: {
    editTitle: '编辑 LLM 配置',
    createTitle: '新建 LLM 配置',
    name: '名称',
    namePlaceholder: '请输入 LLM 配置的名称',
    enabled: '启用',
    isDefault: '默认',
    description: '描述',
    descriptionPlaceholder: '请输入 LLM 配置的描述信息',
    modelType: '模型类型',
    typeChat: '对话模型',
    typeEmbedding: '向量化模型',
    supplierCategory: '提供商类型',
    supplierPlaceholder: '请选择提供商类型',
    model: '模型',
    modelPlaceholder: '请输入模型名称，例如：gpt-4o',
    apiUrlPlaceholder: '请输入接口地址，例如：https://api.openai.com/v1',
    apiKeyPlaceholder: '请输入 API Key',
    show: '显示',
    hide: '隐藏',
    cancel: '取消',
    save: '保存',
    saving: '保存中...'
  },
  supplier: {
    openai: 'OpenAI 兼容',
    claude: 'Claude',
    gemini: 'Gemini',
    qwen: '阿里云千问',
    deepseek: 'DeepSeek',
    kimi: 'Kimi',
    other: '其他'
  },
  modelTypeOption: {
    chat: '对话模型',
    embedding: '向量化模型'
  },
  validation: {
    nameRequired: '请输入名称',
    modelTypeRequired: '请选择模型类型',
    supplierRequired: '请选择提供商类型',
    modelRequired: '请输入模型名称',
    apiUrlRequired: '请输入 API URL',
    apiKeyRequired: '请输入 API Key'
  },
  message: {
    loadFailed: '加载 LLM 配置列表失败：{msg}',
    updateFailed: '更新失败：{msg}',
    createFailed: '新建失败：{msg}',
    toggleFailed: '切换状态失败：{msg}',
    deleteConfirm: '确定要删除配置"{name}"吗？',
    deleteFailed: '删除失败：{msg}'
  },
  skill: {
    title: 'Skill 管理',
    desc: '管理 AI 助手可调用的能力（Skill），包括告警解读、根因分析、巡检报告等',
    add: '新增 Skill',
    listTitle: 'Skill 列表',
    listSub: '已注册的智能运维能力',
    colName: 'Skill 名称',
    colDesc: '描述',
    colVersion: '版本',
    colTrigger: '触发方式',
    colStatus: '状态',
    colActions: '操作',
    edit: '编辑',
    disable: '停用',
    enable: '启用'
  }
}
