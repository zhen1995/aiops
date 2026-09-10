export default {
  severity: {
    p1: 'P1-Critical',
    p2: 'P2-Warning',
    p3: 'P3-Notice'
  },
  events: {
    title: 'Alert Events',
    desc: 'Active and historical alert events automatically detected by the system based on alert rules',
    tabs: {
      active: 'Active Alerts',
      history: 'History'
    },
    filters: {
      timeWindow: 'Time Window',
      last24h: 'Last 24 hours',
      last7d: 'Last 7 days',
      last30d: 'Last 30 days',
      severity: 'Severity',
      all: 'All',
      searchPlaceholder: 'Search rule name or alert target',
      query: 'Search',
      refresh: 'Refresh',
      loading: 'Loading...'
    },
    batch: {
      selectAll: 'Select all on this page',
      selected: '{count} selected',
      cancelSelection: 'Clear selection',
      batchDelete: 'Batch delete',
      deleting: 'Deleting...'
    },
    table: {
      ruleName: 'Rule Name',
      severity: 'Severity',
      target: 'Alert Target',
      triggerTime: 'Trigger Time',
      tags: 'Tags',
      triggerValue: 'Trigger Value',
      actions: 'Actions',
      rca: 'Root Cause',
      delete: 'Delete',
      empty: 'No alert events',
      loading: 'Loading...',
      total: 'Total {total}',
      prev: 'Prev',
      next: 'Next',
      page: 'Page {page}'
    },
    msg: {
      rcaFailed: 'Failed to trigger root cause analysis: {msg}',
      loadFailed: 'Failed to load alert events: {msg}',
      deleteFailed: 'Failed to delete alert event: {msg}',
      batchDeleteFailed: 'Batch delete failed: {msg}',
      confirmDelete: 'Delete alert event "{label}"?',
      confirmBatchDelete: 'Delete the {count} selected alert events?'
    }
  },
  denoise: {
    title: 'Alert Denoising',
    desc: 'Multi-level denoising strategy chain: duplicate and cascading alerts are compressed automatically, and only effective alerts reach on-call staff',
    loading: 'Loading...',
    retry: 'Retry',
    funnelTitle: 'Denoising Funnel',
    funnelSub: 'Raw alerts are compressed level by level through window aggregation and topology suppression, until only effective notifications reach on-call staff',
    funnelEmpty: 'No statistics yet',
    funnelTooltip: '{name}: {value}',
    policyTitle: 'Denoising Policies',
    policySub: 'Policies run in order and can be toggled at any time; changes take effect immediately',
    switchEnable: 'Click to enable',
    switchDisable: 'Click to disable',
    suppressedToday: 'Suppressed today',
    viewRecords: 'View notification records intercepted by this policy today',
    kpi: {
      rawTotal: 'Raw Alert Total',
      rawHint: 'Entered the denoising pipeline today',
      effective: 'Effective Alerts',
      effectiveHint: 'Actually reached on-call staff',
      compressionRate: 'Compression Rate',
      compressionHint: 'Overall denoising effect',
      suppressed: 'Suppressed Alerts',
      suppressedHint: 'Total intercepted by denoising policies'
    },
    effect: {
      windowAggregation: 'Reduces duplicate alerts',
      topologySuppression: 'Reduces cascading alerts'
    },
    msg: {
      loadFailed: 'Failed to load denoising data: {msg}',
      toggleFailed: 'Failed to toggle policy: {msg}'
    }
  },
  rules: {
    title: 'Alert Rules',
    desc: 'Create and maintain alert rules that trigger alerts based on PromQL expressions',
    addRule: 'Add Rule',
    listTitle: 'Rule List',
    listSub: 'Configured alert rules and their enable status',
    keyword: 'Keyword: {kw}',
    refresh: 'Refresh',
    loading: 'Loading...',
    colStatus: 'Status',
    colRuleName: 'Rule Name',
    colSeverity: 'Severity',
    colEvalInterval: 'Eval Interval',
    colDuration: 'Duration (s)',
    colNotifyRule: 'Notify Rule',
    colEnabled: 'Enable Status',
    colCreatedAt: 'Created At',
    colActions: 'Actions',
    everySeconds: 'Every {n} s',
    durationImmediate: '0 (immediate)',
    enabledOn: 'Enabled',
    enabledOff: 'Disabled',
    edit: 'Edit',
    enable: 'Enable',
    disable: 'Disable',
    delete: 'Delete',
    activeBadge: '{count} firing alerts',
    emptyNoMatch: 'No alert rules match the keyword',
    empty: 'No alert rules yet. Click "Add Rule" to create one',
    statusTitle: {
      disabled: 'Rule disabled',
      alerting: 'Alerting ({count} active alerts), click to view',
      noAlert: 'No alerts, click to view recent alerts'
    },
    eventsActive: '· {count} active alerts',
    eventsNone: '· No active alerts',
    eventsEmpty: 'No active alert events under this rule',
    triggerValue: 'Trigger value: ',
    triggerType: {
      all: 'Firing+Recovered',
      firing: 'Firing only',
      recovered: 'Recovered only'
    },
    modal: {
      editTitle: 'Edit Alert Rule',
      createTitle: 'Add Alert Rule'
    },
    form: {
      nameLabel: 'Rule Name',
      namePlaceholder: 'e.g. CPU usage above 90%',
      promqlPlaceholder: 'e.g. 100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 90',
      evalIntervalLabel: 'Eval Interval',
      durationLabel: 'Duration (s)',
      durationPlaceholder: 'e.g. 300',
      durationHint: '0 means the alert fires as soon as one query evaluation matches the condition',
      severityLabel: 'Severity',
      notifyRuleLabel: 'Notify Rule',
      notifyOptional: 'Optional; sends notifications automatically on alert trigger/recovery',
      noNotify: 'Do not notify',
      notifyConfigTitle: 'Notification Config',
      repeatLabel: 'Repeat Interval (minutes)',
      repeatPlaceholder: 'e.g. 60',
      repeatHint: 'If the alert remains unresolved, a reminder is repeated after the given minutes; set 0 to disable repetition',
      maxSendLabel: 'Max Send Count',
      maxSendPlaceholder: 'e.g. 3',
      maxSendHint: 'A value of 0 means no limit on the maximum send count',
      enabledLabel: 'Enable Status',
      cancel: 'Cancel',
      save: 'Save',
      saving: 'Saving...'
    },
    error: {
      nameRequired: 'Please enter a rule name',
      promqlRequired: 'Please enter a PromQL expression',
      evalRequired: 'Please select an eval interval',
      durationNegative: 'Duration cannot be negative'
    },
    msg: {
      loadFailed: 'Failed to load alert rules: {msg}',
      updateFailed: 'Update failed: {msg}',
      createFailed: 'Create failed: {msg}',
      toggleFailed: 'Failed to toggle status: {msg}',
      confirmDelete: 'Delete alert rule "{name}"?',
      deleteFailed: 'Delete failed: {msg}'
    }
  }
}
