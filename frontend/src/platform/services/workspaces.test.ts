import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/workspaceservice', () => ({
  List: vi.fn(),
  Create: vi.fn(),
  Activate: vi.fn(),
}))

const WorkspaceService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/workspaceservice')
const { activateWorkspace, createWorkspace, listWorkspaces } = await import('./workspaces')
const { AppError } = await import('../errors')

const summary = { local_id: 'L1', name: 'Acme', path: '/work/acme', missing: false }

describe('workspace service', () => {
  beforeEach(() => {
    vi.mocked(WorkspaceService.List).mockReset()
    vi.mocked(WorkspaceService.Create).mockReset()
    vi.mocked(WorkspaceService.Activate).mockReset()
  })

  it('lists workspaces as frontend models', async () => {
    vi.mocked(WorkspaceService.List).mockResolvedValue({ active_local_id: 'L1', workspaces: [summary] } as never)
    await expect(listWorkspaces()).resolves.toEqual({
      activeLocalId: 'L1',
      workspaces: [{ localId: 'L1', name: 'Acme', path: '/work/acme', missing: false }],
    })
  })

  it('treats an empty list as no workspaces and no active one', async () => {
    vi.mocked(WorkspaceService.List).mockResolvedValue({ active_local_id: '', workspaces: null } as never)
    await expect(listWorkspaces()).resolves.toEqual({ activeLocalId: null, workspaces: [] })
  })

  it('creates a workspace', async () => {
    vi.mocked(WorkspaceService.Create).mockResolvedValue(summary as never)
    await expect(createWorkspace({ name: 'Acme', path: '/work/acme' })).resolves.toMatchObject({ localId: 'L1' })
    expect(WorkspaceService.Create).toHaveBeenCalledWith({ name: 'Acme', path: '/work/acme' })
  })

  it('activates a workspace', async () => {
    vi.mocked(WorkspaceService.Activate).mockResolvedValue({ summary } as never)
    await expect(activateWorkspace('L1')).resolves.toEqual({ summary: expect.objectContaining({ name: 'Acme' }) })
    expect(WorkspaceService.Activate).toHaveBeenCalledWith({ local_id: 'L1' })
  })

  it('converts Go errors into AppError', async () => {
    const failure = Object.assign(new Error('conflict'), { cause: { error: { code: 'WKS-01409', description: 'not empty' } } })
    vi.mocked(WorkspaceService.Create).mockRejectedValue(failure)
    await expect(createWorkspace({ name: 'Acme', path: '/x' })).rejects.toMatchObject({ code: 'WKS-01409' })
    vi.mocked(WorkspaceService.List).mockRejectedValue(new Error('boom'))
    await expect(listWorkspaces()).rejects.toBeInstanceOf(AppError)
    vi.mocked(WorkspaceService.Activate).mockRejectedValue(failure)
    await expect(activateWorkspace('L1')).rejects.toMatchObject({ code: 'WKS-01409' })
  })

  it('rejects empty responses', async () => {
    vi.mocked(WorkspaceService.Create).mockResolvedValue(null as never)
    await expect(createWorkspace({ name: 'Acme', path: '/x' })).rejects.toMatchObject({ code: 'UNK-00204' })
  })
})
