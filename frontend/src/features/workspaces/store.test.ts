import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft } from '@/platform/mappers/request'
import { defaultLayout, defaultUIState, type UIState } from '@/platform/mappers/ui-state'
import type { WorkspaceSnapshot, WorkspaceSummary } from '@/platform/mappers/workspace'

vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
  openWorkspace: vi.fn(),
  renameWorkspace: vi.fn(),
  removeWorkspace: vi.fn(),
  locateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({
  autoHeaders: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  sendRequest: vi.fn(),
}))

const services = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const uiState = await import('@/platform/services/ui-state')
const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('./store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: false }
const globex = { localId: 'G', name: 'Globex', path: '/w/globex', missing: false }

function snapshot(summary: WorkspaceSummary, state: UIState = defaultUIState()): WorkspaceSnapshot {
  return { summary, tree: [], uiState: state }
}

function store() {
  return useWorkspacesStore.getState()
}

describe('workspaces store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    useTabsStore.getState().reset()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, missingLocalId: null, status: 'idle', error: null })
  })

  it('keeps each workspace layout, clamped, and resets it when leaving', async () => {
    vi.mocked(services.activateWorkspace).mockResolvedValueOnce(
      snapshot(acme, { ...defaultUIState(), layout: { ...defaultLayout(), sidebarWidth: 9000, split: 'side-by-side', consoleVisible: true } }),
    )
    await store().activate('A')
    expect(store().layout).toMatchObject({ sidebarWidth: 600, split: 'side-by-side', consoleVisible: true })
    store().setLayout({ rightPanelWidth: 100, consoleHeight: window.innerHeight * 2, splitRatio: 0.9 })
    expect(store().layout).toMatchObject({ rightPanelWidth: 240, consoleHeight: Math.round(window.innerHeight * 0.6), splitRatio: 0.8, sidebarWidth: 600 })
    await useTabsStore.getState().flushPersist()
    expect(vi.mocked(uiState.saveUIState).mock.calls.at(-1)![1].layout).toMatchObject({ rightPanelWidth: 240, split: 'side-by-side' })
    vi.mocked(services.activateWorkspace).mockResolvedValueOnce(snapshot(globex))
    await store().activate('G')
    expect(store().layout).toEqual(defaultLayout())
    store().setLayout({ sidebarVisible: false })
    store().markMissing('G')
    expect(store().layout).toEqual(defaultLayout())
  })

  it('loads the registered workspaces', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme, globex], activeLocalId: 'G' })
    await store().load()
    expect(store()).toMatchObject({ workspaces: [acme, globex], activeLocalId: 'G', status: 'ready', error: null })
  })

  it('keeps the error when loading fails', async () => {
    vi.mocked(services.listWorkspaces).mockRejectedValue(new AppError('WKS-12500', 'bad app.json'))
    await store().load()
    expect(store()).toMatchObject({ status: 'ready', error: { code: 'WKS-12500' } })
  })

  it('creates a workspace and activates it', async () => {
    useWorkspacesStore.setState({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.createWorkspace).mockResolvedValue(globex)
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot(globex))
    await expect(store().create('Globex', '/w/globex')).resolves.toBe(true)
    expect(services.createWorkspace).toHaveBeenCalledWith({ name: 'Globex', path: '/w/globex' })
    expect(services.activateWorkspace).toHaveBeenCalledWith('G')
    expect(store()).toMatchObject({ workspaces: [acme, globex], activeLocalId: 'G' })
  })

  it('reports a refused creation', async () => {
    vi.mocked(services.createWorkspace).mockRejectedValue(new AppError('WKS-01409', 'not empty'))
    await expect(store().create('Acme', '/w/full')).resolves.toBe(false)
    expect(store().error?.code).toBe('WKS-01409')
    store().clearError()
    expect(store().error).toBeNull()
  })

  it('opens an existing workspace folder and activates it', async () => {
    vi.mocked(services.openWorkspace).mockResolvedValue(acme)
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot(acme))
    await expect(store().open('/w/acme')).resolves.toBe(true)
    expect(store()).toMatchObject({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.openWorkspace).mockRejectedValue(new AppError('WKS-02409', 'overlap'))
    await expect(store().open('/w')).resolves.toBe(false)
    expect(store().error?.code).toBe('WKS-02409')
  })

  it('switching resets every workspace store and loads the snapshot', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    useTabsStore.getState().hydrate('G', defaultUIState())
    const oldTab = useTabsStore.getState().activeTabId!
    useRequestBuilderStore.getState().setUrl(oldTab, 'https://globex')
    useResponseViewerStore.getState().startSending(oldTab)
    useSidebarStore.setState({ query: 'globex only', expanded: { g1: true } })
    const restored: UIState = {
      ...defaultUIState(),
      tabs: [{ tabId: 'acme-tab', requestId: null, pinned: true, baseRevision: null, draft: { ...emptyDraft(), url: 'https://acme' } }],
      activeTabId: 'acme-tab',
    }
    vi.mocked(services.activateWorkspace).mockResolvedValue({
      summary: { ...acme, name: 'Acme Corp' },
      tree: [{ id: 'c1', kind: 'collection', name: 'Billing', method: null, url: null, error: null, warning: null, children: [] }],
      uiState: restored,
    })
    await store().activate('A')
    expect(store()).toMatchObject({ activeLocalId: 'A', workspaces: [{ name: 'Acme Corp' }, globex] })
    expect(useRequestBuilderStore.getState().tabs[oldTab]).toBeUndefined()
    expect(useResponseViewerStore.getState().entries).toEqual({})
    expect(useSidebarStore.getState()).toMatchObject({ query: '', expanded: {}, tree: [{ id: 'c1' }] })
    expect(useTabsStore.getState()).toMatchObject({ localId: 'A', activeTabId: 'acme-tab' })
    expect(useRequestBuilderStore.getState().tabs['acme-tab']!.draft.url).toBe('https://acme')
  })

  it('shows the missing-folder notice when activation finds the folder gone', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-06404', 'missing'))
    await store().activate('A')
    expect(store()).toMatchObject({
      activeLocalId: null,
      missingLocalId: 'A',
      error: null,
      workspaces: [{ localId: 'A', missing: true }, { localId: 'G', missing: false }],
    })
  })

  it('keeps the active workspace when activation fails for another reason', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-04422', 'newer'))
    await store().activate('A')
    expect(store()).toMatchObject({ activeLocalId: 'G', error: { code: 'WKS-04422' } })
  })

  it('renames a workspace', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'A' })
    vi.mocked(services.renameWorkspace).mockResolvedValue({ ...acme, name: 'Acme Corp' })
    await expect(store().rename('A', 'Acme Corp')).resolves.toBe(true)
    expect(store().workspaces[0]?.name).toBe('Acme Corp')
    vi.mocked(services.renameWorkspace).mockRejectedValue(new AppError('WKS-22400', 'blank'))
    await expect(store().rename('A', ' ')).resolves.toBe(false)
    expect(store().error?.code).toBe('WKS-22400')
  })

  it('removes the active workspace and activates the next one', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'A' })
    vi.mocked(services.removeWorkspace).mockResolvedValue({ workspaces: [globex], activeLocalId: 'G' })
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot(globex))
    await expect(store().remove('A', true)).resolves.toBe(true)
    expect(services.removeWorkspace).toHaveBeenCalledWith('A', true)
    expect(store()).toMatchObject({ workspaces: [globex], activeLocalId: 'G' })
  })

  it('shows the welcome screen after removing the last workspace', async () => {
    useWorkspacesStore.setState({ workspaces: [acme], activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    vi.mocked(services.removeWorkspace).mockResolvedValue({ workspaces: [], activeLocalId: null })
    await store().remove('A', false)
    expect(store()).toMatchObject({ workspaces: [], activeLocalId: null, missingLocalId: null })
    expect(useTabsStore.getState().tabs).toEqual([])
  })

  it('removes an inactive workspace without switching', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G' })
    vi.mocked(services.removeWorkspace).mockResolvedValue({ workspaces: [globex], activeLocalId: 'G' })
    await store().remove('A', false)
    expect(services.activateWorkspace).not.toHaveBeenCalled()
    expect(store().activeLocalId).toBe('G')
  })

  it('keeps the workspace when the trash fails', async () => {
    useWorkspacesStore.setState({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.removeWorkspace).mockRejectedValue(new AppError('WKS-05500', 'no trash'))
    await expect(store().remove('A', true)).resolves.toBe(false)
    expect(store()).toMatchObject({ workspaces: [acme], activeLocalId: 'A', error: { code: 'WKS-05500' } })
  })

  it('locates a moved folder and activates the workspace again', async () => {
    useWorkspacesStore.setState({ workspaces: [{ ...acme, missing: true }], activeLocalId: null, missingLocalId: 'A' })
    vi.mocked(dialogs.pickFolder).mockResolvedValueOnce(null).mockResolvedValueOnce('/moved/acme')
    await store().locate('A')
    expect(services.locateWorkspace).not.toHaveBeenCalled()
    vi.mocked(services.locateWorkspace).mockResolvedValue({ ...acme, path: '/moved/acme' })
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot({ ...acme, path: '/moved/acme' }))
    await store().locate('A')
    expect(services.locateWorkspace).toHaveBeenCalledWith('A', '/moved/acme')
    expect(store()).toMatchObject({ activeLocalId: 'A', missingLocalId: null, workspaces: [{ path: '/moved/acme', missing: false }] })
  })

  it('reports a folder that is not a workspace when locating', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/plain')
    vi.mocked(services.locateWorkspace).mockRejectedValue(new AppError('WKS-03404', 'no workspace.yaml'))
    await store().locate('A')
    expect(store().error?.code).toBe('WKS-03404')
    expect(services.activateWorkspace).not.toHaveBeenCalled()
  })

  it('chooses a folder through the native picker', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/new')
    await expect(store().chooseFolder('Choose')).resolves.toBe('/w/new')
    vi.mocked(dialogs.pickFolder).mockRejectedValue(new AppError('DLG-00500', 'no dialogs'))
    await expect(store().chooseFolder('Choose')).resolves.toBeNull()
    expect(store().error?.code).toBe('DLG-00500')
  })

  it('marks a workspace whose folder disappeared and leaves it when it is active', () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'A' })
    store().markMissing('G')
    expect(store().workspaces).toEqual([acme, { ...globex, missing: true }])
    expect(store().activeLocalId).toBe('A')
    store().markMissing('A')
    expect(store()).toMatchObject({ activeLocalId: null, missingLocalId: 'A' })
  })
})
