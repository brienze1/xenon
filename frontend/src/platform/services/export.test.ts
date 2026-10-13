import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exportservice', () => ({
  Options: vi.fn(),
  PickTarget: vi.fn(),
  Export: vi.fn(),
  ShowInFolder: vi.fn(),
}))

const ExportService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exportservice')
const { getExportOptions, pickExportTarget, runExport, showExportInFolder } = await import('./export')

const request = {
  localId: 'A',
  kind: 'collection' as const,
  itemId: 'C',
  format: 'xenon_yaml' as const,
  shape: 'bundle' as const,
  includeSecrets: false,
  passphrase: 'typed but unused',
  passphraseConfirmation: 'typed but unused',
  currentValuesAsInitial: true,
  targetPath: '/out/shop.yaml',
}

describe('export service', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('maps the options of an item, dropping unknown formats and shapes', async () => {
    vi.mocked(ExportService.Options).mockResolvedValue({
      suggested_file_name: 'shop-api',
      has_secrets: true,
      has_current_values: false,
      formats: [
        { format: 'xenon_yaml', suggested_file_name: 'shop-api.zip', shapes: ['zip', 'bundle', 'cube'], include_secrets_allowed: true },
        { format: 'postman', suggested_file_name: 'shop-api.postman_collection.json', shapes: null, include_secrets_allowed: false },
        { format: 'insomnia', suggested_file_name: 'x', shapes: ['file'], include_secrets_allowed: false },
      ],
    })
    const options = await getExportOptions({ localId: 'A', kind: 'collection', itemId: 'C' })
    expect(ExportService.Options).toHaveBeenCalledWith({ local_id: 'A', item_kind: 'collection', item_id: 'C' })
    expect(options).toEqual({
      suggestedFileName: 'shop-api',
      hasSecrets: true,
      hasCurrentValues: false,
      formats: [
        { format: 'xenon_yaml', suggestedFileName: 'shop-api.zip', shapes: ['zip', 'bundle'], includeSecretsAllowed: true },
        { format: 'postman', suggestedFileName: 'shop-api.postman_collection.json', shapes: [], includeSecretsAllowed: false },
      ],
    })
  })

  it('returns the picked path, or null when the save dialog was cancelled', async () => {
    vi.mocked(ExportService.PickTarget).mockResolvedValueOnce({ path: '/out/shop.yaml' }).mockResolvedValueOnce({ path: '' })
    expect(await pickExportTarget('shop-api.zip', 'xenon_yaml', 'bundle')).toBe('/out/shop.yaml')
    expect(ExportService.PickTarget).toHaveBeenCalledWith({ suggested_file_name: 'shop-api.zip', format: 'xenon_yaml', shape: 'bundle' })
    expect(await pickExportTarget('shop-api.zip', 'xenon_yaml', 'zip')).toBeNull()
  })

  it('exports without sending a passphrase unless secrets are included, and maps the summary', async () => {
    const cancel = vi.fn()
    const pending = Object.assign(
      Promise.resolve({
        path: '/out/shop.yaml',
        secrets: 'omitted',
        counts: { collections: 1, folders: 2, requests: 9, examples: 1, environments: 0, variables: 4 },
        dropped: [{ item_path: 'Shop API/Ping', field: 'inheritScripts', reason: 'Postman always runs parent scripts' }],
        warnings: null,
      }),
      { cancel },
    )
    vi.mocked(ExportService.Export).mockReturnValue(pending as never)
    const run = runExport(request)
    expect(ExportService.Export).toHaveBeenCalledWith({
      local_id: 'A',
      item_kind: 'collection',
      item_id: 'C',
      format: 'xenon_yaml',
      shape: 'bundle',
      include_secrets: false,
      passphrase: '',
      passphrase_confirmation: '',
      current_values_as_initial: true,
      target_path: '/out/shop.yaml',
    })
    expect(await run.summary).toEqual({
      path: '/out/shop.yaml',
      secrets: 'omitted',
      counts: { collections: 1, folders: 2, requests: 9, examples: 1, environments: 0, variables: 4 },
      dropped: [{ itemPath: 'Shop API/Ping', field: 'inheritScripts', reason: 'Postman always runs parent scripts' }],
      warnings: [],
    })
    run.cancel()
    expect(cancel).toHaveBeenCalledOnce()
  })

  it('sends the passphrase when secrets are included', () => {
    vi.mocked(ExportService.Export).mockReturnValue(Object.assign(new Promise(() => undefined), { cancel: vi.fn() }) as never)
    runExport({ ...request, includeSecrets: true, passphrase: 'correct horse', passphraseConfirmation: 'correct horse' })
    expect(ExportService.Export).toHaveBeenCalledWith(expect.objectContaining({ include_secrets: true, passphrase: 'correct horse', passphrase_confirmation: 'correct horse' }))
  })

  it('maps backend errors and opens the file manager', async () => {
    vi.mocked(ExportService.ShowInFolder).mockRejectedValueOnce(Object.assign(new Error('x'), { cause: { error: { code: 'EXP-09500', description: 'no file manager' } } }))
    await expect(showExportInFolder('/out/shop.yaml')).rejects.toMatchObject({ code: 'EXP-09500', description: 'no file manager' })
    vi.mocked(ExportService.ShowInFolder).mockResolvedValueOnce()
    await showExportInFolder('/out/shop.yaml')
    expect(ExportService.ShowInFolder).toHaveBeenLastCalledWith({ path: '/out/shop.yaml' })
  })
})
