export default {
  header: {
    title: 'System Config',
    desc: 'Configure the Qdrant vector store address used by the operations knowledge base'
  },
  card: {
    title: 'Knowledge Base Vector Store',
    sub: 'Qdrant vector store service address configuration'
  },
  form: {
    qdrantUrl: 'Qdrant URL',
    qdrantUrlPlaceholder: 'e.g.: http://localhost:6333'
  },
  actions: {
    save: 'Save',
    saving: 'Saving...',
    testConnection: 'Test Connection',
    testing: 'Testing...'
  },
  tips: {
    title: 'Notes',
    item1: 'This address is used by the knowledge base vector store and takes effect immediately after saving.',
    item2: 'After changing the address, documents must be re-uploaded and the vector index rebuilt; existing vectors are not migrated automatically.'
  },
  result: {
    success: 'Connection succeeded',
    failed: 'Connection failed'
  },
  messages: {
    loadFailed: 'Failed to load system config: {message}',
    saveSuccess: 'Saved successfully',
    saveFailed: 'Failed to save: {message}',
    testFailed: 'Test failed: {message}',
    urlRequired: 'Please enter the Qdrant URL'
  }
}
