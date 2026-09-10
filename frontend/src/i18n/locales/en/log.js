export default {
  header: {
    title: 'Log Analysis',
    desc: 'Unstructured log intelligent parsing · Drain template extraction · Log clustering and anomaly detection',
    last24h: 'Last 24 Hours',
    query: 'Query Logs'
  },
  pipeline: {
    title: 'Log Processing Pipeline',
    sub: 'Raw logs are processed in real time through six stages, producing anomaly events and alerts',
    desc1: 'Format recognition and field extraction',
    desc2: 'Online clustering to generate log templates',
    desc3: 'Splitting constant templates and variable parameters',
    desc4: 'Converting log sequences into feature vectors',
    desc5: 'Aggregating similar patterns into clusters',
    desc6: 'Anomaly detection by frequency / new templates'
  },
  trend: {
    title: 'Log Volume Trend',
    sub: 'Total log volume and ERROR log volume (entries/hour) · ERROR surged significantly after 13:00'
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
    trendFlat: 'Stable'
  },
  template: {
    title: 'Drain Template Extraction',
    sub: 'Raw logs → constant templates + variable parameters, updated online in real time',
    raw: 'Raw Log',
    tpl: 'Template',
    params: 'Parameters'
  },
  chart: {
    total: 'Total Logs',
    error: 'ERROR Logs'
  }
}
