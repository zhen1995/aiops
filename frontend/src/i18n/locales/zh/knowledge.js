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
    uploader: '上传人',
    uploadTime: '上传时间',
    actions: '操作',
    reindex: '重新索引',
    delete: '删除',
    emptyNoMatch: '没有匹配关键字的文档',
    emptyNoDocs: '暂无文档，点击右上角“上传文档”按钮添加'
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
  }
}
