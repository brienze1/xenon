import type { ReactNode } from 'react'
import type { AppError } from '@/platform/errors'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import type { SendResult, TestResult, TestSummary } from '@/platform/mappers/request'
import type { Run, RunConfig, RunPlanRow, RunStatus, RunSummaryView, RunTarget } from '@/platform/mappers/runner'
import type { ContainerKind } from '@/platform/mappers/tree'
import type { MissingSecrets, SendConfirmation } from '@/platform/services/requests'

/** The sub-tabs of the Runner section. */
export type RunnerView = 'setup' | 'results' | 'runs'

/** The results filter chips (FR-034). */
export type ResultFilter = 'all' | 'passed' | 'failed' | 'skipped'

export type RunnerStatus = 'idle' | 'loading' | 'ready'

export interface RunnerFailure {
  error: AppError
  nextStep: string
}

/** One collection or folder in the Runner: its plan, the edited configuration and the run it shows. */
export interface RunnerTarget {
  status: RunnerStatus
  loadError: AppError | null
  target: RunTarget | null
  /** The run order as shown (and remembered): order, checks and alias choices. */
  rows: RunPlanRow[]
  config: RunConfig | null
  environments: EnvironmentSummary[]
  activeEnvironmentId: string | null
  /** The remembered data file vanished; Go cleared it and the Setup shows a notice. */
  dataFileMissing: boolean
  view: RunnerView
  /** The run the Results sub-tab shows (the active one, the last started or an opened stored run). */
  runId: string | null
  run: Run | null
  /** The stored runs, newest first; null until the Runs sub-tab first loads them. */
  runs: RunSummaryView[] | null
  failure: RunnerFailure | null
  /** Inline errors of a refused start, by run setting (`iterations`, `profile.ramp_seconds`, `data_file` …). */
  fieldErrors: Record<string, string>
  /** The run's environment asks for a confirmation (SND-05409). */
  confirmation: SendConfirmation | null
  /** Some secrets have no value (SND-06409); Run anyway restarts with the acknowledgement. */
  missingSecrets: MissingSecrets | null
  /** What the user already confirmed for the start in progress. */
  acknowledgement: { confirmedEnvironmentId: string; missingSecrets: boolean }
  /** The configuration of the start in progress (Run again starts a stored run's configuration). */
  pendingConfig: RunConfig | null
  starting: boolean
}

export type ResponseStatus = 'loading' | 'ready' | 'missing' | 'error'

/** The finish toast: the run that ended last. */
export interface RunFinishedNotice {
  itemId: string
  runId: string
  kind: ContainerKind | null
  targetName: string
  status: RunStatus
  passed: number
  failed: number
}

/** What the app renders inside the Runner for a row: the response viewer's test list and response parts. */
export interface RunnerSlots {
  /** A row's tests, rendered by the response viewer's Test Results. */
  tests: (tests: readonly TestResult[], summary: TestSummary) => ReactNode
  /** A persisted response: status line, headers and the body keyed by `responseKey`. */
  response: (responseKey: string, result: SendResult) => ReactNode
}
