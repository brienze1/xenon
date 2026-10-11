import { beforeEach, describe, expect, it, vi } from 'vitest'

const handlers = new Map<string, (event: { data: unknown }) => void>()
const unsubscribe = vi.fn()

vi.mock('@wailsio/runtime', () => ({
  Events: {
    On: vi.fn((name: string, callback: (event: { data: unknown }) => void) => {
      handlers.set(name, callback)
      return unsubscribe
    }),
  },
}))

const events = await import('./events')

describe('platform events', () => {
  beforeEach(() => {
    handlers.clear()
  })

  it('maps workspace:tree-changed payloads', () => {
    const callback = vi.fn()
    expect(events.onWorkspaceTreeChanged(callback)).toBe(unsubscribe)
    handlers.get(events.WORKSPACE_TREE_CHANGED)?.({ data: { local_id: 'L1' } })
    expect(callback).toHaveBeenCalledWith({ localId: 'L1' })
  })

  it('maps request:changed-on-disk payloads', () => {
    const callback = vi.fn()
    events.onRequestChangedOnDisk(callback)
    handlers.get(events.REQUEST_CHANGED_ON_DISK)?.({
      data: { local_id: 'L1', request_id: 'R1', deleted: true, revision: 'abc' },
    })
    expect(callback).toHaveBeenCalledWith({ localId: 'L1', requestId: 'R1', deleted: true, revision: 'abc' })
  })

  it('maps export:progress payloads', () => {
    const callback = vi.fn()
    expect(events.onExportProgress(callback)).toBe(unsubscribe)
    handlers.get(events.EXPORT_PROGRESS)?.({ data: { done: 50, total: 1007 } })
    expect(callback).toHaveBeenCalledWith({ done: 50, total: 1007 })
  })

  it('calls back on export:requested', () => {
    const callback = vi.fn()
    expect(events.onExportRequested(callback)).toBe(unsubscribe)
    handlers.get(events.EXPORT_REQUESTED)?.({ data: { source: 'menu' } })
    expect(callback).toHaveBeenCalledOnce()
  })

  it('maps cookies:changed payloads', () => {
    const callback = vi.fn()
    expect(events.onCookiesChanged(callback)).toBe(unsubscribe)
    handlers.get(events.COOKIES_CHANGED)?.({ data: { local_id: 'L1' } })
    expect(callback).toHaveBeenCalledWith('L1')
  })

  it('maps workspace:folder-missing payloads', () => {
    const callback = vi.fn()
    events.onWorkspaceFolderMissing(callback)
    handlers.get(events.WORKSPACE_FOLDER_MISSING)?.({ data: { local_id: 'L2' } })
    expect(callback).toHaveBeenCalledWith({ localId: 'L2' })
  })

  it('maps variables:changed payloads', () => {
    const callback = vi.fn()
    events.onVariablesChanged(callback)
    handlers.get(events.VARIABLES_CHANGED)?.({ data: { local_id: 'L1', scopes: ['env:e1', 'globals'] } })
    expect(callback).toHaveBeenLastCalledWith({ localId: 'L1', scopes: ['env:e1', 'globals'], reason: null })
    handlers.get(events.VARIABLES_CHANGED)?.({ data: { local_id: 'L1', scopes: null, reason: 'active_environment_removed' } })
    expect(callback).toHaveBeenLastCalledWith({ localId: 'L1', scopes: [], reason: events.ACTIVE_ENVIRONMENT_REMOVED })
  })

  it('maps vault:status-changed payloads', () => {
    const callback = vi.fn()
    events.onVaultStatusChanged(callback)
    handlers.get(events.VAULT_STATUS_CHANGED)?.({
      data: { mode: 'passphrase', needs_passphrase_setup: false, locked_local_ids: ['L1'], broken_local_ids: null },
    })
    expect(callback).toHaveBeenCalledWith({ mode: 'passphrase', needsPassphraseSetup: false, lockedLocalIds: ['L1'], brokenLocalIds: [] })
  })

  it('maps import events', () => {
    const progress = vi.fn()
    const dropped = vi.fn()
    const requested = vi.fn()
    events.onImportProgress(progress)
    events.onImportFilesDropped(dropped)
    events.onImportRequested(requested)
    handlers.get(events.IMPORT_PROGRESS)?.({ data: { source_id: 'S', current_item: 'Acme API', done: 1, total: 2 } })
    expect(progress).toHaveBeenCalledWith({ sourceId: 'S', currentItem: 'Acme API', done: 1, total: 2 })
    handlers.get(events.IMPORT_FILES_DROPPED)?.({ data: { paths: ['/a.json'] } })
    handlers.get(events.IMPORT_FILES_DROPPED)?.({ data: { paths: null } })
    expect(dropped).toHaveBeenNthCalledWith(1, ['/a.json'])
    expect(dropped).toHaveBeenNthCalledWith(2, [])
    handlers.get(events.IMPORT_REQUESTED)?.({ data: { source: 'menu' } })
    expect(requested).toHaveBeenCalledOnce()
  })
})
