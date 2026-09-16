export default {
  header: {
    title: 'System Config',
    desc: 'Configure global settings such as the knowledge base Qdrant vector store and the frontend base URL'
  },
  card: {
    title: 'Knowledge Base Vector Store',
    sub: 'Qdrant vector store service address configuration'
  },
  frontend: {
    title: 'Frontend Base URL',
    sub: 'Used for the "full report" link in inspection notifications and the $.domain variable in alert notify templates',
    urlLabel: 'Frontend URL',
    urlPlaceholder: 'e.g.: http://localhost:5173'
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
    item1: 'The Qdrant address is used by the knowledge base vector store and takes effect immediately after saving.',
    item2: 'After changing the Qdrant address, documents must be re-uploaded and the vector index rebuilt; existing vectors are not migrated automatically.',
    item3: 'The frontend base URL takes effect immediately after saving; both the inspection report link and the $.domain variable in alert notify templates use it.',
    item4: 'In production, set this to a publicly accessible frontend address.'
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
    urlRequired: 'Please enter the Qdrant URL',
    frontendUrlRequired: 'Please enter the frontend base URL',
    frontendUrlInvalid: 'The frontend base URL must start with http:// or https://'
  }
}
