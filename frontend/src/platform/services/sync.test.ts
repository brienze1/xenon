import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/syncservice', () => ({
  Status: vi.fn(),
  Link: vi.fn(),
  Changes: vi.fn(),
  Commit: vi.fn(),
  Push: vi.fn(),
  Pull: vi.fn(),
  SaveSettings: vi.fn(),
  Session: vi.fn(),
  Resolve: vi.fn(),
  FinishPull: vi.fn(),
  CancelPull: vi.fn(),
  CloneAsWorkspace: vi.fn(),
  Unlink: vi.fn(),
}))

const binding = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/syncservice')
const service = await import('./sync')
const { AppError } = await import('../errors')
const { emptyDraft, toRequestDto } = await import('../mappers/request')

const settingsDto = { repository: 'https://git.test/r.git', branch: 'main', subfolder: 'api', account_id: 'a1', clone_path: '/c', auto_pull: false, auto_push: true }
const statusDto = { state: 'ahead', local_changes: 2, ahead: 1, behind: 0, settings: settingsDto }
const status = {
  state: 'ahead',
  localChanges: 2,
  ahead: 1,
  behind: 0,
  lastSyncAt: null,
  lastError: null,
  notice: '',
  settings: { repository: 'https://git.test/r.git', branch: 'main', subfolder: 'api', accountId: 'a1', clonePath: '/c', autoPull: false, autoPush: true },
}

const draft = { ...emptyDraft(), id: 'r1', name: 'Users', method: 'PATCH', url: 'https://a.test/users' }
const conflictDto = {
  item_id: 'i1',
  item_name: 'Users',
  item_kind: 'request',
  kind: 'both_modified',
  resolution: 'unresolved',
  path: 'u.yaml',
  uncommitted: true,
  mine: { text: 'method: PATCH', request: toRequestDto(draft) },
  theirs: {
    text: 'name: Dev',
    environment: { id: 'e', name: 'Dev', description: '', color: 'none', confirm_before_send: false, variables: [{ key: 'token', value: '', description: '', secret: true, enabled: true }] },
  },
  fields: null,
}
const conflict = {
  itemId: 'i1',
  itemName: 'Users',
  itemKind: 'request',
  kind: 'both_modified',
  resolution: 'unresolved',
  path: 'u.yaml',
  uncommitted: true,
  base: null,
  mine: { text: 'method: PATCH', request: draft, environment: null },
  theirs: { text: 'name: Dev', request: null, environment: { name: 'Dev', color: 'none', variables: [{ key: 'token', value: '', description: '', secret: true, enabled: true }] } },
  fields: [],
}
const fieldDto = {
  item_id: 'c1',
  item_name: 'API',
  item_kind: 'collection',
  kind: 'field',
  resolution: 'unresolved',
  path: 'collection.yaml',
  uncommitted: false,
  base: { text: 'description: d' },
  fields: [{ path: 'description', mine: 'Mine', theirs: null, resolution: 'theirs' }],
}

describe('sync service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('reads the status with its last error, sync time and notice', async () => {
    vi.mocked(binding.Status).mockResolvedValue({
      ...statusDto,
      state: 'error',
      last_sync_at: '2026-10-11T12:00:00.000Z',
      last_error: { code: 'GIT-10503', message: 'offline', next_step: 'Try again' },
      notice: 'changed',
    } as never)
    await expect(service.getSyncStatus('L')).resolves.toEqual({
      ...status,
      state: 'error',
      lastSyncAt: '2026-10-11T12:00:00.000Z',
      lastError: { code: 'GIT-10503', message: 'offline', nextStep: 'Try again' },
      notice: 'changed',
    })
    expect(binding.Status).toHaveBeenCalledWith({ local_id: 'L' })
  })

  it('links, commits, pushes and saves the settings', async () => {
    vi.mocked(binding.Link).mockResolvedValue(statusDto as never)
    vi.mocked(binding.Commit).mockResolvedValue(statusDto as never)
    vi.mocked(binding.Push).mockResolvedValue(statusDto as never)
    vi.mocked(binding.SaveSettings).mockResolvedValue(undefined as never)
    await expect(
      service.linkRepository({ localId: 'L', repository: 'r', branch: 'main', subfolder: 'api', accountId: 'a1', clonePath: '', mode: null }),
    ).resolves.toEqual(status)
    expect(binding.Link).toHaveBeenCalledWith({ local_id: 'L', repository: 'r', branch: 'main', subfolder: 'api', account_id: 'a1', clone_path: '', mode: '' })
    await service.linkRepository({ localId: 'L', repository: 'r', branch: 'main', subfolder: '', accountId: 'a1', clonePath: '/c', mode: 'adopt_repo' })
    expect(binding.Link).toHaveBeenLastCalledWith(expect.objectContaining({ mode: 'adopt_repo', clone_path: '/c' }))
    await expect(service.commitChanges('L', 'msg', ['i1'])).resolves.toEqual(status)
    expect(binding.Commit).toHaveBeenCalledWith({ local_id: 'L', message: 'msg', item_ids: ['i1'] })
    await expect(service.pushWorkspace('L')).resolves.toEqual(status)
    await service.saveSyncSettings('L', 'a2', true, false)
    expect(binding.SaveSettings).toHaveBeenCalledWith({ local_id: 'L', account_id: 'a2', auto_pull: true, auto_push: false })
  })

  it('clones a repository as a workspace and unlinks one', async () => {
    const summaryDto = { local_id: 'B', name: 'Backend', path: '/c/api', missing: false }
    const summary = { localId: 'B', name: 'Backend', path: '/c/api', missing: false }
    vi.mocked(binding.CloneAsWorkspace).mockResolvedValue(summaryDto as never)
    vi.mocked(binding.Unlink).mockResolvedValue(summaryDto as never)
    await expect(service.cloneAsWorkspace({ repository: 'r', branch: '', subfolder: 'api', accountId: 'a1', clonePath: '/c', name: 'Backend' })).resolves.toEqual(summary)
    expect(binding.CloneAsWorkspace).toHaveBeenCalledWith({ repository: 'r', branch: '', subfolder: 'api', account_id: 'a1', clone_path: '/c', name: 'Backend' })
    await expect(service.unlinkRepository('B', '/plain')).resolves.toEqual(summary)
    expect(binding.Unlink).toHaveBeenCalledWith({ local_id: 'B', target_folder: '/plain' })
    vi.mocked(binding.Unlink).mockResolvedValue(null as never)
    await expect(service.unlinkRepository('B', '/plain')).rejects.toBeInstanceOf(AppError)
  })

  it('lists changes and pull results with conflicts', async () => {
    vi.mocked(binding.Changes).mockResolvedValue([{ item_id: 'i1', item_name: 'Ping', item_kind: 'request', change: 'modified', paths: null }] as never)
    await expect(service.listChanges('L')).resolves.toEqual([{ itemId: 'i1', itemName: 'Ping', itemKind: 'request', change: 'modified', paths: [] }])
    vi.mocked(binding.Changes).mockResolvedValue(null as never)
    await expect(service.listChanges('L')).resolves.toEqual([])
    vi.mocked(binding.Pull).mockResolvedValue({
      status: { ...statusDto, state: 'conflict' },
      session: { id: 's', local_id: 'L', conflicts: [conflictDto] },
    } as never)
    await expect(service.pullWorkspace('L')).resolves.toEqual({ status: { ...status, state: 'conflict' }, session: { id: 's', localId: 'L', conflicts: [conflict] } })
    vi.mocked(binding.Pull).mockResolvedValue({ status: statusDto } as never)
    await expect(service.pullWorkspace('L')).resolves.toEqual({ status, session: null })
  })

  it('reads the open session and resolves, finishes and cancels it', async () => {
    vi.mocked(binding.Session).mockResolvedValue(null as never)
    await expect(service.getSyncSession('L')).resolves.toBeNull()
    vi.mocked(binding.Session).mockResolvedValue({ id: 's', local_id: 'L', conflicts: null } as never)
    await expect(service.getSyncSession('L')).resolves.toEqual({ id: 's', localId: 'L', conflicts: [] })
    expect(binding.Session).toHaveBeenCalledWith({ local_id: 'L' })
    vi.mocked(binding.Resolve).mockResolvedValue({ ...fieldDto, resolution: 'merged' } as never)
    await expect(service.resolveConflict('s', 'c1', 'theirs', 'description')).resolves.toMatchObject({ resolution: 'merged', fields: [{ path: 'description', mine: 'Mine', theirs: null }] })
    expect(binding.Resolve).toHaveBeenCalledWith({ session_id: 's', item_id: 'c1', resolution: 'theirs', field_path: 'description' })
    vi.mocked(binding.FinishPull).mockResolvedValue(statusDto as never)
    await expect(service.finishPull('s')).resolves.toEqual(status)
    expect(binding.FinishPull).toHaveBeenCalledWith({ session_id: 's' })
    vi.mocked(binding.CancelPull).mockResolvedValue({ ...statusDto, state: 'diverged' } as never)
    await expect(service.cancelPull('s')).resolves.toEqual({ ...status, state: 'diverged' })
    expect(binding.CancelPull).toHaveBeenCalledWith({ session_id: 's' })
    vi.mocked(binding.Resolve).mockRejectedValue({ cause: { error: { code: 'GIT-19400', description: 'no' } } })
    await expect(service.resolveConflict('s', 'c1', 'both', '')).rejects.toMatchObject({ code: 'GIT-19400' })
  })

  it('turns failures and empty answers into AppErrors', async () => {
    vi.mocked(binding.Push).mockRejectedValue({ cause: { error: { code: 'GIT-05409', description: 'busy' } } })
    await expect(service.pushWorkspace('L')).rejects.toMatchObject({ code: 'GIT-05409' })
    vi.mocked(binding.Status).mockResolvedValue(null as never)
    await expect(service.getSyncStatus('L')).rejects.toBeInstanceOf(AppError)
  })
})
