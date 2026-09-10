export default {
  record: {
    pageTitle: 'Notification Records',
    pageDesc: 'Full trace of every alert notification: successful, failed, and denoise-intercepted deliveries',
    filters: {
      status: 'Status',
      strategy: 'Denoise Strategy',
      all: 'All',
      keywordPlaceholder: 'Search by rule name or alert target',
      startTime: 'Start Time',
      endTime: 'End Time',
      search: 'Search',
      reset: 'Reset'
    },
    status: {
      success: 'Sent',
      failed: 'Failed',
      intercepted: 'Intercepted',
      skipped: 'Skipped'
    },
    strategy: {
      windowAggregation: 'Time Window Aggregation',
      topologySuppression: 'Topology Suppression'
    },
    eventType: {
      alert: 'Alert',
      recovery: 'Recovered'
    },
    table: {
      time: 'Notified At',
      ruleName: 'Rule Name',
      eventType: 'Event Type',
      target: 'Alert Target',
      triggerValue: 'Trigger Value',
      status: 'Status',
      media: 'Media',
      actions: 'Actions',
      detail: 'Detail',
      loading: 'Loading...',
      empty: 'No notification records'
    },
    pagination: {
      total: 'Total {count}',
      perPage: '{count} / page',
      prev: 'Prev',
      next: 'Next',
      pageInfo: 'Page {page} / {pages}'
    },
    detail: {
      title: 'Notification Record Detail',
      basicInfo: 'Basic Info',
      notifyChain: 'Notification Chain',
      messageContent: 'Message Content',
      ruleName: 'Rule Name',
      severity: 'Severity',
      eventType: 'Event Type',
      target: 'Alert Target',
      triggerValue: 'Trigger Value',
      triggerTime: 'Trigger Time',
      tags: 'Tags',
      notifyTime: 'Notified At',
      notifyRule: 'Notify Rule',
      media: 'Media',
      mediaType: 'Media Type',
      template: 'Message Template',
      noContent: 'None (not sent)',
      close: 'Close'
    },
    error: {
      loadFailed: 'Failed to load notification records: ',
      loadDetailFailed: 'Failed to load notification record detail: '
    }
  },
  rule: {
    pageTitle: 'Notification Rules',
    pageDesc: 'Configure many-to-many associations between notification media and message templates, bindable to Nightingale alert rules',
    create: '+ New Notification Rule',
    modal: {
      editTitle: 'Edit Notification Rule',
      createTitle: 'New Notification Rule',
      name: 'Rule Name',
      namePlaceholder: 'e.g. Alert Firing - DingTalk',
      remark: 'Remark',
      remarkPlaceholder: 'Optional, describe the purpose of this rule',
      media: 'Notification Media',
      pickOne: 'Pick one',
      mediaEmpty: 'No notification media yet. Add one in the "Notification Media" menu',
      selectMedia: 'Please select a notification medium',
      template: 'Message Template',
      templateEmpty: 'No message templates yet. Add one in the "Message Templates" menu',
      selectTemplate: 'Please select a message template',
      enabled: 'Enabled',
      cancel: 'Cancel',
      save: 'Save',
      saving: 'Saving...'
    },
    list: {
      total: 'Total {count} rules',
      name: 'Rule Name',
      triggerScene: 'Trigger Scene',
      media: 'Notification Media',
      template: 'Message Template',
      status: 'Status',
      actions: 'Actions',
      edit: 'Edit',
      delete: 'Delete',
      empty: 'No notification rules',
      loading: 'Loading...'
    },
    mediaType: {
      dingtalk: 'DingTalk',
      webhook: 'Webhook',
      email: 'Email',
      wecom: 'WeCom'
    },
    trigger: {
      all: 'Firing + Recovered',
      firing: 'Firing Only',
      recovered: 'Recovered Only'
    },
    error: {
      nameRequired: 'Please enter a rule name',
      saveFailed: 'Save failed: ',
      deleteFailed: 'Delete failed: ',
      toggleFailed: 'Toggle failed: ',
      confirmDelete: 'Delete notification rule "{name}"?'
    }
  },
  template: {
    pageTitle: 'Message Templates',
    pageDesc: 'Define alert notification content templates using Go template syntax with alert event variables',
    create: '+ New Template',
    modal: {
      editTitle: 'Edit Message Template',
      createTitle: 'New Message Template',
      name: 'Template Name',
      namePlaceholder: 'e.g. DingTalk - Generic Alert Template',
      scene: 'Applicable Scene',
      sceneAll: 'Generic (Firing + Recovered)',
      sceneFiring: 'Alert Firing Only',
      sceneRecovered: 'Alert Recovered Only',
      mediaType: 'Media Type',
      mediaDingtalk: 'DingTalk',
      mediaWebhook: 'Webhook',
      mediaEmail: 'Email',
      mediaWecom: 'WeCom Work',
      description: 'Description',
      descriptionPlaceholder: 'Template description to help identify its purpose',
      contentLabel: 'Template Content (Go template)',
      contentPlaceholder: '{{$event.RuleName}} alert&#10;Trigger value: {{$event.TriggerValue}}&#10;Site: {{$.domain}}',
      preview: 'Live Preview',
      previewHint: '(Rendered result appears once the template is valid)',
      emptyPreview: 'No content',
      refTitle: 'Variable Reference',
      refHint: '(Click to insert into template)',
      tabVars: 'Alert Event Variables',
      tabFuncs: 'Builtin Functions',
      colVar: 'Variable',
      colFunc: 'Function',
      colDesc: 'Description',
      clickInsert: 'Click to insert: ',
      cancel: 'Cancel',
      save: 'Save Template',
      saving: 'Saving...'
    },
    list: {
      edit: 'Edit',
      delete: 'Delete',
      empty: 'No templates yet. Create one in the top right',
      loading: 'Loading...'
    },
    type: {
      all: 'Generic',
      firing: 'Firing',
      recovered: 'Recovered'
    },
    mediaLabel: {
      dingtalk: 'DingTalk',
      webhook: 'Webhook',
      email: 'Email',
      wecom: 'WeCom'
    },
    vars: {
      event: 'The whole alert event object; useful for debugging all fields',
      labels: 'Event labels map, equivalent to $event.TagsMap',
      value: 'Trigger value, equivalent to $event.TriggerValue',
      domain: 'Site address, used to build detail links',
      ruleName: 'Alert rule name',
      ruleNote: 'Alert rule remark/description',
      id: 'Unique alert event ID',
      severityLabel: 'Alert severity in Chinese (P1-critical / P2-warning / P3-notice)',
      triggerValue: 'Trigger value (result of the PromQL expression)',
      busiGroupName: 'Nightingale business group name',
      cluster: 'Alert cluster identifier',
      triggerTime: 'Trigger time',
      lastEvalTime: 'Most recent PromQL evaluation hit time',
      firstTrigger: 'First trigger time',
      isRecovered: 'Whether recovered (boolean)',
      tagsMap: 'Event labels map, supports lookup by key',
      tagsJson: 'Event labels as JSON string',
      annotations: 'Additional annotations field'
    },
    funcs: {
      timeformat: 'Format time; the second argument is a Go layout',
      timeformatCN: 'Simplified version, outputs yyyy-MM-dd HH:mm:ss',
      timestamp: 'Current time string (often used as "sent at")',
      now: 'Current Unix timestamp in seconds (int64), for duration calculation',
      sub: 'Subtraction, often used to compute alert duration in seconds',
      add: 'Addition',
      mul: 'Multiplication',
      humanizeDuration: 'Seconds → human readable (3m / 2h / 1.5d)',
      humanizeDurationIfc: 'Alias of humanizeDuration, compatible with Nightingale templates',
      durationHuman: 'Simplified version (legacy templates; deprecated, use humanizeDuration)'
    },
    error: {
      loadFailed: 'Load failed: ',
      nameContentRequired: 'Please enter template name and content',
      saveFailed: 'Save failed: ',
      deleteFailed: 'Delete failed: ',
      confirmDelete: 'Delete template "{name}"?'
    }
  },
  medium: {
    pageTitle: 'Notification Media',
    pageDesc: 'Manage notification channels such as DingTalk and Webhook callbacks',
    create: '+ New Medium',
    list: {
      title: 'Media List',
      subtitle: 'Alert notifications are delivered through the following media',
      refresh: 'Refresh',
      loading: 'Loading...',
      name: 'Media Name',
      type: 'Type',
      summary: 'Config Summary',
      status: 'Status',
      actions: 'Actions',
      running: 'Running',
      disabled: 'Disabled',
      edit: 'Edit',
      testing: 'Testing...',
      test: 'Test',
      disable: 'Disable',
      enable: 'Enable',
      delete: 'Delete',
      empty: 'No data yet. Click "New Medium" to add a configuration'
    },
    type: {
      dingtalk: 'DingTalk',
      webhook: 'Webhook Callback'
    },
    modal: {
      editTitle: 'Edit Medium',
      createTitle: 'New Medium',
      name: 'Media Name',
      namePlaceholder: 'e.g. SRE On-Call DingTalk Group',
      type: 'Media Type',
      selectType: 'Please select a type',
      webhookUrl: 'Callback URL',
      webhookUrlPlaceholder: 'Enter the callback URL, e.g. https://example.com/api/notify',
      method: 'HTTP Method',
      timeout: 'Timeout (ms)',
      timeoutPlaceholder: 'Default 5000',
      dingWebhook: 'Robot Webhook',
      secret: 'Signing Secret',
      secretPlaceholder: 'Optional, required when signing is enabled',
      show: 'Show',
      hide: 'Hide',
      remark: 'Remark',
      remarkPlaceholder: 'Optional',
      enabled: 'Enabled',
      cancel: 'Cancel',
      save: 'Save',
      saving: 'Saving...'
    },
    test: {
      title: 'Send Test Message',
      titleWithName: 'Send Test Message — {name}',
      content: 'Message Content',
      contentPlaceholder: 'Enter the test message content',
      sending: 'Sending...',
      send: 'Send Test'
    },
    error: {
      loadFailed: 'Failed to load media list: ',
      nameRequired: 'Please enter a media name',
      typeRequired: 'Please select a media type',
      webhookUrlRequired: 'Please enter the callback URL',
      dingWebhookRequired: 'Please enter the robot webhook URL',
      updateFailed: 'Update failed: ',
      createFailed: 'Create failed: ',
      testContentRequired: 'Please enter the test message content',
      testSent: 'Test message sent to "{name}"',
      testFailed: 'Test to "{name}" failed: {message}',
      toggleFailed: 'Failed to toggle status: ',
      deleteFailed: 'Delete failed: ',
      confirmDelete: 'Delete medium "{name}"?'
    },
    signed: ' (signed)',
    defaultTestMessage: '[AIOPS] Notification medium "test" message, time: {time}'
  }
}
