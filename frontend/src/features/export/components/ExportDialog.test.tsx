import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError, CANCELLED_CODE } from '@/platform/errors'
import type { ExportOptions, ExportSummary } from '@/platform/mappers/export'

vi.mock('@/platform/services/export', () => ({
  getExportOptions: vi.fn(),
  pickExportTarget: vi.fn(),
  runExport: vi.fn(),
  showExportInFolder: vi.fn(),
}))

const service = await import('@/platform/services/export')
const { useExportStore } = await import('@/features/export/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { ExportDialog } = await import('./ExportDialog')

const options: ExportOptions = {
  suggestedFileName: 'shop-api',
  hasSecrets: true,
  hasCurrentValues: true,
  formats: [
    { format: 'xenon_yaml', suggestedFileName: 'shop-api.zip', shapes: ['zip', 'bundle'], includeSecretsAllowed: true },
    { format: 'postman', suggestedFileName: 'shop-api.postman_collection.json', shapes: ['file'], includeSecretsAllowed: false },
  ],
}

const summary: ExportSummary = {
  path: '/out/shop-api.postman_collection.json',
  secrets: 'names_only',
  counts: { collections: 1, folders: 2, requests: 9, examples: 1, environments: 0, variables: 4 },
  dropped: [{ itemPath: 'Shop API/Ping', field: 'inheritScripts', reason: 'Postman always runs the parent scripts' }],
  warnings: [{ itemPath: 'Shop API/Login', code: 'literal_credential', message: 'The basic auth password is typed literally' }],
}

function pendingExport() {
  let resolve: (value: ExportSummary) => void = () => undefined
  let reject: (error: unknown) => void = () => undefined
  const result = new Promise<ExportSummary>((done, fail) => {
    resolve = done
    reject = fail
  })
  const cancel = vi.fn(() => reject(new AppError(CANCELLED_CODE, 'You cancelled the request')))
  return { pending: { summary: result, cancel }, resolve, reject, cancel }
}

async function openShop(rendered = false) {
  vi.mocked(service.getExportOptions).mockResolvedValue(options)
  if (!rendered) {
    render(<ExportDialog />)
  }
  await act(() => useExportStore.getState().openItem('collection', 'C', 'Shop API'))
}

describe('ExportDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useExportStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A', workspaces: [{ localId: 'A', name: 'Acme', path: '/w/acme', missing: false }] })
  })

  afterEach(() => {
    useExportStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null, workspaces: [] })
  })

  it('is hidden until an item is exported and offers the formats and shapes of that item', async () => {
    render(<ExportDialog />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    await openShop(true)
    expect(service.getExportOptions).toHaveBeenCalledWith({ localId: 'A', kind: 'collection', itemId: 'C', name: 'Shop API' })
    expect(screen.getByRole('dialog', { name: 'Export “Shop API”' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Xenon YAML' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Zip of the folder' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Single YAML file' })).not.toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'Include secrets (encrypted with a passphrase)' })).not.toBeChecked()
    expect(screen.getByRole('checkbox', { name: 'Export current values as initial values' })).not.toBeChecked()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('hides Include secrets for Postman with an explanation and only offers its file shape', async () => {
    await openShop()
    await userEvent.click(screen.getByRole('checkbox', { name: 'Include secrets (encrypted with a passphrase)' }))
    await userEvent.click(screen.getByRole('radio', { name: 'Postman' }))
    expect(screen.queryByRole('checkbox', { name: /Include secrets/ })).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Passphrase')).not.toBeInTheDocument()
    expect(screen.queryByRole('radio', { name: 'Single YAML file' })).not.toBeInTheDocument()
    expect(screen.getByText(/Postman exports never carry secret values/)).toBeInTheDocument()
    expect(useExportStore.getState().choices).toMatchObject({ format: 'postman', shape: 'file', includeSecrets: false })
  })

  it('asks for the passphrase twice, at least 8 characters, before an export with secrets', async () => {
    await openShop()
    await userEvent.click(screen.getByRole('checkbox', { name: 'Include secrets (encrypted with a passphrase)' }))
    const exportButton = screen.getByRole('button', { name: 'Export…' })
    expect(exportButton).toBeDisabled()
    await userEvent.type(screen.getByLabelText('Passphrase'), 'short')
    expect(screen.getByRole('alert')).toHaveTextContent('Use at least 8 characters.')
    expect(screen.getByLabelText('Passphrase')).toHaveAttribute('aria-invalid', 'true')
    await userEvent.type(screen.getByLabelText('Passphrase'), ' horse')
    await userEvent.type(screen.getByLabelText('Confirm Passphrase'), 'short horsE')
    expect(screen.getByRole('alert')).toHaveTextContent('The passphrases do not match.')
    expect(exportButton).toBeDisabled()
    await userEvent.clear(screen.getByLabelText('Confirm Passphrase'))
    await userEvent.type(screen.getByLabelText('Confirm Passphrase'), 'short horse')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(exportButton).toBeEnabled()
    const { pending, resolve } = pendingExport()
    vi.mocked(service.pickExportTarget).mockResolvedValue('/out/shop-api.yaml')
    vi.mocked(service.runExport).mockReturnValue(pending)
    await userEvent.click(screen.getByRole('radio', { name: 'Single YAML file' }))
    await userEvent.click(exportButton)
    expect(service.pickExportTarget).toHaveBeenCalledWith('shop-api.zip', 'xenon_yaml', 'bundle')
    expect(service.runExport).toHaveBeenCalledWith({
      localId: 'A',
      kind: 'collection',
      itemId: 'C',
      format: 'xenon_yaml',
      shape: 'bundle',
      includeSecrets: true,
      passphrase: 'short horse',
      passphraseConfirmation: 'short horse',
      currentValuesAsInitial: false,
      targetPath: '/out/shop-api.yaml',
    })
    await act(async () => resolve({ ...summary, path: '/out/shop-api.yaml', secrets: 'included', dropped: [], warnings: [] }))
    expect(await screen.findByText('Secret values are included, encrypted with your passphrase.')).toBeInTheDocument()
    expect(useExportStore.getState().choices).toMatchObject({ passphrase: '', passphraseConfirmation: '' })
  })

  it('exports to the picked file and shows the summary with dropped fields, warnings and Show in Folder', async () => {
    await openShop()
    await userEvent.click(screen.getByRole('radio', { name: 'Postman' }))
    await userEvent.click(screen.getByRole('checkbox', { name: 'Export current values as initial values' }))
    vi.mocked(service.pickExportTarget).mockResolvedValueOnce(null)
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    expect(service.runExport).not.toHaveBeenCalled()
    expect(screen.getByRole('radio', { name: 'Postman' })).toBeChecked()
    const { pending, resolve } = pendingExport()
    vi.mocked(service.pickExportTarget).mockResolvedValueOnce('/out/shop-api.postman_collection.json')
    vi.mocked(service.runExport).mockReturnValue(pending)
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    expect(service.runExport).toHaveBeenCalledWith(expect.objectContaining({ format: 'postman', shape: 'file', currentValuesAsInitial: true, includeSecrets: false, passphrase: '' }))
    expect(screen.getByRole('status')).toHaveTextContent('Writing the export…')
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
    await act(async () => resolve(summary))
    expect(await screen.findByText('Exported 1 collection · 2 folders · 9 requests · 1 example · 4 variables.')).toBeInTheDocument()
    expect(screen.getByText('/out/shop-api.postman_collection.json')).toHaveAttribute('translate', 'no')
    expect(screen.getByText('Secret variables are exported by name with empty values.')).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Dropped or Changed (1)' })).toHaveTextContent('inheritScripts: Postman always runs the parent scripts')
    expect(screen.getByRole('region', { name: 'Needs Attention (1)' })).toHaveTextContent('The basic auth password is typed literally')
    vi.mocked(service.showExportInFolder).mockRejectedValueOnce(new AppError('EXP-09500', 'no file manager'))
    await userEvent.click(screen.getByRole('button', { name: 'Show in Folder' }))
    expect(service.showExportInFolder).toHaveBeenCalledWith('/out/shop-api.postman_collection.json')
    expect(await screen.findByRole('alert')).toHaveTextContent('no file manager')
    await userEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows progress for more than 200 files and goes back to the options when cancelled', async () => {
    await openShop()
    const { pending, cancel } = pendingExport()
    vi.mocked(service.pickExportTarget).mockResolvedValue('/out/shop-api.zip')
    vi.mocked(service.runExport).mockReturnValue(pending)
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    act(() => useExportStore.getState().applyProgress({ done: 50, total: 120 }))
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
    act(() => useExportStore.getState().applyProgress({ done: 150, total: 1007 }))
    expect(screen.getByRole('progressbar', { name: 'Export progress' })).toHaveAttribute('value', '150')
    expect(screen.getByRole('status')).toHaveTextContent('Writing the export… (150 of 1,007 files)')
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(cancel).toHaveBeenCalledOnce()
    expect(await screen.findByText('The export was cancelled; no file was written.')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Xenon YAML' })).toBeChecked()
  })

  it('reports a failed export and a cancellation answered by the backend', async () => {
    await openShop()
    vi.mocked(service.pickExportTarget).mockResolvedValue('/out/shop-api.zip')
    vi.mocked(service.runExport).mockImplementationOnce(() => ({ summary: Promise.reject(new AppError('EXP-02500', 'file operation failed')), cancel: vi.fn() }))
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('file operation failed. Check the save location and the options, then try again.')
    vi.mocked(service.runExport).mockImplementationOnce(() => ({ summary: Promise.reject(new AppError('EXP-05499', 'the export was cancelled')), cancel: vi.fn() }))
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    expect(await screen.findByText('The export was cancelled; no file was written.')).toBeInTheDocument()
    vi.mocked(service.pickExportTarget).mockRejectedValueOnce(new AppError('DLG-00500', 'no dialog'))
    await userEvent.click(screen.getByRole('button', { name: 'Export…' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('no dialog')
  })

  it('reports options that could not be loaded', async () => {
    vi.mocked(service.getExportOptions).mockRejectedValue(new AppError('TRE-03404', 'the item does not exist'))
    render(<ExportDialog />)
    await act(() => useExportStore.getState().openItem('request', 'R', 'Gone'))
    expect(screen.getByRole('alert')).toHaveTextContent('the item does not exist')
    expect(screen.getByRole('button', { name: 'Export…' })).toBeDisabled()
  })
})
