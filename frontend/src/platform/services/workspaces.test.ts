import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/workspaceservice', () => ({
  List: vi.fn(),
  Create: vi.fn(),
  Activate: vi.fn(),
  Open: vi.fn(),
  Rename: vi.fn(),
  Remove: vi.fn(),
  Locate: vi.fn(),
}))

const WorkspaceService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/workspaceservice')
const { activateWorkspace, createWorkspace, listWorkspaces, locateWorkspace, openWorkspace, removeWorkspace, renameWorkspace } = await import('./workspaces')
const { AppError } = await import('../errors')

const summary = { local_id: 'L1', name: 'Acme', path: '/work/acme', missing: false }

describe('workspace service', () => {
  beforeEach(() => {
    vi.resetAllMocks()
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

  it('activates a workspace and maps its tree and UI state', async () => {
    vi.mocked(WorkspaceService.Activate).mockResolvedValue({
      summary,
      tree: [{ id: 'c1', kind: 'collection', name: 'Billing', children: null }],
      ui_state: {
        local_id: 'L1',
        version: 1,
        active_tab_id: 't1',
        preview_tab_id: '',
        tabs: [{ tab_id: 't1', request_id: 'r1', pinned: true, base_revision: 'rev' }],
        sidebar: { view: 'collections', selected_id: '', search: '', expanded: null },
        layout: { split: 'stacked', right_panel_view: 'info', split_ratio: 0.5, sidebar_width: 280, right_panel_width: 320, console_height: 200, sidebar_visible: true, right_panel_visible: true, console_visible: false },
      },
    } as never)
    const snapshot = await activateWorkspace('L1')
    expect(WorkspaceService.Activate).toHaveBeenCalledWith({ local_id: 'L1' })
    expect(snapshot.summary.name).toBe('Acme')
    expect(snapshot.tree[0]).toMatchObject({ id: 'c1', name: 'Billing', children: [] })
    expect(snapshot.uiState.tabs).toEqual([{ tabId: 't1', kind: 'request', requestId: 'r1', exampleId: null, environmentId: null, itemId: null, pinned: true, baseRevision: 'rev', draft: null, aliasId: null, versionDrafts: {} }])
    expect(snapshot.uiState.activeTabId).toBe('t1')
    expect(snapshot.uiState.previewTabId).toBeNull()
  })

  it('opens, renames, removes and locates workspaces', async () => {
    vi.mocked(WorkspaceService.Open).mockResolvedValue(summary as never)
    await expect(openWorkspace('/work/acme')).resolves.toMatchObject({ localId: 'L1' })
    expect(WorkspaceService.Open).toHaveBeenCalledWith({ path: '/work/acme' })
    vi.mocked(WorkspaceService.Rename).mockResolvedValue({ ...summary, name: 'Acme Corp' } as never)
    await expect(renameWorkspace('L1', 'Acme Corp')).resolves.toMatchObject({ name: 'Acme Corp' })
    expect(WorkspaceService.Rename).toHaveBeenCalledWith({ local_id: 'L1', name: 'Acme Corp' })
    vi.mocked(WorkspaceService.Remove).mockResolvedValue({ active_local_id: '', workspaces: [] } as never)
    await expect(removeWorkspace('L1', true)).resolves.toEqual({ activeLocalId: null, workspaces: [] })
    expect(WorkspaceService.Remove).toHaveBeenCalledWith({ local_id: 'L1', move_to_trash: true })
    vi.mocked(WorkspaceService.Locate).mockResolvedValue(summary as never)
    await expect(locateWorkspace('L1', '/moved')).resolves.toMatchObject({ path: '/work/acme' })
    expect(WorkspaceService.Locate).toHaveBeenCalledWith({ local_id: 'L1', path: '/moved' })
    const failure = Object.assign(new Error('x'), { cause: { error: { code: 'WKS-05500', description: 'no trash' } } })
    for (const call of [WorkspaceService.Open, WorkspaceService.Rename, WorkspaceService.Remove, WorkspaceService.Locate]) {
      vi.mocked(call).mockRejectedValue(failure)
    }
    await expect(openWorkspace('/x')).rejects.toMatchObject({ code: 'WKS-05500' })
    await expect(renameWorkspace('L1', 'x')).rejects.toMatchObject({ code: 'WKS-05500' })
    await expect(removeWorkspace('L1', true)).rejects.toMatchObject({ code: 'WKS-05500' })
    await expect(locateWorkspace('L1', '/x')).rejects.toMatchObject({ code: 'WKS-05500' })
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
