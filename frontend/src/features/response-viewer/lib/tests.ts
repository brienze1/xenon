import type { TestResult, TestStatus, TestSummary } from '@/platform/mappers/request'

export type TestFilter = 'all' | TestStatus

export const TEST_FILTERS: ReadonlyArray<{ value: TestFilter; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'passed', label: 'Passed' },
  { value: 'failed', label: 'Failed' },
  { value: 'skipped', label: 'Skipped' },
]

export function filterTests(tests: readonly TestResult[], filter: TestFilter): TestResult[] {
  return filter === 'all' ? [...tests] : tests.filter((test) => test.status === filter)
}

export function filterCount(summary: TestSummary, filter: TestFilter): number {
  return filter === 'all' ? summary.total : summary[filter]
}

/** "Test Results (3/4)": passed out of every test that ran, as Postman counts them. */
export function testResultsTitle(summary: TestSummary): string {
  return summary.total === 0 ? 'Test Results' : `Test Results (${summary.passed}/${summary.total})`
}
