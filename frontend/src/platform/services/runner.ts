import * as RunnerService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/runnerservice'
import { AppError, toAppError } from '../errors'
import { toSendResult, type SendResult } from '../mappers/request'
import {
  toRun,
  toRunConfigDto,
  toRunPlan,
  toRunSummaryView,
  type Run,
  type RunConfig,
  type RunExportFormat,
  type RunPlan,
  type RunSummaryView,
} from '../mappers/runner'

const EMPTY_RESPONSE = 'UNK-00204'

/** No active or stored run with that id, or (from `Response`) no persisted response for that row. */
export const UNKNOWN_RUN = 'RUN-01404'
/** A run is already active on this collection or folder. */
export const RUN_ALREADY_ACTIVE = 'RUN-02409'
/** No row is selected. */
export const EMPTY_SELECTION = 'RUN-03400'
/** A run setting is out of range; its detail names the field (`iterations`, `profile.ramp_seconds` …). */
export const INVALID_RUN_SETTING = 'RUN-04400'
/** The iteration data file is malformed or unreadable. */
export const INVALID_ITERATION_DATA = 'RUN-05400'
/** The collection or folder no longer exists. */
export const UNKNOWN_RUN_TARGET = 'RUN-07404'

/** What the user already agreed to for this run, as for a single send. */
export interface RunAcknowledgement {
  confirmedEnvironmentId: string
  missingSecrets: boolean
}

/** One field error of a refused start, keyed by the run setting it names. */
export function runFieldErrorsOf(error: AppError): Record<string, string> {
  if (error.code === INVALID_ITERATION_DATA) {
    return { data_file: error.details[0]?.messages[0] ?? error.description }
  }
  if (error.code !== INVALID_RUN_SETTING) {
    return {}
  }
  return Object.fromEntries(error.details.filter((detail) => detail.attribute !== '').map((detail) => [detail.attribute, detail.messages.join(' ')]))
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

/** The target's rows, remembered configuration, environments and active run. */
export function planRun(localId: string, itemId: string): Promise<RunPlan> {
  return call(async () => toRunPlan(required(await RunnerService.Plan({ local_id: localId, item_id: itemId }))))
}

/** Remembers the rows and settings of a target on this machine; no workspace file changes. */
export function saveRunnerPreferences(localId: string, itemId: string, config: RunConfig): Promise<void> {
  return call(() => RunnerService.SavePreferences({ local_id: localId, item_id: itemId, config: toRunConfigDto(config) }))
}

/** Starts a run in the background; progress and the end arrive as events. */
export function startRun(localId: string, itemId: string, config: RunConfig, acknowledgement: RunAcknowledgement): Promise<RunSummaryView> {
  return call(async () =>
    toRunSummaryView(
      required(
        await RunnerService.Start({
          local_id: localId,
          item_id: itemId,
          config: toRunConfigDto(config),
          confirmed_environment_id: acknowledgement.confirmedEnvironmentId,
          acknowledge_missing_secrets: acknowledgement.missingSecrets,
        }),
      ),
    ),
  )
}

export function pauseRun(localId: string, runId: string): Promise<RunSummaryView> {
  return call(async () => toRunSummaryView(required(await RunnerService.Pause({ local_id: localId, run_id: runId }))))
}

export function resumeRun(localId: string, runId: string): Promise<RunSummaryView> {
  return call(async () => toRunSummaryView(required(await RunnerService.Resume({ local_id: localId, run_id: runId }))))
}

export function stopRun(localId: string, runId: string): Promise<RunSummaryView> {
  return call(async () => toRunSummaryView(required(await RunnerService.Stop({ local_id: localId, run_id: runId }))))
}

/** The active run, or else the stored run. */
export function getRun(localId: string, runId: string): Promise<Run> {
  return call(async () => toRun(required(await RunnerService.Get({ local_id: localId, run_id: runId }))))
}

/** The stored runs of a target, newest first. */
export function listRuns(localId: string, itemId: string): Promise<RunSummaryView[]> {
  return call(async () => ((await RunnerService.List({ local_id: localId, item_id: itemId })) ?? []).map(toRunSummaryView))
}

/** The active runs of the workspace (to restore the running indicators). */
export function activeRuns(localId: string): Promise<RunSummaryView[]> {
  return call(async () => ((await RunnerService.Active({ local_id: localId })) ?? []).map(toRunSummaryView))
}

/** The persisted response of a run row, while this session holds it (otherwise `RUN-01404`). */
export function runResponse(localId: string, runId: string, index: number): Promise<SendResult> {
  return call(async () => toSendResult(required(await RunnerService.Response({ local_id: localId, run_id: runId, index }))))
}

export function deleteRun(localId: string, runId: string): Promise<void> {
  return call(() => RunnerService.Delete({ local_id: localId, run_id: runId }))
}

/** Deletes every stored run of the target. */
export function clearRuns(localId: string, itemId: string): Promise<void> {
  return call(() => RunnerService.Clear({ local_id: localId, item_id: itemId }))
}

export function exportRun(localId: string, runId: string, format: RunExportFormat, path: string): Promise<void> {
  return call(() => RunnerService.Export({ local_id: localId, run_id: runId, format, path }))
}
