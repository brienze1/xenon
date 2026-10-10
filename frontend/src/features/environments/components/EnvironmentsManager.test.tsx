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
  createEnvironment: vi.fn(),
  updateEnvironment: vi.fn(),
  duplicateEnvironment: vi.fn(),
  deleteEnvironment: vi.fn(),
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
const { EnvironmentsManager } = await import('./EnvironmentsManager')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const
const broken = { id: 'error:x.yaml', name: 'x.yaml', color: 'none', confirmBeforeSend: false, error: 'x.yaml cannot be read' } as const
const detail = (id: string, name: string): Environment => ({ id, name, description: '', color: 'none', confirmBeforeSend: false, variables: [] })

function open() {
  useEnvironmentsStore.setState({ environments: [dev, hml, broken], activeId: 'e1', status: 'ready' })
  render(<EnvironmentsManager />)
  act(() => useEnvironmentsStore.getState().openEditor('manager'))
  return screen.getByRole('dialog', { name: 'Environments' })
}

describe('EnvironmentsManager', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [], warnings: [] }))
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('lists the environments, shows a broken one with its error and edits the selected one', async () => {
    const dialog = open()
    const list = within(dialog).getByRole('list', { name: 'Environments' })
    expect(within(list).getAllByRole('button').map((item) => item.textContent)).toEqual(['dev', 'hml'])
    expect(within(list).getByText('x.yaml cannot be read')).toBeInTheDocument()
    expect(within(list).getByRole('button', { name: 'dev' })).toHaveAttribute('aria-current', 'true')
    expect(await screen.findByRole('table', { name: 'dev variables' })).toBeInTheDocument()
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'env:e1')
    await userEvent.click(within(list).getByRole('button', { name: 'hml' }))
    expect(await screen.findByRole('table', { name: 'hml variables' })).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'Environment name' })).toHaveValue('hml')
  })

  it('creates an environment and selects it', async () => {
    open()
    vi.mocked(services.createEnvironment).mockResolvedValue(detail('e3', 'qa'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [dev, hml, { ...dev, id: 'e3', name: 'qa' }], activeId: 'e1' })
    const create = screen.getByRole('button', { name: 'Create' })
    expect(create).toBeDisabled()
    await userEvent.type(screen.getByRole('textbox', { name: 'New environment name' }), 'qa{Enter}')
    expect(services.createEnvironment).toHaveBeenCalledWith('A', 'qa')
    expect(await screen.findByRole('table', { name: 'qa variables' })).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'New environment name' })).toHaveValue('')
  })

  it('shows a duplicate name inline', async () => {
    open()
    vi.mocked(services.createEnvironment).mockRejectedValue(new AppError('ENV-02409', 'name taken'))
    await userEvent.type(screen.getByRole('textbox', { name: 'New environment name' }), 'DEV')
    await userEvent.click(screen.getByRole('button', { name: 'Create' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('An environment named “DEV” already exists. Choose another name.')
  })

  it('renames and duplicates the selected environment', async () => {
    open()
    vi.mocked(services.getEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.updateEnvironment).mockResolvedValue(detail('e1', 'local'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [{ ...dev, name: 'local' }, hml], activeId: 'e1' })
    const name = screen.getByRole('textbox', { name: 'Environment name' })
    await userEvent.clear(name)
    await userEvent.type(name, 'local')
    await userEvent.click(screen.getByRole('button', { name: 'Rename' }))
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', expect.objectContaining({ id: 'e1', name: 'local' }))
    expect(await screen.findByRole('button', { name: 'local' })).toBeInTheDocument()
    vi.mocked(services.duplicateEnvironment).mockResolvedValue(detail('e4', 'local Copy'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [{ ...dev, name: 'local' }, { ...dev, id: 'e4', name: 'local Copy' }, hml], activeId: 'e1' })
    await userEvent.click(screen.getByRole('button', { name: 'Duplicate' }))
    expect(services.duplicateEnvironment).toHaveBeenCalledWith('A', 'e1')
    expect(await screen.findByRole('table', { name: 'local Copy variables' })).toBeInTheDocument()
  })

  it('deletes only after confirmation', async () => {
    open()
    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    expect(screen.getByText(/local current values and secret values are removed/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(services.deleteEnvironment).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    vi.mocked(services.deleteEnvironment).mockResolvedValue(null)
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [hml], activeId: null })
    await userEvent.click(screen.getByRole('button', { name: 'Delete Environment' }))
    expect(services.deleteEnvironment).toHaveBeenCalledWith('A', 'e1')
    expect(await screen.findByRole('table', { name: 'hml variables' })).toBeInTheDocument()
    expect(useEnvironmentsStore.getState().activeId).toBeNull()
  })

  it('closes with the Close button', async () => {
    open()
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useEnvironmentsStore.getState().editor).toBeNull()
  })

  it('picks a neon colour and turns "Confirm before send" on for the selected environment', async () => {
    vi.mocked(services.getEnvironment).mockResolvedValue({ ...detail('e1', 'dev'), description: 'local', color: 'green' })
    vi.mocked(services.updateEnvironment).mockResolvedValue(detail('e1', 'dev'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [{ ...dev, color: 'red', confirmBeforeSend: true }, hml], activeId: 'e1' })
    const dialog = open()
    const colours = within(dialog).getByRole('group', { name: 'Colour' })
    expect(within(colours).getAllByRole('radio').map((radio) => radio.getAttribute('aria-label'))).toEqual([
      'No colour',
      'Neon green',
      'Neon cyan',
      'Neon magenta',
      'Neon yellow',
      'Neon red',
    ])
    expect(within(colours).getByRole('radio', { name: 'Neon green' })).toBeChecked()
    await userEvent.click(within(colours).getByRole('radio', { name: 'Neon red' }))
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', { id: 'e1', name: 'dev', description: 'local', color: 'red', confirmBeforeSend: false })
    expect(await within(dialog).findByRole('checkbox', { name: 'Confirm before send' })).toBeChecked()
    await userEvent.click(within(dialog).getByRole('checkbox', { name: 'Confirm before send' }))
    expect(services.updateEnvironment).toHaveBeenLastCalledWith('A', { id: 'e1', name: 'dev', description: 'local', color: 'red', confirmBeforeSend: false })
  })

  it('shows a failed colour change', async () => {
    vi.mocked(services.getEnvironment).mockRejectedValue(new AppError('ENV-01404', 'environment gone'))
    const dialog = open()
    await userEvent.click(within(dialog).getByRole('checkbox', { name: 'Confirm before send' }))
    expect(await within(dialog).findByText(/environment gone/)).toBeInTheDocument()
  })
})
