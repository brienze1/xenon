import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/events', () => ({ onWorkspaceFolderMissing: vi.fn(() => () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(async () => []),
  sendRequest: vi.fn(),
  lastResult: vi.fn(async () => null),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))

const services = await import('@/platform/services/workspaces')
const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
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
    expect(screen.getByRole('textbox', { name: 'Request URL' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'No response yet' })).toBeInTheDocument()
    expect(screen.getByRole('contentinfo')).toHaveTextContent('/work/acme')
  })

  it('sends the active tab with Cmd/Ctrl+Enter', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme })
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    render(<App />)
    await screen.findByRole('textbox', { name: 'Request URL' })
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', metaKey: true, ctrlKey: false }))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', ctrlKey: true, metaKey: false }))
    const tabId = useRequestBuilderStore.getState().activeTabId
    expect(requests.sendRequest).toHaveBeenCalledTimes(1)
    expect(requests.sendRequest).toHaveBeenCalledWith('A', tabId, expect.objectContaining({ method: 'GET' }))
  })
})
