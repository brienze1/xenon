import type { RunRowOutcome, RunStatus, RunType } from '@/platform/mappers/runner'

export const STATUS_LABELS: Record<RunStatus, string> = {
  running: 'Running',
  paused: 'Paused',
  completed: 'Completed',
  stopped: 'Stopped',
  stopped_on_error: 'Stopped on error',
}

export const TYPE_LABELS: Record<RunType, string> = { functional: 'Functional', performance: 'Performance' }

export const OUTCOME_LABELS: Record<RunRowOutcome, string> = {
  passed: 'Passed',
  failed: 'Failed',
  skipped: 'Skipped',
  error: 'Error',
  cancelled: 'Cancelled',
}

/** How the finish toast names the status: "<target> run completed · …". */
export const FINISHED_LABELS: Record<RunStatus, string> = {
  running: 'running',
  paused: 'paused',
  completed: 'completed',
  stopped: 'stopped',
  stopped_on_error: 'stopped on error',
}
