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
})
