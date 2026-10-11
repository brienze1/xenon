import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { defaultUIState } from '@/platform/mappers/ui-state'

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
vi.mock('@/platform/services/export', () => ({ getExportOptions: vi.fn(), pickExportTarget: vi.fn(), runExport: vi.fn(), showExportInFolder: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), releaseResponse: vi.fn(async () => undefined), getRequest: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const requests = await import('@/platform/services/requests')
const uiState = await import('@/platform/services/ui-state')
const { useTabsStore } = await import('@/features/tabs/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const exportService = await import('@/platform/services/export')
const { useExportStore } = await import('@/features/export/store')
const { WorkspaceSwitcher } = await import('./WorkspaceSwitcher')
const { MissingFolderNotice } = await import('./MissingFolderNotice')

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: true }
const globex = { localId: 'G', name: 'Globex', path: '/w/globex', missing: false }

function snapshot(summary: typeof acme) {
  return { summary, tree: [], uiState: defaultUIState() }
}

async function openMenu(name = 'Workspace: Globex') {
  await userEvent.click(screen.getByRole('button', { name }))
}

function resetWorkspaceStores() {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

function mockDefaults() {
  vi.mocked(requests.autoHeaders).mockResolvedValue([])
  vi.mocked(requests.releaseResponse).mockResolvedValue()
  vi.mocked(uiState.saveUIState).mockResolvedValue()
}

describe('WorkspaceSwitcher', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.resetAllMocks()
    mockDefaults()
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G', missingLocalId: null, status: 'ready', error: null })
  })

  afterEach(() => resetWorkspaceStores())

  it('shows the active workspace and switches to another one', async () => {
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot({ ...acme, missing: false }))
    render(<WorkspaceSwitcher />)
    await openMenu()
    expect(screen.getByRole('menuitemradio', { name: 'Globex' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('menuitemradio', { name: /^Acme/ })).toHaveAttribute('aria-checked', 'false')
    expect(screen.getByText('(missing)')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('menuitemradio', { name: /^Acme/ }))
    expect(services.activateWorkspace).toHaveBeenCalledWith('A')
  })

  it('exports the active workspace, and not a missing one', async () => {
    vi.mocked(exportService.getExportOptions).mockResolvedValue({ suggestedFileName: 'globex', formats: [], hasSecrets: false, hasCurrentValues: false })
    const { unmount } = render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Export Workspace…' }))
    expect(exportService.getExportOptions).toHaveBeenCalledWith({ localId: 'G', kind: 'workspace', itemId: '', name: 'Globex' })
    useExportStore.getState().reset()
    unmount()
    useWorkspacesStore.setState({ activeLocalId: null, missingLocalId: 'A' })
    render(<WorkspaceSwitcher />)
    await openMenu('Workspace: Acme')
    expect(screen.queryByRole('menuitem', { name: 'Export Workspace…' })).not.toBeInTheDocument()
  })

  it('says when there is no workspace', () => {
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null })
    render(<WorkspaceSwitcher />)
    expect(screen.getByRole('button', { name: 'Workspace: None' })).toHaveTextContent('No workspace')
  })

  it('creates a new workspace in a chosen folder', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/initech')
    const initech = { localId: 'I', name: 'Initech', path: '/w/initech', missing: false }
    vi.mocked(services.createWorkspace).mockResolvedValue(initech)
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot(initech))
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'New Workspace…' }))
    const dialog = screen.getByRole('dialog', { name: 'New Workspace' })
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Workspace name' }), 'Initech')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Choose Folder…' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Create Workspace' }))
    expect(services.createWorkspace).toHaveBeenCalledWith({ name: 'Initech', path: '/w/initech' })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useWorkspacesStore.getState().activeLocalId).toBe('I')
  })

  it('validates the new workspace form on submit', async () => {
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'New Workspace…' }))
    const dialog = screen.getByRole('dialog', { name: 'New Workspace' })
    const nameInput = within(dialog).getByRole('textbox', { name: 'Workspace name' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Create Workspace' }))
    expect(nameInput).toHaveFocus()
    expect(nameInput).toHaveAccessibleDescription('Enter a name for the workspace.')
    await userEvent.type(nameInput, 'Initech')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Create Workspace' }))
    expect(within(dialog).getByRole('button', { name: 'Choose Folder…' })).toHaveFocus()
    expect(within(dialog).getByRole('alert')).toHaveTextContent('Choose a folder for the workspace.')
    expect(services.createWorkspace).not.toHaveBeenCalled()
  })

  it('opens an existing workspace folder and reports a refusal', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w')
    vi.mocked(services.openWorkspace).mockRejectedValue(new AppError('WKS-02409', 'overlap'))
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Open Workspace…' }))
    expect(services.openWorkspace).toHaveBeenCalledWith('/w')
    expect(await screen.findByRole('alert')).toHaveTextContent('Another workspace already uses that folder')
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Error' }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('renames the active workspace', async () => {
    vi.mocked(services.renameWorkspace).mockResolvedValue({ ...globex, name: 'Globex Corp' })
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename Workspace…' }))
    const input = screen.getByRole('textbox', { name: 'Workspace name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Globex Corp')
    await userEvent.click(screen.getByRole('button', { name: 'Rename Workspace' }))
    expect(services.renameWorkspace).toHaveBeenCalledWith('G', 'Globex Corp')
    expect(screen.getByRole('button', { name: 'Workspace: Globex Corp' })).toBeInTheDocument()
  })

  it('refuses a blank workspace name on rename', async () => {
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename Workspace…' }))
    const input = screen.getByRole('textbox', { name: 'Workspace name' })
    await userEvent.clear(input)
    await userEvent.click(screen.getByRole('button', { name: 'Rename Workspace' }))
    expect(input).toHaveFocus()
    expect(input).toHaveAccessibleDescription('Enter a name for the workspace.')
    expect(screen.getByRole('alert')).toHaveTextContent('Enter a name for the workspace.')
    expect(services.renameWorkspace).not.toHaveBeenCalled()
  })

  it('removes with the trash and offers keeping the files when the trash fails', async () => {
    vi.mocked(services.removeWorkspace)
      .mockRejectedValueOnce(new AppError('WKS-05500', 'no trash'))
      .mockResolvedValueOnce({ workspaces: [acme], activeLocalId: null })
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Remove Workspace…' }))
    const dialog = screen.getByRole('dialog', { name: 'Remove “Globex”?' })
    await userEvent.click(within(dialog).getByRole('radio', { name: 'Remove and move the folder to the trash' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove Workspace' }))
    expect(services.removeWorkspace).toHaveBeenCalledWith('G', true)
    expect(within(dialog).getByRole('alert')).toHaveTextContent('Xenon couldn’t move the folder to the trash')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove from Xenon' }))
    expect(services.removeWorkspace).toHaveBeenLastCalledWith('G', false)
    expect(useWorkspacesStore.getState()).toMatchObject({ workspaces: [acme], activeLocalId: null })
  })
})

describe('WorkspaceSwitcher progress', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.resetAllMocks()
    mockDefaults()
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G', missingLocalId: null, status: 'ready', error: null })
  })

  afterEach(() => resetWorkspaceStores())

  function deferred<T>() {
    let resolve: (value: T) => void = () => undefined
    const promise = new Promise<T>((done) => (resolve = done))
    return { promise, resolve }
  }

  it('labels the create button while the workspace is being created', async () => {
    const pending = deferred<typeof globex>()
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/initech')
    vi.mocked(services.createWorkspace).mockReturnValue(pending.promise)
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot(globex))
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'New Workspace…' }))
    const dialog = screen.getByRole('dialog', { name: 'New Workspace' })
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Workspace name' }), 'Initech')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Choose Folder…' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Create Workspace' }))
    expect(within(dialog).getByRole('button', { name: 'Creating…' })).toBeDisabled()
    await act(async () => pending.resolve({ ...globex, localId: 'I' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('labels the rename button while the workspace is being renamed', async () => {
    const pending = deferred<typeof globex>()
    vi.mocked(services.renameWorkspace).mockReturnValue(pending.promise)
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename Workspace…' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Workspace name' }), ' Corp')
    await userEvent.click(screen.getByRole('button', { name: 'Rename Workspace' }))
    expect(screen.getByRole('button', { name: 'Renaming…' })).toBeDisabled()
    await act(async () => pending.resolve({ ...globex, name: 'Globex Corp' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('labels the remove button while the workspace is being removed', async () => {
    const pending = deferred<{ workspaces: (typeof acme)[]; activeLocalId: string | null }>()
    vi.mocked(services.removeWorkspace).mockReturnValue(pending.promise)
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Remove Workspace…' }))
    const dialog = screen.getByRole('dialog', { name: 'Remove “Globex”?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove Workspace' }))
    expect(within(dialog).getByRole('button', { name: 'Removing…' })).toBeDisabled()
    await act(async () => pending.resolve({ workspaces: [acme], activeLocalId: null }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})

describe('MissingFolderNotice', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.resetAllMocks()
    mockDefaults()
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: null, missingLocalId: 'A', status: 'ready', error: null })
  })

  afterEach(() => resetWorkspaceStores())

  it('offers to locate the folder or remove the workspace', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/moved/acme')
    vi.mocked(services.locateWorkspace).mockResolvedValue({ ...acme, path: '/moved/acme', missing: false })
    vi.mocked(services.activateWorkspace).mockResolvedValue(snapshot({ ...acme, missing: false }))
    render(<MissingFolderNotice localId="A" />)
    expect(screen.getByRole('region', { name: 'Workspace folder missing' })).toHaveTextContent('/w/acme')
    await userEvent.click(screen.getByRole('button', { name: 'Locate Folder…' }))
    expect(services.locateWorkspace).toHaveBeenCalledWith('A', '/moved/acme')
    expect(useWorkspacesStore.getState()).toMatchObject({ activeLocalId: 'A', missingLocalId: null })
  })

  it('asks for confirmation before removing the workspace from Xenon, keeping its files', async () => {
    vi.mocked(services.removeWorkspace).mockResolvedValue({ workspaces: [globex], activeLocalId: null })
    render(<MissingFolderNotice localId="A" />)
    await userEvent.click(screen.getByRole('button', { name: 'Remove from Xenon…' }))
    expect(services.removeWorkspace).not.toHaveBeenCalled()
    const dialog = screen.getByRole('dialog', { name: 'Remove “Acme”?' })
    expect(within(dialog).getByRole('radio', { name: 'Remove from Xenon (the files stay on disk)' })).toBeChecked()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove Workspace' }))
    expect(services.removeWorkspace).toHaveBeenCalledWith('A', false)
    expect(useWorkspacesStore.getState().missingLocalId).toBeNull()
  })
})
