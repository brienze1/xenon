import { beforeEach, describe, expect, it, vi } from 'vitest'

const unsubscribe = vi.fn()
let folderMissing: ((event: { localId: string }) => void) | undefined

vi.mock('@/platform/events', () => ({
  onWorkspaceFolderMissing: vi.fn((callback: (event: { localId: string }) => void) => {
    folderMissing = callback
    return unsubscribe
  }),
}))
vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

vi.mock('@/platform/services/requests', () => ({ lastResult: vi.fn(), releaseResponse: vi.fn(), saveResponseBody: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { bootstrap, openLastWorkspace } = await import('./bootstrap')

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: false }

describe('bootstrap', () => {
  beforeEach(() => {
    vi.mocked(services.listWorkspaces).mockReset()
    vi.mocked(services.activateWorkspace).mockReset()
    unsubscribe.mockReset()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'idle', error: null })
  })

  it('loads the workspaces and reopens the last active one', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme })
    await openLastWorkspace()
    expect(services.activateWorkspace).toHaveBeenCalledWith('A')
    expect(useWorkspacesStore.getState().activeLocalId).toBe('A')
  })

  it('shows the welcome state when nothing was active', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    await openLastWorkspace()
    expect(services.activateWorkspace).not.toHaveBeenCalled()
    expect(useWorkspacesStore.getState().status).toBe('ready')
  })

  it('subscribes to backend events and unsubscribes on cleanup', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = bootstrap()
    await vi.waitFor(() => expect(useWorkspacesStore.getState().status).toBe('ready'))
    folderMissing?.({ localId: 'A' })
    expect(useWorkspacesStore.getState().workspaces[0]?.missing).toBe(true)
    cleanup()
    expect(unsubscribe).toHaveBeenCalledOnce()
  })

  it('starts fresh request and response state when the workspace changes', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = bootstrap()
    await vi.waitFor(() => expect(useWorkspacesStore.getState().status).toBe('ready'))
    const firstTab = useRequestBuilderStore.getState().activeTabId
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    expect(useRequestBuilderStore.getState().activeTabId).toBe(firstTab)
    useResponseViewerStore.getState().startSending(firstTab)
    useWorkspacesStore.setState({ activeLocalId: 'B' })
    expect(useRequestBuilderStore.getState().activeTabId).not.toBe(firstTab)
    expect(useResponseViewerStore.getState().entries).toEqual({})
    cleanup()
  })
})
