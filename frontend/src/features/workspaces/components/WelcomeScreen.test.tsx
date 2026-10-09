import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), releaseResponse: vi.fn(async () => undefined), getRequest: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { WelcomeScreen } = await import('./WelcomeScreen')

describe('WelcomeScreen', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'ready', error: null })
  })

  it('creates a workspace in the folder chosen with the native picker', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/work/acme')
    const acme = { localId: 'A', name: 'Acme', path: '/work/acme', missing: false }
    vi.mocked(services.createWorkspace).mockResolvedValue(acme)
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme, tree: [], uiState: defaultUIState() })
    render(<WelcomeScreen />)
    const submit = screen.getByRole('button', { name: 'Create Workspace' })
    expect(submit).toBeEnabled()
    await userEvent.type(screen.getByRole('textbox', { name: 'Workspace name' }), 'Acme')
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder…' }))
    expect(dialogs.pickFolder).toHaveBeenCalledWith('Choose the folder for the new workspace')
    expect(screen.getByText('/work/acme')).toBeInTheDocument()
    await userEvent.click(submit)
    expect(services.createWorkspace).toHaveBeenCalledWith({ name: 'Acme', path: '/work/acme' })
    expect(useWorkspacesStore.getState().activeLocalId).toBe('A')
  })

  it('validates on submit and focuses the first invalid field', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/work/acme')
    render(<WelcomeScreen />)
    const nameInput = screen.getByRole('textbox', { name: 'Workspace name' })
    await userEvent.click(screen.getByRole('button', { name: 'Create Workspace' }))
    expect(services.createWorkspace).not.toHaveBeenCalled()
    expect(nameInput).toHaveFocus()
    expect(nameInput).toHaveAttribute('aria-invalid', 'true')
    expect(nameInput).toHaveAccessibleDescription('Enter a name for the workspace.')
    expect(screen.getByRole('button', { name: 'Choose Folder…' })).toHaveAccessibleDescription('Choose a folder for the workspace.')
    await userEvent.type(nameInput, 'Acme')
    expect(nameInput).toHaveAttribute('aria-invalid', 'false')
    await userEvent.click(screen.getByRole('button', { name: 'Create Workspace' }))
    expect(screen.getByRole('button', { name: 'Choose Folder…' })).toHaveFocus()
    expect(services.createWorkspace).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder…' }))
    expect(screen.queryByText('Choose a folder for the workspace.')).not.toBeInTheDocument()
  })

  it('keeps the form when the picker is cancelled', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue(null)
    render(<WelcomeScreen />)
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder…' }))
    expect(screen.getByText('No folder chosen')).toBeInTheDocument()
  })

  it('explains why a folder was refused', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/work/full')
    vi.mocked(services.createWorkspace).mockRejectedValue(new AppError('WKS-01409', 'not empty'))
    render(<WelcomeScreen />)
    await userEvent.type(screen.getByRole('textbox', { name: 'Workspace name' }), 'Acme')
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder…' }))
    await userEvent.click(screen.getByRole('button', { name: 'Create Workspace' }))
    expect(screen.getByRole('alert')).toHaveTextContent('That folder is not empty')
    expect(screen.getByRole('button', { name: 'Create Workspace' })).toBeEnabled()
  })
})
