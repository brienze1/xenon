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
}))

const binding = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/syncservice')
const service = await import('./sync')
const { AppError } = await import('../errors')

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

  it('lists changes and pull results with conflicts', async () => {
    vi.mocked(binding.Changes).mockResolvedValue([{ item_id: 'i1', item_name: 'Ping', item_kind: 'request', change: 'modified', paths: null }] as never)
    await expect(service.listChanges('L')).resolves.toEqual([{ itemId: 'i1', itemName: 'Ping', itemKind: 'request', change: 'modified', paths: [] }])
    vi.mocked(binding.Changes).mockResolvedValue(null as never)
    await expect(service.listChanges('L')).resolves.toEqual([])
    vi.mocked(binding.Pull).mockResolvedValue({
      status: { ...statusDto, state: 'conflict' },
      session: { id: 's', conflicts: [{ item_id: 'i1', item_name: 'Users', item_kind: 'request', kind: 'both_modified', resolution: 'unresolved', path: 'u.yaml' }] },
    } as never)
    await expect(service.pullWorkspace('L')).resolves.toEqual({
      status: { ...status, state: 'conflict' },
      conflicts: [{ itemId: 'i1', itemName: 'Users', itemKind: 'request', kind: 'both_modified', path: 'u.yaml' }],
    })
    vi.mocked(binding.Pull).mockResolvedValue({ status: statusDto } as never)
    await expect(service.pullWorkspace('L')).resolves.toEqual({ status, conflicts: [] })
  })

  it('turns failures and empty answers into AppErrors', async () => {
    vi.mocked(binding.Push).mockRejectedValue({ cause: { error: { code: 'GIT-05409', description: 'busy' } } })
    await expect(service.pushWorkspace('L')).rejects.toMatchObject({ code: 'GIT-05409' })
    vi.mocked(binding.Status).mockResolvedValue(null as never)
    await expect(service.getSyncStatus('L')).rejects.toBeInstanceOf(AppError)
  })
})
