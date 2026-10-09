import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultUIState } from '@/platform/mappers/ui-state'
import type { RequestChangedOnDisk } from '@/platform/events'

const unsubscribe = vi.fn()
const handlers: {
  folderMissing?: (event: { localId: string }) => void
  treeChanged?: (event: { localId: string }) => void
  changedOnDisk?: (event: RequestChangedOnDisk) => void
} = {}

vi.mock('@/platform/events', () => ({
  onWorkspaceFolderMissing: vi.fn((callback) => {
    handlers.folderMissing = callback
    return unsubscribe
  }),
  onWorkspaceTreeChanged: vi.fn((callback) => {
    handlers.treeChanged = callback
    return unsubscribe
  }),
  onRequestChangedOnDisk: vi.fn((callback) => {
    handlers.changedOnDisk = callback
    return unsubscribe
  }),
}))
vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
  autoHeaders: vi.fn(),
  getRequest: vi.fn(),
  applyUrl: vi.fn(),
}))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn(), getHistoryEntry: vi.fn(), clearHistory: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const tree = await import('@/platform/services/tree')
const uiState = await import('@/platform/services/ui-state')
const requests = await import('@/platform/services/requests')
const history = await import('@/platform/services/history')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore, PERSIST_DELAY_MS } = await import('@/features/tabs/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { bootstrap, openLastWorkspace } = await import('./bootstrap')

function resetWorkspaceStores() {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: false }

async function started() {
  const cleanup = bootstrap()
  await vi.waitFor(() => expect(useWorkspacesStore.getState().status).toBe('ready'))
  return cleanup
}

describe('bootstrap', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(requests.applyUrl).mockImplementation(async (draft) => draft)
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    vi.mocked(history.listHistory).mockResolvedValue([])
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, missingLocalId: null, status: 'idle', error: null })
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
    resetWorkspaceStores()
  })

  it('loads the workspaces and reopens the last active one', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme, tree: [], uiState: defaultUIState() })
    await openLastWorkspace()
    expect(services.activateWorkspace).toHaveBeenCalledWith('A')
    expect(useWorkspacesStore.getState().activeLocalId).toBe('A')
    expect(useTabsStore.getState().tabs).toHaveLength(1)
  })

  it('shows the welcome state when nothing was active', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    await openLastWorkspace()
    expect(services.activateWorkspace).not.toHaveBeenCalled()
    expect(useWorkspacesStore.getState().status).toBe('ready')
  })

  it('subscribes to backend events and unsubscribes on cleanup', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    handlers.folderMissing?.({ localId: 'A' })
    expect(useWorkspacesStore.getState().workspaces[0]?.missing).toBe(true)
    cleanup()
    expect(unsubscribe).toHaveBeenCalledTimes(3)
  })

  it('reloads the sidebar when the active workspace tree changed outside Xenon', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    handlers.treeChanged?.({ localId: 'other' })
    expect(tree.getTree).not.toHaveBeenCalled()
    handlers.treeChanged?.({ localId: 'A' })
    expect(tree.getTree).toHaveBeenCalledWith('A')
    cleanup()
  })

  it('hands request changes on disk to the tabs', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    const handle = vi.spyOn(useTabsStore.getState(), 'handleChangedOnDisk').mockResolvedValue()
    const event = { localId: 'A', requestId: 'r1', deleted: true, revision: '' }
    handlers.changedOnDisk?.(event)
    expect(handle).toHaveBeenCalledWith(event)
    handle.mockRestore()
    cleanup()
  })

  it('pins an edited preview tab and saves the UI state after edits', async () => {
    vi.useFakeTimers()
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: { ...(await import('@/platform/mappers/request')).emptyDraft(), id: 'r1' }, revision: 'rev' })
    const cleanup = bootstrap()
    await vi.advanceTimersByTimeAsync(0)
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    await useTabsStore.getState().openPreview('r1')
    const preview = useTabsStore.getState().previewTabId!
    useRequestBuilderStore.getState().setUrl(preview, 'https://edited')
    expect(useTabsStore.getState().previewTabId).toBeNull()
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS * 3)
    expect(uiState.saveUIState).toHaveBeenCalledTimes(1)
    expect(uiState.saveUIState).toHaveBeenCalledWith('A', expect.objectContaining({ previewTabId: null }))
    window.dispatchEvent(new Event('beforeunload'))
    expect(uiState.saveUIState).toHaveBeenCalledTimes(2)
    cleanup()
  })

  it('refreshes the shown history when a send finishes', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(history.listHistory).mockResolvedValue([])
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    const viewer = useResponseViewerStore.getState()
    viewer.startSending('t1')
    viewer.setFailure('t1', new (await import('@/platform/errors')).AppError('x', 'y'))
    expect(history.listHistory).not.toHaveBeenCalled()
    useSidebarStore.getState().setView('history')
    expect(history.listHistory).toHaveBeenCalledTimes(1)
    viewer.startSending('t1')
    expect(history.listHistory).toHaveBeenCalledTimes(1)
    viewer.setFailure('t1', new (await import('@/platform/errors')).AppError('x', 'y'))
    expect(history.listHistory).toHaveBeenCalledTimes(2)
    cleanup()
  })

  it('saves the UI state when the layout changes', async () => {
    vi.useFakeTimers()
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = bootstrap()
    await vi.advanceTimersByTimeAsync(0)
    useWorkspacesStore.getState().setLayout({ sidebarWidth: 400 })
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS)
    expect(uiState.saveUIState).not.toHaveBeenCalled()
    useTabsStore.getState().hydrate('A', defaultUIState())
    useWorkspacesStore.getState().setLayout({ sidebarWidth: 420 })
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS)
    expect(uiState.saveUIState).toHaveBeenCalledWith('A', expect.objectContaining({ layout: expect.objectContaining({ sidebarWidth: 420 }) }))
    cleanup()
  })
})
