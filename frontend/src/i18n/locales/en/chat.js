export default {
  session: {
    newButton: '+ New Chat',
    noMessage: 'No messages yet',
    defaultTitle: 'New Chat'
  },
  header: {
    title: 'AI Ops Assistant',
    subtitle: 'LLM-powered intelligent Q&A, alert interpretation and root cause analysis',
    currentModel: 'Current model',
    defaultModel: 'Default model',
    noDefaultModel: 'No default model configured'
  },
  message: {
    userAvatar: 'Me',
    assistantName: 'AIOPS Assistant',
    userName: 'Ops Admin',
    thinking: 'Thinking...',
    copy: 'Copy',
    copied: 'Copied',
    error: '\n[Error: {err}]',
    aborted: '_[Aborted]_'
  },
  toolbar: {
    p0Alerts: "Today's P0 Alerts",
    rootCause: 'Root Cause Analysis',
    inspectionSummary: 'Inspection Summary'
  },
  input: {
    placeholder: 'Ask a question, e.g.: How is the system doing today?',
    send: 'Send',
    abort: 'Stop'
  },
  rca: {
    invalidParam: 'Invalid root cause analysis parameters',
    statusRecovered: 'Recovered',
    statusFiring: 'Firing',
    severityP1: 'P1-Critical',
    severityP2: 'P2-Warning',
    severityP3: 'P3-Notice',
    prompt: `Please perform a root cause analysis on the following alert event:
- Rule name: {ruleName}
- Alert target: {targetIdent}
- Trigger time: {time}
- Severity: {severity}
- Status: {status}
- Tags: {tags}
- Trigger value: {triggerValue}

Please query the relevant Prometheus metrics and Elasticsearch logs to investigate the outage cause, and provide the root cause, evidence chain and remediation suggestions.`
  }
}
