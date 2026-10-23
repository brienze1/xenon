import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { RunnerSlots } from '@/features/runner/model'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  planRun: vi.fn(),
  saveRunnerPreferences: vi.fn(async () => undefined),
  startRun: vi.fn(),
  getRun: vi.fn(),
  listRuns: vi.fn(),
  pauseRun: vi.fn(),
  stopRun: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFile: vi.fn(), pickSaveFile: vi.fn(), pickFolder: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { run, runPlan, runProgress, runSummaryView } = await import('@/features/runner/test-fixtures')
const { RunnerSection } = await import('./RunnerSection')

const slots: RunnerSlots = { tests: () => null, response: () => null }

const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight')

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, get: () => 600 })
})

afterAll(() => {
  if (offsetHeight) {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', offsetHeight)
  }
})

function confirmError() {
  return new AppError('SND-05409', 'confirm', [
    { attribute: 'environment_id', messages: ['e2'] },
    { attribute: 'name', messages: ['prod'] },
    { attribute: 'color', messages: ['red'] },
  ])
}

async function renderSection() {
  render(<RunnerSection itemId="c1" slots={slots} />)
  return screen.findByRole('tablist', { name: 'Runner' })
}

describe('RunnerSection', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
    vi.mocked(service.planRun).mockResolvedValue(runPlan())
    vi.mocked(service.saveRunnerPreferences).mockResolvedValue()
  })

  it('opens on Setup with the run order and Start run, and has Setup, Results and Runs sub-tabs', async () => {
    const tabs = await renderSection()
    expect(within(tabs).getAllByRole('tab').map((tab) => tab.textContent)).toEqual(['Setup', 'Results', 'Runs'])
    expect(within(tabs).getByRole('tab', { name: 'Setup' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByText('2 of 2 selected')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Start run' })).toBeEnabled()
    expect(service.planRun).toHaveBeenCalledWith('L1', 'c1')
  })

  it('starts a run, switches to Results with live progress, then shows the results', async () => {
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running', rows: [] }))
    await renderSection()
    await userEvent.click(screen.getByRole('button', { name: 'Start run' }))
    expect(await screen.findByRole('region', { name: 'Run progress' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Results' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('button', { name: 'Pause' })).toBeInTheDocument()
    vi.mocked(service.getRun).mockResolvedValue(run())
    act(() =>
      useRunnerStore.getState().applyFinished({ localId: 'L1', runId: 'R1', itemId: 'c1', targetName: 'Orders API', type: 'functional', status: 'completed', summary: run().summary }),
    )
    expect(await screen.findByLabelText('Run summary')).toHaveTextContent('Completed')
    expect(screen.queryByRole('region', { name: 'Run progress' })).not.toBeInTheDocument()
  })

  it('asks once before running with a confirm-before-send environment; Cancel starts nothing', async () => {
    vi.mocked(service.startRun).mockRejectedValueOnce(confirmError())
    await renderSection()
    await userEvent.click(screen.getByRole('button', { name: 'Start run' }))
    const dialog = await screen.findByRole('dialog', { name: 'Run with prod?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(service.startRun).toHaveBeenCalledTimes(1)
    vi.mocked(service.startRun).mockRejectedValueOnce(confirmError()).mockResolvedValueOnce(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    await userEvent.click(screen.getByRole('button', { name: 'Start run' }))
    await userEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Run with prod' }))
    expect(vi.mocked(service.startRun).mock.calls[2]![3]).toEqual({ confirmedEnvironmentId: 'e2', missingSecrets: false })
  })

  it('runs anyway with missing secrets', async () => {
    vi.mocked(service.startRun)
      .mockRejectedValueOnce(new AppError('SND-06409', 'secrets', [{ attribute: 'missing_secrets', messages: ['api_key'] }]))
      .mockResolvedValueOnce(runSummaryView())
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    await renderSection()
    await userEvent.click(screen.getByRole('button', { name: 'Start run' }))
    const dialog = await screen.findByRole('dialog', { name: 'Run Without Some Secrets?' })
    expect(dialog).toHaveTextContent('{{api_key}}')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Run Anyway' }))
    expect(vi.mocked(service.startRun).mock.calls[1]![3]).toEqual({ confirmedEnvironmentId: '', missingSecrets: true })
  })

  it('shows a refused start as a dismissible failure', async () => {
    vi.mocked(service.startRun).mockRejectedValue(new AppError('RUN-02409', 'a run is already active on this collection or folder'))
    await renderSection()
    await userEvent.click(screen.getByRole('button', { name: 'Start run' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('a run is already active on this collection or folder')
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows Results empty before any run, and loads the stored runs on Runs', async () => {
    vi.mocked(service.listRuns).mockResolvedValue([runSummaryView({ status: 'completed' })])
    await renderSection()
    await userEvent.click(screen.getByRole('tab', { name: 'Results' }))
    expect(screen.getByRole('region', { name: 'No run yet' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('tab', { name: 'Runs' }))
    expect(await screen.findByRole('region', { name: 'Stored runs' })).toBeInTheDocument()
    expect(service.listRuns).toHaveBeenCalledWith('L1', 'c1')
  })

  it('opens on Results with the progress of a run already active on the target', async () => {
    vi.mocked(service.planRun).mockResolvedValue(runPlan({ activeRun: runSummaryView({ progress: runProgress({ currentName: 'Get order' }) }) }))
    vi.mocked(service.getRun).mockResolvedValue(run({ status: 'running' }))
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    useRunnerStore.getState().showRunner('c1')
    await renderSection()
    expect(screen.getByRole('tab', { name: 'Results' })).toHaveAttribute('aria-selected', 'true')
    expect(await screen.findByRole('region', { name: 'Run progress' })).toHaveTextContent('Get order')
  })

  it('shows a load error with Try Again', async () => {
    vi.mocked(service.planRun).mockRejectedValueOnce(new AppError('RUN-07404', 'no collection or folder with the id c1'))
    render(<RunnerSection itemId="c1" slots={slots} />)
    expect(await screen.findByRole('alert')).toHaveTextContent('no collection or folder with the id c1')
    await userEvent.click(screen.getByRole('button', { name: 'Try Again' }))
    expect(await screen.findByRole('tablist', { name: 'Runner' })).toBeInTheDocument()
  })
})
