import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  pauseRun: vi.fn(),
  resumeRun: vi.fn(),
  stopRun: vi.fn(),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { runProgress, runSummaryView } = await import('@/features/runner/test-fixtures')
const { RunProgress } = await import('./RunProgress')

describe('RunProgress', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows the status, current request and iteration, the bar, counts and elapsed time, with Pause and Stop', async () => {
    const run = runSummaryView({ progress: runProgress({ currentName: 'Get order', iteration: 1, completed: 3, planned: 4, testsPassed: 5, testsFailed: 2, elapsedMs: 1500 }) })
    useRunnerStore.setState({ active: { c1: run } })
    vi.mocked(service.pauseRun).mockResolvedValue(runSummaryView({ status: 'paused' }))
    vi.mocked(service.stopRun).mockResolvedValue(runSummaryView({ status: 'stopped' }))
    render(<RunProgress itemId="c1" run={run} />)
    expect(screen.getByRole('status')).toHaveTextContent('Running')
    expect(screen.getByText(/Get order/).parentElement).toHaveTextContent('Get order · iteration 2')
    expect(screen.getByRole('progressbar', { name: 'Completed requests' })).toHaveAttribute('aria-valuenow', '3')
    expect(screen.getByText('3 of 4 requests')).toBeInTheDocument()
    expect(screen.getByText('5 passed')).toBeInTheDocument()
    expect(screen.getByText('2 failed')).toBeInTheDocument()
    expect(screen.getByText(/1[.,]50\ss/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Pause' }))
    expect(service.pauseRun).toHaveBeenCalledWith('L1', 'R1')
    await userEvent.click(screen.getByRole('button', { name: 'Stop' }))
    expect(service.stopRun).toHaveBeenCalledWith('L1', 'R1')
  })

  it('offers Resume while paused', async () => {
    const run = runSummaryView({ status: 'paused' })
    useRunnerStore.setState({ active: { c1: run } })
    vi.mocked(service.resumeRun).mockResolvedValue(runSummaryView())
    render(<RunProgress itemId="c1" run={run} />)
    expect(screen.getByRole('status')).toHaveTextContent('Paused')
    await userEvent.click(screen.getByRole('button', { name: 'Resume' }))
    expect(service.resumeRun).toHaveBeenCalledWith('L1', 'R1')
  })

  it('offers only Stop for a performance run and shows its live metrics', () => {
    const run = runSummaryView({ type: 'performance', progress: runProgress({ metrics: { activeUsers: 7, requests: 1234, avgMs: 42, throughput: 12.5, errorRate: 0.02 } }) })
    render(<RunProgress itemId="c1" run={run} />)
    expect(screen.queryByRole('button', { name: 'Pause' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Stop' })).toBeInTheDocument()
    expect(screen.getByText('Active users').nextSibling).toHaveTextContent('7')
    expect(screen.getByText('Throughput').nextSibling).toHaveTextContent('12.5 req/s')
    expect(screen.getByText('Error rate').nextSibling).toHaveTextContent('2.0%')
  })
})
