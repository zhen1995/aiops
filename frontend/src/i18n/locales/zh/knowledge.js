export default {
  header: {
    title: '运维知识库',
    desc: '管理运维文档、Runbook、RCA 报告等知识库文件，支持 Word、PDF、Markdown、TXT 上传与维护',
    upload: '上传文档'
  },
  stats: {
    total: '知识库文档',
    indexed: '已索引',
    pending: '待处理',
    unit: '个'
  },
  list: {
    title: '文档列表',
    sub: '已上传的运维知识库文件，向量化后即可被 AI 助手检索引用'
  },
  table: {
    fileName: '文件名',
    type: '类型',
    size: '大小',
    status: '状态',
    failReason: '失败原因',
    uploader: '上传人',
    uploadTime: '上传时间',
    actions: '操作',
    reindex: '重新索引',
    delete: '删除',
    viewReason: '查看原因',
    emptyNoMatch: '没有匹配关键字的文档',
    emptyNoDocs: '暂无文档，点击右上角“上传文档”按钮添加'
  },
  failDialog: {
    title: '索引失败原因',
    empty: '未记录失败原因',
    close: '关闭'
  },
  status: {
    indexed: '已索引',
    pending: '待索引',
    indexing: '索引中',
    failed: '索引失败'
  },
  alert: {
    partialUploadFailed: '部分文件上传失败：',
    uploadFailed: '上传失败',
    reindexFailed: '重新索引失败',
    deleteConfirm: '确定删除该文档吗？删除后不可恢复。',
    deleteFailed: '删除失败'
  },
  wiki: {
    tabs: {
      wiki: 'Wiki 浏览',
      docs: '文档源',
      review: '待审核'
    },
    status: {
      draft: '草稿',
      published: '已发布',
      archived: '已归档'
    },
    category: {
      title: '知识分类',
      add: '新建分类',
      name: '分类名称',
      parent: '父分类',
      none: '无（顶级分类）',
      empty: '暂无分类，点击右上角 + 新建',
      confirmDelete: '确定删除分类"{name}"吗？子分类将上移，其下词条变为未分类。',
      edit: '编辑分类'
    },
    entry: {
      add: '新建词条',
      edit: '编辑词条',
      detail: '词条详情',
      colTitle: '标题',
      colStatus: '状态',
      colCategory: '分类',
      colSummary: '摘要',
      colUpdatedAt: '更新时间',
      colActions: '操作',
      titleLabel: '词条标题',
      titlePlaceholder: '例如：MySQL 主从延迟处理',
      categoryLabel: '所属分类',
      uncategorized: '未分类',
      tagsLabel: '标签（逗号分隔）',
      tagsPlaceholder: '例如：MySQL, 主从复制, 故障处理',
      summaryLabel: '摘要',
      summaryPlaceholder: '一句话概括词条内容（50 字以内）',
      contentLabel: '正文（Markdown）',
      contentPlaceholder: '支持 Markdown 语法',
      source: '来源',
      publish: '发布',
      archive: '归档',
      withdraw: '退回草稿',
      confirmDelete: '确定删除词条"{title}"吗？',
      empty: '暂无词条',
      searchPlaceholder: '搜索标题/摘要',
      generate: '生成词条',
      generating: '生成中...',
      generateDone: '已生成 {count} 个词条草稿，请到「待审核」页签查看'
    },
    error: {
      titleRequired: '请输入词条标题',
      contentRequired: '请输入词条正文',
      categoryNameRequired: '请输入分类名称'
    },
    msg: {
      loadCategoriesFailed: '加载分类失败：{msg}',
      loadEntriesFailed: '加载词条失败：{msg}',
      saveEntryFailed: '保存词条失败：{msg}',
      deleteEntryFailed: '删除词条失败：{msg}',
      statusFailed: '状态变更失败：{msg}',
      generateFailed: '生成词条失败：{msg}',
      saveCategoryFailed: '保存分类失败：{msg}',
      deleteCategoryFailed: '删除分类失败：{msg}'
    }
  },
}
