import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { idleUpdateStatus } from '@/platform/mappers/update-status'
import { updateStatus } from '@/features/updates/test-fixtures'

vi.mock('@/platform/services/updates', () => ({
  getUpdateStatus: vi.fn(),
  checkForUpdates: vi.fn(),
  checkOnStartup: vi.fn(),
  startUpdate: vi.fn(),
  restartToUpdate: vi.fn(),
  openReleasePage: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(), releaseResponse: vi.fn() }))

const service = await import('@/platform/services/updates')
const { useUpdatesStore } = await import('@/features/updates/store')
const { UpdateReadyToast } = await import('./UpdateReadyToast')

function push(patch: Parameters<typeof updateStatus>[0]) {
  act(() => useUpdatesStore.getState().applyEvent(updateStatus(patch)))
}

function toast() {
  return screen.queryByRole('status')
}

describe('UpdateReadyToast', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useUpdatesStore.setState({ status: idleUpdateStatus(), toast: 'unseen' })
  })

  it('shows when an automatic download is ready, and stays until answered', () => {
    vi.useFakeTimers()
    try {
      render(<UpdateReadyToast />)
      expect(toast()).not.toBeInTheDocument()
      push({ state: 'downloading', automatic: true })
      expect(toast()).not.toBeInTheDocument()
      push({ state: 'ready', automatic: true })
      expect(toast()).toHaveTextContent('Xenon 0.1.7 is ready to install.')
      expect(toast()).toHaveAttribute('aria-live', 'polite')
      act(() => vi.advanceTimersByTime(10 * 60 * 1000))
      expect(toast()).toBeInTheDocument()
    } finally {
      vi.useRealTimers()
    }
  })

  it('never shows for a failed automatic download or a manual one', () => {
    render(<UpdateReadyToast />)
    push({ state: 'failed', reason: 'download_failed', automatic: true })
    expect(toast()).not.toBeInTheDocument()
    push({ state: 'ready', automatic: false })
    expect(toast()).not.toBeInTheDocument()
  })

  it('restarts into the update from Restart now', async () => {
    vi.mocked(service.restartToUpdate).mockResolvedValue(updateStatus({ state: 'installing' }))
    render(<UpdateReadyToast />)
    push({ state: 'ready', automatic: true })
    await userEvent.click(screen.getByRole('button', { name: 'Restart now' }))
    expect(service.restartToUpdate).toHaveBeenCalledTimes(1)
    expect(toast()).not.toBeInTheDocument()
  })

  it.each(['Later', 'Close'])('closes on %s and does not come back in this run', async (name) => {
    render(<UpdateReadyToast />)
    push({ state: 'ready', automatic: true })
    await userEvent.click(screen.getByRole('button', { name }))
    expect(toast()).not.toBeInTheDocument()
    expect(service.restartToUpdate).not.toHaveBeenCalled()
    push({ state: 'ready', automatic: true })
    expect(toast()).not.toBeInTheDocument()
  })
})
