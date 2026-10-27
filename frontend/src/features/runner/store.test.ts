import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/runner', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/platform/services/runner')>()
  return {
    ...actual,
    planRun: vi.fn(),
    saveRunnerPreferences: vi.fn(),
    startRun: vi.fn(),
    pauseRun: vi.fn(),
    resumeRun: vi.fn(),
    stopRun: vi.fn(),
    getRun: vi.fn(),
    listRuns: vi.fn(),
    activeRuns: vi.fn(),
    runResponse: vi.fn(),
    deleteRun: vi.fn(),
    clearRuns: vi.fn(),
    exportRun: vi.fn(),
  }
})
vi.mock('@/platform/services/dialogs', () => ({ pickFile: vi.fn(), pickSaveFile: vi.fn(), pickFolder: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { SAVE_DELAY_MS, useRunnerStore, isRunActive } = await import('./store')
const { planRow, run, runConfig, runPlan, runProgress, runSummary, runSummaryView } = await import('./test-fixtures')

function target(itemId = 'c1') {
  return useRunnerStore.getState().targets[itemId]!
}

async function loaded(plan = runPlan()) {
  vi.mocked(service.planRun).mockResolvedValue(plan)
  await useRunnerStore.getState().ensure('c1')
}

function progressOf(runId: string, progress: Parameters<typeof runProgress>[0] = {}, itemId = 'c1') {
  return { localId: 'L1', runId, itemId, type: 'functional' as const, status: 'running' as const, progress: runProgress(progress) }
}

/** A loaded target whose Results show the stored run R0. */
async function showingStoredRun() {
  await loaded()
  vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R0', status: 'completed' }))
  await useRunnerStore.getState().openRun('c1', 'R0')
}

describe('runner store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads the plan once and keeps the rows, configuration and environments', async () => {
    await loaded()
    await useRunnerStore.getState().ensure('c1')
    expect(service.planRun).toHaveBeenCalledTimes(1)
    expect(service.planRun).toHaveBeenCalledWith('L1', 'c1')
    expect(target()).toMatchObject({ status: 'ready', view: 'setup', rows: [{ requestId: 'r1' }, { requestId: 'r2' }], activeEnvironmentId: 'e1' })
  })

  it('shows a load error', async () => {
    vi.mocked(service.planRun).mockRejectedValue(new AppError('RUN-07404', 'gone'))
    await useRunnerStore.getState().load('c1')
    expect(target()).toMatchObject({ status: 'ready', loadError: { code: 'RUN-07404' } })
  })

  it('restores the active run of the target and its results from the plan', async () => {
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    await loaded(runPlan({ activeRun: runSummaryView() }))
    expect(isRunActive(useRunnerStore.getState(), 'c1')).toBe(true)
    expect(target()).toMatchObject({ runId: 'R1', run: { id: 'R1' } })
  })

  it('remembers row and setting changes once, 300 ms after the last edit', async () => {
    await loaded()
    vi.useFakeTimers()
    const store = useRunnerStore.getState()
    store.setRows('c1', [planRow('r2', 1), planRow('r1', 2, { selected: false })])
    store.updateConfig('c1', (config) => ({ ...config, functional: { ...config.functional, iterations: 3 } }))
    vi.advanceTimersByTime(SAVE_DELAY_MS - 1)
    expect(service.saveRunnerPreferences).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    expect(service.saveRunnerPreferences).toHaveBeenCalledTimes(1)
    const [, , config] = vi.mocked(service.saveRunnerPreferences).mock.calls[0]!
    expect(config.rows).toEqual([
      { requestId: 'r2', aliasId: '', selected: true },
      { requestId: 'r1', aliasId: '', selected: false },
    ])
    expect(config.functional.iterations).toBe(3)
  })

  it('reset drops pending saves', async () => {
    await loaded()
    vi.useFakeTimers()
    useRunnerStore.getState().setRows('c1', [])
    useRunnerStore.getState().reset()
    vi.advanceTimersByTime(SAVE_DELAY_MS)
    expect(service.saveRunnerPreferences).not.toHaveBeenCalled()
    expect(useRunnerStore.getState().targets).toEqual({})
  })

  it('picks a CSV or JSON data file for the run type and clears it', async () => {
    await loaded(runPlan({ dataFileMissing: true }))
    vi.mocked(dialogs.pickFile).mockResolvedValue('/data/users.csv')
    await useRunnerStore.getState().pickDataFile('c1')
    expect(dialogs.pickFile).toHaveBeenCalledWith('Choose Iteration Data', [{ displayName: 'Data files (CSV, JSON)', pattern: '*.csv;*.json' }])
    expect(target().config?.functional.dataFile).toBe('/data/users.csv')
    expect(target().dataFileMissing).toBe(false)
    useRunnerStore.getState().clearDataFile('c1')
    expect(target().config?.functional.dataFile).toBe('')
    useRunnerStore.getState().updateConfig('c1', (config) => ({ ...config, type: 'performance' }))
    await useRunnerStore.getState().pickDataFile('c1')
    expect(target().config?.performance.dataFile).toBe('/data/users.csv')
  })

  it('starts a run with the shown rows and switches to Results', async () => {
    await loaded()
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    await useRunnerStore.getState().start('c1')
    const [localId, itemId, config, acknowledgement] = vi.mocked(service.startRun).mock.calls[0]!
    expect([localId, itemId, acknowledgement]).toEqual(['L1', 'c1', { confirmedEnvironmentId: '', missingSecrets: false }])
    expect(config.rows.map((row) => row.requestId)).toEqual(['r1', 'r2'])
    expect(target()).toMatchObject({ view: 'results', runId: 'R1', run: { id: 'R1' }, starting: false })
    expect(useRunnerStore.getState().active.c1?.id).toBe('R1')
  })

  it('asks to confirm a confirm-before-send environment once, then starts with the confirmation', async () => {
    await loaded()
    vi.mocked(service.startRun).mockRejectedValueOnce(
      new AppError('SND-05409', 'confirm', [
        { attribute: 'environment_id', messages: ['e2'] },
        { attribute: 'name', messages: ['prod'] },
        { attribute: 'color', messages: ['red'] },
      ]),
    )
    await useRunnerStore.getState().start('c1')
    expect(target().confirmation).toEqual({ environmentId: 'e2', name: 'prod', color: 'red' })
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run())
    await useRunnerStore.getState().confirmStart('c1')
    expect(vi.mocked(service.startRun).mock.calls[1]![3]).toEqual({ confirmedEnvironmentId: 'e2', missingSecrets: false })
    expect(target().confirmation).toBeNull()
  })

  it('cancelling the confirmation starts nothing', async () => {
    await loaded()
    vi.mocked(service.startRun).mockRejectedValue(new AppError('SND-05409', 'confirm', [{ attribute: 'environment_id', messages: ['e2'] }]))
    await useRunnerStore.getState().start('c1')
    useRunnerStore.getState().dismissConfirmation('c1')
    expect(target().confirmation).toBeNull()
    expect(service.startRun).toHaveBeenCalledTimes(1)
  })

  it('runs anyway with missing secrets acknowledged, keeping a confirmation already given', async () => {
    await loaded()
    vi.mocked(service.startRun).mockRejectedValueOnce(new AppError('SND-06409', 'secrets', [{ attribute: 'missing_secrets', messages: ['token'] }]))
    await useRunnerStore.getState().start('c1', { acknowledgement: { confirmedEnvironmentId: 'e2' } })
    expect(target().missingSecrets).toEqual({ missing: ['token'], locked: [] })
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run())
    await useRunnerStore.getState().runAnyway('c1')
    expect(vi.mocked(service.startRun).mock.calls[1]![3]).toEqual({ confirmedEnvironmentId: 'e2', missingSecrets: true })
  })

  it('shows invalid settings inline on Setup and other refusals as a failure', async () => {
    await loaded()
    useRunnerStore.getState().showView('c1', 'results')
    vi.mocked(service.startRun).mockRejectedValueOnce(new AppError('RUN-04400', 'bad', [{ attribute: 'iterations', messages: ['must be at least 1'] }]))
    await useRunnerStore.getState().start('c1')
    expect(target()).toMatchObject({ view: 'setup', fieldErrors: { iterations: 'must be at least 1' }, failure: null })
    useRunnerStore.getState().updateConfig('c1', (config) => config)
    expect(target().fieldErrors).toEqual({})
    vi.mocked(service.startRun).mockRejectedValueOnce(new AppError('RUN-02409', 'busy'))
    await useRunnerStore.getState().start('c1')
    expect(target().failure).toMatchObject({ error: { code: 'RUN-02409' }, nextStep: 'Wait for the active run to finish, or stop it.' })
  })

  it('does not start twice while a run is active', async () => {
    await loaded()
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    await useRunnerStore.getState().start('c1')
    expect(service.startRun).not.toHaveBeenCalled()
  })

  it('pauses, resumes and stops the active run', async () => {
    await loaded()
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    vi.mocked(service.pauseRun).mockResolvedValue(runSummaryView({ status: 'paused' }))
    vi.mocked(service.resumeRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.stopRun).mockRejectedValue(new AppError('RUN-01404', 'gone'))
    await useRunnerStore.getState().pause('c1')
    expect(service.pauseRun).toHaveBeenCalledWith('L1', 'R1')
    expect(useRunnerStore.getState().active.c1?.status).toBe('paused')
    await useRunnerStore.getState().resume('c1')
    expect(useRunnerStore.getState().active.c1?.status).toBe('running')
    await useRunnerStore.getState().stop('c1')
    expect(target().failure?.error.code).toBe('RUN-01404')
  })

  it('applies progress to the indicator and reads the shown functional run again', async () => {
    await loaded()
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    useRunnerStore.getState().showRunner('c1')
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'paused' }))
    useRunnerStore.getState().applyProgress({ localId: 'L1', runId: 'R1', itemId: 'c1', type: 'functional', status: 'paused', progress: runProgress({ completed: 2 }) })
    expect(useRunnerStore.getState().active.c1).toMatchObject({ status: 'paused', progress: { completed: 2 } })
    await vi.waitFor(() => expect(target().run?.status).toBe('paused'))
    expect(target().view).toBe('results')
  })

  it('ignores progress of another workspace and adopts a run started elsewhere', async () => {
    await loaded()
    useRunnerStore.getState().applyProgress({ localId: 'L2', runId: 'R9', itemId: 'c1', type: 'functional', status: 'running', progress: runProgress() })
    expect(useRunnerStore.getState().active).toEqual({})
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R9' }))
    useRunnerStore.getState().applyProgress({ localId: 'L1', runId: 'R9', itemId: 'c1', type: 'functional', status: 'running', progress: runProgress() })
    expect(useRunnerStore.getState().active.c1).toMatchObject({ id: 'R9', targetName: 'Orders API' })
    expect(target().runId).toBe('R9')
  })

  it('follows a run of the target started elsewhere', async () => {
    await showingStoredRun()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    expect(target()).toMatchObject({ runId: 'R2', view: 'results' })
    await vi.waitFor(() => expect(target().run?.id).toBe('R2'))
    expect(service.getRun).toHaveBeenLastCalledWith('L1', 'R2')
  })

  it('follows a new run once; an older run opened during it stays shown', async () => {
    await showingStoredRun()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    await vi.waitFor(() => expect(target().run?.id).toBe('R2'))
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R0', status: 'completed' }))
    await useRunnerStore.getState().openRun('c1', 'R0')
    useRunnerStore.getState().applyProgress(progressOf('R2', { completed: 2 }))
    expect(target()).toMatchObject({ runId: 'R0', run: { id: 'R0' } })
    expect(useRunnerStore.getState().active.c1).toMatchObject({ id: 'R2', progress: { completed: 2 } })
  })

  it.each(['setup', 'runs'] as const)('follows a run started elsewhere from %s', async (view) => {
    await loaded()
    vi.mocked(service.listRuns).mockResolvedValue([])
    useRunnerStore.getState().showView('c1', view)
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    expect(target()).toMatchObject({ runId: 'R2', view: 'results' })
  })

  it('progress of another target leaves the tab alone', async () => {
    await showingStoredRun()
    useRunnerStore.getState().applyProgress(progressOf('R2', {}, 'c2'))
    expect(target()).toMatchObject({ runId: 'R0', run: { id: 'R0' } })
  })

  it('progress after the run finished does not switch the tab', async () => {
    await showingStoredRun()
    useRunnerStore.getState().applyFinished({ localId: 'L1', runId: 'R2', itemId: 'c1', targetName: 'Orders API', type: 'functional', status: 'completed', summary: runSummary() })
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    expect(target()).toMatchObject({ runId: 'R0', run: { id: 'R0' } })
    expect(isRunActive(useRunnerStore.getState(), 'c1')).toBe(false)
  })

  it('pauses, resumes and stops the followed run', async () => {
    await showingStoredRun()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    vi.mocked(service.pauseRun).mockResolvedValue(runSummaryView({ id: 'R2', status: 'paused' }))
    vi.mocked(service.resumeRun).mockResolvedValue(runSummaryView({ id: 'R2' }))
    vi.mocked(service.stopRun).mockResolvedValue(runSummaryView({ id: 'R2', status: 'stopped' }))
    await useRunnerStore.getState().pause('c1')
    await useRunnerStore.getState().resume('c1')
    await useRunnerStore.getState().stop('c1')
    expect(service.pauseRun).toHaveBeenCalledWith('L1', 'R2')
    expect(service.resumeRun).toHaveBeenCalledWith('L1', 'R2')
    expect(service.stopRun).toHaveBeenCalledWith('L1', 'R2')
  })

  it("reads a followed run again when it finishes and shows the notice", async () => {
    await showingStoredRun()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    await vi.waitFor(() => expect(target().run?.id).toBe('R2'))
    vi.mocked(service.getRun).mockClear()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2', status: 'completed' }))
    useRunnerStore.getState().applyFinished({ localId: 'L1', runId: 'R2', itemId: 'c1', targetName: 'Orders API', type: 'functional', status: 'completed', summary: runSummary() })
    expect(service.getRun).toHaveBeenCalledWith('L1', 'R2')
    expect(useRunnerStore.getState().notice?.runId).toBe('R2')
    await vi.waitFor(() => expect(target().run?.status).toBe('completed'))
  })

  it('reads the Runs list again when a run it lacks reports progress', async () => {
    await loaded()
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ id: 'R0', status: 'completed' })])
    await useRunnerStore.getState().loadRuns('c1')
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ id: 'R2' }), runSummaryView({ id: 'R0', status: 'completed' })])
    useRunnerStore.getState().applyProgress(progressOf('R2'))
    await vi.waitFor(() => expect(target().runs).toHaveLength(2))
    useRunnerStore.getState().applyProgress(progressOf('R2', { completed: 2 }))
    expect(service.listRuns).toHaveBeenCalledTimes(2)
  })

  it('Clear runs keeps the run in progress', async () => {
    await loaded()
    useRunnerStore.setState({ active: { c1: runSummaryView({ id: 'R2' }) } })
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ id: 'R2' }), runSummaryView({ id: 'R0', status: 'completed' })])
    await useRunnerStore.getState().loadRuns('c1')
    await useRunnerStore.getState().clearRuns('c1')
    expect(target().runs?.map((listed) => listed.id)).toEqual(['R2'])
  })

  it('on finish clears the indicator, shows the notice, and reloads the shown run and the Runs list', async () => {
    await loaded()
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    useRunnerStore.getState().showRunner('c1')
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ status: 'completed' })])
    await useRunnerStore.getState().loadRuns('c1')
    vi.mocked(service.getRun).mockResolvedValue(run())
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ status: 'completed' }), runSummaryView({ id: 'R0', status: 'stopped' })])
    useRunnerStore.getState().applyFinished({ localId: 'L1', runId: 'R1', itemId: 'c1', targetName: 'Orders API', type: 'functional', status: 'completed', summary: runSummary() })
    expect(isRunActive(useRunnerStore.getState(), 'c1')).toBe(false)
    expect(useRunnerStore.getState().notice).toEqual({ itemId: 'c1', runId: 'R1', kind: 'collection', targetName: 'Orders API', status: 'completed', passed: 3, failed: 1 })
    await vi.waitFor(() => expect(target().runs).toHaveLength(2))
    expect(target().run?.status).toBe('completed')
  })

  it('does not bring back the indicator of a run that ended before Start answered', async () => {
    await loaded()
    let answer: (value: ReturnType<typeof runSummaryView>) => void = () => undefined
    vi.mocked(service.startRun).mockReturnValue(new Promise((resolve) => (answer = resolve)))
    vi.mocked(service.getRun).mockResolvedValue(run())
    const starting = useRunnerStore.getState().start('c1')
    useRunnerStore.getState().applyFinished({ localId: 'L1', runId: 'R1', itemId: 'c1', targetName: 'Orders API', type: 'functional', status: 'completed', summary: runSummary() })
    answer(runSummaryView())
    await starting
    expect(isRunActive(useRunnerStore.getState(), 'c1')).toBe(false)
  })

  it('opens, deletes and clears stored runs', async () => {
    await loaded()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R0' }))
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ id: 'R0', status: 'completed' }), runSummaryView({ id: 'R2', status: 'stopped' })])
    useRunnerStore.getState().showView('c1', 'runs')
    await vi.waitFor(() => expect(target().runs).toHaveLength(2))
    await useRunnerStore.getState().openRun('c1', 'R0')
    expect(target()).toMatchObject({ view: 'results', runId: 'R0', run: { id: 'R0' } })
    await useRunnerStore.getState().deleteRun('c1', 'R0')
    expect(service.deleteRun).toHaveBeenCalledWith('L1', 'R0')
    expect(target()).toMatchObject({ runs: [{ id: 'R2' }], runId: null, run: null })
    await useRunnerStore.getState().clearRuns('c1')
    expect(service.clearRuns).toHaveBeenCalledWith('L1', 'c1')
    expect(target().runs).toEqual([])
  })

  it('exports the shown run to the chosen file and skips a cancelled dialog', async () => {
    await loaded()
    vi.mocked(service.getRun).mockResolvedValue(run())
    await useRunnerStore.getState().openRun('c1', 'R1')
    vi.mocked(dialogs.pickSaveFile).mockResolvedValueOnce('/out/run.xml').mockResolvedValueOnce(null)
    await useRunnerStore.getState().exportRun('c1', 'junit')
    expect(dialogs.pickSaveFile).toHaveBeenCalledWith('Export Run', 'Orders API run.xml')
    expect(service.exportRun).toHaveBeenCalledWith('L1', 'R1', 'junit', '/out/run.xml')
    await useRunnerStore.getState().exportRun('c1', 'json')
    expect(service.exportRun).toHaveBeenCalledTimes(1)
  })

  it('runs again with the shown run configuration', async () => {
    await loaded()
    const stored = run({ config: runConfig({ environmentId: 'none', rows: [{ requestId: 'r2', aliasId: 'a1', selected: true }] }) })
    vi.mocked(service.getRun).mockResolvedValue(stored)
    await useRunnerStore.getState().openRun('c1', 'R1')
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView({ id: 'R5' }))
    await useRunnerStore.getState().runAgain('c1')
    expect(vi.mocked(service.startRun).mock.calls[0]![2]).toEqual(stored.config)
    expect(target().runId).toBe('R5')
  })

  it('loads a persisted response into the response viewer, or notes that it is gone', async () => {
    const result = { statusCode: 200, bodyUrl: '/body' } as SendResult
    vi.mocked(service.runResponse).mockResolvedValueOnce(result).mockRejectedValueOnce(new AppError('RUN-01404', 'gone'))
    await useRunnerStore.getState().loadResponse('R1', 0)
    expect(useRunnerStore.getState().responses['run:R1:0']).toBe('ready')
    expect(useResponseViewerStore.getState().entries['run:R1:0']?.result).toMatchObject({ statusCode: 200, tabId: 'run:R1:0' })
    await useRunnerStore.getState().loadResponse('R1', 0)
    expect(service.runResponse).toHaveBeenCalledTimes(1)
    await useRunnerStore.getState().loadResponse('R1', 1)
    expect(useRunnerStore.getState().responses['run:R1:1']).toBe('missing')
  })

  it('restores the active runs of the workspace', async () => {
    vi.mocked(service.activeRuns).mockResolvedValue([runSummaryView({ itemId: 'f1', targetKind: 'folder' })])
    await useRunnerStore.getState().restoreActive()
    expect(service.activeRuns).toHaveBeenCalledWith('L1')
    expect(Object.keys(useRunnerStore.getState().active)).toEqual(['f1'])
  })

  it('Run… shows Results while a run is active, else Setup', () => {
    useRunnerStore.getState().showRunner('c1')
    expect(target().view).toBe('setup')
    useRunnerStore.setState({ active: { f1: runSummaryView({ itemId: 'f1' }) } })
    useRunnerStore.getState().showRunner('f1')
    expect(target('f1')).toMatchObject({ view: 'results', runId: 'R1', status: 'idle' })
  })
})

describe('runner store refreshAll', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  function deferred<T>() {
    let resolve!: (value: T) => void
    const promise = new Promise<T>((done) => {
      resolve = done
    })
    return { promise, resolve }
  }

  const added = runPlan({
    rows: [planRow('r1', 1, { name: 'Sign in', method: 'POST', folderPath: ['Auth'] }), planRow('r2', 2), planRow('r3', 3, { folderPath: ['Payments'] })],
    activeEnvironmentId: 'e2',
    dataFileMissing: true,
  })

  function rowIds(): string[] {
    return target().rows.map((row) => row.requestId)
  }

  it('reads every target shown this session again and takes the reconciled rows, settings and environments', async () => {
    await loaded()
    vi.mocked(service.planRun).mockResolvedValue(added)
    await useRunnerStore.getState().refreshAll()
    expect(service.planRun).toHaveBeenCalledTimes(2)
    expect(service.planRun).toHaveBeenLastCalledWith('L1', 'c1')
    expect(rowIds()).toEqual(['r1', 'r2', 'r3'])
    expect(target().rows[0]).toMatchObject({ number: 1, name: 'Sign in' })
    expect(target().rows[2]).toMatchObject({ selected: true, folderPath: ['Payments'] })
    expect(target()).toMatchObject({ status: 'ready', loadError: null, activeEnvironmentId: 'e2', dataFileMissing: true, config: added.config, target: added.target })
  })

  it('leaves targets that were never read alone', async () => {
    useRunnerStore.getState().showRunner('c1')
    await useRunnerStore.getState().refreshAll()
    expect(service.planRun).not.toHaveBeenCalled()
    expect(target()).toMatchObject({ status: 'idle', rows: [] })
  })

  it('keeps the shown view and run', async () => {
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'completed' }))
    await loaded(runPlan({ activeRun: runSummaryView() }))
    useRunnerStore.getState().showView('c1', 'results')
    vi.mocked(service.planRun).mockResolvedValue(added)
    await useRunnerStore.getState().refreshAll()
    expect(target()).toMatchObject({ view: 'results', runId: 'R1', run: { id: 'R1' } })
    expect(useRunnerStore.getState().active).toEqual({})
  })

  it('remembers a pending edit before reading the rows again', async () => {
    await loaded()
    vi.useFakeTimers()
    vi.mocked(service.saveRunnerPreferences).mockResolvedValue(undefined)
    useRunnerStore.getState().setRows('c1', [planRow('r2', 1), planRow('r1', 2, { selected: false })])
    vi.mocked(service.planRun).mockResolvedValue(runPlan({ rows: [planRow('r2', 1), planRow('r1', 2, { selected: false }), planRow('r3', 3)] }))
    await useRunnerStore.getState().refreshAll()
    expect(service.saveRunnerPreferences).toHaveBeenCalledTimes(1)
    const [, , config] = vi.mocked(service.saveRunnerPreferences).mock.calls[0]!
    expect(config.rows).toEqual([
      { requestId: 'r2', aliasId: '', selected: true },
      { requestId: 'r1', aliasId: '', selected: false },
    ])
    expect(vi.mocked(service.saveRunnerPreferences).mock.invocationCallOrder[0]!).toBeLessThan(vi.mocked(service.planRun).mock.invocationCallOrder[1]!)
    vi.advanceTimersByTime(SAVE_DELAY_MS)
    expect(service.saveRunnerPreferences).toHaveBeenCalledTimes(1)
    expect(rowIds()).toEqual(['r2', 'r1', 'r3'])
    expect(target().rows[1]?.selected).toBe(false)
  })

  it('drops a read the user edited during, remembers the edit and reads again', async () => {
    await loaded()
    vi.mocked(service.saveRunnerPreferences).mockResolvedValue(undefined)
    const stale = deferred<typeof added>()
    vi.mocked(service.planRun).mockReturnValueOnce(stale.promise).mockResolvedValueOnce(runPlan({ rows: [planRow('r2', 1, { selected: false }), planRow('r1', 2), planRow('r3', 3)] }))
    const refreshed = useRunnerStore.getState().refreshAll()
    await vi.waitFor(() => expect(service.planRun).toHaveBeenCalledTimes(2))
    useRunnerStore.getState().setRows('c1', [planRow('r2', 1, { selected: false }), planRow('r1', 2)])
    stale.resolve(added)
    await refreshed
    expect(service.saveRunnerPreferences).toHaveBeenCalledTimes(1)
    expect(vi.mocked(service.saveRunnerPreferences).mock.calls[0]![2].rows[0]).toEqual({ requestId: 'r2', aliasId: '', selected: false })
    expect(service.planRun).toHaveBeenCalledTimes(3)
    expect(rowIds()).toEqual(['r2', 'r1', 'r3'])
    expect(target().rows[0]?.selected).toBe(false)
  })

  it('coalesces a burst of tree changes into one read in flight and one trailing read', async () => {
    await loaded()
    const first = deferred<typeof added>()
    vi.mocked(service.planRun).mockReturnValueOnce(first.promise).mockResolvedValueOnce(added)
    const refreshes = [useRunnerStore.getState().refreshAll()]
    await vi.waitFor(() => expect(service.planRun).toHaveBeenCalledTimes(2))
    refreshes.push(useRunnerStore.getState().refreshAll(), useRunnerStore.getState().refreshAll())
    first.resolve(runPlan({ rows: [planRow('r1', 1)] }))
    await Promise.all(refreshes)
    expect(service.planRun).toHaveBeenCalledTimes(3)
    expect(rowIds()).toEqual(['r1', 'r2', 'r3'])
  })

  it('shows the load error instead of the rows when the read fails', async () => {
    await loaded()
    vi.mocked(service.planRun).mockRejectedValue(new AppError('RUN-07404', 'gone'))
    await useRunnerStore.getState().refreshAll()
    expect(target()).toMatchObject({ config: null, rows: [], target: null, loadError: { code: 'RUN-07404' } })
  })

  it('ignores a read that ends after the workspace changed', async () => {
    await loaded()
    const late = deferred<typeof added>()
    vi.mocked(service.planRun).mockReturnValueOnce(late.promise)
    const refreshed = useRunnerStore.getState().refreshAll()
    await vi.waitFor(() => expect(service.planRun).toHaveBeenCalledTimes(2))
    useWorkspacesStore.setState({ activeLocalId: 'L2' })
    late.resolve(added)
    await refreshed
    expect(rowIds()).toEqual(['r1', 'r2'])
  })

  it('ignores a read that ends after the store was reset', async () => {
    await loaded()
    const late = deferred<typeof added>()
    vi.mocked(service.planRun).mockReturnValueOnce(late.promise)
    const refreshed = useRunnerStore.getState().refreshAll()
    await vi.waitFor(() => expect(service.planRun).toHaveBeenCalledTimes(2))
    useRunnerStore.getState().reset()
    late.resolve(added)
    await refreshed
    expect(useRunnerStore.getState().targets).toEqual({})
  })

  it('keeps the newest read when a slow first read ends after a refresh', async () => {
    const slow = deferred<typeof added>()
    vi.mocked(service.planRun).mockReturnValueOnce(slow.promise).mockResolvedValueOnce(added)
    const first = useRunnerStore.getState().load('c1')
    await useRunnerStore.getState().refreshAll()
    slow.resolve(runPlan())
    await first
    expect(rowIds()).toEqual(['r1', 'r2', 'r3'])
  })

  it('starts the next run with the refreshed rows', async () => {
    await loaded()
    vi.mocked(service.planRun).mockResolvedValue(runPlan({ rows: [planRow('r2', 1, { aliasId: 'a1' }), planRow('r3', 2, { selected: false })] }))
    await useRunnerStore.getState().refreshAll()
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    await useRunnerStore.getState().start('c1')
    const [, , config] = vi.mocked(service.startRun).mock.calls[0]!
    expect(config.rows).toEqual([
      { requestId: 'r2', aliasId: 'a1', selected: true },
      { requestId: 'r3', aliasId: '', selected: false },
    ])
  })
})
