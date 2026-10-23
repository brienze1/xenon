import type { PerformanceReport, RequestStats } from '@/platform/mappers/runner'
import { formatDuration } from '@/shared/lib/format-duration'
import { formatNumber } from '@/shared/lib/number-format'
import { seriesPath, type SeriesKey } from '@/features/runner/lib/time-series'
import styles from './PerformanceResults.module.css'

const CHART_WIDTH = 600
const PANEL_HEIGHT = 56
const PANEL_GAP = 20

interface Series {
  key: SeriesKey
  label: string
  stepped: boolean
}

const SERIES: readonly Series[] = [
  { key: 'avgMs', label: 'Response time (ms)', stepped: false },
  { key: 'requests', label: 'Throughput (req/s)', stepped: false },
  { key: 'errors', label: 'Errors', stepped: false },
  { key: 'activeUsers', label: 'Active users', stepped: true },
]

interface StatsRowProps {
  stats: RequestStats
  label: string
}

function StatsRow({ stats, label }: StatsRowProps) {
  return (
    <tr>
      <th scope="row" translate="no">
        {label}
      </th>
      <td>{formatNumber(stats.count, 0)}</td>
      <td>{formatDuration(stats.avgMs)}</td>
      <td>{formatDuration(stats.p90Ms)}</td>
      <td>{formatDuration(stats.p95Ms)}</td>
      <td>{formatDuration(stats.p99Ms)}</td>
      <td>{formatDuration(stats.minMs)}</td>
      <td>{formatDuration(stats.maxMs)}</td>
      <td>{formatNumber(stats.throughput, 1)}/s</td>
      <td>{formatNumber(stats.errorRate * 100, 1)}%</td>
      <td translate="no">{stats.statusCounts.map((entry) => `${entry.status}: ${entry.count}`).join(', ')}</td>
    </tr>
  )
}

interface PerformanceResultsProps {
  report: PerformanceReport
}

/** A performance run's statistics (overall and per request) and its per-second time series. */
export function PerformanceResults({ report }: PerformanceResultsProps) {
  const height = SERIES.length * (PANEL_HEIGHT + PANEL_GAP)
  return (
    <div className={styles.performance}>
      <div className={styles.tableScroll}>
        <table className={styles.table}>
          <caption className="visually-hidden">Performance statistics</caption>
          <thead>
            <tr>
              <th scope="col">Request</th>
              <th scope="col">Requests</th>
              <th scope="col">Avg</th>
              <th scope="col">P90</th>
              <th scope="col">P95</th>
              <th scope="col">P99</th>
              <th scope="col">Min</th>
              <th scope="col">Max</th>
              <th scope="col">Throughput</th>
              <th scope="col">Error rate</th>
              <th scope="col">Status codes</th>
            </tr>
          </thead>
          <tbody>
            <StatsRow stats={report.overall} label="All requests" />
            {report.perRequest.map((stats) => (
              <StatsRow key={stats.requestId || stats.name} stats={stats} label={stats.name} />
            ))}
          </tbody>
        </table>
      </div>
      <figure className={styles.figure}>
        <svg viewBox={`0 0 ${CHART_WIDTH} ${height}`} role="img" aria-label={`Time series over ${report.series.length} seconds`} className={styles.chart}>
          {SERIES.map((series, index) => {
            const top = index * (PANEL_HEIGHT + PANEL_GAP) + PANEL_GAP
            return (
              <g key={series.key} transform={`translate(0 ${top})`} className={styles.panel} data-series={series.key}>
                <text x={0} y={-6} className={styles.label}>
                  {series.label}
                </text>
                <line x1={0} y1={PANEL_HEIGHT} x2={CHART_WIDTH} y2={PANEL_HEIGHT} className={styles.axis} />
                <path d={seriesPath(report.series, series.key, series.stepped, CHART_WIDTH, PANEL_HEIGHT)} className={styles.line} />
              </g>
            )
          })}
        </svg>
      </figure>
    </div>
  )
}
