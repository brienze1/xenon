import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const { useWorkspacesStore } = await import('../store')
const { WorkspaceSwitcher } = await import('./WorkspaceSwitcher')

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: true }
const globex = { localId: 'G', name: 'Globex', path: '/w/globex', missing: false }

describe('WorkspaceSwitcher', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('shows the active workspace and switches to another one', async () => {
    useWorkspacesStore.setState({ workspaces: [acme, globex], activeLocalId: 'G', status: 'ready', error: null })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: { ...acme, missing: false } })
    render(<WorkspaceSwitcher />)
    await userEvent.click(screen.getByRole('button', { name: 'Workspace: Globex' }))
    expect(screen.getByRole('menuitemradio', { name: 'Globex' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('menuitemradio', { name: /^Acme/ })).toHaveAttribute('aria-checked', 'false')
    expect(screen.getByText('(missing)')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('menuitemradio', { name: /^Acme/ }))
    expect(services.activateWorkspace).toHaveBeenCalledWith('A')
  })

  it('says when there is no workspace', () => {
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'ready', error: null })
    render(<WorkspaceSwitcher />)
    expect(screen.getByRole('button', { name: 'Workspace: none' })).toHaveTextContent('No workspace')
  })
})
