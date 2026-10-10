import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/importservice', () => ({
  Detect: vi.fn(),
  PickSources: vi.fn(),
  Clashes: vi.fn(),
  Import: vi.fn(),
  Discard: vi.fn(),
}))

const ImportService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/importservice')
const { detectImport, discardImport, findImportClashes, pickImportSources, runImport } = await import('./import')

const counts = { collections: 1, folders: 3, requests: 14, examples: 3, variables: 3, scripts: 7, environments: 0 }
const warningDto = {
  code: 'unsupported_auth',
  severity: 'warning',
  message: 'Auth type hawk is not supported yet',
  item_key: '0',
  item_id: 'R1',
  local_id: 'A',
  path: 'Acme API/Users/Propfind',
  location: '',
}
const usageDto = { api: 'pm.cookies.jar', occurrences: [{ item_key: '0', item_id: 'R2', local_id: 'A', path: 'Acme API/Raw', phase: 'preRequest', line: 2 }] }
const plan = {
  sourceId: 'S',
  applyToAll: null,
  items: [{ key: '0', include: true, targetLocalId: 'A', targetParentId: '', resolution: null }],
}

describe('import service', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('detects pasted text or paths and maps the preview', async () => {
    vi.mocked(ImportService.Detect).mockResolvedValue({
      source_id: 'S',
      kind: 'postman_collection',
      version: '2.1',
      items: [{ key: '0', kind: 'collection', name: 'Acme API', counts, warnings_count: 17 }],
      warnings: [warningDto, { ...warningDto, severity: 'info' }],
      pm_usages: [usageDto, { api: 'pm.require', occurrences: null }],
    } as never)
    const preview = await detectImport({ paths: ['/a.json'] })
    expect(ImportService.Detect).toHaveBeenCalledWith({ paths: ['/a.json'], text: '' })
    expect(preview).toEqual({
      sourceId: 'S',
      kind: 'postman_collection',
      version: '2.1',
      items: [{ key: '0', kind: 'collection', name: 'Acme API', counts, warningsCount: 17 }],
      warnings: [
        { code: 'unsupported_auth', severity: 'warning', message: 'Auth type hawk is not supported yet', itemKey: '0', itemId: 'R1', localId: 'A', path: 'Acme API/Users/Propfind', location: '' },
        { code: 'unsupported_auth', severity: 'info', message: 'Auth type hawk is not supported yet', itemKey: '0', itemId: 'R1', localId: 'A', path: 'Acme API/Users/Propfind', location: '' },
      ],
      pmUsages: [
        { api: 'pm.cookies.jar', occurrences: [{ itemId: 'R2', localId: 'A', path: 'Acme API/Raw', phase: 'preRequest', line: 2 }] },
        { api: 'pm.require', occurrences: [] },
      ],
    })
    vi.mocked(ImportService.Detect).mockResolvedValue({ source_id: 'T', kind: 'curl', version: '', items: null, warnings: null, pm_usages: null } as never)
    await expect(detectImport({ text: 'curl https://x' })).resolves.toMatchObject({ items: [], warnings: [], pmUsages: [] })
    expect(ImportService.Detect).toHaveBeenLastCalledWith({ paths: [], text: 'curl https://x' })
  })

  it('reports refused sources as AppError', async () => {
    vi.mocked(ImportService.Detect).mockRejectedValue({ cause: { error: { code: 'IMP-00415', description: 'nope' } } })
    await expect(detectImport({ paths: ['/x'] })).rejects.toMatchObject({ code: 'IMP-00415', description: 'nope' })
    vi.mocked(ImportService.Detect).mockResolvedValue(null as never)
    await expect(detectImport({ paths: ['/x'] })).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('picks sources', async () => {
    vi.mocked(ImportService.PickSources).mockResolvedValue({ paths: ['/a', '/b'] } as never)
    await expect(pickImportSources()).resolves.toEqual(['/a', '/b'])
    vi.mocked(ImportService.PickSources).mockResolvedValue(null as never)
    await expect(pickImportSources()).resolves.toEqual([])
  })

  it('finds clashes for a plan', async () => {
    vi.mocked(ImportService.Clashes).mockResolvedValue({
      clashes: [{ item_key: '0', kind: 'collection', existing_id: 'C', existing_name: 'Acme API', name: 'Acme API' }],
    } as never)
    await expect(findImportClashes({ ...plan, applyToAll: 'keep_both', items: [{ ...plan.items[0]!, resolution: 'skip' }] })).resolves.toEqual([
      { itemKey: '0', kind: 'collection', existingId: 'C', existingName: 'Acme API', name: 'Acme API' },
    ])
    expect(ImportService.Clashes).toHaveBeenCalledWith({
      source_id: 'S',
      apply_to_all: 'keep_both',
      items: [{ key: '0', include: true, target_local_id: 'A', target_parent_id: '', resolution: 'skip' }],
    })
    vi.mocked(ImportService.Clashes).mockResolvedValue({ clashes: null } as never)
    await expect(findImportClashes(plan)).resolves.toEqual([])
  })

  it('runs a cancellable import and maps the result', async () => {
    const cancel = vi.fn()
    const pending = Promise.resolve({
      items: [
        { key: '0', status: 'imported', local_id: 'A', id: 'C', name: 'Acme API' },
        { key: '1', status: 'weird', local_id: 'A', id: '', name: 'Other' },
      ],
      workspaces: [{ local_id: 'A', created: counts, replaced: 1, skipped: 0 }],
      warnings: null,
      pm_usages: null,
      failed: [{ item_key: '1', name: 'Other', code: 'IMP-29500', message: 'permission denied' }],
    })
    vi.mocked(ImportService.Import).mockReturnValue(Object.assign(pending, { cancel }) as never)
    const run = runImport(plan)
    expect(ImportService.Import).toHaveBeenCalledWith({
      source_id: 'S',
      apply_to_all: '',
      items: [{ key: '0', include: true, target_local_id: 'A', target_parent_id: '', resolution: '' }],
    })
    await expect(run.result).resolves.toEqual({
      items: [
        { key: '0', status: 'imported', localId: 'A', id: 'C', name: 'Acme API' },
        { key: '1', status: 'failed', localId: 'A', id: '', name: 'Other' },
      ],
      workspaces: [{ localId: 'A', created: counts, replaced: 1, skipped: 0 }],
      warnings: [],
      pmUsages: [],
      failed: [{ itemKey: '1', name: 'Other', code: 'IMP-29500', message: 'permission denied' }],
    })
    run.cancel()
    expect(cancel).toHaveBeenCalledOnce()
  })

  it('discards a source', async () => {
    vi.mocked(ImportService.Discard).mockResolvedValue(undefined as never)
    await discardImport('S')
    expect(ImportService.Discard).toHaveBeenCalledWith({ source_id: 'S' })
  })
})
