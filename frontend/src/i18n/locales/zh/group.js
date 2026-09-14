export default {
  title: '业务分组',
  desc: '创建和维护业务分组，用于对告警规则进行分组管理',
  addGroup: '新建分组',
  listTitle: '分组列表',
  listSub: '已创建的业务分组及关联的告警规则数量',
  refresh: '刷新',
  loading: '加载中...',
  colName: '分组名称',
  colRuleCount: '告警规则数',
  colCreatedAt: '创建时间',
  colActions: '操作',
  edit: '编辑',
  delete: '删除',
  empty: '暂无业务分组，请点击"新建分组"创建',
  modal: {
    editTitle: '编辑业务分组',
    createTitle: '新建业务分组'
  },
  form: {
    nameLabel: '分组名称',
    namePlaceholder: '例如：基础资源组',
    cancel: '取消',
    save: '保存',
    saving: '保存中...'
  },
  error: {
    nameRequired: '请输入分组名称'
  },
  msg: {
    loadFailed: '加载业务分组失败：{msg}',
    updateFailed: '更新失败：{msg}',
    createFailed: '新建失败：{msg}',
    confirmDelete: '确定要删除业务分组"{name}"吗？该分组下的告警规则将变为未分组。',
    deleteFailed: '删除失败：{msg}'
  }
}
