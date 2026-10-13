import { fireEvent, render, screen, within } from '@testing-library/react'
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
vi.mock('@/platform/services/variables', () => ({ CONFIRM_REVEAL_REQUIRED: 'VAR-01409', getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409' }))
vi.mock('@/platform/services/export', () => ({ getExportOptions: vi.fn(), pickExportTarget: vi.fn(), runExport: vi.fn(), showExportInFolder: vi.fn() }))

const services = await import('@/platform/services/environments')
const exportService = await import('@/platform/services/export')
const { useExportStore } = await import('@/features/export/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { EnvironmentsView } = await import('./EnvironmentsView')

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const
const broken = { id: 'error:x.yaml', name: 'x.yaml', color: 'none', confirmBeforeSend: false, error: 'x.yaml cannot be read' } as const
const detail = (id: string, name: string): Environment => ({ id, name, description: '', color: 'none', confirmBeforeSend: false, variables: [] })

const onOpen = vi.fn()

function renderView(selected: Parameters<typeof EnvironmentsView>[0]['selected'] = null) {
  render(<EnvironmentsView selected={selected} onOpen={onOpen} />)
  return screen.getByRole('list', { name: 'Environments' })
}

function openMenu(name: string) {
  fireEvent.contextMenu(screen.getByRole('button', { name }), { clientX: 10, clientY: 10 })
}

describe('EnvironmentsView', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ environments: [dev, hml, broken], activeId: 'e1', status: 'ready' })
  })

  afterEach(() => {
    useEnvironmentsStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
    useExportStore.getState().reset()
  })

  it('lists the globals and every environment with its colour, marks the active one and shows a broken file', () => {
    const list = renderView({ kind: 'environment', environmentId: 'e2' })
    expect(within(list).getAllByRole('button').map((item) => item.textContent)).toEqual(['◎Globals', 'devActive', 'hml'])
    expect(within(list).getByRole('button', { name: 'hml' })).toHaveAttribute('aria-current', 'true')
    expect(list).toHaveTextContent('x.yaml cannot be read. Fix or remove the file in the workspace folder.')
  })

  it('previews an entry on a click and pins it on a double click or Mod+Enter', async () => {
    renderView()
    await userEvent.click(screen.getByRole('button', { name: 'dev, active' }))
    expect(onOpen).toHaveBeenLastCalledWith({ kind: 'environment', environmentId: 'e1' }, 'preview')
    await userEvent.dblClick(screen.getByRole('button', { name: 'Globals' }))
    expect(onOpen).toHaveBeenLastCalledWith({ kind: 'globals' }, 'pinned')
    screen.getByRole('button', { name: 'hml' }).focus()
    await userEvent.keyboard('{Enter}')
    expect(onOpen).toHaveBeenLastCalledWith({ kind: 'environment', environmentId: 'e2' }, 'preview')
    await userEvent.keyboard('{Control>}{Enter}{/Control}')
    expect(onOpen).toHaveBeenLastCalledWith({ kind: 'environment', environmentId: 'e2' }, 'pinned')
  })

  it('creates an environment and opens it in a pinned tab', async () => {
    vi.mocked(services.createEnvironment).mockResolvedValue(detail('e3', 'qa'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [dev, hml, { ...dev, id: 'e3', name: 'qa' }], activeId: 'e1' })
    renderView()
    await userEvent.click(screen.getByRole('button', { name: 'New Environment' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    expect(input).toHaveValue('New Environment')
    await userEvent.clear(input)
    await userEvent.type(input, 'qa{Enter}')
    expect(services.createEnvironment).toHaveBeenCalledWith('A', 'qa')
    expect(onOpen).toHaveBeenCalledWith({ kind: 'environment', environmentId: 'e3' }, 'pinned')
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
  })

  it('shows a duplicate name while creating and cancels with Escape', async () => {
    vi.mocked(services.createEnvironment).mockRejectedValue(new AppError('ENV-02409', 'name taken'))
    renderView()
    await userEvent.click(screen.getByRole('button', { name: 'New Environment' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'DEV{Enter}')
    expect(await screen.findByRole('alert')).toHaveTextContent('An environment named “DEV” already exists. Choose another name.')
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
    expect(onOpen).not.toHaveBeenCalled()
  })

  it('renames an environment from its context menu', async () => {
    vi.mocked(services.getEnvironment).mockResolvedValue(detail('e2', 'hml'))
    vi.mocked(services.updateEnvironment).mockResolvedValue(detail('e2', 'staging'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [dev, { ...hml, name: 'staging' }], activeId: 'e1' })
    renderView()
    openMenu('hml')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'staging{Enter}')
    expect(services.updateEnvironment).toHaveBeenCalledWith('A', expect.objectContaining({ id: 'e2', name: 'staging' }))
    expect(await screen.findByRole('button', { name: 'staging' })).toBeInTheDocument()
  })

  it('duplicates, exports and activates an environment from its context menu', async () => {
    vi.mocked(services.duplicateEnvironment).mockResolvedValue(detail('e4', 'hml Copy'))
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [dev, hml, { ...hml, id: 'e4', name: 'hml Copy' }], activeId: 'e1' })
    vi.mocked(services.setActiveEnvironment).mockResolvedValue('e2')
    vi.mocked(exportService.getExportOptions).mockResolvedValue({ suggestedFileName: 'hml', formats: [], hasSecrets: false, hasCurrentValues: false })
    renderView()
    openMenu('hml')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Duplicate' }))
    expect(services.duplicateEnvironment).toHaveBeenCalledWith('A', 'e2')
    expect(onOpen).toHaveBeenLastCalledWith({ kind: 'environment', environmentId: 'e4' }, 'preview')
    openMenu('hml')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Export…' }))
    expect(exportService.getExportOptions).toHaveBeenCalledWith({ localId: 'A', kind: 'environment', itemId: 'e2', name: 'hml' })
    openMenu('dev, active')
    expect(screen.getByRole('menuitem', { name: 'Set as Active' })).toHaveAttribute('aria-disabled', 'true')
    await userEvent.keyboard('{Escape}')
    openMenu('hml')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Set as Active' }))
    expect(services.setActiveEnvironment).toHaveBeenCalledWith('A', 'e2')
  })

  it('shows a failed duplicate in a dismissable banner', async () => {
    vi.mocked(services.duplicateEnvironment).mockRejectedValue(new AppError('ENV-05500', 'disk full'))
    renderView()
    openMenu('dev, active')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Duplicate' }))
    expect(await screen.findByText(/disk full/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Error' }))
    expect(screen.queryByText(/disk full/)).not.toBeInTheDocument()
  })

  it('deletes an environment only after confirmation', async () => {
    vi.mocked(services.deleteEnvironment).mockResolvedValue(null)
    vi.mocked(services.listEnvironments).mockResolvedValue({ environments: [hml], activeId: null })
    renderView()
    openMenu('dev, active')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    const dialog = screen.getByRole('dialog', { name: 'Delete “dev”?' })
    expect(dialog).toHaveTextContent('Its file, local current values and secret values are removed.')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(services.deleteEnvironment).not.toHaveBeenCalled()
    openMenu('dev, active')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    await userEvent.click(screen.getByRole('button', { name: 'Delete Environment' }))
    expect(services.deleteEnvironment).toHaveBeenCalledWith('A', 'e1')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'dev, active' })).not.toBeInTheDocument()
  })

  it('invites to create the first environment', () => {
    useEnvironmentsStore.setState({ environments: [] })
    renderView()
    expect(screen.getByText(/No environments yet/)).toBeInTheDocument()
  })

  it('filters the list by name as the search is typed', async () => {
    const list = renderView()
    await userEvent.type(screen.getByRole('searchbox', { name: 'Search environments' }), 'HM')
    expect(within(list).getAllByRole('button').map((item) => item.textContent)).toEqual(['hml'])
    await userEvent.clear(screen.getByRole('searchbox', { name: 'Search environments' }))
    await userEvent.type(screen.getByRole('searchbox', { name: 'Search environments' }), 'nothing')
    expect(within(list).queryAllByRole('button')).toEqual([])
    expect(screen.getByText('No environments match “nothing”.')).toBeInTheDocument()
  })

  it('opens the import dialog from the toolbar', async () => {
    const { useImportStore } = await import('@/features/import-export/store')
    renderView()
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(useImportStore.getState().step).toBe('choose')
    useImportStore.getState().close()
  })
})
