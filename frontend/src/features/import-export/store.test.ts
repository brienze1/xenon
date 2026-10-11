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

const service = await import('@/platform/services/import')
const tree = await import('@/platform/services/tree')
const { useImportStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')

const counts = { collections: 1, folders: 0, requests: 1, examples: 0, variables: 0, scripts: 0, environments: 0 }
const preview: ImportPreview = {
  sourceId: 'S',
  kind: 'postman_collection',
  version: '2.1',
  items: [{ key: '0', kind: 'collection', name: 'Acme API', counts, warningsCount: 0 }],
  warnings: [],
  pmUsages: [],
  encryptedSecrets: false,
}
const result: ImportResult = {
  items: [{ key: '0', status: 'imported', localId: 'A', id: 'C', name: 'Acme API' }],
  workspaces: [{ localId: 'A', created: counts, replaced: 0, skipped: 0 }],
  warnings: [],
  pmUsages: [],
  failed: [],
}
const clash = { itemKey: '0', kind: 'collection', existingId: 'X', existingName: 'Acme API', name: 'Acme API' }

function state() {
  return useImportStore.getState()
}

function resolvedImport(value: ImportResult = result) {
  return { result: Promise.resolve(value), cancel: vi.fn() }
}

describe('import store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    state().reset()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    vi.mocked(service.discardImport).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([])
  })

  afterEach(() => {
    state().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('opens on the source step once and closes back to closed', () => {
    state().open()
    expect(state().step).toBe('choose')
    useImportStore.setState({ step: 'preview' })
    state().open()
    expect(state().step).toBe('preview')
    state().close()
    expect(state().step).toBe('closed')
  })

  it('detects a source, targets the active workspace and loads its locations', async () => {
    vi.mocked(service.detectImport).mockResolvedValue(preview)
    vi.mocked(tree.getTree).mockResolvedValue([
      { id: 'C', kind: 'collection', name: 'Main', method: null, url: null, error: null, warning: null, children: [] },
    ])
    state().open()
    await state().detect({ paths: ['/a.json'] })
    expect(state().step).toBe('preview')
    expect(state().targets).toEqual({ '0': { include: true, localId: 'A', parentId: '' } })
    await vi.waitFor(() => expect(state().locations.A).toEqual([{ id: 'C', label: 'Main' }]))
  })

  it('shows refused sources on the source step and discards replaced previews', async () => {
    useImportStore.setState({ step: 'preview', preview })
    vi.mocked(service.detectImport).mockRejectedValue(new AppError('IMP-00415', 'not an export'))
    await state().detect({ text: 'hello' })
    expect(service.discardImport).toHaveBeenCalledWith('S')
    expect(state().step).toBe('choose')
    expect(state().error?.code).toBe('IMP-00415')
  })

  it('picks sources and detects them, ignoring an empty pick and reporting picker failures', async () => {
    vi.mocked(service.pickImportSources).mockResolvedValueOnce([])
    await state().pickSources()
    expect(service.detectImport).not.toHaveBeenCalled()
    vi.mocked(service.pickImportSources).mockResolvedValueOnce(['/a.json', '/b.json'])
    vi.mocked(service.detectImport).mockResolvedValue(preview)
    await state().pickSources()
    expect(service.detectImport).toHaveBeenCalledWith({ paths: ['/a.json', '/b.json'] })
    vi.mocked(service.pickImportSources).mockRejectedValueOnce(new AppError('DLG-03500', 'no picker'))
    await state().pickSources()
    expect(state().error?.code).toBe('DLG-03500')
  })

  it('changes a target workspace, clearing its location and loading the new locations', async () => {
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: 'F' } } })
    state().setTarget('0', { parentId: 'G' })
    expect(state().targets['0']).toEqual({ include: true, localId: 'A', parentId: 'G' })
    state().setTarget('0', { localId: 'B' })
    expect(state().targets['0']).toEqual({ include: true, localId: 'B', parentId: '' })
    await vi.waitFor(() => expect(tree.getTree).toHaveBeenCalledWith('B'))
    state().setTarget('missing', { include: false })
    expect(state().targets.missing).toBeUndefined()
    vi.mocked(tree.getTree).mockRejectedValue(new AppError('WKS-07404', 'gone'))
    await state().loadLocations('B')
    expect(state().error?.code).toBe('WKS-07404')
  })

  it('imports straight away without clashes and reloads the active sidebar', async () => {
    const load = vi.spyOn(useSidebarStore.getState(), 'load').mockResolvedValue()
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    vi.mocked(service.runImport).mockReturnValue(resolvedImport())
    await state().start()
    expect(service.runImport).toHaveBeenCalledWith({
      sourceId: 'S',
      applyToAll: null,
      items: [{ key: '0', include: true, targetLocalId: 'A', targetParentId: '', resolution: null }],
    })
    expect(state().step).toBe('summary')
    expect(state().result).toEqual(result)
    expect(load).toHaveBeenCalled()
    load.mockRestore()
  })

  it('asks about clashes, one item at a time or for all of them', async () => {
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'B', parentId: '' }, '1': { include: true, localId: 'B', parentId: '' } } })
    vi.mocked(service.findImportClashes).mockResolvedValue([clash, { ...clash, itemKey: '1' }])
    vi.mocked(service.runImport).mockReturnValue(resolvedImport({ ...result, items: [{ ...result.items[0]!, localId: 'B' }] }))
    await state().start()
    expect(state().step).toBe('clashes')
    await state().resolve('keep_both', false)
    expect(state().resolutions).toEqual({ '0': 'keep_both' })
    expect(service.runImport).not.toHaveBeenCalled()
    await state().resolve('skip', false)
    expect(vi.mocked(service.runImport).mock.calls[0]![0].items.map((item) => item.resolution)).toEqual(['keep_both', 'skip'])
    expect(state().step).toBe('summary')

    useImportStore.setState({ step: 'clashes', clashes: [clash], resolutions: {}, applyToAll: null })
    await state().resolve('replace', true)
    expect(vi.mocked(service.runImport).mock.calls[1]![0].applyToAll).toBe('replace')
  })

  it('keeps the preview with the error when clash detection or the import fails', async () => {
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    vi.mocked(service.findImportClashes).mockRejectedValue(new AppError('WKS-07404', 'gone'))
    await state().start()
    expect(state().step).toBe('preview')
    expect(state().error?.code).toBe('WKS-07404')
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    vi.mocked(service.runImport).mockReturnValue({ result: Promise.reject(new AppError('IMP-10499', 'cancelled')), cancel: vi.fn() })
    await state().start()
    expect(state().step).toBe('preview')
    expect(state().error?.code).toBe('IMP-10499')
    useImportStore.setState({ preview: null })
    await state().start()
    expect(service.runImport).toHaveBeenCalledTimes(1)
  })

  it('cancels a running import and follows its progress', async () => {
    const cancel = vi.fn()
    let finish: (value: ImportResult) => void = () => undefined
    vi.mocked(service.runImport).mockReturnValue({ result: new Promise((resolve) => (finish = resolve)), cancel })
    vi.mocked(service.findImportClashes).mockResolvedValue([])
    useImportStore.setState({ step: 'preview', preview, targets: { '0': { include: true, localId: 'A', parentId: '' } } })
    const running = state().start()
    await vi.waitFor(() => expect(state().step).toBe('importing'))
    state().applyProgress({ sourceId: 'other', done: 5, total: 9, currentItem: 'x' })
    expect(state().progress).toBeNull()
    state().applyProgress({ sourceId: 'S', done: 0, total: 1, currentItem: 'Acme API' })
    expect(state().progress).toEqual({ done: 0, total: 1, currentItem: 'Acme API' })
    state().cancel()
    expect(cancel).toHaveBeenCalledOnce()
    finish(result)
    await running
    state().applyProgress({ sourceId: 'S', done: 1, total: 1, currentItem: '' })
    expect(state().progress).toEqual({ done: 0, total: 1, currentItem: 'Acme API' })
  })

  it('goes back to the source step discarding the detected source', () => {
    useImportStore.setState({ step: 'preview', preview })
    state().back()
    expect(state().step).toBe('choose')
    expect(state().preview).toBeNull()
    expect(service.discardImport).toHaveBeenCalledWith('S')
  })

  it('opens an imported item, switching workspace when needed', async () => {
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    const load = vi.spyOn(useSidebarStore.getState(), 'load').mockResolvedValue()
    const openPinned = vi.spyOn(useTabsStore.getState(), 'openPinned').mockResolvedValue()
    useSidebarStore.setState({
      tree: [
        {
          id: 'C',
          kind: 'collection',
          name: 'Acme API',
          method: null,
          url: null,
          error: null,
          warning: null,
          children: [{ id: 'R', kind: 'request', name: 'Ping', method: 'GET', url: '', error: null, warning: null, children: [] }],
        },
      ],
    })
    useImportStore.setState({ step: 'summary', result })
    await state().openItem('B', 'R')
    expect(activate).toHaveBeenCalledWith('B')
    expect(openPinned).toHaveBeenCalledWith('R')
    expect(useSidebarStore.getState().selectedId).toBe('R')
    expect(useSidebarStore.getState().expanded.C).toBe(true)
    expect(state().step).toBe('closed')
    await state().openItem('A', 'C')
    expect(activate).toHaveBeenCalledTimes(1)
    expect(openPinned).toHaveBeenCalledTimes(1)
    await state().openItem('A', 'missing')
    expect(useSidebarStore.getState().selectedId).toBe('C')
    expect(load).toHaveBeenCalledTimes(3)
    activate.mockRestore()
    load.mockRestore()
    openPinned.mockRestore()
    useSidebarStore.getState().reset()
  })
})
