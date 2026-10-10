import { act, render, screen, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('./bootstrap', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./bootstrap')>()),
  bootstrap: vi.fn(() => () => undefined),
}))
vi.mock('@/platform/events', () => ({
  onWorkspaceFolderMissing: vi.fn(() => () => undefined),
  onWorkspaceTreeChanged: vi.fn(() => () => undefined),
  onRequestChangedOnDisk: vi.fn(() => () => undefined),
}))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
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
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(async (draft: unknown) => draft),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(async () => []),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  lastResult: vi.fn(async () => null),
  releaseResponse: vi.fn(async () => undefined),
  saveResponseBody: vi.fn(),
}))

const services = await import('@/platform/services/workspaces')
const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { bootstrap, openLastWorkspace } = await import('./bootstrap')
const { App } = await import('./App')

function resetWorkspaceStores() {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

async function renderApp() {
  render(<App />)
  await act(() => openLastWorkspace())
}

const acme = { localId: 'A', name: 'Acme', path: '/work/acme', missing: false }

function press(key: string, modifiers: KeyboardEventInit = {}) {
  act(() => {
    window.dispatchEvent(new KeyboardEvent('keydown', { key, ctrlKey: true, metaKey: false, ...modifiers }))
  })
}

function openTabs() {
  return within(screen.getByRole('tablist', { name: 'Open requests' })).getAllByRole('tab')
}

async function renderWorkspace() {
  vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
  vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme, tree: [], uiState: defaultUIState() })
  await renderApp()
  await screen.findByRole('textbox', { name: 'Request URL' })
}

describe('App', () => {
  it('bootstraps once per app load, not once per mount', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    const first = render(<App />)
    first.unmount()
    render(<App />)
    expect(bootstrap).toHaveBeenCalledTimes(1)
    await act(() => openLastWorkspace())
  })

  beforeEach(() => {
    vi.mocked(services.listWorkspaces).mockReset()
    vi.mocked(services.activateWorkspace).mockReset()
    vi.mocked(requests.sendRequest).mockReset()
    resetWorkspaceStores()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, missingLocalId: null, status: 'idle', error: null })
  })

  afterEach(() => resetWorkspaceStores())

  it('shows the welcome screen when there is no workspace', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    await renderApp()
    expect(await screen.findByRole('form', { name: 'Create workspace' })).toBeInTheDocument()
    expect(screen.queryByRole('navigation', { name: 'Sidebar' })).not.toBeInTheDocument()
  })

  it('reopens the last active workspace in the full window layout', async () => {
    await renderWorkspace()
    expect(screen.getByRole('button', { name: 'Workspace: Acme' })).toBeInTheDocument()
    expect(screen.getByRole('navigation', { name: 'Sidebar' })).toBeInTheDocument()
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toBeInTheDocument()
    expect(screen.getByRole('tablist', { name: 'Open requests' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'No response yet' })).toBeInTheDocument()
    expect(screen.getByRole('contentinfo')).toHaveTextContent('/work/acme')
  })

  it('shows the missing-folder notice for a workspace whose folder is gone', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    const { AppError } = await import('@/platform/errors')
    vi.mocked(services.activateWorkspace).mockRejectedValue(new AppError('WKS-06404', 'missing'))
    await renderApp()
    expect(await screen.findByRole('region', { name: 'Workspace folder missing' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Workspace: Acme' })).toBeInTheDocument()
  })

  it('sends the active tab with Cmd/Ctrl+Enter', async () => {
    await renderWorkspace()
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    press('Enter', { metaKey: true, ctrlKey: false })
    press('Enter')
    const tabId = useTabsStore.getState().activeTabId
    expect(requests.sendRequest).toHaveBeenCalledTimes(1)
    expect(requests.sendRequest).toHaveBeenCalledWith('A', tabId, expect.objectContaining({ method: 'GET' }), false)
  })

  it('opens, saves and closes tabs with the keyboard', async () => {
    await renderWorkspace()
    press('t')
    expect(openTabs()).toHaveLength(2)
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://new'))
    press('s')
    expect(await screen.findByRole('dialog', { name: 'Save Request' })).toBeInTheDocument()
    act(() => useTabsStore.getState().cancelSave())
    press('S', { shiftKey: true })
    expect(await screen.findByRole('dialog', { name: 'Save Request As' })).toBeInTheDocument()
    act(() => useTabsStore.getState().cancelSave())
    press('w')
    expect(await screen.findByRole('dialog', { name: 'Save Changes?' })).toBeInTheDocument()
    act(() => void useTabsStore.getState().resolveClose('discard'))
    expect(openTabs()).toHaveLength(1)
  })
  it('shows an empty side panel once the last tab is closed', async () => {
    await renderWorkspace()
    expect(screen.getByRole('textbox', { name: 'Name' })).toBeInTheDocument()
    act(() => useTabsStore.getState().requestClose())
    expect(screen.getByRole('region', { name: 'No request selected' })).toHaveTextContent('Open a request to see its details here.')
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
  })
})
