export default {
  header: {
    title: 'Operations Knowledge Base',
    desc: 'Manage operations documents, runbooks, RCA reports and other knowledge base files. Supports uploading and maintaining Word, PDF, Markdown and TXT files.',
    upload: 'Upload Document'
  },
  stats: {
    total: 'Knowledge Base Documents',
    indexed: 'Indexed',
    pending: 'Pending',
    unit: 'docs'
  },
  list: {
    title: 'Documents',
    sub: 'Uploaded knowledge base files. Once vectorized, they can be retrieved and cited by the AI assistant.'
  },
  table: {
    fileName: 'File Name',
    type: 'Type',
    size: 'Size',
    status: 'Status',
    uploader: 'Uploader',
    uploadTime: 'Uploaded At',
    actions: 'Actions',
    reindex: 'Reindex',
    delete: 'Delete',
    emptyNoMatch: 'No documents match the keyword',
    emptyNoDocs: 'No documents yet. Click the "Upload Document" button in the top right to add one.'
  },
  status: {
    indexed: 'Indexed',
    pending: 'Pending',
    indexing: 'Indexing',
    failed: 'Indexing Failed'
  },
  alert: {
    partialUploadFailed: 'Some files failed to upload:',
    uploadFailed: 'Upload failed',
    reindexFailed: 'Reindex failed',
    deleteConfirm: 'Are you sure you want to delete this document? This cannot be undone.',
    deleteFailed: 'Delete failed'
  }
}
