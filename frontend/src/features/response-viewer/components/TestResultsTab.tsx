import { useState } from 'react'
import type { TestResult, TestSummary } from '@/platform/mappers/request'
import { cn } from '@/shared/lib/cn'
import { formatNumber } from '@/shared/lib/number-format'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { filterCount, filterTests, TEST_FILTERS, type TestFilter } from '@/features/response-viewer/lib/tests'
import styles from './TestResultsTab.module.css'

const STATUS_LABELS: Record<TestResult['status'], string> = { passed: 'Pass', failed: 'Fail', skipped: 'Skip' }

const NO_TESTS = (
  <EmptyState.Root label="No tests" size="compact">
    <EmptyState.Title>No Tests in This Send</EmptyState.Title>
    <EmptyState.Description>Add pm.test(…) to a script to check the response automatically.</EmptyState.Description>
  </EmptyState.Root>
)

interface TestResultsTabProps {
  tests: readonly TestResult[]
  summary: TestSummary
}

export function TestResultsTab({ tests, summary }: TestResultsTabProps) {
  const [filter, setFilter] = useState<TestFilter>('all')
  if (tests.length === 0) {
    return NO_TESTS
  }
  const visible = filterTests(tests, filter)
  return (
    <div className={styles.results}>
      <div role="group" aria-label="Filter tests" className={styles.filters}>
        {TEST_FILTERS.map((option) => (
          <button
            key={option.value}
            type="button"
            aria-pressed={filter === option.value}
            className={styles.filter}
            onClick={() => setFilter(option.value)}
          >
            {option.label} ({formatNumber(filterCount(summary, option.value), 0)})
          </button>
        ))}
      </div>
      {visible.length === 0 ? (
        <p className={styles.none}>No {filter} tests.</p>
      ) : (
        <ul aria-label="Test results" className={styles.list}>
          {visible.map((test) => (
            <li key={test.index} className={styles.row}>
              <span className={cn(styles.status, styles[test.status])}>{STATUS_LABELS[test.status]}</span>
              <div className={styles.detail}>
                <span className={styles.name}>{test.name}</span>
                {test.errorMessage ? <span className={styles.message}>{test.errorMessage}</span> : null}
              </div>
              <span className={styles.script} translate="no">
                {test.scriptLabel}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
