import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { ImportPreview, ImportResult } from '@/platform/mappers/import'

vi.mock('@/platform/services/import', () => ({
  detectImport: vi.fn(),
  discardImport: vi.fn(),
  findImportClashes: vi.fn(),
  pickImportSources: vi.fn(),
  runImport: vi.fn(),
}))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn() }))

const service = await import('@/platform/services/import')
const tree = await import('@/platform/services/tree')
const dialogs = await import('@/platform/services/dialogs')
const workspaces = await import('@/platform/services/workspaces')
const { useImportStore } = await import('@/features/import-export/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { ImportDialog } = await import('./ImportDialog')

const counts = { collections: 1, folders: 3, requests: 14, examples: 3, variables: 3, scripts: 7, environments: 0 }
const preview: ImportPreview = {
  sourceId: 'S',
  kind: 'postman_collection',
  version: '2.1',
  items: [{ key: '0', kind: 'collection', name: 'Acme API', counts, warningsCount: 17 }],
  warnings: [],
  pmUsages: [],
  encryptedSecrets: false,
}
const result: ImportResult = {
  items: [{ key: '0', status: 'imported', localId: 'A', id: 'C', name: 'Acme API' }],
  workspaces: [{ localId: 'A', created: counts, replaced: 1, skipped: 2 }],
  warnings: [
    {
      code: 'unsupported_auth',
      severity: 'warning',
      message: 'Auth type hawk is not supported yet',
      itemKey: '0',
      itemId: 'R1',
      localId: 'A',
      path: 'Acme API/Users/Propfind',
      location: '',
    },
    { code: 'variable_kept', severity: 'info', message: 'base_url keeps its value', itemKey: '0', itemId: '', localId: 'A', path: 'Main', location: '' },
  ],
  pmUsages: [{ api: 'pm.cookies.jar', occurrences: [{ itemId: 'R2', localId: 'A', path: 'Acme API/Raw JSON', phase: 'preRequest', line: 2 }] }],
  failed: [{ itemKey: '1', name: 'CRM API', code: 'IMP-29500', message: 'permission denied' }],
}

describe('ImportDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useImportStore.getState().reset()
    useWorkspacesStore.setState({
      activeLocalId: 'A',
      workspaces: [
        { localId: 'A', name: 'Acme', path: '/w/acme', missing: false },
        { localId: 'B', name: 'Globex', path: '/w/globex', missing: false },
        { localId: 'M', name: 'Missing', path: '/w/gone', missing: true },
      ],
    })
    vi.mocked(service.discardImport).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([
      {
        id: 'C',
        kind: 'collection',
        name: 'Main',
        method: null,
        url: null,
        error: null,
        warning: null,
        children: [{ id: 'F', kind: 'folder', name: 'Target', method: null, url: null, error: null, warning: null, children: [] }],
      },
    ])
  })

  afterEach(() => {
    useImportStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null, workspaces: [] })
  })

  it('is hidden until opened and offers the picker, a drop zone and a paste box', async () => {
    render(<ImportDialog />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    act(() => useImportStore.getState().open())
    expect(screen.getByRole('dialog', { name: 'Import' })).toBeInTheDocument()
    expect(screen.getByText(/Drop Postman or Bruno exports/).closest('[data-file-drop-target]')).not.toBeNull()
    vi.mocked(service.pickImportSources).mockResolvedValue([])
    await userEvent.click(screen.getByRole('button', { name: 'Choose Files…' }))
    expect(service.pickImportSources).toHaveBeenCalledOnce()
    expect(screen.getByRole('button', { name: 'Read Pasted Text' })).toBeDisabled()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('reads pasted text, shows a refused source and the preview of an accepted one', async () => {
    render(<ImportDialog />)
    act(() => useImportStore.getState().open())
    vi.mocked(service.detectImport).mockRejectedValueOnce(new AppError('IMP-00415', 'pasted text is not a Postman, Bruno or cURL export'))
    await userEvent.type(screen.getByRole('textbox', { name: 'Or paste Postman JSON or a cURL command' }), 'hello')
    await userEvent.click(screen.getByRole('button', { name: 'Read Pasted Text' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('pasted text is not a Postman, Bruno or cURL export. Check the source and the target workspace, then try again.')
    vi.mocked(service.detectImport).mockResolvedValueOnce(preview)
    await userEvent.click(screen.getByRole('button', { name: 'Read Pasted Text' }))
    expect(await screen.findByText('Postman collection v2.1')).toBeInTheDocument()
    expect(service.detectImport).toHaveBeenLastCalledWith({ text: 'hello' })
    expect(screen.getByText('Acme API')).toBeInTheDocument()
    expect(screen.getByText('1 collection · 3 folders · 14 requests · 3 examples · 3 variables · 7 scripts · 17 to review')).toBeInTheDocument()
    expect(screen.getByRole('combobox', { name: 'Workspace for Acme API' })).toHaveValue('A')
    expect(screen.queryByRole('option', { name: 'Missing' })).not.toBeInTheDocument()
    expect(await screen.findByRole('option', { name: 'Into Main / Target' })).toBeInTheDocument()
  })

  it('imports into the chosen location, asks about clashes and shows the summary with links', async () => {
    const openItem = vi.spyOn(useImportStore.getState(), 'openItem').mockResolvedValue()
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    act(() => void useImportStore.getState().loadLocations('A'))
    render(<ImportDialog />)
    await userEvent.selectOptions(await screen.findByRole('combobox', { name: 'Location for Acme API' }), 'F')
    vi.mocked(service.findImportClashes).mockResolvedValue([
      { itemKey: '0', kind: 'request', existingId: 'X', existingName: 'Health', name: 'Health' },
      { itemKey: '0', kind: 'variable', existingId: 'C', existingName: 'base_url', name: 'base_url' },
    ])
    vi.mocked(service.runImport).mockReturnValue({ result: Promise.resolve(result), cancel: vi.fn() })
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(await screen.findByText(/already exists/)).toHaveTextContent('A request named “Health” already exists')
    expect(screen.getByText('Health')).toHaveAttribute('translate', 'no')
    await userEvent.click(screen.getByRole('checkbox', { name: 'Apply to all 2 clashes' }))
    await userEvent.click(screen.getByRole('button', { name: 'Keep Both' }))
    expect(service.runImport).toHaveBeenCalledWith({
      sourceId: 'S',
      applyToAll: 'keep_both',
      items: [{ key: '0', include: true, targetLocalId: 'A', targetParentId: 'F', resolution: null }],
    })
    expect(await screen.findByText(': 1 collection · 3 folders · 14 requests · 3 examples · 3 variables · 7 scripts, 1 replaced, 2 skipped')).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Not Imported' })).toHaveTextContent('CRM API: permission denied')
    expect(screen.getByRole('region', { name: 'Needs Attention (2)' })).toHaveTextContent('Auth type hawk is not supported yet')
    expect(screen.getByRole('region', { name: 'pm APIs Outside Xenon’s Script Subset' })).toHaveTextContent('pm.cookies.jar')
    expect(screen.getAllByRole('button', { name: /^Open / })).toHaveLength(2)
    await userEvent.click(screen.getByRole('button', { name: 'Open Acme API/Users/Propfind' }))
    expect(openItem).toHaveBeenCalledWith('A', 'R1')
    await userEvent.click(screen.getByRole('button', { name: 'Open Acme API/Raw JSON' }))
    expect(openItem).toHaveBeenLastCalledWith('A', 'R2')
    openItem.mockRestore()
    await userEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('excludes items, skips a clash per item and shows progress with cancel', async () => {
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    render(<ImportDialog />)
    const include = screen.getByRole('checkbox', { name: 'Import Acme API' })
    await userEvent.click(include)
    expect(screen.getByRole('button', { name: 'Import' })).toBeDisabled()
    await userEvent.click(include)
    vi.mocked(service.findImportClashes).mockResolvedValue([{ itemKey: '0', kind: 'collection', existingId: 'X', existingName: 'Acme API', name: 'Acme API' }])
    const cancel = vi.fn()
    vi.mocked(service.runImport).mockReturnValue({ result: new Promise(() => undefined), cancel })
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Skip' }))
    expect(vi.mocked(service.runImport).mock.calls[0]![0].items[0]!.resolution).toBe('skip')
    expect(await screen.findByText('Preparing the import…')).toBeInTheDocument()
    act(() => useImportStore.getState().applyProgress({ sourceId: 'S', done: 0, total: 1, currentItem: 'Acme API' }))
    expect(screen.getByRole('status')).toHaveTextContent('Importing Acme API… (0 of 1)')
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(cancel).toHaveBeenCalled()
  })

  it('names a pasted cURL request and imports it into a collection or folder', async () => {
    vi.mocked(service.detectImport).mockResolvedValue({
      ...preview,
      kind: 'curl',
      version: '',
      items: [{ key: '0', kind: 'request', name: 'POST /users', counts: { ...counts, collections: 0, folders: 0, requests: 1, examples: 0, variables: 0, scripts: 0 }, warningsCount: 0 }],
    })
    render(<ImportDialog />)
    act(() => useImportStore.getState().open())
    await userEvent.type(screen.getByRole('textbox', { name: 'Or paste Postman JSON or a cURL command' }), 'curl -X POST https://api.test/users')
    await userEvent.click(screen.getByRole('button', { name: 'Read Pasted Text' }))
    expect(await screen.findByText('cURL command')).toBeInTheDocument()
    const location = screen.getByRole('combobox', { name: 'Location for POST /users' })
    expect(screen.getByRole('option', { name: 'Choose a collection or folder' })).toBeDisabled()
    expect(screen.queryByRole('option', { name: 'New collection' })).not.toBeInTheDocument()
    await vi.waitFor(() => expect(location).toHaveValue('C'))
    const name = screen.getByRole('textbox', { name: 'Name for POST /users' })
    expect(name).toHaveValue('POST /users')
    await userEvent.clear(name)
    expect(screen.getByRole('button', { name: 'Import' })).toBeEnabled()
    await userEvent.type(name, 'Invite user')
    await userEvent.selectOptions(location, 'F')
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    vi.mocked(service.runImport).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(service.runImport).toHaveBeenCalledWith({
      sourceId: 'S',
      applyToAll: null,
      items: [{ key: '0', include: true, targetLocalId: 'A', targetParentId: 'F', name: 'Invite user', resolution: null }],
    })
  })

  it('keeps Import disabled while a request has no collection or folder to go into', async () => {
    vi.mocked(tree.getTree).mockResolvedValue([])
    const request = { key: '0', kind: 'request', name: 'GET /x', counts, warningsCount: 0 }
    useImportStore.setState({ step: 'preview', preview: { ...preview, kind: 'curl', items: [request] }, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    act(() => void useImportStore.getState().loadLocations('A'))
    render(<ImportDialog />)
    expect(screen.getByRole('button', { name: 'Import' })).toBeDisabled()
    await userEvent.click(screen.getByRole('checkbox', { name: 'Import GET /x' }))
    expect(screen.getByRole('textbox', { name: 'Name for GET /x' })).toBeDisabled()
  })

  it('says when a detected source has nothing to import yet and goes back', async () => {
    useImportStore.setState({ step: 'preview', preview: { ...preview, kind: 'postman_environment', version: '', items: [] }, targets: {} })
    render(<ImportDialog />)
    expect(screen.getByText('Postman environment')).toBeInTheDocument()
    expect(screen.getByText('There is nothing to import in this source.')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Back' }))
    expect(screen.getByRole('button', { name: 'Choose Files…' })).toBeInTheDocument()
  })

  it('asks for the passphrase of encrypted secrets, retries a wrong one and can import without them', async () => {
    useImportStore.setState({ step: 'preview', preview: { ...preview, kind: 'xenon_zip', encryptedSecrets: true }, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    render(<ImportDialog />)
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    vi.mocked(service.runImport).mockImplementationOnce(() => ({ result: Promise.reject(new AppError('IMP-50401', 'the export has encrypted secrets')), cancel: vi.fn() }))
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(await screen.findByText(/This export carries secret values encrypted with a passphrase/)).toBeInTheDocument()
    const submit = screen.getByRole('button', { name: 'Import' })
    expect(submit).toBeDisabled()
    vi.mocked(service.runImport).mockImplementationOnce(() => ({ result: Promise.reject(new AppError('IMP-08401', 'wrong passphrase')), cancel: vi.fn() }))
    await userEvent.type(screen.getByLabelText('Passphrase'), 'wrong horse{Enter}')
    expect(vi.mocked(service.runImport).mock.calls[1]![0]).toMatchObject({ passphrase: 'wrong horse' })
    expect(await screen.findByRole('alert')).toHaveTextContent('That passphrase does not open the export’s secrets.')
    expect(screen.getByLabelText('Passphrase')).toHaveValue('')
    expect(screen.getByLabelText('Passphrase')).toHaveAttribute('aria-invalid', 'true')
    vi.mocked(service.runImport).mockReturnValueOnce({ result: Promise.resolve(result), cancel: vi.fn() })
    await userEvent.type(screen.getByLabelText('Passphrase'), 'correct horse')
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(vi.mocked(service.runImport).mock.calls[2]![0]).toMatchObject({ passphrase: 'correct horse' })
    expect(vi.mocked(service.runImport).mock.calls[2]![0]).not.toHaveProperty('skipSecrets')
    expect(await screen.findByRole('button', { name: 'Done' })).toBeInTheDocument()
    expect(useImportStore.getState().passphrase).toBe('')
    useImportStore.setState({ step: 'passphrase', passphraseRequest: 'missing' })
    vi.mocked(service.runImport).mockReturnValueOnce({ result: Promise.resolve(result), cancel: vi.fn() })
    await userEvent.click(await screen.findByRole('button', { name: 'Import Without Secrets' }))
    expect(vi.mocked(service.runImport).mock.calls[3]![0]).toMatchObject({ skipSecrets: true })
    expect(vi.mocked(service.runImport).mock.calls[3]![0]).not.toHaveProperty('passphrase')
  })

  it('imports a workspace export as a new workspace in a chosen folder and reloads the workspace list', async () => {
    const workspaceItem = { key: '0', kind: 'workspace', name: 'Acme', counts, warningsCount: 0 }
    vi.mocked(workspaces.listWorkspaces).mockResolvedValue({
      workspaces: [{ localId: 'A', name: 'Acme', path: '/w/acme', missing: false }, { localId: 'N', name: 'Acme', path: '/w/restored', missing: false }],
      activeLocalId: 'A',
    })
    useImportStore.setState({ step: 'preview', preview: { ...preview, kind: 'xenon_workspace_zip', items: [workspaceItem] }, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    render(<ImportDialog />)
    expect(screen.queryByRole('combobox', { name: 'Workspace for Acme' })).not.toBeInTheDocument()
    expect(screen.getByText('New workspace')).toBeInTheDocument()
    expect(screen.getByText('Choose an empty folder')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Import' })).toBeDisabled()
    vi.mocked(dialogs.pickFolder).mockResolvedValueOnce(null)
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder for Acme' }))
    expect(screen.getByRole('button', { name: 'Import' })).toBeDisabled()
    vi.mocked(dialogs.pickFolder).mockResolvedValueOnce('/w/restored')
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder for Acme' }))
    expect(await screen.findByText('/w/restored')).toHaveAttribute('translate', 'no')
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    vi.mocked(service.runImport).mockReturnValue({
      result: Promise.resolve({ ...result, items: [{ key: '0', status: 'imported', localId: 'N', id: 'W', name: 'Acme' }], workspaces: [{ localId: 'N', created: counts, replaced: 0, skipped: 0 }], warnings: [], pmUsages: [], failed: [] }),
      cancel: vi.fn(),
    })
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(service.runImport).toHaveBeenCalledWith({
      sourceId: 'S',
      applyToAll: null,
      items: [{ key: '0', include: true, targetLocalId: '', targetParentId: '', targetFolder: '/w/restored', resolution: null }],
    })
    await vi.waitFor(() => expect(workspaces.listWorkspaces).toHaveBeenCalled())
    expect(await screen.findByText(': 1 collection · 3 folders · 14 requests · 3 examples · 3 variables · 7 scripts')).toBeInTheDocument()
  })

  it('reports a folder picker that failed', async () => {
    const workspaceItem = { key: '0', kind: 'workspace', name: 'Acme', counts, warningsCount: 0 }
    useImportStore.setState({ step: 'preview', preview: { ...preview, kind: 'xenon_workspace_zip', items: [workspaceItem] }, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    render(<ImportDialog />)
    vi.mocked(dialogs.pickFolder).mockRejectedValueOnce(new AppError('DLG-00500', 'no dialog'))
    await userEvent.click(screen.getByRole('button', { name: 'Choose Folder for Acme' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('no dialog')
  })
})
