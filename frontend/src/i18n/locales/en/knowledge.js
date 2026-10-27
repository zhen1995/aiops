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
    failReason: 'Failure Reason',
    uploader: 'Uploader',
    uploadTime: 'Uploaded At',
    actions: 'Actions',
    reindex: 'Reindex',
    delete: 'Delete',
    viewReason: 'View Reason',
    emptyNoMatch: 'No documents match the keyword',
    emptyNoDocs: 'No documents yet. Click the "Upload Document" button in the top right to add one.'
  },
  failDialog: {
    title: 'Indexing Failure Reason',
    empty: 'No failure reason recorded',
    close: 'Close'
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
  },
  wiki: {
    tabs: {
      wiki: 'Wiki',
      docs: 'Documents',
      review: 'Review'
    },
    status: {
      draft: 'Draft',
      published: 'Published',
      archived: 'Archived'
    },
    category: {
      title: 'Categories',
      add: 'New Category',
      name: 'Category Name',
      parent: 'Parent Category',
      none: 'None (top-level)',
      empty: 'No categories yet. Click + at the top right to create one.',
      confirmDelete: 'Delete category "{name}"? Child categories move up and entries become uncategorized.',
      edit: 'Edit Category'
    },
    entry: {
      add: 'New Entry',
      edit: 'Edit Entry',
      detail: 'Entry Detail',
      colTitle: 'Title',
      colStatus: 'Status',
      colCategory: 'Category',
      colSummary: 'Summary',
      colUpdatedAt: 'Updated At',
      colActions: 'Actions',
      titleLabel: 'Entry Title',
      titlePlaceholder: 'e.g. MySQL replication lag handling',
      categoryLabel: 'Category',
      uncategorized: 'Uncategorized',
      tagsLabel: 'Tags (comma separated)',
      tagsPlaceholder: 'e.g. MySQL, replication, troubleshooting',
      summaryLabel: 'Summary',
      summaryPlaceholder: 'One-sentence summary (within 50 chars)',
      contentLabel: 'Content (Markdown)',
      contentPlaceholder: 'Markdown supported',
      source: 'Source',
      publish: 'Publish',
      archive: 'Archive',
      withdraw: 'Withdraw to Draft',
      confirmDelete: 'Delete entry "{title}"?',
      empty: 'No entries yet',
      searchPlaceholder: 'Search title/summary',
      generate: 'Generate Entries',
      generating: 'Generating...',
      generateDone: '{count} draft entries generated. Check the Review tab.'
    },
    error: {
      titleRequired: 'Please enter the entry title',
      contentRequired: 'Please enter the entry content',
      categoryNameRequired: 'Please enter the category name'
    },
    msg: {
      loadCategoriesFailed: 'Failed to load categories: {msg}',
      loadEntriesFailed: 'Failed to load entries: {msg}',
      saveEntryFailed: 'Failed to save entry: {msg}',
      deleteEntryFailed: 'Failed to delete entry: {msg}',
      statusFailed: 'Failed to change status: {msg}',
      generateFailed: 'Failed to generate entries: {msg}',
      saveCategoryFailed: 'Failed to save category: {msg}',
      deleteCategoryFailed: 'Failed to delete category: {msg}'
    }
  },
}
