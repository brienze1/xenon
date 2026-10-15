import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { idleUpdateStatus } from '@/platform/mappers/update-status'
import { updateStatus } from './test-fixtures'

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
const { useTabsStore } = await import('@/features/tabs/store')
const { useUpdatesStore } = await import('./store')

function updates() {
  return useUpdatesStore.getState()
}

describe('updates store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useUpdatesStore.setState({ status: idleUpdateStatus(), toast: 'unseen' })
  })

  afterEach(() => vi.restoreAllMocks())

  it('loads, checks and runs the startup check, keeping the returned status', async () => {
    vi.mocked(service.getUpdateStatus).mockResolvedValue(updateStatus({ state: 'idle' }))
    await updates().load()
    expect(updates().status.currentVersion).toBe('0.1.4')
    vi.mocked(service.checkForUpdates).mockResolvedValue(updateStatus({ state: 'up_to_date' }))
    await updates().check()
    expect(updates().status.state).toBe('up_to_date')
    vi.mocked(service.checkOnStartup).mockResolvedValue(updateStatus({ state: 'downloading', automatic: true }))
    await updates().checkOnStartup()
    expect(updates().status.state).toBe('downloading')
  })

  it('never lets a call result overwrite a status pushed while it ran', async () => {
    let resolve: (value: ReturnType<typeof updateStatus>) => void = () => undefined
    vi.mocked(service.checkOnStartup).mockReturnValue(new Promise((done) => (resolve = done)))
    const running = updates().checkOnStartup()
    updates().applyEvent(updateStatus({ state: 'ready', automatic: true }))
    resolve(updateStatus({ state: 'downloading', automatic: true }))
    await running
    expect(updates().status.state).toBe('ready')
  })

  it('keeps the status when a call fails', async () => {
    useUpdatesStore.setState({ status: updateStatus({ state: 'available' }) })
    vi.mocked(service.checkForUpdates).mockRejectedValue(new Error('boom'))
    await updates().check()
    expect(updates().status.state).toBe('available')
  })

  it('opens the release page through Go', async () => {
    vi.mocked(service.openReleasePage).mockResolvedValue(updateStatus({ state: 'available' }))
    await updates().openReleasePage()
    expect(service.openReleasePage).toHaveBeenCalledTimes(1)
  })

  it('saves the tabs UI state before restarting', async () => {
    const order: string[] = []
    vi.spyOn(useTabsStore.getState(), 'flushPersist').mockImplementation(async () => {
      await Promise.resolve()
      order.push('flush')
    })
    vi.mocked(service.restartToUpdate).mockImplementation(async () => {
      order.push('restart')
      return updateStatus({ state: 'installing' })
    })
    await updates().restart()
    expect(order).toEqual(['flush', 'restart'])
    expect(updates().status.state).toBe('installing')
  })

  it('restarts into the update once it is downloaded and verified', async () => {
    const order: string[] = []
    vi.spyOn(useTabsStore.getState(), 'flushPersist').mockImplementation(async () => {
      order.push('flush')
    })
    vi.mocked(service.startUpdate).mockImplementation(async () => {
      order.push('update')
      return updateStatus({ state: 'ready' })
    })
    vi.mocked(service.restartToUpdate).mockImplementation(async () => {
      order.push('restart')
      return updateStatus({ state: 'installing' })
    })
    await updates().update()
    expect(order).toEqual(['update', 'flush', 'restart'])
  })

  it('does not restart when the update fails', async () => {
    const flush = vi.spyOn(useTabsStore.getState(), 'flushPersist').mockResolvedValue()
    vi.mocked(service.startUpdate).mockResolvedValue(updateStatus({ state: 'failed', reason: 'unverified' }))
    await updates().update()
    expect(updates().status).toMatchObject({ state: 'failed', reason: 'unverified' })
    vi.mocked(service.startUpdate).mockRejectedValue(new Error('boom'))
    await updates().update()
    expect(flush).not.toHaveBeenCalled()
    expect(service.restartToUpdate).not.toHaveBeenCalled()
  })

  it('opens the toast once per run, only for an automatic ready status', () => {
    updates().applyEvent(updateStatus({ state: 'failed', reason: 'download_failed', automatic: true }))
    expect(updates().toast).toBe('unseen')
    updates().applyEvent(updateStatus({ state: 'ready', automatic: false }))
    expect(updates().toast).toBe('unseen')
    updates().applyEvent(updateStatus({ state: 'ready', automatic: true }))
    expect(updates().toast).toBe('open')
    updates().closeToast()
    expect(updates().toast).toBe('closed')
    updates().applyEvent(updateStatus({ state: 'ready', automatic: true }))
    expect(updates().toast).toBe('closed')
  })
})
