import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Environment } from '@/platform/mappers/environment'

vi.mock('@/platform/services/environments', () => ({
  DUPLICATE_ENVIRONMENT_NAME: 'ENV-02409',
  listEnvironments: vi.fn(),
  setActiveEnvironment: vi.fn(),
  getEnvironment: vi.fn(),
  updateEnvironment: vi.fn(),
}))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(),
  saveVariableTable: vi.fn(),
  resetVariables: vi.fn(),
  persistVariables: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  revealSecret: vi.fn(),
  copySecret: vi.fn(),
}))

const services = await import('@/platform/services/environments')
const variables = await import('@/platform/services/variables')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { EnvironmentTab } = await import('./EnvironmentTab')
const { GlobalsTab } = await import('./GlobalsTab')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const
const broken = { id: 'error:x.yaml', name: 'x.yaml', color: 'none', confirmBeforeSend: false, error: 'x.yaml cannot be read' } as const
const detail = (id: string, name: string): Environment => ({ id, name, description: 'local', color: 'green', confirmBeforeSend: false, variables: [] })

const ROW = {
  key: 'url',
  initialValue: 'A',
  currentValue: 'A',
  hasCurrentOverride: false,
  secret: false,
  secretState: null,
  enabled: true,
  description: '',
  duplicate: false,
  surroundingSpaces: false,
}

describe('EnvironmentTab', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ environments: [dev, hml, broken], activeId: 'e1', status: 'ready' })
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [ROW], warnings: [] }))
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('shows the name, colour, "Confirm before send" and the variables of the environment', async () => {
    render(<EnvironmentTab environmentId="e1" />)
    const tab = screen.getByRole('region', { name: 'dev environment' })
    expect(within(tab).getByRole('heading', { name: 'dev' })).toBeInTheDocument()
    expect(within(tab).getByText('Active')).toBeInTheDocument()
    expect(within(tab).getByRole('textbox', { name: 'Environment name' })).toHaveValue('dev')
    expect(within(tab).getByRole('radio', { name: 'Neon green' })).toBeChecked()
    expect(within(tab).getByRole('checkbox', { name: 'Confirm before send' })).not.toBeChecked()
    expect(await within(tab).findByRole('table', { name: 'dev variables' })).toBeInTheDocument()
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'env:e1')
  })

  it('edits the variables in place and leaves saving and discarding to the tab, as request tabs do', async () => {
    render(<EnvironmentTab environmentId="e1" />)
    await userEvent.type(await screen.findByRole('textbox', { name: 'Description of url' }), 'x')
    expect(useEnvironmentsStore.getState().drafts['env:e1']?.dirty).toBe(true)
    expect(screen.queryByRole('button', { name: 'Discard Changes' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Save Variables' })).not.toBeInTheDocument()
  })

  it('has no save buttons on the globals tab either', async () => {
    render(<GlobalsTab />)
    expect(await screen.findByRole('table', { name: 'Global variables' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Save Variables' })).not.toBeInTheDocument()
  })

  it('keeps unsaved edits while the tab is hidden', async () => {
    const { unmount } = render(<EnvironmentTab environmentId="e1" />)
    await userEvent.type(await screen.findByRole('textbox', { name: 'Description of url' }), 'kept')
    unmount()
    render(<EnvironmentTab environmentId="e1" />)
    expect(await screen.findByRole('textbox', { name: 'Description of url' })).toHaveValue('kept')
  })

  it('renames the environment and follows a rename made elsewhere', async () => {
    vi.mocked(services.getEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.updateEnvironment).mockResolvedValue(detail('e1', 'local'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [{ ...dev, name: 'local' }, hml], activeId: 'e1' })
    render(<EnvironmentTab environmentId="e1" />)
    const name = screen.getByRole('textbox', { name: 'Environment name' })
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(services.getEnvironment).not.toHaveBeenCalled()
    await userEvent.clear(name)
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(screen.getByRole('alert')).toHaveTextContent('Type a name for the environment.')
    await userEvent.type(name, 'local')
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', expect.objectContaining({ id: 'e1', name: 'local' }))
    expect(await screen.findByRole('heading', { name: 'local' })).toBeInTheDocument()
    act(() => useEnvironmentsStore.setState({ environments: [{ ...dev, name: 'qa' }, hml] }))
    expect(screen.getByRole('textbox', { name: 'Environment name' })).toHaveValue('qa')
  })

  it('shows a duplicate name and a failed colour change', async () => {
    vi.mocked(services.getEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.updateEnvironment).mockRejectedValue(new AppError('ENV-02409', 'taken'))
    render(<EnvironmentTab environmentId="e1" />)
    const name = screen.getByRole('textbox', { name: 'Environment name' })
    await userEvent.clear(name)
    await userEvent.type(name, 'HML')
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(await screen.findByText('An environment named “HML” already exists. Choose another name.')).toBeInTheDocument()
    vi.mocked(services.getEnvironment).mockRejectedValue(new AppError('ENV-01404', 'environment gone'))
    await userEvent.click(screen.getByRole('checkbox', { name: 'Confirm before send' }))
    expect(await screen.findByText(/environment gone/)).toBeInTheDocument()
  })

  it('changes the colour and "Confirm before send"', async () => {
    vi.mocked(services.getEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.updateEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [{ ...dev, color: 'red' }, hml], activeId: 'e1' })
    render(<EnvironmentTab environmentId="e1" />)
    await userEvent.click(screen.getByRole('radio', { name: 'Neon red' }))
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', { id: 'e1', name: 'dev', description: 'local', color: 'red', confirmBeforeSend: false })
    await userEvent.click(screen.getByRole('checkbox', { name: 'Confirm before send' }))
    expect(services.updateEnvironment).toHaveBeenLastCalledWith('A', { id: 'e1', name: 'dev', description: 'local', color: 'red', confirmBeforeSend: true })
  })

  it('tells when the environment is gone or its file is broken', () => {
    const { unmount } = render(<EnvironmentTab environmentId="gone" />)
    expect(screen.getByRole('region', { name: 'Environment not found' })).toBeInTheDocument()
    unmount()
    render(<EnvironmentTab environmentId="error:x.yaml" />)
    expect(screen.getByText(/x.yaml cannot be read/)).toBeInTheDocument()
  })

  it('shows the globals with their variables', async () => {
    render(<GlobalsTab />)
    expect(screen.getByRole('heading', { name: 'Globals' })).toBeInTheDocument()
    expect(await screen.findByRole('table', { name: 'Global variables' })).toBeInTheDocument()
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'globals')
  })
})
