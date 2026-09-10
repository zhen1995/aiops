export default {
  header: {
    title: 'Ops Overview Dashboard',
    desc: 'Multi-source monitoring data fusion · real-time overview of anomaly detection / root cause analysis / alert noise reduction',
    range24h: 'Last 24 hours',
    rangeCustom: 'Custom range',
    to: 'to',
    confirm: 'Apply',
    retry: 'Retry',
    noData: 'No data'
  },
  kpi: {
    activeAlerts: {
      label: 'Active Alerts',
      hint: 'vs yesterday',
      help: 'Total alerts still in unresolved (firing) state, regardless of trigger time. The percentage below compares with the unresolved alert count 24 hours ago.'
    },
    anomalyToday: {
      label: 'Anomalies Today',
      hint: 'anomalies hit',
      help: 'Alert events triggered since 00:00 today; "+N" is the difference versus the same period yesterday.'
    },
    compressionRate: {
      label: 'Alert Compression Rate',
      hint: 'Target ≥80%',
      help: 'Noise-reduction interceptions as a proportion of (alert events + interceptions), reflecting noise-reduction effectiveness. Target ≥80%.'
    },
    mttr: {
      label: 'Avg MTTR',
      hint: 'Target <12min',
      help: 'Average time from trigger to recovery (minutes) for alerts resolved within the range. Target <12min.'
    }
  },
  trend: {
    title: 'Alert Trend (Raw vs Denoised)',
    sub: 'Noise reduction active in real time, compression rate {rate}%',
    legendRaw: 'Raw alerts',
    legendDenoised: 'Denoised alerts'
  },
  severity: {
    title: 'Alert Severity Distribution',
    sub: 'All alerts in range grouped by severity'
  },
  health: {
    title: 'Service Health',
    sub: 'Composite score based on multi-metric correlation detection',
    abnormalMetrics: '{count} abnormal metrics',
    trendWorse: '↓ Worsening',
    trendBetter: '↑ Improving',
    trendStable: '→ Stable'
  },
  alerts: {
    title: 'Latest High-Severity Alerts',
    sub: 'Latest alert events in range (sorted by severity)',
    view: 'View alert events',
    colLevel: 'Severity',
    colContent: 'Alert Content',
    colService: 'Service',
    colTime: 'Time',
    p1: 'P1-Critical',
    p2: 'P2-Warning',
    p3: 'P3-Notice'
  },
  error: {
    loadFailed: 'Failed to load overview data'
  }
}
