import type { SendResult, TestResult, TestSummary } from '@/platform/mappers/request'
import { TestResultsTab } from '@/features/response-viewer/components/TestResultsTab'
import type { RunnerSlots } from '@/features/runner/model'
import { RunRowResponse } from './RunRowResponse'

/** The response viewer parts the Runner shows for a result row. */
export const RUNNER_SLOTS: RunnerSlots = {
  tests: (tests: readonly TestResult[], summary: TestSummary) => <TestResultsTab tests={tests} summary={summary} />,
  response: (responseKey: string, result: SendResult) => <RunRowResponse responseKey={responseKey} result={result} />,
}
