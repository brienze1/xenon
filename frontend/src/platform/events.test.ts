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

  it('maps the AI access events', () => {
    const status = vi.fn()
    expect(events.onAIAccessStatusChanged(status)).toBe(unsubscribe)
    handlers.get(events.AI_ACCESS_STATUS_CHANGED)?.({ data: { state: 'error', reason: 'port 47821 is in use', code: 'MCP-01500', clients: 0 } })
    expect(status).toHaveBeenCalledWith({ state: 'error', address: null, reason: 'port 47821 is in use', code: 'MCP-01500', clients: 0 })
    const requested = vi.fn()
    events.onApprovalRequested(requested)
    handlers.get(events.AI_APPROVAL_REQUESTED)?.({
      data: { id: 'a1', client: 'Claude Code', method: 'POST', url: '{{base}}/orders', environment: { id: 'e1', name: 'Production', color: 'red' }, created_at: '2026-10-13T10:00:00Z' },
    })
    expect(requested).toHaveBeenCalledWith({ id: 'a1', client: 'Claude Code', method: 'POST', url: '{{base}}/orders', environment: { id: 'e1', name: 'Production', color: 'red' }, createdAt: '2026-10-13T10:00:00Z', summary: null })
    const resolved = vi.fn()
    events.onApprovalResolved(resolved)
    handlers.get(events.AI_APPROVAL_RESOLVED)?.({ data: { id: 'a1', outcome: 'denied' } })
    expect(resolved).toHaveBeenCalledWith({ id: 'a1', outcome: 'denied' })
  })

  it('maps history:changed and workspaces:changed payloads', () => {
    const history = vi.fn()
    events.onHistoryChanged(history)
    handlers.get(events.HISTORY_CHANGED)?.({ data: { local_id: 'L1' } })
    expect(history).toHaveBeenCalledWith('L1')
    const workspaces = vi.fn()
    events.onWorkspacesChanged(workspaces)
    handlers.get(events.WORKSPACES_CHANGED)?.({ data: { active_local_id: 'L2' } })
    handlers.get(events.WORKSPACES_CHANGED)?.({ data: { active_local_id: '' } })
    expect(workspaces.mock.calls).toEqual([['L2'], [null]])
  })

  it('maps git:host-key-unknown payloads', () => {
    const callback = vi.fn()
    expect(events.onGitHostKeyUnknown(callback)).toBe(unsubscribe)
    handlers.get(events.GIT_HOST_KEY_UNKNOWN)?.({ data: { repository: 'ssh://git@h/r.git', host: 'h', fingerprint: 'SHA256:x', key_type: 'ssh-ed25519' } })
    expect(callback).toHaveBeenCalledWith({ repository: 'ssh://git@h/r.git', host: 'h', fingerprint: 'SHA256:x', keyType: 'ssh-ed25519' })
  })

  it('maps git:status-changed and git:progress payloads', () => {
    const status = vi.fn()
    expect(events.onGitStatusChanged(status)).toBe(unsubscribe)
    handlers.get(events.GIT_STATUS_CHANGED)?.({
      data: {
        local_id: 'L1',
        status: {
          state: 'behind',
          local_changes: 0,
          ahead: 0,
          behind: 2,
          settings: { repository: 'r', branch: 'main', subfolder: '', account_id: 'a', clone_path: '/c', auto_pull: false, auto_push: false },
        },
      },
    })
    expect(status).toHaveBeenCalledWith({ localId: 'L1', status: expect.objectContaining({ state: 'behind', behind: 2, lastError: null }) })
    const progress = vi.fn()
    events.onGitProgress(progress)
    handlers.get(events.GIT_PROGRESS)?.({ data: { local_id: 'L1', phase: 'fetch', done: 0, total: 3 } })
    expect(progress).toHaveBeenCalledWith({ localId: 'L1', phase: 'fetch', done: 0, total: 3 })
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

  it('maps update:status payloads', () => {
    const callback = vi.fn()
    expect(events.onUpdateStatus(callback)).toBe(unsubscribe)
    handlers.get(events.UPDATE_STATUS)?.({
      data: { state: 'downloading', reason: '', current_version: '0.1.4', latest_version: '0.1.7', release_url: 'https://x/v0.1.7', folder: '', received: 5, total: 10, automatic: true },
    })
    expect(callback).toHaveBeenCalledWith({
      state: 'downloading',
      reason: null,
      currentVersion: '0.1.4',
      latestVersion: '0.1.7',
      releaseUrl: 'https://x/v0.1.7',
      folder: '',
      received: 5,
      total: 10,
      automatic: true,
    })
  })
  it('maps the runner progress and finished events', () => {
    const progress = vi.fn()
    const finished = vi.fn()
    expect(events.onRunProgress(progress)).toBe(unsubscribe)
    expect(events.onRunFinished(finished)).toBe(unsubscribe)
    const progressDto = { current_row: 2, current_name: 'Get order', iteration: 0, completed: 1, planned: 6, tests_passed: 3, tests_failed: 1, elapsed_ms: 420, metrics: null }
    handlers.get(events.RUNNER_PROGRESS)?.({ data: { local_id: 'L1', run_id: 'R1', item_id: 'c1', type: 'functional', status: 'paused', progress: progressDto } })
    expect(progress).toHaveBeenCalledWith({
      localId: 'L1',
      runId: 'R1',
      itemId: 'c1',
      type: 'functional',
      status: 'paused',
      progress: { currentRow: 2, currentName: 'Get order', iteration: 0, completed: 1, planned: 6, testsPassed: 3, testsFailed: 1, elapsedMs: 420, metrics: null },
    })
    const summary = { iterations: 1, total_requests: 6, tests_passed: 10, tests_failed: 1, tests_skipped: 0, avg_response_ms: 120, duration_ms: 1840, passed: 5, failed: 1, skipped: 0, errors: 0, cancelled: 0 }
    handlers.get(events.RUNNER_FINISHED)?.({ data: { local_id: 'L1', run_id: 'R1', item_id: 'c1', target_name: 'Orders API', type: 'functional', status: 'stopped_on_error', summary } })
    expect(finished).toHaveBeenCalledWith(expect.objectContaining({ runId: 'R1', targetName: 'Orders API', status: 'stopped_on_error', summary: expect.objectContaining({ testsPassed: 10, testsFailed: 1 }) }))
  })
})
