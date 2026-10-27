export default {
  title: 'Business Groups',
  edit: 'Edit',
  delete: 'Delete',
  refresh: 'Refresh',
  loading: 'Loading...',
  panel: {
    allRules: 'All Rules',
    searchPlaceholder: 'Search group name',
    noMatch: 'No matching groups',
    empty: 'No groups yet. Click + at the top right to create one.',
    filtering: 'Group: {name}'
  },
  modal: {
    editTitle: 'Edit Business Group',
    createTitle: 'New Business Group'
  },
  form: {
    nameLabel: 'Group Name',
    namePlaceholder: 'e.g. Infrastructure',
    parentLabel: 'Parent Group',
    parentNone: 'None (top-level)',
    cancel: 'Cancel',
    save: 'Save',
    saving: 'Saving...'
  },
  rename: {
    saveFailed: 'Rename failed: {msg}'
  },
  error: {
    nameRequired: 'Please enter a group name'
  },
  msg: {
    loadFailed: 'Failed to load business groups: {msg}',
    updateFailed: 'Update failed: {msg}',
    createFailed: 'Create failed: {msg}',
    confirmDelete: 'Delete business group "{name}"? Alert rules in this group will become ungrouped, and child groups will move up to its parent.',
    deleteFailed: 'Delete failed: {msg}'
  }
}
