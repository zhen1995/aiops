export default {
  title: 'Business Groups',
  desc: 'Create and manage business groups to organize alert rules',
  addGroup: 'New Group',
  listTitle: 'Groups',
  listSub: 'Business groups and the number of associated alert rules',
  refresh: 'Refresh',
  loading: 'Loading...',
  colName: 'Group Name',
  colRuleCount: 'Alert Rules',
  colCreatedAt: 'Created At',
  colActions: 'Actions',
  edit: 'Edit',
  delete: 'Delete',
  empty: 'No business groups yet. Click "New Group" to create one.',
  modal: {
    editTitle: 'Edit Business Group',
    createTitle: 'New Business Group'
  },
  form: {
    nameLabel: 'Group Name',
    namePlaceholder: 'e.g. Infrastructure',
    cancel: 'Cancel',
    save: 'Save',
    saving: 'Saving...'
  },
  error: {
    nameRequired: 'Please enter a group name'
  },
  msg: {
    loadFailed: 'Failed to load business groups: {msg}',
    updateFailed: 'Update failed: {msg}',
    createFailed: 'Create failed: {msg}',
    confirmDelete: 'Delete business group "{name}"? Alert rules in this group will become ungrouped.',
    deleteFailed: 'Delete failed: {msg}'
  }
}
