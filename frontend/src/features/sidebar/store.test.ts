import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { ItemSettings, TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  searchTree: vi.fn(),
  createCollection: vi.fn(),
  createFolder: vi.fn(),
  createRequest: vi.fn(),
  renameItem: vi.fn(),
  duplicateItem: vi.fn(),
  deleteItem: vi.fn(),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({ parseBulk: vi.fn(), formatBulk: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { SEARCH_DELAY_MS, NEW_COLLECTION_NAME, NEW_FOLDER_NAME, NEW_REQUEST_NAME, selectLocation, useSidebarStore } = await import('./store')

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: `name-${id}`, method: null, url: null, error: null, warning: null, children }
}

const sample = [node('c1', 'collection', [node('r1', 'request')])]

function state() {
  return useSidebarStore.getState()
}

describe('sidebar store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.getState().reset()
    vi.mocked(tree.getTree).mockResolvedValue(sample)
  })

  afterEach(() => vi.useRealTimers())

  it('loads the tree of the active workspace', async () => {
    await state().load()
    expect(tree.getTree).toHaveBeenCalledWith('ws')
    expect(state().tree).toEqual(sample)
    expect(state().status).toBe('ready')
  })

  it('does nothing without an active workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().load()
    expect(tree.getTree).not.toHaveBeenCalled()
    expect(state().status).toBe('ready')
  })

  it('keeps the error when loading fails', async () => {
    vi.mocked(tree.getTree).mockRejectedValue(new AppError('WKS-04422', 'newer'))
    await state().load()
    expect(state().error?.code).toBe('WKS-04422')
    state().clearError()
    expect(state().error).toBeNull()
  })

  it('debounces the search and drops stale answers', async () => {
    vi.useFakeTimers()
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: ['r1'], ancestorIds: ['c1'] })
    state().setQuery('pi')
    state().setQuery('pin')
    expect(tree.searchTree).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(SEARCH_DELAY_MS)
    expect(tree.searchTree).toHaveBeenCalledTimes(1)
    expect(tree.searchTree).toHaveBeenCalledWith('ws', 'pin')
    expect(state().search).toEqual({ matchIds: ['r1'], ancestorIds: ['c1'] })
    state().setQuery('  ')
    expect(state().search).toBeNull()
  })

  it('ignores a search answer that arrives after the query changed', async () => {
    vi.useFakeTimers()
    let resolve: (value: { matchIds: string[]; ancestorIds: string[] }) => void = () => {}
    vi.mocked(tree.searchTree).mockReturnValue(new Promise((done) => (resolve = done)))
    state().setQuery('a')
    await vi.advanceTimersByTimeAsync(SEARCH_DELAY_MS)
    state().setQuery('')
    resolve({ matchIds: ['r1'], ancestorIds: [] })
    await vi.runAllTimersAsync()
    expect(state().search).toBeNull()
  })

  it('toggles and selects items', () => {
    state().toggle('c1')
    expect(state().expanded.c1).toBe(true)
    state().toggle('c1')
    expect(state().expanded.c1).toBe(false)
    state().select('r1')
    expect(state().selectedId).toBe('r1')
  })

  it('creates items with a default name and starts renaming them', async () => {
    vi.mocked(tree.createCollection).mockResolvedValue(node('c9', 'collection'))
    vi.mocked(tree.createFolder).mockResolvedValue(node('f9', 'folder'))
    vi.mocked(tree.createRequest).mockResolvedValue(node('r9', 'request'))
    await state().createCollection()
    expect(tree.createCollection).toHaveBeenCalledWith('ws', NEW_COLLECTION_NAME)
    expect(state().renamingId).toBe('c9')
    await state().createFolder('c1')
    expect(tree.createFolder).toHaveBeenCalledWith('ws', 'c1', NEW_FOLDER_NAME)
    expect(state().expanded.c1).toBe(true)
    await state().createRequest('c1')
    expect(tree.createRequest).toHaveBeenCalledWith('ws', 'c1', NEW_REQUEST_NAME)
    expect(state().selectedId).toBe('r9')
    expect(tree.getTree).toHaveBeenCalledTimes(3)
  })

  it('renames only when the name really changes', async () => {
    await state().load()
    state().startRename('r1')
    expect(state().renamingId).toBe('r1')
    await expect(state().rename('r1', 'name-r1')).resolves.toBe(false)
    await expect(state().rename('r1', '   ')).resolves.toBe(false)
    expect(tree.renameItem).not.toHaveBeenCalled()
    vi.mocked(tree.renameItem).mockResolvedValue(node('r1', 'request'))
    await expect(state().rename('r1', 'Pong')).resolves.toBe(true)
    expect(tree.renameItem).toHaveBeenCalledWith('ws', 'r1', 'Pong')
    expect(state().renamingId).toBeNull()
    state().startRename('c1')
    state().cancelRename()
    expect(state().renamingId).toBeNull()
  })

  it('duplicates and selects the copy', async () => {
    vi.mocked(tree.duplicateItem).mockResolvedValue(node('r2', 'request'))
    await state().duplicate('r1')
    expect(tree.duplicateItem).toHaveBeenCalledWith('ws', 'r1')
    expect(state().selectedId).toBe('r2')
  })

  it('deletes and returns the deleted request ids', async () => {
    vi.mocked(tree.deleteItem).mockResolvedValue(['r1'])
    state().select('c1')
    await expect(state().remove('c1')).resolves.toEqual(['r1'])
    expect(state().selectedId).toBeNull()
    vi.mocked(tree.deleteItem).mockRejectedValue(new AppError('TRE-02404', 'gone'))
    await expect(state().remove('c1')).resolves.toEqual([])
    expect(state().error?.code).toBe('TRE-02404')
  })

  it('loads and saves item settings', async () => {
    const settings: ItemSettings = {
      id: 'c1',
      kind: 'collection',
      name: 'Billing',
      description: '',
      auth: { type: 'none', bearer: { token: '' }, basic: { username: '', password: '' } },
      headers: [],
    }
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    vi.mocked(tree.saveItemSettings).mockResolvedValue(settings)
    await expect(state().loadSettings('c1')).resolves.toEqual(settings)
    await expect(state().saveSettings(settings)).resolves.toBe(true)
    vi.mocked(tree.saveItemSettings).mockRejectedValue(new AppError('TRE-02404', 'gone'))
    await expect(state().saveSettings(settings)).resolves.toBe(false)
  })

  it('converts headers to and from bulk text, keeping the error on failure', async () => {
    const rows = [{ id: 'h1', key: 'X-Tenant', value: 'acme', description: '', enabled: true }]
    vi.mocked(requests.formatBulk).mockResolvedValue('X-Tenant:acme')
    vi.mocked(requests.parseBulk).mockResolvedValue(rows)
    await expect(state().formatBulk(rows)).resolves.toBe('X-Tenant:acme')
    expect(requests.formatBulk).toHaveBeenCalledWith(rows)
    await expect(state().parseBulk('X-Tenant:acme')).resolves.toEqual(rows)
    expect(requests.parseBulk).toHaveBeenCalledWith('X-Tenant:acme')
    vi.mocked(requests.parseBulk).mockRejectedValue(new AppError('REQ-01400', 'bad bulk'))
    await expect(state().parseBulk('nonsense')).resolves.toBeNull()
    expect(state().error?.code).toBe('REQ-01400')
  })

  it('describes the location of an item', () => {
    expect(selectLocation(sample, 'r1')).toEqual(['name-c1', 'name-r1'])
    expect(selectLocation(sample, 'x')).toEqual([])
  })
})
