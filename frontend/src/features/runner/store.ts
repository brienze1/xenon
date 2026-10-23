import { create } from 'zustand'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError, toAppError } from '@/platform/errors'
import {
  runResponseKey,
  type RunConfig,
  type RunExportFormat,
  type RunFinishedEvent,
  type RunPlanRow,
  type RunProgressEvent,
  type RunSummary,
  type RunSummaryView,
} from '@/platform/mappers/runner'
import { pickFile, pickSaveFile } from '@/platform/services/dialogs'
import { confirmationOf, missingSecretsOf } from '@/platform/services/requests'
import {
  activeRuns,
  clearRuns,
  deleteRun,
  EMPTY_SELECTION,
  exportRun,
  getRun,
  listRuns,
  pauseRun,
  planRun,
  resumeRun,
  RUN_ALREADY_ACTIVE,
  runFieldErrorsOf,
  runResponse,
  saveRunnerPreferences,
  startRun,
  stopRun,
  UNKNOWN_RUN,
  type RunAcknowledgement,
} from '@/platform/services/runner'
import { configWithRows } from './lib/run-order'
import type { ResponseStatus, RunFinishedNotice, RunnerFailure, RunnerTarget, RunnerView } from './model'

/** Run order and settings changes are remembered after this pause in editing. */
export const SAVE_DELAY_MS = 300

const SAVE_NEXT_STEP = 'Your changes still apply to the next run; reopen the Runner to try remembering them again.'
const START_NEXT_STEP = 'Check the run settings, then start the run again.'
const CONTROL_NEXT_STEP = 'Try again. If the run already ended, its results are in Runs.'
const RUNS_NEXT_STEP = 'Try again. Stored runs live in the app data folder of this machine.'
const EXPORT_NEXT_STEP = 'Choose a folder you can write to, then export again.'

export const DATA_FILE_FILTERS = [{ displayName: 'Data files (CSV, JSON)', pattern: '*.csv;*.json' }]

const NEXT_STEPS: Record<string, string> = {
  [EMPTY_SELECTION]: 'Check at least one request in the run order.',
  [RUN_ALREADY_ACTIVE]: 'Wait for the active run to finish, or stop it.',
}

const EMPTY_SUMMARY: RunSummary = {
  iterations: 0,
  totalRequests: 0,
  testsPassed: 0,
  testsFailed: 0,
  testsSkipped: 0,
  avgResponseMs: 0,
  durationMs: 0,
  passed: 0,
  failed: 0,
  skipped: 0,
  errors: 0,
  cancelled: 0,
}

interface StartOptions {
  acknowledgement?: Partial<RunAcknowledgement>
  /** Runs this configuration instead of the Setup's (Run again). */
  config?: RunConfig
}

interface RunnerState {
  targets: Record<string, RunnerTarget>
  /** The active runs of the workspace, by collection or folder id (the tab indicators). */
  active: Record<string, RunSummaryView>
  /** Persisted run responses loaded into the response viewer, by `run:<runId>:<index>`. */
  responses: Record<string, ResponseStatus>
  /** The run that ended last, for the finish toast. */
  notice: RunFinishedNotice | null
  /** Loads the target once (when its Runner is first shown). */
  ensure: (itemId: string) => Promise<void>
  load: (itemId: string) => Promise<void>
  showView: (itemId: string, view: RunnerView) => void
  /** Run…: shows Results while a run is active on the target, else Setup. */
  showRunner: (itemId: string) => void
  setRows: (itemId: string, rows: RunPlanRow[]) => void
  updateConfig: (itemId: string, update: (config: RunConfig) => RunConfig) => void
  pickDataFile: (itemId: string) => Promise<void>
  clearDataFile: (itemId: string) => void
  dismissDataFileMissing: (itemId: string) => void
  dismissFailure: (itemId: string) => void
  start: (itemId: string, options?: StartOptions) => Promise<void>
  confirmStart: (itemId: string) => Promise<void>
  dismissConfirmation: (itemId: string) => void
  runAnyway: (itemId: string) => Promise<void>
  dismissMissingSecrets: (itemId: string) => void
  pause: (itemId: string) => Promise<void>
  resume: (itemId: string) => Promise<void>
  stop: (itemId: string) => Promise<void>
  /** Opens a stored (or the active) run in Results. */
  openRun: (itemId: string, runId: string) => Promise<void>
  loadRuns: (itemId: string) => Promise<void>
  deleteRun: (itemId: string, runId: string) => Promise<void>
  clearRuns: (itemId: string) => Promise<void>
  exportRun: (itemId: string, format: RunExportFormat) => Promise<void>
  /** Starts a new run with the shown run's target and configuration. */
  runAgain: (itemId: string) => Promise<void>
  loadResponse: (runId: string, index: number) => Promise<void>
  applyProgress: (event: RunProgressEvent) => void
  applyFinished: (event: RunFinishedEvent) => void
  /** Reads the workspace's active runs (after a reload or a workspace switch). */
  restoreActive: () => Promise<void>
  dismissNotice: () => void
  reset: () => void
}

function blankTarget(view: RunnerView = 'setup'): RunnerTarget {
  return {
    status: 'idle',
    loadError: null,
    target: null,
    rows: [],
    config: null,
    environments: [],
    activeEnvironmentId: null,
    dataFileMissing: false,
    view,
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
  }
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function failureOf(error: AppError, nextStep: string): RunnerFailure {
  return { error, nextStep: NEXT_STEPS[error.code] ?? nextStep }
}

/** Sets the data file of the configuration's run type. */
function withDataFile(config: RunConfig, dataFile: string): RunConfig {
  return config.type === 'performance'
    ? { ...config, performance: { ...config.performance, dataFile } }
    : { ...config, functional: { ...config.functional, dataFile } }
}

const saveTimers = new Map<string, ReturnType<typeof setTimeout>>()
/** Runs whose Results are being read again; a progress event during the read asks for one more. */
const refreshing = new Map<string, boolean>()
/** Runs that ended before their Start call answered, so the answer does not bring their indicator back. */
const finishedRuns = new Set<string>()

function clearTimers(): void {
  for (const timer of saveTimers.values()) {
    clearTimeout(timer)
  }
  saveTimers.clear()
  refreshing.clear()
  finishedRuns.clear()
}

export const useRunnerStore = create<RunnerState>()((set, get) => {
  function patch(itemId: string, update: Partial<RunnerTarget> | ((target: RunnerTarget) => Partial<RunnerTarget>)) {
    set((state) => {
      const target = state.targets[itemId] ?? blankTarget()
      return { targets: { ...state.targets, [itemId]: { ...target, ...(typeof update === 'function' ? update(target) : update) } } }
    })
  }

  function setActive(itemId: string, run: RunSummaryView | null) {
    set((state) => {
      const active = { ...state.active }
      if (run === null) {
        delete active[itemId]
      } else {
        active[itemId] = run
      }
      return { active }
    })
  }

  /** Runs `operation` in the active workspace; null when there is none or the workspace changed meanwhile. */
  async function inWorkspace<T>(operation: (localId: string) => Promise<T>): Promise<{ ok: true; value: T } | { ok: false; error: AppError } | null> {
    const localId = activeLocalId()
    if (localId === null) {
      return null
    }
    try {
      const value = await operation(localId)
      return activeLocalId() === localId ? { ok: true, value } : null
    } catch (error) {
      return activeLocalId() === localId ? { ok: false, error: toAppError(error) } : null
    }
  }

  function scheduleSave(itemId: string) {
    clearTimeout(saveTimers.get(itemId))
    saveTimers.set(
      itemId,
      setTimeout(() => {
        saveTimers.delete(itemId)
        void save(itemId)
      }, SAVE_DELAY_MS),
    )
  }

  async function save(itemId: string) {
    const target = get().targets[itemId]
    if (!target?.config) {
      return
    }
    const config = configWithRows(target.config, target.rows)
    const saved = await inWorkspace((localId) => saveRunnerPreferences(localId, itemId, config))
    if (saved && !saved.ok) {
      patch(itemId, { failure: failureOf(saved.error, SAVE_NEXT_STEP) })
    }
  }

  async function fetchRun(itemId: string, runId: string) {
    const loaded = await inWorkspace((localId) => getRun(localId, runId))
    if (loaded === null || get().targets[itemId]?.runId !== runId) {
      return
    }
    patch(itemId, loaded.ok ? { run: loaded.value } : { failure: failureOf(loaded.error, RUNS_NEXT_STEP) })
  }

  /** Reads the shown run again after progress; at most one read at a time per run, plus one trailing read. */
  async function refreshRun(itemId: string, runId: string) {
    if (refreshing.has(runId)) {
      refreshing.set(runId, true)
      return
    }
    refreshing.set(runId, false)
    try {
      await fetchRun(itemId, runId)
    } finally {
      const again = refreshing.get(runId) === true
      refreshing.delete(runId)
      if (again) {
        void refreshRun(itemId, runId)
      }
    }
  }

  async function control(itemId: string, operation: (localId: string, runId: string) => Promise<RunSummaryView>) {
    const runId = get().active[itemId]?.id
    if (!runId) {
      return
    }
    const result = await inWorkspace((localId) => operation(localId, runId))
    if (result === null) {
      return
    }
    if (!result.ok) {
      patch(itemId, { failure: failureOf(result.error, CONTROL_NEXT_STEP) })
      return
    }
    if (get().active[itemId]?.id === runId) {
      setActive(itemId, result.value)
    }
  }

  return {
    targets: {},
    active: {},
    responses: {},
    notice: null,

    ensure: async (itemId) => {
      const target = get().targets[itemId]
      if (target && target.status !== 'idle') {
        return
      }
      await get().load(itemId)
    },

    load: async (itemId) => {
      patch(itemId, { status: 'loading' })
      const loaded = await inWorkspace((localId) => planRun(localId, itemId))
      if (loaded === null) {
        return
      }
      if (!loaded.ok) {
        patch(itemId, { status: 'ready', loadError: loaded.error })
        return
      }
      const plan = loaded.value
      setActive(itemId, plan.activeRun)
      const runId = get().targets[itemId]?.runId ?? plan.activeRun?.id ?? null
      patch(itemId, {
        status: 'ready',
        loadError: null,
        target: plan.target,
        rows: plan.rows,
        config: plan.config,
        environments: plan.environments,
        activeEnvironmentId: plan.activeEnvironmentId,
        dataFileMissing: plan.dataFileMissing,
        runId,
      })
      if (runId !== null && get().targets[itemId]?.run?.id !== runId) {
        await fetchRun(itemId, runId)
      }
    },

    showView: (itemId, view) => {
      patch(itemId, { view })
      if (view === 'runs' && get().targets[itemId]?.runs == null) {
        void get().loadRuns(itemId)
      }
    },

    showRunner: (itemId) => {
      const active = get().active[itemId]
      patch(itemId, (target) => (active ? { view: 'results', runId: active.id, run: target.run?.id === active.id ? target.run : null } : { view: 'setup' }))
      if (active && get().targets[itemId]?.status !== 'idle') {
        void fetchRun(itemId, active.id)
      }
    },

    setRows: (itemId, rows) => {
      patch(itemId, { rows })
      scheduleSave(itemId)
    },

    updateConfig: (itemId, update) => {
      const config = get().targets[itemId]?.config
      if (!config) {
        return
      }
      patch(itemId, { config: update(config), fieldErrors: {} })
      scheduleSave(itemId)
    },

    pickDataFile: async (itemId) => {
      const picked = await inWorkspace(() => pickFile('Choose Iteration Data', DATA_FILE_FILTERS))
      if (picked === null) {
        return
      }
      if (!picked.ok) {
        patch(itemId, { failure: failureOf(picked.error, START_NEXT_STEP) })
        return
      }
      const path = picked.value
      if (path !== null) {
        get().updateConfig(itemId, (config) => withDataFile(config, path))
        patch(itemId, { dataFileMissing: false })
      }
    },

    clearDataFile: (itemId) => get().updateConfig(itemId, (config) => withDataFile(config, '')),

    dismissDataFileMissing: (itemId) => patch(itemId, { dataFileMissing: false }),

    dismissFailure: (itemId) => patch(itemId, { failure: null }),

    start: async (itemId, options = {}) => {
      const target = get().targets[itemId]
      const config = options.config ?? (target?.config ? configWithRows(target.config, target.rows) : null)
      if (!config || target?.starting || get().active[itemId]) {
        return
      }
      const acknowledgement: RunAcknowledgement = { confirmedEnvironmentId: '', missingSecrets: false, ...options.acknowledgement }
      patch(itemId, { starting: true, failure: null, fieldErrors: {}, confirmation: null, missingSecrets: null, acknowledgement, pendingConfig: config })
      const started = await inWorkspace((localId) => startRun(localId, itemId, config, acknowledgement))
      if (started === null) {
        patch(itemId, { starting: false })
        return
      }
      if (!started.ok) {
        const error = started.error
        const confirmation = confirmationOf(error)
        const missingSecrets = missingSecretsOf(error)
        const fieldErrors = runFieldErrorsOf(error)
        if (confirmation) {
          patch(itemId, { starting: false, confirmation })
        } else if (missingSecrets) {
          patch(itemId, { starting: false, missingSecrets })
        } else if (Object.keys(fieldErrors).length > 0) {
          patch(itemId, { starting: false, fieldErrors, view: 'setup' })
        } else {
          patch(itemId, { starting: false, failure: failureOf(error, START_NEXT_STEP) })
        }
        return
      }
      const run = started.value
      if (!finishedRuns.has(run.id)) {
        setActive(itemId, run)
      }
      patch(itemId, { starting: false, pendingConfig: null, view: 'results', runId: run.id, run: null })
      await fetchRun(itemId, run.id)
    },

    confirmStart: async (itemId) => {
      const target = get().targets[itemId]
      if (!target?.confirmation) {
        return
      }
      const acknowledgement = { confirmedEnvironmentId: target.confirmation.environmentId, missingSecrets: target.acknowledgement.missingSecrets }
      await get().start(itemId, { acknowledgement, config: target.pendingConfig ?? undefined })
    },

    dismissConfirmation: (itemId) => patch(itemId, { confirmation: null, pendingConfig: null }),

    runAnyway: async (itemId) => {
      const target = get().targets[itemId]
      if (!target?.missingSecrets) {
        return
      }
      const acknowledgement = { confirmedEnvironmentId: target.acknowledgement.confirmedEnvironmentId, missingSecrets: true }
      await get().start(itemId, { acknowledgement, config: target.pendingConfig ?? undefined })
    },

    dismissMissingSecrets: (itemId) => patch(itemId, { missingSecrets: null, pendingConfig: null }),

    pause: (itemId) => control(itemId, pauseRun),

    resume: (itemId) => control(itemId, resumeRun),

    stop: (itemId) => control(itemId, stopRun),

    openRun: async (itemId, runId) => {
      patch(itemId, (target) => ({ view: 'results', runId, run: target.run?.id === runId ? target.run : null }))
      await fetchRun(itemId, runId)
    },

    loadRuns: async (itemId) => {
      const loaded = await inWorkspace((localId) => listRuns(localId, itemId))
      if (loaded === null) {
        return
      }
      patch(itemId, loaded.ok ? { runs: loaded.value } : { runs: [], failure: failureOf(loaded.error, RUNS_NEXT_STEP) })
    },

    deleteRun: async (itemId, runId) => {
      const deleted = await inWorkspace((localId) => deleteRun(localId, runId))
      if (deleted === null) {
        return
      }
      if (!deleted.ok) {
        patch(itemId, { failure: failureOf(deleted.error, RUNS_NEXT_STEP) })
        return
      }
      patch(itemId, (target) => ({
        runs: target.runs?.filter((run) => run.id !== runId) ?? null,
        ...(target.runId === runId ? { runId: null, run: null } : {}),
      }))
    },

    clearRuns: async (itemId) => {
      const cleared = await inWorkspace((localId) => clearRuns(localId, itemId))
      if (cleared === null) {
        return
      }
      if (!cleared.ok) {
        patch(itemId, { failure: failureOf(cleared.error, RUNS_NEXT_STEP) })
        return
      }
      const activeId = get().active[itemId]?.id ?? null
      patch(itemId, (target) => ({ runs: [], ...(target.runId !== null && target.runId !== activeId ? { runId: null, run: null } : {}) }))
    },

    exportRun: async (itemId, format) => {
      const run = get().targets[itemId]?.run
      if (!run) {
        return
      }
      const extension = format === 'junit' ? 'xml' : 'json'
      const exported = await inWorkspace(async (localId) => {
        const path = await pickSaveFile('Export Run', `${run.target.name} run.${extension}`)
        if (path !== null) {
          await exportRun(localId, run.id, format, path)
        }
      })
      if (exported && !exported.ok) {
        patch(itemId, { failure: failureOf(exported.error, EXPORT_NEXT_STEP) })
      }
    },

    runAgain: async (itemId) => {
      const run = get().targets[itemId]?.run
      if (run) {
        await get().start(itemId, { config: run.config })
      }
    },

    loadResponse: async (runId, index) => {
      const key = runResponseKey(runId, index)
      const status = get().responses[key]
      if (status === 'loading' || status === 'ready') {
        return
      }
      const setStatus = (next: ResponseStatus) => set((state) => ({ responses: { ...state.responses, [key]: next } }))
      setStatus('loading')
      const loaded = await inWorkspace((localId) => runResponse(localId, runId, index))
      if (loaded === null) {
        setStatus('error')
        return
      }
      if (!loaded.ok) {
        setStatus(loaded.error.code === UNKNOWN_RUN ? 'missing' : 'error')
        return
      }
      useResponseViewerStore.getState().setResult(key, { ...loaded.value, tabId: key })
      setStatus('ready')
    },

    applyProgress: (event) => {
      if (event.localId !== activeLocalId() || finishedRuns.has(event.runId)) {
        return
      }
      const known = get().active[event.itemId]
      const target = get().targets[event.itemId]
      setActive(
        event.itemId,
        known?.id === event.runId
          ? { ...known, status: event.status, progress: event.progress }
          : {
              id: event.runId,
              itemId: event.itemId,
              targetKind: target?.target?.kind ?? 'collection',
              targetName: target?.target?.name ?? '',
              type: event.type,
              status: event.status,
              environmentName: '',
              startedAt: '',
              finishedAt: '',
              progress: event.progress,
              summary: EMPTY_SUMMARY,
            },
      )
      if (!target || target.status === 'idle') {
        return
      }
      if (target.runId === null) {
        patch(event.itemId, { runId: event.runId })
      }
      if ((target.runId ?? event.runId) !== event.runId) {
        return
      }
      if (target.run?.id === event.runId) {
        patch(event.itemId, (current) => (current.run ? { run: { ...current.run, status: event.status, progress: event.progress } } : {}))
      }
      if (event.type === 'functional' || target.run?.id !== event.runId) {
        void refreshRun(event.itemId, event.runId)
      }
    },

    applyFinished: (event) => {
      if (event.localId !== activeLocalId()) {
        return
      }
      finishedRuns.add(event.runId)
      const state = get()
      const kind = state.active[event.itemId]?.targetKind ?? state.targets[event.itemId]?.target?.kind ?? null
      if (state.active[event.itemId]?.id === event.runId) {
        setActive(event.itemId, null)
      }
      set({
        notice: {
          itemId: event.itemId,
          runId: event.runId,
          kind,
          targetName: event.targetName,
          status: event.status,
          passed: event.summary.testsPassed,
          failed: event.summary.testsFailed,
        },
      })
      const target = get().targets[event.itemId]
      if (!target || target.status === 'idle') {
        return
      }
      if (target.runId === event.runId) {
        void fetchRun(event.itemId, event.runId)
      }
      if (target.runs !== null) {
        void get().loadRuns(event.itemId)
      }
    },

    restoreActive: async () => {
      const loaded = await inWorkspace((localId) => activeRuns(localId))
      if (loaded?.ok) {
        set({ active: Object.fromEntries(loaded.value.map((run) => [run.itemId, run])) })
      }
    },

    dismissNotice: () => set({ notice: null }),

    reset: () => {
      clearTimers()
      set({ targets: {}, active: {}, responses: {}, notice: null })
    },
  }
})

/** Whether a run is active on the collection or folder (its tab shows the running indicator). */
export function isRunActive(state: Pick<RunnerState, 'active'>, itemId: string): boolean {
  return state.active[itemId] !== undefined
}
