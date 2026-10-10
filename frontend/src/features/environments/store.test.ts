import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { EnvironmentList } from '@/platform/mappers/environment'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))

const services = await import('@/platform/services/environments')
const { useEnvironmentsStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const prod = { id: 'e2', name: 'prod', color: 'red', confirmBeforeSend: true, error: null } as const
const list: EnvironmentList = { environments: [dev, prod], activeId: 'e1' }

function store() {
  return useEnvironmentsStore.getState()
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => {
    resolve = done
  })
  return { promise, resolve }
}

describe('environments store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
  })

  afterEach(() => useWorkspacesStore.setState({ activeLocalId: null }))

  it('loads the environments and the active one of the active workspace', async () => {
    vi.mocked(services.listEnvironments).mockResolvedValue(list)
    const loading = store().load()
    expect(store().status).toBe('loading')
    await loading
    expect(services.listEnvironments).toHaveBeenCalledWith('A')
    expect(store()).toMatchObject({ environments: [dev, prod], activeId: 'e1', status: 'ready', error: null })
  })

  it('does nothing without an active workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await store().load()
    await store().setActive('e1')
    expect(services.listEnvironments).not.toHaveBeenCalled()
    expect(services.setActiveEnvironment).not.toHaveBeenCalled()
    expect(store().status).toBe('idle')
  })

  it('ignores answers that arrive after a workspace switch', async () => {
    const pending = deferred<EnvironmentList>()
    vi.mocked(services.listEnvironments).mockReturnValue(pending.promise)
    const loading = store().load()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    pending.resolve(list)
    await loading
    expect(store().environments).toEqual([])
    vi.mocked(services.listEnvironments).mockRejectedValue(new AppError('ENV-05500', 'disk'))
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    const failing = store().load()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    await failing
    expect(store().error).toBeNull()
  })

  it('keeps the error of a failed load', async () => {
    vi.mocked(services.listEnvironments).mockRejectedValue(new AppError('ENV-05500', 'disk'))
    await store().load()
    expect(store()).toMatchObject({ status: 'ready', error: { code: 'ENV-05500' } })
    store().clearError()
    expect(store().error).toBeNull()
  })

  it('sets the active environment and "No environment"', async () => {
    useEnvironmentsStore.setState({ environments: [dev, prod], activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockResolvedValueOnce('e2').mockResolvedValueOnce(null)
    await store().setActive('e2')
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e2')
    expect(store().activeId).toBe('e2')
    await store().setActive(null)
    expect(services.setActiveEnvironment).toHaveBeenLastCalledWith('A', null)
    expect(store().activeId).toBeNull()
  })

  it('keeps the previous active environment when setting fails or the workspace changed', async () => {
    useEnvironmentsStore.setState({ environments: [dev, prod], activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockRejectedValue(new AppError('ENV-01404', 'gone'))
    await store().setActive('e2')
    expect(store()).toMatchObject({ activeId: 'e1', error: { code: 'ENV-01404' } })
    const pending = deferred<string | null>()
    vi.mocked(services.setActiveEnvironment).mockReturnValue(pending.promise)
    const setting = store().setActive('e2')
    expect(store().error).toBeNull()
    useWorkspacesStore.setState({ activeLocalId: 'G' })
    pending.resolve('e2')
    await setting
    expect(store().activeId).toBe('e1')
  })

  it('opens and closes the switcher and resets everything', () => {
    useEnvironmentsStore.setState({ environments: [dev], activeId: 'e1', status: 'ready', error: new AppError('X', 'y') })
    store().openSwitcher()
    expect(store().switcherOpen).toBe(true)
    store().setSwitcherOpen(false)
    expect(store().switcherOpen).toBe(false)
    store().openSwitcher()
    store().reset()
    expect(store()).toMatchObject({ environments: [], activeId: null, status: 'idle', error: null, switcherOpen: false })
  })
})
