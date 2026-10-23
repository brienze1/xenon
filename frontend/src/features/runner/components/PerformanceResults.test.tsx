import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { RequestStats } from '@/platform/mappers/runner'
import { seriesPath } from '@/features/runner/lib/time-series'
import { PerformanceResults } from './PerformanceResults'

function stats(name: string, extra: Partial<RequestStats> = {}): RequestStats {
  return {
    requestId: name,
    name,
    count: 100,
    avgMs: 40,
    p90Ms: 80,
    p95Ms: 90,
    p99Ms: 120,
    minMs: 5,
    maxMs: 200,
    throughput: 3.3,
    errorRate: 0.01,
    statusCounts: [
      { status: '200', count: 99 },
      { status: '500', count: 1 },
    ],
    ...extra,
  }
}

const series = [
  { second: 0, activeUsers: 1, requests: 2, avgMs: 10, errors: 0 },
  { second: 1, activeUsers: 2, requests: 4, avgMs: 20, errors: 1 },
]

describe('PerformanceResults', () => {
  it('shows the overall and per-request statistics', () => {
    render(<PerformanceResults report={{ overall: stats('All', { count: 200 }), perRequest: [stats('Login'), stats('Get order')], series }} />)
    const table = screen.getByRole('table', { name: 'Performance statistics' })
    const rows = within(table).getAllByRole('row')
    expect(within(rows[0]!).getAllByRole('columnheader').map((cell) => cell.textContent)).toEqual([
      'Request',
      'Requests',
      'Avg',
      'P90',
      'P95',
      'P99',
      'Min',
      'Max',
      'Throughput',
      'Error rate',
      'Status codes',
    ])
    expect(rows.slice(1).map((row) => within(row).getByRole('rowheader').textContent)).toEqual(['All requests', 'Login', 'Get order'])
    expect(rows[1]).toHaveTextContent('200')
    expect(rows[2]).toHaveTextContent('1.0%')
    expect(rows[2]).toHaveTextContent('200: 99, 500: 1')
  })

  it('draws the time series, with active users as a stepped line', () => {
    const { container } = render(<PerformanceResults report={{ overall: stats('All'), perRequest: [], series }} />)
    expect(screen.getByRole('img', { name: 'Time series over 2 seconds' })).toBeInTheDocument()
    expect(container.querySelector('[data-series="activeUsers"] path')).toHaveAttribute('d', 'M0,28H600V0')
    expect(container.querySelector('[data-series="avgMs"] path')).toHaveAttribute('d', 'M0,28L600,0')
    expect(seriesPath([], 'errors', false, 10, 10)).toBe('')
  })
})
