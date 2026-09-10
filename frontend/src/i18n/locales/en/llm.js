export default {
  header: {
    title: 'LLM Management',
    desc: 'Manage LLM integration configurations, supporting multiple vendors and models',
    add: 'Add Model'
  },
  list: {
    title: 'Model List',
    sub: 'Connected LLM services and their running status',
    refresh: 'Refresh',
    loading: 'Loading...'
  },
  table: {
    name: 'Name',
    modelType: 'Model Type',
    supplierCategory: 'Provider Type',
    model: 'Model',
    endpoint: 'Endpoint',
    apiKey: 'API Key',
    status: 'Status',
    actions: 'Actions',
    default: 'Default',
    statusRunning: 'Running',
    statusStopped: 'Disabled',
    edit: 'Edit',
    disable: 'Disable',
    enable: 'Enable',
    delete: 'Delete',
    empty: 'No data yet. Click "Add Model" to create a configuration'
  },
  modal: {
    editTitle: 'Edit LLM Configuration',
    createTitle: 'New LLM Configuration',
    name: 'Name',
    namePlaceholder: 'Enter a name for the LLM configuration',
    enabled: 'Enabled',
    isDefault: 'Default',
    description: 'Description',
    descriptionPlaceholder: 'Enter a description for the LLM configuration',
    modelType: 'Model Type',
    typeChat: 'Chat Model',
    typeEmbedding: 'Embedding Model',
    supplierCategory: 'Provider Type',
    supplierPlaceholder: 'Select a provider type',
    model: 'Model',
    modelPlaceholder: 'Enter the model name, e.g.: gpt-4o',
    apiUrlPlaceholder: 'Enter the API endpoint, e.g.: https://api.openai.com/v1',
    apiKeyPlaceholder: 'Enter the API Key',
    show: 'Show',
    hide: 'Hide',
    cancel: 'Cancel',
    save: 'Save',
    saving: 'Saving...'
  },
  supplier: {
    openai: 'OpenAI Compatible',
    claude: 'Claude',
    gemini: 'Gemini',
    qwen: 'Alibaba Cloud Qwen',
    deepseek: 'DeepSeek',
    kimi: 'Kimi',
    other: 'Other'
  },
  modelTypeOption: {
    chat: 'Chat Model',
    embedding: 'Embedding Model'
  },
  validation: {
    nameRequired: 'Please enter a name',
    modelTypeRequired: 'Please select a model type',
    supplierRequired: 'Please select a provider type',
    modelRequired: 'Please enter a model name',
    apiUrlRequired: 'Please enter the API URL',
    apiKeyRequired: 'Please enter the API Key'
  },
  message: {
    loadFailed: 'Failed to load LLM configurations: {msg}',
    updateFailed: 'Update failed: {msg}',
    createFailed: 'Create failed: {msg}',
    toggleFailed: 'Failed to toggle status: {msg}',
    deleteConfirm: 'Are you sure you want to delete the configuration "{name}"?',
    deleteFailed: 'Delete failed: {msg}'
  },
  skill: {
    title: 'Skill Management',
    desc: 'Manage capabilities (Skills) available to the AI assistant, including alert interpretation, root cause analysis, inspection reports, and more',
    add: 'Add Skill',
    listTitle: 'Skill List',
    listSub: 'Registered intelligent operations capabilities',
    colName: 'Skill Name',
    colDesc: 'Description',
    colVersion: 'Version',
    colTrigger: 'Trigger',
    colStatus: 'Status',
    colActions: 'Actions',
    edit: 'Edit',
    disable: 'Disable',
    enable: 'Enable'
  }
}
