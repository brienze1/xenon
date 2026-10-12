import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth } from '@/platform/mappers/request'
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
  listEnvironments: vi.fn(async () => ({ environments: [{ id: 'e1', name: 'dev', color: 'cyan', confirmBeforeSend: false, error: null }], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(async (_localId: string, scope: string) => ({ scope, rows: [], warnings: [] })),
  saveVariableTable: vi.fn(),
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
vi.mock('@/platform/services/tree', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/tree')>()),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
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
const tree = await import('@/platform/services/tree')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSearchStore } = await import('@/features/search/store')
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
    expect(requests.sendRequest).toHaveBeenCalledWith('A', tabId, expect.objectContaining({ method: 'GET' }), { confirmedEnvironmentId: '', missingSecrets: false, expiredToken: false })
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
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
  })

  it('reaches the stage-5 shortcuts: search, tab cycling and reopening, sidebar and URL focus', async () => {
    await renderWorkspace()
    press('k', { code: 'KeyK' })
    expect(await screen.findByRole('dialog', { name: 'Search Requests' })).toBeInTheDocument()
    act(() => useSearchStore.getState().close())
    press('t', { code: 'KeyT' })
    const [first, second] = openTabs()
    expect(second).toHaveAttribute('aria-selected', 'true')
    press('Tab', { code: 'Tab' })
    expect(openTabs()[0]).toHaveAttribute('aria-selected', 'true')
    press('1', { code: 'Digit2' })
    expect(openTabs()[1]).toHaveAttribute('aria-selected', 'true')
    act(() => useRequestBuilderStore.getState().setUrl(useTabsStore.getState().activeTabId!, 'https://closed'))
    press('w', { code: 'KeyW', altKey: true })
    expect(openTabs()).toHaveLength(1)
    press('T', { code: 'KeyT', shiftKey: true })
    expect(openTabs()).toHaveLength(2)
    expect(useRequestBuilderStore.getState().tabs[useTabsStore.getState().activeTabId!]!.draft.url).toBe('https://closed')
    press('l', { code: 'KeyL' })
    expect(document.activeElement).toHaveAttribute('aria-label', 'Request URL')
    const visible = useWorkspacesStore.getState().layout.sidebarVisible
    press('\\', { code: 'Backslash' })
    expect(useWorkspacesStore.getState().layout.sidebarVisible).toBe(!visible)
    expect(first).toBeDefined()
  })

  it('opens the globals as a tab from the switcher and shows the environments in the sidebar for "Manage Environments…"', async () => {
    await renderWorkspace()
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Globals' }))
    expect(openTabs()).toHaveLength(2)
    expect(await screen.findByRole('table', { name: 'Global variables' })).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    act(() => useWorkspacesStore.getState().setLayout({ sidebarVisible: false }))
    await userEvent.click(screen.getByRole('button', { name: 'Environment: No environment' }))
    await userEvent.click(screen.getByRole('menuitem', { name: 'Manage Environments…' }))
    expect(useSidebarStore.getState().view).toBe('environments')
    expect(useWorkspacesStore.getState().layout.sidebarVisible).toBe(true)
    const environments = await screen.findByRole('list', { name: 'Environments' })
    expect(within(environments).getByRole('button', { name: 'Globals' })).toHaveAttribute('aria-current', 'true')
    await userEvent.click(within(environments).getByRole('button', { name: 'dev' }))
    expect(await screen.findByRole('table', { name: 'dev variables' })).toBeInTheDocument()
    expect(openTabs()).toHaveLength(3)
    expect(within(environments).getByRole('button', { name: 'dev' })).toHaveAttribute('aria-current', 'true')
  })

  it('switches the sidebar view from the view menu with the keyboard', async () => {
    await renderWorkspace()
    const trigger = screen.getByRole('button', { name: 'Sidebar view: Collections' })
    trigger.focus()
    await userEvent.keyboard('{Enter}')
    expect(screen.getAllByRole('menuitemradio').map((item) => item.textContent)).toEqual(['Collections', 'Environments', 'History'])
    expect(screen.getByRole('menuitemradio', { name: 'Collections' })).toHaveAttribute('aria-checked', 'true')
    await userEvent.keyboard('{ArrowDown}{Enter}')
    expect(useSidebarStore.getState().view).toBe('environments')
    expect(screen.getByRole('button', { name: 'Sidebar view: Environments' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'New Collection' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'New Environment' })).toBeInTheDocument()
  })

  it('opens a collection as a tab from the sidebar, saves it with Cmd/Ctrl+S and shows no request details', async () => {
    const billing = { id: 'c1', kind: 'collection' as const, name: 'Billing', method: null, url: null, error: null, warning: null, children: [] }
    const settings = { id: 'c1', kind: 'collection' as const, name: 'Billing', description: '', auth: { ...emptyAuth(), type: 'inherit' as const }, headers: [], scripts: { preRequest: '', postResponse: '', disabled: [] }, inheritScripts: null }
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    vi.mocked(tree.saveItemSettings).mockImplementation(async (_localId, saved) => saved)
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme, tree: [billing], uiState: defaultUIState() })
    const offsetHeight = vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(600)
    await renderApp()
    await userEvent.click(await screen.findByRole('treeitem', { name: /Billing/ }))
    offsetHeight.mockRestore()
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    expect(openTabs()).toHaveLength(2)
    expect(openTabs()[1]).toHaveAttribute('title', 'Billing (preview)')
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    await userEvent.type(within(tab).getByRole('textbox', { name: 'Description' }), 'Billing API')
    expect(within(openTabs()[1]!).getByRole('img', { name: 'Unsaved changes' })).toBeInTheDocument()
    press('s')
    await vi.waitFor(() => expect(tree.saveItemSettings).toHaveBeenCalledWith('A', expect.objectContaining({ id: 'c1', description: 'Billing API' })))
    await vi.waitFor(() => expect(within(openTabs()[1]!).queryByRole('img', { name: 'Unsaved changes' })).not.toBeInTheDocument())
  })
})

