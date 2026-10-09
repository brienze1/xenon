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
})
