import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('../store')
const { WelcomeScreen } = await import('./WelcomeScreen')

describe('WelcomeScreen', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'ready', error: null })
  })

  it('creates a workspace in the folder chosen with the native picker', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/work/acme')
    vi.mocked(services.createWorkspace).mockResolvedValue({ localId: 'A', name: 'Acme', path: '/work/acme', missing: false })
    render(<WelcomeScreen />)
    const submit = screen.getByRole('button', { name: 'Create workspace' })
    expect(submit).toBeDisabled()
    await userEvent.type(screen.getByRole('textbox', { name: 'Workspace name' }), 'Acme')
    expect(submit).toBeDisabled()
    await userEvent.click(screen.getByRole('button', { name: 'Choose folder…' }))
    expect(dialogs.pickFolder).toHaveBeenCalledWith('Choose the folder for the new workspace')
    expect(screen.getByText('/work/acme')).toBeInTheDocument()
    await userEvent.click(submit)
    expect(services.createWorkspace).toHaveBeenCalledWith({ name: 'Acme', path: '/work/acme' })
    expect(useWorkspacesStore.getState().activeLocalId).toBe('A')
  })

  it('keeps the form when the picker is cancelled', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue(null)
    render(<WelcomeScreen />)
    await userEvent.click(screen.getByRole('button', { name: 'Choose folder…' }))
    expect(screen.getByText('No folder chosen')).toBeInTheDocument()
  })

  it('explains why a folder was refused', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/work/full')
    vi.mocked(services.createWorkspace).mockRejectedValue(new AppError('WKS-01409', 'not empty'))
    render(<WelcomeScreen />)
    await userEvent.type(screen.getByRole('textbox', { name: 'Workspace name' }), 'Acme')
    await userEvent.click(screen.getByRole('button', { name: 'Choose folder…' }))
    await userEvent.click(screen.getByRole('button', { name: 'Create workspace' }))
    expect(screen.getByRole('alert')).toHaveTextContent('That folder is not empty')
    expect(screen.getByRole('button', { name: 'Create workspace' })).toBeEnabled()
  })
})
