import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { defaultUIState } from '@/platform/mappers/ui-state'

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
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), releaseResponse: vi.fn(async () => undefined), getRequest: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
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

describe('WorkspaceSwitcher', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G', missingLocalId: null, status: 'ready', error: null })
  })

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

  it('says when there is no workspace', () => {
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null })
    render(<WorkspaceSwitcher />)
    expect(screen.getByRole('button', { name: 'Workspace: none' })).toHaveTextContent('No workspace')
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
    expect(within(dialog).getByText('Choose a folder for the workspace.')).toBeInTheDocument()
    expect(services.createWorkspace).not.toHaveBeenCalled()
  })

  it('opens an existing workspace folder and reports a refusal', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w')
    vi.mocked(services.openWorkspace).mockRejectedValue(new AppError('WKS-02409', 'overlap'))
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Open Workspace…' }))
    expect(services.openWorkspace).toHaveBeenCalledWith('/w')
    expect(await screen.findByRole('alert')).toHaveTextContent('already used by another workspace')
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
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
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(services.renameWorkspace).toHaveBeenCalledWith('G', 'Globex Corp')
    expect(screen.getByRole('button', { name: 'Workspace: Globex Corp' })).toBeInTheDocument()
  })

  it('refuses a blank workspace name on rename', async () => {
    render(<WorkspaceSwitcher />)
    await openMenu()
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename Workspace…' }))
    const input = screen.getByRole('textbox', { name: 'Workspace name' })
    await userEvent.clear(input)
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(input).toHaveFocus()
    expect(input).toHaveAccessibleDescription('Enter a name for the workspace.')
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
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove' }))
    expect(services.removeWorkspace).toHaveBeenCalledWith('G', true)
    expect(within(dialog).getByRole('alert')).toHaveTextContent('could not be moved to the trash')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove from Xenon' }))
    expect(services.removeWorkspace).toHaveBeenLastCalledWith('G', false)
    expect(useWorkspacesStore.getState()).toMatchObject({ workspaces: [acme], activeLocalId: null })
  })
})

describe('MissingFolderNotice', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: null, missingLocalId: 'A', status: 'ready', error: null })
  })

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

  it('removes the workspace from Xenon keeping its files', async () => {
    vi.mocked(services.removeWorkspace).mockResolvedValue({ workspaces: [globex], activeLocalId: null })
    render(<MissingFolderNotice localId="A" />)
    await userEvent.click(screen.getByRole('button', { name: 'Remove from Xenon' }))
    expect(services.removeWorkspace).toHaveBeenCalledWith('A', false)
    expect(useWorkspacesStore.getState().missingLocalId).toBeNull()
  })
})
