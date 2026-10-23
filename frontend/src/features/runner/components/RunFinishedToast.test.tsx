import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const { useRunnerStore } = await import('@/features/runner/store')
const { RUN_TOAST_MS, RunFinishedToast } = await import('./RunFinishedToast')

const notice = { itemId: 'c1', runId: 'R1', kind: 'collection', targetName: 'Orders API', status: 'stopped_on_error', passed: 10, failed: 1 } as const

describe('RunFinishedToast', () => {
  beforeEach(() => {
    useRunnerStore.getState().reset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('names the target, the status and the test totals, and opens the results', async () => {
    const onOpenResults = vi.fn()
    useRunnerStore.setState({ notice })
    render(<RunFinishedToast onOpenResults={onOpenResults} />)
    expect(screen.getByRole('status')).toHaveTextContent('Orders API run stopped on error · 10 passed · 1 failed')
    await userEvent.click(screen.getByRole('button', { name: 'Open results' }))
    expect(onOpenResults).toHaveBeenCalledWith(notice)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('closes itself after 8 seconds, or from ×', async () => {
    vi.useFakeTimers()
    useRunnerStore.setState({ notice: { ...notice, status: 'completed' } })
    render(<RunFinishedToast onOpenResults={vi.fn()} />)
    expect(screen.getByRole('status')).toHaveTextContent('Orders API run completed')
    act(() => vi.advanceTimersByTime(RUN_TOAST_MS - 1))
    expect(screen.getByRole('status')).toBeInTheDocument()
    act(() => vi.advanceTimersByTime(1))
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    vi.useRealTimers()
    act(() => useRunnerStore.setState({ notice }))
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(useRunnerStore.getState().notice).toBeNull()
    expect(RUN_TOAST_MS).toBe(8000)
  })
})
