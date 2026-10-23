import type {
  Run as RunDto,
  RunConfig as RunConfigDto,
  RunPerformance as RunPerformanceDto,
  RunPerformanceMetrics as RunPerformanceMetricsDto,
  RunPerformanceSecond as RunPerformanceSecondDto,
  RunRequestStats as RunRequestStatsDto,
  RunPlan as RunPlanDto,
  RunPlanRow as RunPlanRowDto,
  RunProgress as RunProgressDto,
  RunRowResult as RunRowResultDto,
  RunSummary as RunSummaryDto,
  RunSummaryView as RunSummaryViewDto,
  RunTarget as RunTargetDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import type {
  RunFinished as RunFinishedDto,
  RunProgressChanged as RunProgressChangedDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/publisher/dto/models'
import { toConsoleEntry, type ConsoleEntry } from './console'
import { toEnvironmentSummary, type EnvironmentSummary } from './environment'
import { toScriptFailure, type ScriptFailure, type TestResult, type TestSummary } from './request'
import type { ContainerKind } from './tree'

export const RUN_TYPES = ['functional', 'performance'] as const
export const RUN_STATUSES = ['running', 'paused', 'completed', 'stopped', 'stopped_on_error'] as const
export const RUN_ROW_OUTCOMES = ['passed', 'failed', 'skipped', 'error', 'cancelled'] as const
export const LOAD_PROFILE_KINDS = ['fixed', 'ramp_up', 'spike', 'peak'] as const
export const RUN_EXPORT_FORMATS = ['json', 'junit'] as const

export type RunType = (typeof RUN_TYPES)[number]
export type RunStatus = (typeof RUN_STATUSES)[number]
export type RunRowOutcome = (typeof RUN_ROW_OUTCOMES)[number]
export type LoadProfileKind = (typeof LOAD_PROFILE_KINDS)[number]
export type RunExportFormat = (typeof RUN_EXPORT_FORMATS)[number]

/** In `environmentId`, the active environment of the workspace. */
export const ACTIVE_ENVIRONMENT = ''
/** In `environmentId`, no environment at all. */
export const NO_ENVIRONMENT = 'none'

/** One remembered row of the run order: which request (and alias version) runs, and whether it is checked. */
export interface RunRowChoice {
  requestId: string
  aliasId: string
  selected: boolean
}

export interface FunctionalSettings {
  iterations: number
  dataFile: string
  delayMs: number
  persistResponses: boolean
  disableLogs: boolean
  stopOnError: boolean
  keepVariableValues: boolean
  skipStoredCookies: boolean
  saveCookies: boolean
}

export interface LoadProfile {
  kind: LoadProfileKind
  initialUsers: number
  rampSeconds: number
  baseUsers: number
  spikeSeconds: number
  peakSeconds: number
}

export interface PerformanceSettings {
  virtualUsers: number
  durationSeconds: number
  dataFile: string
  profile: LoadProfile
}

export interface RunConfig {
  type: RunType
  rows: RunRowChoice[]
  functional: FunctionalSettings
  performance: PerformanceSettings
  /** `''` runs with the active environment, `'none'` without one. */
  environmentId: string
}

export interface RunTarget {
  itemId: string
  kind: ContainerKind
  name: string
}

export interface RunPlanAlias {
  id: string
  name: string
}

/** A row of the run order with what the list shows about its request. */
export interface RunPlanRow extends RunRowChoice {
  number: number
  name: string
  method: string
  folderPath: string[]
  aliases: RunPlanAlias[]
}

export interface RunMetrics {
  activeUsers: number
  requests: number
  avgMs: number
  throughput: number
  errorRate: number
}

export interface RunProgress {
  currentRow: number
  currentName: string
  iteration: number
  completed: number
  planned: number
  testsPassed: number
  testsFailed: number
  elapsedMs: number
  /** The live metrics of a performance run; null for a functional run. */
  metrics: RunMetrics | null
}

export interface RunSummary {
  iterations: number
  totalRequests: number
  testsPassed: number
  testsFailed: number
  testsSkipped: number
  avgResponseMs: number
  durationMs: number
  passed: number
  failed: number
  skipped: number
  errors: number
  cancelled: number
}

/** A run in a list: the active runs of a workspace or the stored runs of a target. */
export interface RunSummaryView {
  id: string
  itemId: string
  targetKind: ContainerKind
  targetName: string
  type: RunType
  status: RunStatus
  environmentName: string
  startedAt: string
  finishedAt: string
  progress: RunProgress
  summary: RunSummary
}

export interface RunPlan {
  target: RunTarget
  rows: RunPlanRow[]
  config: RunConfig
  activeEnvironmentId: string | null
  environments: EnvironmentSummary[]
  activeRun: RunSummaryView | null
  /** The remembered data file no longer exists; Go cleared it. */
  dataFileMissing: boolean
}

export interface RunSentRequest {
  method: string
  url: string
  headers: Array<{ key: string; value: string }>
  body: string
}

export interface RunRowResult {
  index: number
  iteration: number
  requestId: string
  aliasId: string
  aliasName: string
  name: string
  method: string
  folderPath: string[]
  url: string
  status: number
  durationMs: number
  sizeBytes: number
  outcome: RunRowOutcome
  tests: TestResult[]
  testSummary: TestSummary
  scriptErrors: ScriptFailure[]
  console: ConsoleEntry[]
  sentRequest: RunSentRequest
  error: string
  /** Set when the run persisted responses and this session still holds this row's. */
  responseBodyUrl: string
}

export interface StatusCount {
  status: string
  count: number
}

export interface RequestStats {
  requestId: string
  name: string
  count: number
  avgMs: number
  p90Ms: number
  p95Ms: number
  p99Ms: number
  minMs: number
  maxMs: number
  throughput: number
  errorRate: number
  statusCounts: StatusCount[]
}

export interface SeriesPoint {
  second: number
  activeUsers: number
  requests: number
  avgMs: number
  errors: number
}

export interface PerformanceReport {
  overall: RequestStats
  perRequest: RequestStats[]
  series: SeriesPoint[]
}

export interface Run {
  id: string
  target: RunTarget
  config: RunConfig
  environmentName: string
  dataFileName: string
  status: RunStatus
  startedAt: string
  finishedAt: string
  progress: RunProgress
  summary: RunSummary
  rows: RunRowResult[]
  performance: PerformanceReport | null
  notices: string[]
}

export interface RunProgressEvent {
  localId: string
  runId: string
  itemId: string
  type: RunType
  status: RunStatus
  progress: RunProgress
}

export interface RunFinishedEvent {
  localId: string
  runId: string
  itemId: string
  targetName: string
  type: RunType
  status: RunStatus
  summary: RunSummary
}

/** The response-viewer key of a persisted run response: `run:<runId>:<index>`. */
export function runResponseKey(runId: string, index: number): string {
  return `run:${runId}:${index}`
}

function oneOf<T extends string>(values: readonly T[], value: string | undefined, fallback: T): T {
  return (values as readonly string[]).includes(value ?? '') ? (value as T) : fallback
}

function containerKind(kind: string): ContainerKind {
  return kind === 'folder' ? 'folder' : 'collection'
}

export function toRunConfig(dto: RunConfigDto): RunConfig {
  const functional = dto.functional
  const performance = dto.performance
  const profile = performance.profile
  return {
    type: oneOf(RUN_TYPES, dto.type, 'functional'),
    rows: (dto.rows ?? []).map((row) => ({ requestId: row.request_id, aliasId: row.alias_id, selected: row.selected })),
    functional: {
      iterations: functional.iterations,
      dataFile: functional.data_file,
      delayMs: functional.delay_ms,
      persistResponses: functional.persist_responses,
      disableLogs: functional.disable_logs,
      stopOnError: functional.stop_on_error,
      keepVariableValues: functional.keep_variable_values,
      skipStoredCookies: functional.skip_stored_cookies,
      saveCookies: functional.save_cookies,
    },
    performance: {
      virtualUsers: performance.virtual_users,
      durationSeconds: performance.duration_seconds,
      dataFile: performance.data_file,
      profile: {
        kind: oneOf(LOAD_PROFILE_KINDS, profile.kind, 'fixed'),
        initialUsers: profile.initial_users,
        rampSeconds: profile.ramp_seconds,
        baseUsers: profile.base_users,
        spikeSeconds: profile.spike_seconds,
        peakSeconds: profile.peak_seconds,
      },
    },
    environmentId: dto.environment_id,
  }
}

export function toRunConfigDto(config: RunConfig): RunConfigDto {
  const { functional, performance } = config
  const profile = performance.profile
  return {
    type: config.type,
    environment_id: config.environmentId,
    rows: config.rows.map((row) => ({ request_id: row.requestId, alias_id: row.aliasId, selected: row.selected })),
    functional: {
      iterations: functional.iterations,
      data_file: functional.dataFile,
      delay_ms: functional.delayMs,
      persist_responses: functional.persistResponses,
      disable_logs: functional.disableLogs,
      stop_on_error: functional.stopOnError,
      keep_variable_values: functional.keepVariableValues,
      skip_stored_cookies: functional.skipStoredCookies,
      save_cookies: functional.saveCookies,
    },
    performance: {
      virtual_users: performance.virtualUsers,
      duration_seconds: performance.durationSeconds,
      data_file: performance.dataFile,
      profile: {
        kind: profile.kind,
        initial_users: profile.initialUsers,
        ramp_seconds: profile.rampSeconds,
        base_users: profile.baseUsers,
        spike_seconds: profile.spikeSeconds,
        peak_seconds: profile.peakSeconds,
      },
    },
  }
}

function toRunTarget(dto: RunTargetDto): RunTarget {
  return { itemId: dto.item_id, kind: containerKind(dto.kind), name: dto.name }
}

function toRunPlanRow(dto: RunPlanRowDto): RunPlanRow {
  return {
    number: dto.number,
    requestId: dto.request_id,
    aliasId: dto.alias_id,
    selected: dto.selected,
    name: dto.name,
    method: dto.method,
    folderPath: dto.folder_path ?? [],
    aliases: (dto.aliases ?? []).map((alias) => ({ id: alias.id, name: alias.name })),
  }
}

function toRunMetrics(dto: RunPerformanceMetricsDto | null): RunMetrics | null {
  if (!dto) {
    return null
  }
  return { activeUsers: dto.active_users, requests: dto.requests, avgMs: dto.avg_ms, throughput: dto.throughput, errorRate: dto.error_rate }
}

export function toRunProgress(dto: RunProgressDto): RunProgress {
  return {
    currentRow: dto.current_row,
    currentName: dto.current_name,
    iteration: dto.iteration,
    completed: dto.completed,
    planned: dto.planned,
    testsPassed: dto.tests_passed,
    testsFailed: dto.tests_failed,
    elapsedMs: dto.elapsed_ms,
    metrics: toRunMetrics(dto.metrics),
  }
}

export function toRunSummary(dto: RunSummaryDto): RunSummary {
  return {
    iterations: dto.iterations,
    totalRequests: dto.total_requests,
    testsPassed: dto.tests_passed,
    testsFailed: dto.tests_failed,
    testsSkipped: dto.tests_skipped,
    avgResponseMs: dto.avg_response_ms,
    durationMs: dto.duration_ms,
    passed: dto.passed,
    failed: dto.failed,
    skipped: dto.skipped,
    errors: dto.errors,
    cancelled: dto.cancelled,
  }
}

export function toRunSummaryView(dto: RunSummaryViewDto): RunSummaryView {
  return {
    id: dto.id,
    itemId: dto.item_id,
    targetKind: containerKind(dto.target_kind),
    targetName: dto.target_name,
    type: oneOf(RUN_TYPES, dto.type, 'functional'),
    status: oneOf(RUN_STATUSES, dto.status, 'completed'),
    environmentName: dto.environment_name,
    startedAt: dto.started_at,
    finishedAt: dto.finished_at,
    progress: toRunProgress(dto.progress),
    summary: toRunSummary(dto.summary),
  }
}

export function toRunPlan(dto: RunPlanDto): RunPlan {
  return {
    target: toRunTarget(dto.target),
    rows: (dto.rows ?? []).map(toRunPlanRow),
    config: toRunConfig(dto.config),
    activeEnvironmentId: dto.active_environment_id ? dto.active_environment_id : null,
    environments: (dto.environments ?? []).map(toEnvironmentSummary),
    activeRun: dto.active_run ? toRunSummaryView(dto.active_run) : null,
    dataFileMissing: dto.data_file_missing,
  }
}

function toRunRowResult(dto: RunRowResultDto): RunRowResult {
  const tests: TestResult[] = (dto.tests ?? []).map((test, index) => ({
    index,
    name: test.name,
    status: test.status === 'passed' || test.status === 'skipped' ? test.status : 'failed',
    errorMessage: test.error,
    scriptLabel: '',
    async: false,
  }))
  const count = (status: TestResult['status']) => tests.filter((test) => test.status === status).length
  return {
    index: dto.index,
    iteration: dto.iteration,
    requestId: dto.request_id,
    aliasId: dto.alias_id,
    aliasName: dto.alias_name,
    name: dto.name,
    method: dto.method,
    folderPath: dto.folder_path ?? [],
    url: dto.url,
    status: dto.status,
    durationMs: dto.duration_ms,
    sizeBytes: dto.size_bytes,
    outcome: oneOf(RUN_ROW_OUTCOMES, dto.outcome, 'error'),
    tests,
    testSummary: { total: tests.length, passed: count('passed'), failed: count('failed'), skipped: count('skipped') },
    scriptErrors: (dto.script_errors ?? []).map(toScriptFailure),
    console: (dto.console ?? []).map(toConsoleEntry),
    sentRequest: {
      method: dto.sent_request.method,
      url: dto.sent_request.url,
      headers: (dto.sent_request.headers ?? []).map(({ key, value }) => ({ key, value })),
      body: dto.sent_request.body,
    },
    error: dto.error,
    responseBodyUrl: dto.response_body_url,
  }
}

function toRequestStats(dto: RunRequestStatsDto): RequestStats {
  return {
    requestId: dto.request_id,
    name: dto.name,
    count: dto.count,
    avgMs: dto.avg_ms,
    p90Ms: dto.p90_ms,
    p95Ms: dto.p95_ms,
    p99Ms: dto.p99_ms,
    minMs: dto.min_ms,
    maxMs: dto.max_ms,
    throughput: dto.throughput,
    errorRate: dto.error_rate,
    statusCounts: Object.entries(dto.status_counts ?? {}).map(([status, count]) => ({ status, count: count ?? 0 })),
  }
}

function toSeriesPoint(dto: RunPerformanceSecondDto): SeriesPoint {
  return { second: dto.second, activeUsers: dto.active_users, requests: dto.requests, avgMs: dto.avg_ms, errors: dto.errors }
}

export function toPerformanceReport(dto: RunPerformanceDto | null): PerformanceReport | null {
  if (!dto) {
    return null
  }
  return { overall: toRequestStats(dto.overall), perRequest: (dto.per_request ?? []).map(toRequestStats), series: (dto.series ?? []).map(toSeriesPoint) }
}

export function toRun(dto: RunDto): Run {
  return {
    id: dto.id,
    target: toRunTarget(dto.target),
    config: toRunConfig(dto.config),
    environmentName: dto.environment_name,
    dataFileName: dto.data_file_name,
    status: oneOf(RUN_STATUSES, dto.status, 'completed'),
    startedAt: dto.started_at,
    finishedAt: dto.finished_at,
    progress: toRunProgress(dto.progress),
    summary: toRunSummary(dto.summary),
    rows: (dto.rows ?? []).map(toRunRowResult),
    performance: toPerformanceReport(dto.performance),
    notices: dto.notices ?? [],
  }
}

export function toRunProgressEvent(dto: RunProgressChangedDto): RunProgressEvent {
  return {
    localId: dto.local_id,
    runId: dto.run_id,
    itemId: dto.item_id,
    type: oneOf(RUN_TYPES, dto.type, 'functional'),
    status: oneOf(RUN_STATUSES, dto.status, 'running'),
    progress: toRunProgress(dto.progress),
  }
}

export function toRunFinishedEvent(dto: RunFinishedDto): RunFinishedEvent {
  return {
    localId: dto.local_id,
    runId: dto.run_id,
    itemId: dto.item_id,
    targetName: dto.target_name,
    type: oneOf(RUN_TYPES, dto.type, 'functional'),
    status: oneOf(RUN_STATUSES, dto.status, 'completed'),
    summary: toRunSummary(dto.summary),
  }
}
