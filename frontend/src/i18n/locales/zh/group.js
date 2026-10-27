export default {
  title: '业务分组',
  edit: '编辑',
  delete: '删除',
  refresh: '刷新',
  loading: '加载中...',
  panel: {
    allRules: '全部规则',
    searchPlaceholder: '搜索分组名称',
    noMatch: '无匹配分组',
    empty: '暂无分组，点击右上角 + 新建',
    filtering: '分组：{name}'
  },
  modal: {
    editTitle: '编辑业务分组',
    createTitle: '新建业务分组'
  },
  form: {
    nameLabel: '分组名称',
    namePlaceholder: '例如：基础资源组',
    parentLabel: '父分组',
    parentNone: '无（顶级分组）',
    cancel: '取消',
    save: '保存',
    saving: '保存中...'
  },
  rename: {
    saveFailed: '重命名失败：{msg}'
  },
  error: {
    nameRequired: '请输入分组名称'
  },
  msg: {
    loadFailed: '加载业务分组失败：{msg}',
    updateFailed: '更新失败：{msg}',
    createFailed: '新建失败：{msg}',
    confirmDelete: '确定要删除业务分组"{name}"吗？该分组下的告警规则将变为未分组，子分组将上移到其上级分组。',
    deleteFailed: '删除失败：{msg}'
  }
}
