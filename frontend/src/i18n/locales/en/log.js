export default {
  header: {
    title: 'Log Analysis',
    desc: 'Unstructured log intelligent parsing · Drain template extraction · Log clustering and anomaly detection',
    last1h: 'Last 1 Hour',
    last6h: 'Last 6 Hours',
    last12h: 'Last 12 Hours',
    last24h: 'Last 24 Hours',
    query: 'Query',
    querying: 'Querying…',
    serviceAll: 'All Services',
    serviceSelected: '{count} services selected',
    selectAll: 'Select All',
    clear: 'Clear'
  },
  pipeline: {
    title: 'Log Processing Pipeline',
    sub: 'Raw logs are processed in real time through six stages, producing anomaly events and alerts',
    step1: 'Parsing',
    step2: 'Template Extraction (Drain)',
    step3: 'Parameter Splitting',
    step4: 'Vectorization',
    step5: 'Clustering',
    step6: 'Anomaly Detection',
    desc1: 'Format recognition and field extraction',
    desc2: 'Online clustering to generate log templates',
    desc3: 'Splitting constant templates and variable parameters',
    desc4: 'Converting log sequences into feature vectors',
    desc5: 'Aggregating similar patterns into clusters',
    desc6: 'Anomaly detection by frequency / new templates'
  },
  trend: {
    title: 'Log Volume Trend',
    sub: 'Total log volume and ERROR log volume (entries/hour)'
  },
  cluster: {
    title: 'Log Clustering',
    sub: 'Similar logs are automatically aggregated into pattern clusters, sorted by frequency and trend',
    total: '{count} active clusters',
    colId: 'Cluster ID',
    colPattern: 'Log Pattern',
    colCount: 'Count',
    colLevel: 'Level',
    colServices: 'Related Services',
    colTrend: 'Trend',
    colFirstSeen: 'First Seen',
    trendSpike: 'Surge',
    trendRising: 'Rising',
    trendFlat: 'Stable',
    filterAnomaly: 'Error Levels (ERROR/WARN)',
    filterError: 'ERROR',
    filterWarn: 'WARN',
    filterInfo: 'INFO',
    filterAll: 'All Levels'
  },
  template: {
    title: 'Drain Template Extraction',
    sub: 'Raw logs → constant templates + variable parameters, updated online in real time',
    raw: 'Raw Log',
    tpl: 'Template',
    params: 'Parameters',
    count: '{count} entries'
  },
  chart: {
    total: 'Total Logs',
    error: 'ERROR Logs'
  },
  common: {
    empty: 'No data yet. It will be shown after the log analysis task runs.',
    loadFailed: 'Failed to load, please retry later',
    copy: 'Copy',
    copied: 'Copied'
  }
}
