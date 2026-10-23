import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  getRun: vi.fn(),
  deleteRun: vi.fn(async () => undefined),
  clearRuns: vi.fn(async () => undefined),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { readyTarget, run, runPlan, runSummary, runSummaryView } = await import('@/features/runner/test-fixtures')
const { RunsList } = await import('./RunsList')

const stored = [
  runSummaryView({ id: 'R2', status: 'stopped_on_error', type: 'functional', environmentName: 'dev', startedAt: '2026-10-23T11:00:00Z', summary: runSummary({ totalRequests: 6, testsPassed: 10, testsFailed: 1, testsSkipped: 2, durationMs: 1840 }) }),
  runSummaryView({ id: 'R1', status: 'completed', type: 'performance', environmentName: '', startedAt: '2026-10-23T10:00:00Z' }),
]

function seed(runs: typeof stored | null = stored) {
  useRunnerStore.setState({ targets: { c1: readyTarget(runPlan(), { view: 'runs', runs }) } })
}

describe('RunsList', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('lists the stored runs newest first with their type, status, environment, totals and duration', () => {
    seed()
    render(<RunsList itemId="c1" />)
    const items = within(screen.getByRole('region', { name: 'Stored runs' })).getAllByRole('listitem')
    expect(items).toHaveLength(2)
    expect(items[0]).toHaveTextContent('Functional')
    expect(items[0]).toHaveTextContent('Stopped on error')
    expect(items[0]).toHaveTextContent('dev')
    expect(items[0]).toHaveTextContent('6 requests')
    expect(items[0]).toHaveTextContent('10 passed · 1 failed · 2 skipped')
    expect(items[0]).toHaveTextContent('1.84')
    expect(items[1]).toHaveTextContent('Performance')
    expect(items[1]).toHaveTextContent('No environment')
  })

  it('opens a run in Results', async () => {
    seed()
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2' }))
    render(<RunsList itemId="c1" />)
    const [first] = within(screen.getByRole('region', { name: 'Stored runs' })).getAllByRole('listitem')
    await userEvent.click(within(first!).getAllByRole('button')[0]!)
    expect(useRunnerStore.getState().targets.c1).toMatchObject({ view: 'results', runId: 'R2' })
  })

  it('deletes a run from its ⋯ menu', async () => {
    seed()
    render(<RunsList itemId="c1" />)
    const [first] = within(screen.getByRole('region', { name: 'Stored runs' })).getAllByRole('listitem')
    await userEvent.click(within(first!).getByRole('button', { name: /More actions for the run of/ }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    expect(service.deleteRun).toHaveBeenCalledWith('L1', 'R2')
    expect(within(screen.getByRole('region', { name: 'Stored runs' })).getAllByRole('listitem')).toHaveLength(1)
  })

  it('clears every run after a confirmation, and Cancel keeps them', async () => {
    seed()
    render(<RunsList itemId="c1" />)
    await userEvent.click(screen.getByRole('button', { name: 'Clear runs' }))
    const dialog = screen.getByRole('dialog', { name: 'Clear Stored Runs?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(service.clearRuns).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Clear runs' }))
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Clear runs' }))
    expect(service.clearRuns).toHaveBeenCalledWith('L1', 'c1')
    expect(await screen.findByRole('region', { name: 'No stored runs' })).toBeInTheDocument()
  })

  it('shows loading until the runs are read', () => {
    seed(null)
    render(<RunsList itemId="c1" />)
    expect(screen.getByRole('status')).toHaveTextContent('Loading runs…')
  })
})
