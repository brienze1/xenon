import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { pingChange, syncStatus, usersChange } from './test-fixtures'

vi.mock('@/platform/services/sync', () => ({
  getSyncStatus: vi.fn(),
  linkRepository: vi.fn(),
  listChanges: vi.fn(),
  commitChanges: vi.fn(),
  pushWorkspace: vi.fn(),
  pullWorkspace: vi.fn(),
  saveSyncSettings: vi.fn(),
}))

const service = await import('@/platform/services/sync')
const { useSyncStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

describe('sync store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(service.listChanges).mockResolvedValue([pingChange, usersChange])
    useSyncStore.getState().reset('L')
  })

  it('refreshes the status of the current workspace only', async () => {
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus({ state: 'behind', behind: 1 }))
    await useSyncStore.getState().refresh()
    expect(useSyncStore.getState().status?.state).toBe('behind')
    useSyncStore.getState().applyStatus('other', syncStatus())
    expect(useSyncStore.getState().status?.state).toBe('behind')
    useSyncStore.getState().reset(null)
    await useSyncStore.getState().refresh()
    expect(service.getSyncStatus).toHaveBeenCalledTimes(1)
  })

  it('drops a status that arrives after a workspace switch', async () => {
    let answer: (status: ReturnType<typeof syncStatus>) => void = () => undefined
    vi.mocked(service.getSyncStatus).mockReturnValue(new Promise((resolve) => (answer = resolve)))
    const refreshing = useSyncStore.getState().refresh()
    useSyncStore.getState().reset('other')
    answer(syncStatus())
    await refreshing
    expect(useSyncStore.getState()).toMatchObject({ localId: 'other', status: null, busy: null })
  })

  it('opens the panel with the changes and commits only the checked ones, then pushes', async () => {
    useSyncStore.setState({ status: syncStatus({ state: 'local_changes', localChanges: 2 }) })
    useSyncStore.getState().openPanel()
    await vi.waitFor(() => expect(useSyncStore.getState().changes).toHaveLength(2))
    useSyncStore.getState().toggle('r2')
    useSyncStore.getState().setMessage('Ping with HEAD')
    vi.mocked(service.commitChanges).mockResolvedValue(syncStatus({ state: 'ahead', ahead: 1 }))
    vi.mocked(service.pushWorkspace).mockResolvedValue(syncStatus())
    vi.mocked(service.listChanges).mockResolvedValue([usersChange])
    await useSyncStore.getState().commit(true)
    expect(service.commitChanges).toHaveBeenCalledWith('L', 'Ping with HEAD', ['r1'])
    expect(service.pushWorkspace).toHaveBeenCalledWith('L')
    expect(useSyncStore.getState()).toMatchObject({ message: '', excluded: [], busy: null, changes: [usersChange], status: { state: 'up_to_date' } })
    useSyncStore.getState().toggle('r2')
    useSyncStore.getState().toggle('r2')
    expect(useSyncStore.getState().excluded).toEqual([])
  })

  it('keeps the error with its next step when an operation fails and refuses a second one meanwhile', async () => {
    let fail: (error: unknown) => void = () => undefined
    vi.mocked(service.pushWorkspace).mockReturnValue(new Promise((_, reject) => (fail = reject)))
    const pushing = useSyncStore.getState().push()
    expect(useSyncStore.getState().busy).toBe('push')
    await useSyncStore.getState().pull()
    expect(service.pullWorkspace).not.toHaveBeenCalled()
    useSyncStore.getState().applyProgress({ localId: 'L', phase: 'push', done: 0, total: 1 })
    useSyncStore.getState().applyProgress({ localId: 'other', phase: 'fetch', done: 0, total: 1 })
    expect(useSyncStore.getState().progress?.phase).toBe('push')
    fail(new AppError('GIT-10503', 'offline', [{ attribute: 'next_step', messages: ['Try again'] }]))
    await pushing
    expect(useSyncStore.getState()).toMatchObject({ busy: null, progress: null, error: { code: 'GIT-10503' } })
    useSyncStore.getState().applyProgress({ localId: 'L', phase: 'late', done: 1, total: 1 })
    expect(useSyncStore.getState().progress).toBeNull()
  })

  it('pulls and keeps the conflicting items', async () => {
    vi.mocked(service.pullWorkspace).mockResolvedValue({
      status: syncStatus({ state: 'conflict' }),
      conflicts: [{ itemId: 'r1', itemName: 'Ping', itemKind: 'request', kind: 'both_modified', path: 'ping.yaml' }],
    })
    await useSyncStore.getState().pull()
    expect(useSyncStore.getState()).toMatchObject({ status: { state: 'conflict' }, conflicts: [{ itemName: 'Ping' }], changes: [pingChange, usersChange] })
  })

  it('links, reopens the moved workspace and shows the panel; a refusal stays with the dialog', async () => {
    const load = vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    useSyncStore.getState().openLink()
    vi.mocked(service.linkRepository).mockRejectedValueOnce(new AppError('GIT-01409', 'has a workspace'))
    const input = { repository: 'r', branch: 'main', subfolder: 'api', accountId: 'a1', clonePath: '', mode: null }
    await expect(useSyncStore.getState().link(input)).resolves.toMatchObject({ code: 'GIT-01409' })
    expect(useSyncStore.getState().linkOpen).toBe(true)
    vi.mocked(service.linkRepository).mockResolvedValue(syncStatus({ state: 'local_changes' }))
    await expect(useSyncStore.getState().link({ ...input, mode: 'replace_with_mine' })).resolves.toBeNull()
    expect(service.linkRepository).toHaveBeenLastCalledWith({ ...input, mode: 'replace_with_mine', localId: 'L' })
    expect(load).toHaveBeenCalled()
    expect(activate).toHaveBeenCalledWith('L')
    expect(useSyncStore.getState()).toMatchObject({ linkOpen: false, panelOpen: true, status: { state: 'local_changes' } })
    useSyncStore.getState().closeLink()
    useSyncStore.getState().closePanel()
    expect(useSyncStore.getState()).toMatchObject({ linkOpen: false, panelOpen: false })
  })

  it('saves the account choice and rereads the status', async () => {
    await useSyncStore.getState().saveSettings('a2')
    expect(service.saveSyncSettings).not.toHaveBeenCalled()
    useSyncStore.setState({ status: syncStatus({ settings: { ...syncStatus().settings, accountId: '', autoPush: true } }) })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus())
    await useSyncStore.getState().saveSettings('a2')
    expect(service.saveSyncSettings).toHaveBeenCalledWith('L', 'a2', false, true)
    expect(useSyncStore.getState().status?.settings.accountId).toBe('a1')
  })

  it('does not list changes of a workspace without a repository and reports a failed listing', async () => {
    useSyncStore.setState({ status: syncStatus({ state: 'not_linked' }) })
    await useSyncStore.getState().loadChanges()
    expect(service.listChanges).not.toHaveBeenCalled()
    useSyncStore.setState({ status: syncStatus() })
    vi.mocked(service.listChanges).mockRejectedValue(new AppError('GIT-09409', 'not linked'))
    await useSyncStore.getState().loadChanges()
    expect(useSyncStore.getState().error?.code).toBe('GIT-09409')
  })
})
