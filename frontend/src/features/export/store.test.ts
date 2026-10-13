import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ExportOptions } from '@/platform/mappers/export'

vi.mock('@/platform/services/export', () => ({
  getExportOptions: vi.fn(),
  pickExportTarget: vi.fn(),
  runExport: vi.fn(),
  showExportInFolder: vi.fn(),
}))

const service = await import('@/platform/services/export')
const { useExportStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

const options: ExportOptions = {
  suggestedFileName: 'acme',
  hasSecrets: false,
  hasCurrentValues: false,
  formats: [
    { format: 'xenon_yaml', suggestedFileName: 'acme.zip', shapes: ['zip'], includeSecretsAllowed: true },
    { format: 'postman', suggestedFileName: 'acme-postman.zip', shapes: ['zip'], includeSecretsAllowed: false },
  ],
}

describe('export store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useExportStore.getState().reset()
  })

  afterEach(() => {
    useExportStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null, workspaces: [] })
  })

  it('exports the active workspace only when its folder is there', async () => {
    vi.mocked(service.getExportOptions).mockResolvedValue(options)
    await useExportStore.getState().openWorkspace()
    useWorkspacesStore.setState({ activeLocalId: 'A', workspaces: [{ localId: 'A', name: 'Acme', path: '/w/acme', missing: true }] })
    await useExportStore.getState().openWorkspace()
    await useExportStore.getState().openItem('collection', 'C', 'Shop API')
    expect(service.getExportOptions).toHaveBeenCalledOnce()
    useExportStore.getState().reset()
    useWorkspacesStore.setState({ workspaces: [{ localId: 'A', name: 'Acme', path: '/w/acme', missing: false }] })
    await useExportStore.getState().openWorkspace()
    expect(useExportStore.getState()).toMatchObject({ step: 'configure', target: { localId: 'A', kind: 'workspace', itemId: '', name: 'Acme' } })
    expect(useExportStore.getState().choices).toMatchObject({ format: 'xenon_yaml', shape: 'zip' })
  })

  it('ignores options that arrive after the dialog was closed and a second open while one is shown', async () => {
    useWorkspacesStore.setState({ activeLocalId: 'A', workspaces: [] })
    let answer: (value: ExportOptions) => void = () => undefined
    vi.mocked(service.getExportOptions).mockReturnValueOnce(new Promise((resolve) => (answer = resolve)))
    const opening = useExportStore.getState().openItem('request', 'R', 'Ping')
    await useExportStore.getState().openItem('environment', 'E', 'local')
    expect(service.getExportOptions).toHaveBeenCalledOnce()
    useExportStore.getState().close()
    answer(options)
    await opening
    expect(useExportStore.getState().step).toBe('closed')
  })

  it('closing a running export cancels it and its late result is dropped', async () => {
    useWorkspacesStore.setState({ activeLocalId: 'A', workspaces: [] })
    vi.mocked(service.getExportOptions).mockResolvedValue(options)
    await useExportStore.getState().openItem('collection', 'C', 'Shop API')
    let finish: () => void = () => undefined
    const cancel = vi.fn()
    vi.mocked(service.pickExportTarget).mockResolvedValue('/out/acme.zip')
    vi.mocked(service.runExport).mockReturnValue({
      summary: new Promise((resolve) => (finish = () => resolve({ path: '/out/acme.zip', secrets: 'omitted', counts: { collections: 0, folders: 0, requests: 0, examples: 0, environments: 0, variables: 0 }, dropped: [], warnings: [] }))),
      cancel,
    })
    const running = useExportStore.getState().start()
    await vi.waitFor(() => expect(useExportStore.getState().step).toBe('exporting'))
    useExportStore.getState().close()
    expect(cancel).toHaveBeenCalledOnce()
    finish()
    await running
    expect(useExportStore.getState().step).toBe('closed')
    useExportStore.getState().setFormat('postman')
    useExportStore.getState().setShape('zip')
    useExportStore.getState().setChoice({ currentValuesAsInitial: true })
    await useExportStore.getState().showInFolder()
    expect(useExportStore.getState().choices).toBeNull()
    expect(service.showExportInFolder).not.toHaveBeenCalled()
  })
})
