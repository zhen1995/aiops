export default {
  header: {
    title: 'Data Sources',
    desc: 'Manage data source connections such as Prometheus / ElasticSearch / Pyroscope',
    add: 'Add Data Source'
  },
  list: {
    title: 'Data Sources',
    sub: 'Configured data sources and their status',
    refresh: 'Refresh',
    loading: 'Loading...',
    empty: 'No data yet. Click "Add Data Source" to add a configuration',
    columns: {
      name: 'Name',
      type: 'Type',
      url: 'URL',
      timeout: 'Timeout (ms)',
      status: 'Status',
      actions: 'Actions'
    }
  },
  status: {
    enabled: 'Enabled',
    disabled: 'Disabled'
  },
  actions: {
    edit: 'Edit',
    enable: 'Enable',
    disable: 'Disable',
    delete: 'Delete'
  },
  modal: {
    createTitle: 'Add Data Source',
    editTitle: 'Edit Data Source',
    cancel: 'Cancel',
    save: 'Save',
    saving: 'Saving...',
    form: {
      name: 'Name',
      namePlaceholder: 'Enter data source name',
      type: 'Type',
      typePlaceholder: 'Select a type',
      url: 'URL',
      urlPlaceholder: 'Enter the HTTP address, e.g.: http://localhost:9090',
      timeout: 'Timeout (ms)',
      timeoutPlaceholder: 'Default 5000',
      skipSsl: 'Skip SSL Verification',
      username: 'Username',
      password: 'Password',
      remark: 'Remark',
      enabled: 'Enabled',
      optional: 'Optional',
      show: 'Show',
      hide: 'Hide'
    }
  },
  messages: {
    loadFailed: 'Failed to load data sources: {message}',
    nameRequired: 'Please enter a name',
    typeRequired: 'Please select a type',
    urlRequired: 'Please enter the URL',
    createFailed: 'Failed to create: {message}',
    updateFailed: 'Failed to update: {message}',
    toggleFailed: 'Failed to toggle status: {message}',
    deleteConfirm: 'Are you sure you want to delete the data source "{name}"?',
    deleteFailed: 'Failed to delete: {message}'
  }
}
