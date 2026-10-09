import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/events', () => ({ onWorkspaceFolderMissing: vi.fn(() => () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const services = await import('@/platform/services/workspaces')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { App } = await import('./App')

const acme = { localId: 'A', name: 'Acme', path: '/work/acme', missing: false }

describe('App', () => {
  beforeEach(() => {
    vi.mocked(services.listWorkspaces).mockReset()
    vi.mocked(services.activateWorkspace).mockReset()
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, status: 'idle', error: null })
  })

  it('shows the welcome screen when there is no workspace', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    render(<App />)
    expect(await screen.findByRole('form', { name: 'Create workspace' })).toBeInTheDocument()
    expect(screen.queryByRole('navigation', { name: 'Sidebar' })).not.toBeInTheDocument()
  })

  it('reopens the last active workspace in the full window layout', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme })
    render(<App />)
    expect(await screen.findByRole('button', { name: 'Workspace: Acme' })).toBeInTheDocument()
    expect(screen.getByRole('navigation', { name: 'Sidebar' })).toBeInTheDocument()
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Acme' })).toBeInTheDocument()
    expect(screen.getByRole('contentinfo')).toHaveTextContent('/work/acme')
  })
})
