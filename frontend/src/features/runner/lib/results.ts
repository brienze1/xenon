import type { RunRowResult } from '@/platform/mappers/runner'
import type { ResultFilter } from '@/features/runner/model'

export const RESULT_FILTERS: ReadonlyArray<{ filter: ResultFilter; label: string }> = [
  { filter: 'all', label: 'All' },
  { filter: 'passed', label: 'Passed' },
  { filter: 'failed', label: 'Failed' },
  { filter: 'skipped', label: 'Skipped' },
]

/** FR-034: Passed is no error and every test passed; Failed is a failed test, an error or a cancelled row; Skipped is skipped by a script. */
export function matchesFilter(row: RunRowResult, filter: ResultFilter): boolean {
  switch (filter) {
    case 'all':
      return true
    case 'passed':
      return row.outcome === 'passed'
    case 'failed':
      return row.outcome === 'failed' || row.outcome === 'error' || row.outcome === 'cancelled'
    case 'skipped':
      return row.outcome === 'skipped'
  }
}

export function filterCounts(rows: readonly RunRowResult[]): Record<ResultFilter, number> {
  const counts: Record<ResultFilter, number> = { all: rows.length, passed: 0, failed: 0, skipped: 0 }
  for (const row of rows) {
    if (matchesFilter(row, 'passed')) {
      counts.passed++
    } else if (matchesFilter(row, 'failed')) {
      counts.failed++
    } else if (matchesFilter(row, 'skipped')) {
      counts.skipped++
    }
  }
  return counts
}

/** One line of the virtualized results list: an iteration heading or a row result. */
export type ResultItem = { kind: 'iteration'; key: string; iteration: number; count: number } | { kind: 'row'; key: string; row: RunRowResult }

/** The filtered rows grouped by iteration, flattened for the virtualized list. */
export function resultItems(rows: readonly RunRowResult[], filter: ResultFilter): ResultItem[] {
  const items: ResultItem[] = []
  let heading: Extract<ResultItem, { kind: 'iteration' }> | null = null
  for (const row of rows) {
    if (!matchesFilter(row, filter)) {
      continue
    }
    if (heading === null || heading.iteration !== row.iteration) {
      heading = { kind: 'iteration', key: `iteration-${row.iteration}-${row.index}`, iteration: row.iteration, count: 0 }
      items.push(heading)
    }
    heading.count++
    items.push({ kind: 'row', key: `row-${row.index}`, row })
  }
  return items
}
