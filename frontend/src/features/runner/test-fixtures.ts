import type { Run, RunConfig, RunPlan, RunPlanRow, RunProgress, RunRowResult, RunSummary, RunSummaryView } from '@/platform/mappers/runner'
import type { RunnerTarget } from './model'

export function planRow(requestId: string, number: number, extra: Partial<RunPlanRow> = {}): RunPlanRow {
  return { number, requestId, aliasId: '', selected: true, name: `Request ${requestId}`, method: 'GET', folderPath: [], aliases: [], ...extra }
}

export function runConfig(extra: Partial<RunConfig> = {}): RunConfig {
  return {
    type: 'functional',
    rows: [],
    functional: {
      iterations: 1,
      dataFile: '',
      delayMs: 0,
      persistResponses: false,
      disableLogs: false,
      stopOnError: false,
      keepVariableValues: true,
      skipStoredCookies: false,
      saveCookies: false,
    },
    performance: {
      virtualUsers: 10,
      durationSeconds: 600,
      dataFile: '',
      profile: { kind: 'fixed', initialUsers: 0, rampSeconds: 0, baseUsers: 0, spikeSeconds: 0, peakSeconds: 0 },
    },
    environmentId: '',
    ...extra,
  }
}

export function runPlan(extra: Partial<RunPlan> = {}): RunPlan {
  return {
    target: { itemId: 'c1', kind: 'collection', name: 'Orders API' },
    rows: [planRow('r1', 1, { name: 'Login', method: 'POST', folderPath: ['Auth'] }), planRow('r2', 2, { name: 'Get order' })],
    config: runConfig(),
    activeEnvironmentId: 'e1',
    environments: [
      { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null },
      { id: 'e2', name: 'prod', color: 'red', confirmBeforeSend: true, error: null },
    ],
    activeRun: null,
    dataFileMissing: false,
    ...extra,
  }
}

export function runProgress(extra: Partial<RunProgress> = {}): RunProgress {
  return { currentRow: 1, currentName: 'Login', iteration: 0, completed: 1, planned: 2, testsPassed: 1, testsFailed: 0, elapsedMs: 1200, metrics: null, ...extra }
}

export function runSummary(extra: Partial<RunSummary> = {}): RunSummary {
  return {
    iterations: 1,
    totalRequests: 2,
    testsPassed: 3,
    testsFailed: 1,
    testsSkipped: 0,
    avgResponseMs: 120,
    durationMs: 1840,
    passed: 1,
    failed: 1,
    skipped: 0,
    errors: 0,
    cancelled: 0,
    ...extra,
  }
}

export function runSummaryView(extra: Partial<RunSummaryView> = {}): RunSummaryView {
  return {
    id: 'R1',
    itemId: 'c1',
    targetKind: 'collection',
    targetName: 'Orders API',
    type: 'functional',
    status: 'running',
    environmentName: 'dev',
    startedAt: '2026-10-23T10:00:00Z',
    finishedAt: '',
    progress: runProgress(),
    summary: runSummary(),
    ...extra,
  }
}

export function runRow(index: number, extra: Partial<RunRowResult> = {}): RunRowResult {
  return {
    index,
    iteration: 0,
    requestId: `r${index + 1}`,
    aliasId: '',
    aliasName: '',
    name: `Request ${index + 1}`,
    method: 'GET',
    folderPath: [],
    url: `https://api.test/${index + 1}`,
    status: 200,
    durationMs: 50,
    sizeBytes: 128,
    outcome: 'passed',
    tests: [],
    testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 },
    scriptErrors: [],
    console: [],
    sentRequest: { method: 'GET', url: `https://api.test/${index + 1}`, headers: [], body: '' },
    error: '',
    responseBodyUrl: '',
    ...extra,
  }
}

export function run(extra: Partial<Run> = {}): Run {
  return {
    id: 'R1',
    target: { itemId: 'c1', kind: 'collection', name: 'Orders API' },
    config: runConfig(),
    environmentName: 'dev',
    dataFileName: '',
    status: 'completed',
    startedAt: '2026-10-23T10:00:00Z',
    finishedAt: '2026-10-23T10:00:02Z',
    progress: runProgress({ completed: 2 }),
    summary: runSummary(),
    rows: [runRow(0, { name: 'Login', method: 'POST' }), runRow(1, { name: 'Get order', outcome: 'failed', status: 500 })],
    performance: null,
    notices: [],
    ...extra,
  }
}

/** A loaded Runner target as the store holds it after `load`. */
export function readyTarget(plan: RunPlan = runPlan(), extra: Partial<RunnerTarget> = {}): RunnerTarget {
  return {
    status: 'ready',
    loadError: null,
    target: plan.target,
    rows: plan.rows,
    config: plan.config,
    environments: plan.environments,
    activeEnvironmentId: plan.activeEnvironmentId,
    dataFileMissing: plan.dataFileMissing,
    view: 'setup',
    runId: null,
    run: null,
    runs: null,
    failure: null,
    fieldErrors: {},
    confirmation: null,
    missingSecrets: null,
    acknowledgement: { confirmedEnvironmentId: '', missingSecrets: false },
    pendingConfig: null,
    starting: false,
    ...extra,
  }
}
