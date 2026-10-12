import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { fieldConflict, pingChange, requestConflict, syncSession, syncStatus, usersChange } from './test-fixtures'

vi.mock('@/platform/services/sync', () => ({
  getSyncStatus: vi.fn(),
  linkRepository: vi.fn(),
  listChanges: vi.fn(),
  commitChanges: vi.fn(),
  pushWorkspace: vi.fn(),
  pullWorkspace: vi.fn(),
  saveSyncSettings: vi.fn(),
  getSyncSession: vi.fn(),
  resolveConflict: vi.fn(),
  finishPull: vi.fn(),
  cancelPull: vi.fn(),
  cloneAsWorkspace: vi.fn(),
  unlinkRepository: vi.fn(),
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

  it('opens the conflict view when a pull finds conflicts and keeps the panel otherwise', async () => {
    useSyncStore.getState().openPanel()
    vi.mocked(service.pullWorkspace).mockResolvedValue({ status: syncStatus({ state: 'conflict' }), session: syncSession() })
    await useSyncStore.getState().pull()
    expect(useSyncStore.getState()).toMatchObject({
      status: { state: 'conflict' },
      session: { id: 's1' },
      selectedConflictId: 'r1',
      conflictsOpen: true,
      panelOpen: false,
      changes: [pingChange, usersChange],
    })
    vi.mocked(service.pullWorkspace).mockResolvedValue({ status: syncStatus(), session: null })
    useSyncStore.setState({ panelOpen: true })
    await useSyncStore.getState().pull()
    expect(useSyncStore.getState()).toMatchObject({ session: null, selectedConflictId: null, conflictsOpen: false, panelOpen: true })
  })

  it('opens the session of a push whose replay conflicts', async () => {
    vi.mocked(service.pushWorkspace).mockResolvedValue(syncStatus({ state: 'conflict' }))
    vi.mocked(service.getSyncSession).mockResolvedValue(syncSession([fieldConflict, requestConflict()]))
    await useSyncStore.getState().push()
    expect(service.getSyncSession).toHaveBeenCalledWith('L')
    expect(useSyncStore.getState()).toMatchObject({ conflictsOpen: true, selectedConflictId: 'c1' })
    vi.mocked(service.pushWorkspace).mockResolvedValue(syncStatus())
    await useSyncStore.getState().push()
    expect(service.getSyncSession).toHaveBeenCalledTimes(1)
  })

  it('reopens the conflict view from the session, or the panel when there is none', async () => {
    vi.mocked(service.getSyncSession).mockResolvedValue(syncSession([fieldConflict, requestConflict()]))
    useSyncStore.setState({ selectedConflictId: 'r1' })
    await useSyncStore.getState().openConflicts()
    expect(useSyncStore.getState()).toMatchObject({ conflictsOpen: true, selectedConflictId: 'r1' })
    useSyncStore.getState().closeConflicts()
    expect(useSyncStore.getState()).toMatchObject({ conflictsOpen: false, session: { id: 's1' } })
    vi.mocked(service.getSyncSession).mockResolvedValue(null)
    await useSyncStore.getState().openConflicts()
    expect(useSyncStore.getState()).toMatchObject({ conflictsOpen: false, session: null, panelOpen: true })
  })

  it('resolves conflicts in any order, then finishes back to the panel', async () => {
    await useSyncStore.getState().resolve('r1', 'mine')
    await useSyncStore.getState().finishPull()
    await useSyncStore.getState().cancelPull()
    expect(service.resolveConflict).not.toHaveBeenCalled()
    useSyncStore.setState({ session: syncSession([fieldConflict, requestConflict()]), conflictsOpen: true })
    useSyncStore.getState().selectConflict('r1')
    vi.mocked(service.resolveConflict).mockResolvedValue(requestConflict({ resolution: 'both' }))
    await useSyncStore.getState().resolve('r1', 'both')
    expect(service.resolveConflict).toHaveBeenCalledWith('s1', 'r1', 'both', '')
    vi.mocked(service.resolveConflict).mockResolvedValue({ ...fieldConflict, resolution: 'merged' })
    await useSyncStore.getState().resolve('c1', 'theirs', 'description')
    expect(service.resolveConflict).toHaveBeenLastCalledWith('s1', 'c1', 'theirs', 'description')
    expect(useSyncStore.getState().session?.conflicts.map((conflict) => conflict.resolution)).toEqual(['merged', 'both'])
    expect(useSyncStore.getState().selectedConflictId).toBe('r1')
    vi.mocked(service.finishPull).mockResolvedValue(syncStatus({ state: 'ahead', ahead: 1 }))
    await useSyncStore.getState().finishPull()
    expect(service.finishPull).toHaveBeenCalledWith('s1')
    expect(service.listChanges).toHaveBeenCalledWith('L')
    expect(useSyncStore.getState()).toMatchObject({ session: null, conflictsOpen: false, panelOpen: true, status: { state: 'ahead' }, busy: null })
  })

  it('cancels the pull and keeps a failed resolution as the error', async () => {
    useSyncStore.setState({ session: syncSession(), conflictsOpen: true })
    vi.mocked(service.resolveConflict).mockRejectedValue(new AppError('GIT-16404', 'the pull is not open any more'))
    await useSyncStore.getState().resolve('r1', 'mine')
    expect(useSyncStore.getState()).toMatchObject({ error: { code: 'GIT-16404' }, session: { conflicts: [{ resolution: 'unresolved' }] } })
    vi.mocked(service.cancelPull).mockResolvedValue(syncStatus({ state: 'diverged' }))
    await useSyncStore.getState().cancelPull()
    expect(service.cancelPull).toHaveBeenCalledWith('s1')
    expect(useSyncStore.getState()).toMatchObject({ session: null, conflictsOpen: false, panelOpen: true, status: { state: 'diverged' }, error: null })
  })

  it('drops a resolution answered for a session that was replaced meanwhile', async () => {
    useSyncStore.setState({ session: syncSession() })
    let answer: (conflict: ReturnType<typeof requestConflict>) => void = () => undefined
    vi.mocked(service.resolveConflict).mockReturnValue(new Promise((resolve) => (answer = resolve)))
    const resolving = useSyncStore.getState().resolve('r1', 'mine')
    useSyncStore.setState({ session: { ...syncSession(), id: 's2' } })
    answer(requestConflict({ resolution: 'mine' }))
    await resolving
    expect(useSyncStore.getState().session).toMatchObject({ id: 's2', conflicts: [{ resolution: 'unresolved' }] })
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
    await useSyncStore.getState().saveSettings({ accountId: 'a2' })
    expect(service.saveSyncSettings).not.toHaveBeenCalled()
    useSyncStore.setState({ status: syncStatus({ settings: { ...syncStatus().settings, accountId: '', autoPush: true } }) })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus())
    await useSyncStore.getState().saveSettings({ accountId: 'a2' })
    expect(service.saveSyncSettings).toHaveBeenCalledWith('L', 'a2', false, true)
    expect(useSyncStore.getState().status?.settings.accountId).toBe('a1')
  })

  it('saves the automatic sync options and keeps the other settings', async () => {
    useSyncStore.setState({ status: syncStatus() })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus())
    await useSyncStore.getState().saveSettings({ autoPull: true })
    expect(service.saveSyncSettings).toHaveBeenLastCalledWith('L', 'a1', true, false)
    await useSyncStore.getState().saveSettings({ autoPush: true })
    expect(service.saveSyncSettings).toHaveBeenLastCalledWith('L', 'a1', false, true)
  })

  it('unlinks into a folder, then reloads and reactivates the workspace', async () => {
    const load = vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    useSyncStore.setState({ status: syncStatus(), changes: [pingChange], message: 'wip' })
    vi.mocked(service.unlinkRepository).mockResolvedValue({ localId: 'L', name: 'Acme', path: '/plain', missing: false })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus({ state: 'not_linked' }))
    await expect(useSyncStore.getState().unlink('/plain')).resolves.toBeNull()
    expect(service.unlinkRepository).toHaveBeenCalledWith('L', '/plain')
    expect(load).toHaveBeenCalled()
    expect(activate).toHaveBeenCalledWith('L')
    expect(useSyncStore.getState()).toMatchObject({ changes: [], message: '', status: { state: 'not_linked' } })
    vi.mocked(service.unlinkRepository).mockRejectedValue(new AppError('GIT-43409', 'the folder is not empty'))
    await expect(useSyncStore.getState().unlink('/busy')).resolves.toMatchObject({ code: 'GIT-43409' })
    expect(useSyncStore.getState().error?.code).toBe('GIT-43409')
    load.mockRestore()
    activate.mockRestore()
  })

  it('clones one repository at a time and reports a failed clone', async () => {
    const input = { repository: 'r', branch: '', subfolder: 'api', accountId: 'a1', clonePath: '', name: '' }
    let finish: (summary: { localId: string; name: string; path: string; missing: boolean }) => void = () => undefined
    vi.mocked(service.cloneAsWorkspace).mockImplementation(() => new Promise((resolve) => (finish = resolve)))
    const load = vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    useSyncStore.getState().openClone()
    const first = useSyncStore.getState().cloneWorkspace(input)
    expect(useSyncStore.getState().cloning).toBe(true)
    await expect(useSyncStore.getState().cloneWorkspace(input)).resolves.toBeNull()
    finish({ localId: 'B', name: 'Backend', path: '/b', missing: false })
    await expect(first).resolves.toBeNull()
    expect(service.cloneAsWorkspace).toHaveBeenCalledTimes(1)
    expect(activate).toHaveBeenCalledWith('B')
    expect(useSyncStore.getState()).toMatchObject({ cloning: false, cloneOpen: false })
    vi.mocked(service.cloneAsWorkspace).mockRejectedValue(new AppError('GIT-65409', 'the folder is not empty'))
    await expect(useSyncStore.getState().cloneWorkspace(input)).resolves.toMatchObject({ code: 'GIT-65409' })
    expect(useSyncStore.getState().cloning).toBe(false)
    load.mockRestore()
    activate.mockRestore()
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
