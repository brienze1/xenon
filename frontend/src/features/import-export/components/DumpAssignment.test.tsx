import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { ImportPreview } from '@/platform/mappers/import'

vi.mock('@/platform/services/import', () => ({
  detectImport: vi.fn(),
  discardImport: vi.fn(),
  findImportClashes: vi.fn(),
  pickImportSources: vi.fn(),
  runImport: vi.fn(),
}))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn() }))
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

const service = await import('@/platform/services/import')
const tree = await import('@/platform/services/tree')
const workspaces = await import('@/platform/services/workspaces')
const dialogs = await import('@/platform/services/dialogs')
const { useImportStore } = await import('@/features/import-export/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { ImportDialog } = await import('./ImportDialog')

const none = { collections: 0, folders: 0, requests: 0, examples: 0, variables: 0, scripts: 0, environments: 0 }
const preview: ImportPreview = {
  sourceId: 'S',
  kind: 'postman_dump',
  version: '',
  items: [
    { key: '0', kind: 'collection', name: 'Billing API', counts: { ...none, collections: 1, requests: 2 }, warningsCount: 0 },
    { key: '1', kind: 'collection', name: 'CRM API', counts: { ...none, collections: 1, requests: 1 }, warningsCount: 0 },
    { key: '2', kind: 'environment', name: 'crm-prod', counts: { ...none, environments: 1, variables: 1 }, warningsCount: 0 },
  ],
  warnings: [
    {
      code: 'skipped_part',
      severity: 'warning',
      message: "The dump's mock part (1 file) is not a collection or an environment and was skipped",
      itemKey: '',
      itemId: '',
      localId: '',
      path: 'mock',
      location: '',
    },
  ],
  pmUsages: [],
  encryptedSecrets: false,
}
const initech = { localId: 'I', name: 'Initech', path: '/w/initech', missing: false }

function targets() {
  return useImportStore.getState().targets
}

describe('DumpAssignment', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useImportStore.getState().reset()
    useWorkspacesStore.setState({
      activeLocalId: 'A',
      error: null,
      workspaces: [
        { localId: 'A', name: 'Acme', path: '/w/acme', missing: false },
        { localId: 'B', name: 'Globex', path: '/w/globex', missing: false },
      ],
    })
    vi.mocked(service.discardImport).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([])
    useImportStore.setState({
      step: 'preview',
      preview,
      targets: {
        '0': { include: true, localId: 'A', parentId: '' },
        '1': { include: true, localId: 'A', parentId: '' },
        '2': { include: true, localId: 'A', parentId: '' },
      },
    })
  })

  afterEach(() => {
    useImportStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null, workspaces: [], error: null })
  })

  it('lists every item with its kind and counts, a workspace picker each and the skipped parts', () => {
    render(<ImportDialog />)
    expect(screen.getByText('Postman data dump')).toBeInTheDocument()
    const rows = screen.getAllByRole('row').slice(1)
    expect(rows).toHaveLength(3)
    expect(within(rows[2]!).getByText('Environment')).toBeInTheDocument()
    expect(within(rows[2]!).getByText('1 variable · 1 environment')).toBeInTheDocument()
    expect(within(rows[0]!).getByText('Collection')).toBeInTheDocument()
    expect(screen.getByRole('combobox', { name: 'Workspace for CRM API' })).toHaveValue('A')
    expect(screen.queryByRole('combobox', { name: 'Location for CRM API' })).not.toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Not Imported' })).toHaveTextContent("The dump's mock part (1 file)")
  })

  it('assigns each item to its own workspace and the included ones in bulk', async () => {
    render(<ImportDialog />)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Workspace for CRM API' }), 'B')
    expect(targets()['1']).toEqual({ include: true, localId: 'B', parentId: '' })
    await userEvent.click(screen.getByRole('checkbox', { name: 'Import Billing API' }))
    expect(screen.getByRole('combobox', { name: 'Workspace for Billing API' })).toBeDisabled()
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Workspace for the included items' }), 'B')
    await userEvent.click(screen.getByRole('button', { name: 'Assign' }))
    expect(targets()).toEqual({
      '0': { include: false, localId: 'A', parentId: '' },
      '1': { include: true, localId: 'B', parentId: '' },
      '2': { include: true, localId: 'B', parentId: '' },
    })
  })

  it('creates a new workspace inline for one item without leaving the dialog', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/initech')
    vi.mocked(workspaces.createWorkspace).mockRejectedValueOnce(new AppError('WKS-01409', 'the folder is not empty'))
    render(<ImportDialog />)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Workspace for crm-prod' }), 'New workspace…')
    const form = screen.getByRole('form', { name: 'New workspace' })
    const create = within(form).getByRole('button', { name: 'Create Workspace' })
    expect(create).toBeDisabled()
    await userEvent.type(within(form).getByRole('textbox', { name: 'Workspace name' }), 'Initech')
    await userEvent.click(within(form).getByRole('button', { name: 'Choose Folder…' }))
    expect(within(form).getByText('/w/initech')).toBeInTheDocument()
    await userEvent.click(create)
    expect(await within(form).findByRole('alert')).toHaveTextContent('the folder is not empty. Choose another folder or name, then try again.')
    expect(within(form).getByRole('textbox', { name: 'Workspace name' })).toHaveAttribute('aria-invalid', 'true')
    vi.mocked(workspaces.createWorkspace).mockResolvedValueOnce(initech)
    await userEvent.click(create)
    await vi.waitFor(() => expect(screen.queryByRole('form', { name: 'New workspace' })).not.toBeInTheDocument())
    expect(workspaces.createWorkspace).toHaveBeenLastCalledWith({ name: 'Initech', path: '/w/initech' })
    expect(workspaces.activateWorkspace).not.toHaveBeenCalled()
    expect(targets()['2']).toEqual({ include: true, localId: 'I', parentId: '' })
    expect(screen.getByRole('combobox', { name: 'Workspace for Billing API' })).toHaveTextContent('Initech')
  })

  it('creates a new workspace for every included item from the bulk picker and can cancel the form', async () => {
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/initech')
    vi.mocked(workspaces.createWorkspace).mockResolvedValue(initech)
    render(<ImportDialog />)
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Workspace for the included items' }), 'New workspace…')
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('form', { name: 'New workspace' })).not.toBeInTheDocument()
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Workspace for the included items' }), 'New workspace…')
    const form = screen.getByRole('form', { name: 'New workspace' })
    await userEvent.type(within(form).getByRole('textbox', { name: 'Workspace name' }), 'Initech')
    await userEvent.click(within(form).getByRole('button', { name: 'Choose Folder…' }))
    await userEvent.click(within(form).getByRole('button', { name: 'Create Workspace' }))
    await vi.waitFor(() => expect(targets()['0']?.localId).toBe('I'))
    expect(Object.values(targets()).map((target) => target.localId)).toEqual(['I', 'I', 'I'])
  })

  it('places environments and globals without a location in a single import', () => {
    useImportStore.setState({ preview: { ...preview, kind: 'postman_environment', warnings: [] } })
    render(<ImportDialog />)
    expect(screen.getByRole('combobox', { name: 'Location for Billing API' })).toBeInTheDocument()
    expect(screen.queryByRole('combobox', { name: 'Location for crm-prod' })).not.toBeInTheDocument()
    expect(screen.getByText('Environments')).toBeInTheDocument()
  })
})
