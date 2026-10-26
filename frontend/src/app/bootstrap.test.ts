import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultUIState } from '@/platform/mappers/ui-state'
import type { ExportProgress, GitProgress, GitStatusChanged, HostKeyUnknown, ImportProgress, RequestChangedOnDisk, VariablesChanged } from '@/platform/events'
import type { SyncStatus } from '@/platform/mappers/sync'
import type { ConsoleEntriesEvent } from '@/platform/mappers/console'
import type { VaultStatus } from '@/platform/mappers/vault'
import type { AIAccessRuntime, PendingApproval } from '@/platform/mappers/ai-access'
import { idleUpdateStatus, type UpdateStatus } from '@/platform/mappers/update-status'
import type { RunFinishedEvent, RunProgressEvent } from '@/platform/mappers/runner'

const unsubscribe = vi.fn()
const handlers: {
  folderMissing?: (event: { localId: string }) => void
  treeChanged?: (event: { localId: string }) => void
  changedOnDisk?: (event: RequestChangedOnDisk) => void
  variablesChanged?: (event: VariablesChanged) => void
  vaultStatusChanged?: (status: VaultStatus) => void
  importProgress?: (progress: ImportProgress) => void
  importRequested?: () => void
  filesDropped?: (paths: string[]) => void
  consoleEntries?: (event: ConsoleEntriesEvent) => void
  cookiesChanged?: (localId: string) => void
  exportProgress?: (progress: ExportProgress) => void
  exportRequested?: () => void
  hostKeyUnknown?: (prompt: HostKeyUnknown) => void
  gitStatusChanged?: (event: GitStatusChanged) => void
  gitProgress?: (event: GitProgress) => void
  historyChanged?: (localId: string) => void
  workspacesChanged?: (activeLocalId: string | null) => void
  approvalRequested?: (approval: PendingApproval) => void
  approvalResolved?: (resolved: { id: string; outcome: string }) => void
  aiAccessStatusChanged?: (status: AIAccessRuntime) => void
  updateStatus?: (status: UpdateStatus) => void
  runProgress?: (event: RunProgressEvent) => void
  runFinished?: (event: RunFinishedEvent) => void
} = {}

vi.mock('@/platform/events', () => ({
  onHistoryChanged: vi.fn((callback) => {
    handlers.historyChanged = callback
    return unsubscribe
  }),
  onWorkspacesChanged: vi.fn((callback) => {
    handlers.workspacesChanged = callback
    return unsubscribe
  }),
  onApprovalRequested: vi.fn((callback) => {
    handlers.approvalRequested = callback
    return unsubscribe
  }),
  onApprovalResolved: vi.fn((callback) => {
    handlers.approvalResolved = callback
    return unsubscribe
  }),
  onAIAccessStatusChanged: vi.fn((callback) => {
    handlers.aiAccessStatusChanged = callback
    return unsubscribe
  }),
  onConsoleEntries: vi.fn((callback) => {
    handlers.consoleEntries = callback
    return unsubscribe
  }),
  onCookiesChanged: vi.fn((callback) => {
    handlers.cookiesChanged = callback
    return unsubscribe
  }),
  onGitHostKeyUnknown: vi.fn((callback) => {
    handlers.hostKeyUnknown = callback
    return unsubscribe
  }),
  onGitStatusChanged: vi.fn((callback) => {
    handlers.gitStatusChanged = callback
    return unsubscribe
  }),
  onGitProgress: vi.fn((callback) => {
    handlers.gitProgress = callback
    return unsubscribe
  }),
  onWorkspaceFolderMissing: vi.fn((callback) => {
    handlers.folderMissing = callback
    return unsubscribe
  }),
  onWorkspaceTreeChanged: vi.fn((callback) => {
    handlers.treeChanged = callback
    return unsubscribe
  }),
  onRequestChangedOnDisk: vi.fn((callback) => {
    handlers.changedOnDisk = callback
    return unsubscribe
  }),
  onVariablesChanged: vi.fn((callback) => {
    handlers.variablesChanged = callback
    return unsubscribe
  }),
  onVaultStatusChanged: vi.fn((callback) => {
    handlers.vaultStatusChanged = callback
    return unsubscribe
  }),
  onImportProgress: vi.fn((callback) => {
    handlers.importProgress = callback
    return unsubscribe
  }),
  onImportRequested: vi.fn((callback) => {
    handlers.importRequested = callback
    return unsubscribe
  }),
  onImportFilesDropped: vi.fn((callback) => {
    handlers.filesDropped = callback
    return unsubscribe
  }),
  onExportProgress: vi.fn((callback) => {
    handlers.exportProgress = callback
    return unsubscribe
  }),
  onExportRequested: vi.fn((callback) => {
    handlers.exportRequested = callback
    return unsubscribe
  }),
  onUpdateStatus: vi.fn((callback) => {
    handlers.updateStatus = callback
    return unsubscribe
  }),
  onRunProgress: vi.fn((callback) => {
    handlers.runProgress = callback
    return unsubscribe
  }),
  onRunFinished: vi.fn((callback) => {
    handlers.runFinished = callback
    return unsubscribe
  }),
}))
vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  activeRuns: vi.fn(async () => []),
  getRun: vi.fn(),
  listRuns: vi.fn(),
}))
vi.mock('@/platform/services/updates', () => ({
  getUpdateStatus: vi.fn(),
  checkForUpdates: vi.fn(),
  checkOnStartup: vi.fn(),
  startUpdate: vi.fn(),
  restartToUpdate: vi.fn(),
  openReleasePage: vi.fn(),
}))
vi.mock('@/platform/services/export', () => ({
  getExportOptions: vi.fn(),
  pickExportTarget: vi.fn(),
  runExport: vi.fn(),
  showExportInFolder: vi.fn(),
}))
vi.mock('@/platform/services/import', () => ({
  detectImport: vi.fn(),
  discardImport: vi.fn(),
  findImportClashes: vi.fn(),
  pickImportSources: vi.fn(),
  runImport: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409', getVaultStatus: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
  reorderEnvironment: vi.fn(),
}))
vi.mock('@/platform/services/workspaces', () => ({
  listWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  activateWorkspace: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
  autoHeaders: vi.fn(),
  getRequest: vi.fn(),
  applyUrl: vi.fn(),
}))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), getItemSettings: vi.fn(), getDocumentation: vi.fn() }))
vi.mock('@/platform/services/sync', () => ({ getSyncStatus: vi.fn(), listChanges: vi.fn(async () => []) }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn(), getHistoryEntry: vi.fn(), clearHistory: vi.fn() }))
vi.mock('@/platform/services/ai-access', () => ({
  APPROVAL_ALREADY_RESOLVED: 'MCP-05409',
  APPROVAL_UNKNOWN: 'MCP-07400',
  listPendingApprovals: vi.fn(),
  answerApproval: vi.fn(),
  getAIAccessStatus: vi.fn(),
  copyText: vi.fn(),
}))

const services = await import('@/platform/services/workspaces')
const tree = await import('@/platform/services/tree')
const uiState = await import('@/platform/services/ui-state')
const requests = await import('@/platform/services/requests')
const history = await import('@/platform/services/history')
const vault = await import('@/platform/services/vault')
const variables = await import('@/platform/services/variables')
const environments = await import('@/platform/services/environments')
const sync = await import('@/platform/services/sync')
const runner = await import('@/platform/services/runner')
const { useRunnerStore } = await import('@/features/runner/store')
const aiAccess = await import('@/platform/services/ai-access')
const updates = await import('@/platform/services/updates')
const { useUpdatesStore } = await import('@/features/updates/store')
const { useAIAccessStore } = await import('@/features/ai-access/store')
const { useSettingsStore } = await import('@/features/settings/store')
const { useSyncStore } = await import('@/features/sync/store')
const { isScopeDirty, useEnvironmentsStore } = await import('@/features/environments/store')
const { useItemSettingsStore } = await import('@/features/item-settings/store')
const { useDocumentationStore } = await import('@/features/documentation/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useConsoleStore } = await import('@/features/console/store')
const { useCookiesStore } = await import('@/features/cookies/store')
const { useGitAccountsStore } = await import('@/features/settings/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore, PERSIST_DELAY_MS } = await import('@/features/tabs/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useImportStore } = await import('@/features/import-export/store')
const importService = await import('@/platform/services/import')
const exportService = await import('@/platform/services/export')
const { useExportStore } = await import('@/features/export/store')
const { bootstrap, openLastWorkspace } = await import('./bootstrap')

function resetWorkspaceStores() {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

const acme = { localId: 'A', name: 'Acme', path: '/w/acme', missing: false }
const lockedVault: VaultStatus = { mode: 'passphrase', needsPassphraseSetup: false, lockedLocalIds: ['A'], brokenLocalIds: [] }

const syncStatus: SyncStatus = {
  state: 'up_to_date',
  localChanges: 0,
  ahead: 0,
  behind: 0,
  lastSyncAt: null,
  lastError: null,
  notice: '',
  settings: { repository: 'r', branch: 'main', subfolder: '', accountId: 'a', clonePath: '/c', autoPull: false, autoPush: false },
}

async function started() {
  const cleanup = bootstrap()
  await vi.waitFor(() => expect(useWorkspacesStore.getState().status).toBe('ready'))
  return cleanup
}

const consoleEntry = {
  id: 'c1',
  sendId: 's',
  kind: 'log' as const,
  level: 'log' as const,
  summary: 'hi',
  createdAt: '2026-01-01T00:00:00.000Z',
  arguments: ['hi'],
  masked: false,
  scriptLabel: null,
  line: null,
  column: null,
  request: null,
  response: null,
  values: null,
}

describe('bootstrap', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(requests.applyUrl).mockImplementation(async (draft) => draft)
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    vi.mocked(history.listHistory).mockResolvedValue([])
    vi.mocked(vault.getVaultStatus).mockResolvedValue(lockedVault)
    vi.mocked(runner.activeRuns).mockResolvedValue([])
    vi.mocked(environments.listEnvironments).mockResolvedValue({ environments: [], activeId: null })
    vi.mocked(sync.getSyncStatus).mockResolvedValue(syncStatus)
    vi.mocked(sync.listChanges).mockResolvedValue([])
    vi.mocked(aiAccess.listPendingApprovals).mockResolvedValue([])
    vi.mocked(updates.getUpdateStatus).mockResolvedValue({ ...idleUpdateStatus(), currentVersion: '0.1.4' })
    vi.mocked(updates.checkOnStartup).mockResolvedValue({ ...idleUpdateStatus(), state: 'up_to_date', currentVersion: '0.1.4', latestVersion: '0.1.4' })
    useUpdatesStore.setState({ status: idleUpdateStatus(), toast: 'unseen' })
    useAIAccessStore.setState({ pending: [], answering: null, error: null })
    useSyncStore.getState().reset(null)
    useWorkspacesStore.setState({ workspaces: [], activeLocalId: null, missingLocalId: null, status: 'idle', error: null })
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
    resetWorkspaceStores()
  })

  it('loads the workspaces and reopens the last active one', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: 'A' })
    vi.mocked(services.activateWorkspace).mockResolvedValue({ summary: acme, tree: [], uiState: defaultUIState() })
    await openLastWorkspace()
    expect(services.activateWorkspace).toHaveBeenCalledWith('A')
    expect(useWorkspacesStore.getState().activeLocalId).toBe('A')
    expect(useTabsStore.getState().tabs).toHaveLength(1)
  })

  it('shows the welcome state when nothing was active', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    await openLastWorkspace()
    expect(services.activateWorkspace).not.toHaveBeenCalled()
    expect(useWorkspacesStore.getState().status).toBe('ready')
  })

  it('subscribes to backend events and unsubscribes on cleanup', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    handlers.folderMissing?.({ localId: 'A' })
    expect(useWorkspacesStore.getState().workspaces[0]?.missing).toBe(true)
    handlers.consoleEntries?.({ entries: [consoleEntry], cleared: false })
    expect(useConsoleStore.getState().entries.map((entry) => entry.id)).toEqual(['c1'])
    const jarChanged = vi.spyOn(useCookiesStore.getState(), 'jarChanged')
    handlers.cookiesChanged?.('A')
    expect(jarChanged).toHaveBeenCalledWith('A')
    const prompt = { repository: 'ssh://git@h/r.git', host: 'h', fingerprint: 'SHA256:x', keyType: 'ssh-ed25519' }
    handlers.hostKeyUnknown?.(prompt)
    expect(useGitAccountsStore.getState().hostKeyPrompt).toEqual(prompt)
    useGitAccountsStore.setState({ hostKeyPrompt: null })
    cleanup()
    expect(unsubscribe).toHaveBeenCalledTimes(23)
  })

  it('runs the startup update check exactly once and follows update:status', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [], activeLocalId: null })
    const cleanup = await started()
    await vi.waitFor(() => expect(useUpdatesStore.getState().status.state).toBe('up_to_date'))
    expect(updates.getUpdateStatus).toHaveBeenCalledTimes(1)
    expect(updates.checkOnStartup).toHaveBeenCalledTimes(1)
    expect(updates.checkForUpdates).not.toHaveBeenCalled()
    handlers.updateStatus?.({ ...idleUpdateStatus(), state: 'ready', currentVersion: '0.1.4', latestVersion: '0.1.7', automatic: true })
    expect(useUpdatesStore.getState()).toMatchObject({ status: { state: 'ready' }, toast: 'open' })
    cleanup()
    expect(updates.checkOnStartup).toHaveBeenCalledTimes(1)
  })

  it('opens the import dialog from the menu and from dropped files only when a workspace is active', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(importService.detectImport).mockResolvedValue({ sourceId: 'S', kind: 'postman_collection', version: '2.1', items: [], warnings: [], pmUsages: [], encryptedSecrets: false })
    const cleanup = await started()
    handlers.importRequested?.()
    handlers.filesDropped?.(['/x.json'])
    expect(useImportStore.getState().step).toBe('closed')
    expect(importService.detectImport).not.toHaveBeenCalled()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    handlers.importRequested?.()
    expect(useImportStore.getState().step).toBe('choose')
    handlers.filesDropped?.(['/x.json'])
    await vi.waitFor(() => expect(useImportStore.getState().step).toBe('preview'))
    expect(importService.detectImport).toHaveBeenCalledWith({ paths: ['/x.json'] })
    useImportStore.setState({ step: 'importing' })
    handlers.importProgress?.({ sourceId: 'S', done: 1, total: 2, currentItem: 'Acme API' })
    expect(useImportStore.getState().progress).toEqual({ done: 1, total: 2, currentItem: 'Acme API' })
    useImportStore.getState().reset()
    cleanup()
  })

  it('opens the workspace export from the menu only when a workspace is active and follows its progress', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(exportService.getExportOptions).mockResolvedValue({ suggestedFileName: 'acme', formats: [], hasSecrets: false, hasCurrentValues: false })
    const cleanup = await started()
    handlers.exportRequested?.()
    expect(useExportStore.getState().step).toBe('closed')
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    handlers.exportRequested?.()
    await vi.waitFor(() => expect(useExportStore.getState().step).toBe('configure'))
    expect(exportService.getExportOptions).toHaveBeenCalledWith({ localId: 'A', kind: 'workspace', itemId: '', name: 'Acme' })
    handlers.exportProgress?.({ done: 50, total: 1007 })
    expect(useExportStore.getState().progress).toBeNull()
    useExportStore.setState({ step: 'exporting' })
    handlers.exportProgress?.({ done: 50, total: 1007 })
    expect(useExportStore.getState().progress).toEqual({ done: 50, total: 1007 })
    useExportStore.getState().reset()
    cleanup()
  })

  it('reloads the sidebar when the active workspace tree changed outside Xenon', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    handlers.treeChanged?.({ localId: 'other' })
    expect(tree.getTree).not.toHaveBeenCalled()
    handlers.treeChanged?.({ localId: 'A' })
    expect(tree.getTree).toHaveBeenCalledWith('A')
    cleanup()
  })

  it('reads the Runner rows again whenever the sidebar tree changes, from an event or from Xenon itself', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(tree.getTree).mockResolvedValue([])
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    const refreshRunner = vi.spyOn(useRunnerStore.getState(), 'refreshAll').mockResolvedValue()
    handlers.treeChanged?.({ localId: 'other' })
    expect(refreshRunner).not.toHaveBeenCalled()
    handlers.treeChanged?.({ localId: 'A' })
    await vi.waitFor(() => expect(refreshRunner).toHaveBeenCalledTimes(1))
    useSidebarStore.getState().setTree([])
    expect(refreshRunner).toHaveBeenCalledTimes(2)
    refreshRunner.mockRestore()
    cleanup()
  })

  it('keeps following MCP writes through the tree and request-on-disk events', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    handlers.treeChanged?.({ localId: 'A' })
    expect(tree.getTree).toHaveBeenCalledWith('A')
    const handle = vi.spyOn(useTabsStore.getState(), 'handleChangedOnDisk').mockResolvedValue()
    const refreshDocs = vi.spyOn(useDocumentationStore.getState(), 'refreshAll').mockResolvedValue()
    const event = { localId: 'A', requestId: 'r1', deleted: false, revision: 'rev-2' }
    handlers.changedOnDisk?.(event)
    expect(handle).toHaveBeenCalledWith(event)
    expect(refreshDocs).toHaveBeenCalled()
    handle.mockRestore()
    refreshDocs.mockRestore()
    cleanup()
  })

  it('reloads the shown history when the active workspace history changed', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useSidebarStore.getState().setView('history')
    await vi.waitFor(() => expect(history.listHistory).toHaveBeenCalledTimes(1))
    handlers.historyChanged?.('other')
    expect(history.listHistory).toHaveBeenCalledTimes(1)
    handlers.historyChanged?.('A')
    expect(history.listHistory).toHaveBeenCalledTimes(2)
    expect(history.listHistory).toHaveBeenLastCalledWith('A', { query: '', source: 'all' })
    cleanup()
  })

  it('follows workspace changes made through MCP', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    const follow = vi.spyOn(useWorkspacesStore.getState(), 'follow').mockResolvedValue()
    handlers.workspacesChanged?.('A')
    expect(follow).toHaveBeenCalledWith('A')
    follow.mockRestore()
    cleanup()
  })

  it('queues AI approvals from start and from events, and follows the AI access status', async () => {
    const approval = { id: 'a1', client: 'Claude Code', method: 'POST', url: '{{base}}/orders', environment: { id: 'e1', name: 'Production', color: 'red' }, createdAt: 't', summary: null }
    vi.mocked(aiAccess.listPendingApprovals).mockResolvedValue([approval])
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    await vi.waitFor(() => expect(useAIAccessStore.getState().pending).toEqual([approval]))
    handlers.approvalRequested?.({ ...approval, id: 'a2' })
    expect(useAIAccessStore.getState().pending.map((pending) => pending.id)).toEqual(['a1', 'a2'])
    handlers.approvalResolved?.({ id: 'a1', outcome: 'timed_out' })
    expect(useAIAccessStore.getState().pending.map((pending) => pending.id)).toEqual(['a2'])
    handlers.aiAccessStatusChanged?.({ state: 'running', address: 'http://127.0.0.1:47821/mcp', reason: null, code: null, clients: 1 })
    expect(useSettingsStore.getState().aiStatus).toMatchObject({ state: 'running', address: 'http://127.0.0.1:47821/mcp', clients: 1 })
    cleanup()
  })

  it('closes the full-page Settings once a workspace is active', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useSettingsStore.getState().openPage()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    expect(useSettingsStore.getState().pageOpen).toBe(false)
    cleanup()
  })

  it('hands request changes on disk to the tabs', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    const handle = vi.spyOn(useTabsStore.getState(), 'handleChangedOnDisk').mockResolvedValue()
    const event = { localId: 'A', requestId: 'r1', deleted: true, revision: '' }
    handlers.changedOnDisk?.(event)
    expect(handle).toHaveBeenCalledWith(event)
    handle.mockRestore()
    cleanup()
  })

  it('pins an edited preview tab and saves the UI state after edits', async () => {
    vi.useFakeTimers()
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: { ...(await import('@/platform/mappers/request')).emptyDraft(), id: 'r1' }, revision: 'rev', aliases: [] })
    const cleanup = bootstrap()
    await vi.advanceTimersByTimeAsync(0)
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    await useTabsStore.getState().openPreview('r1')
    const preview = useTabsStore.getState().previewTabId!
    useRequestBuilderStore.getState().setUrl(preview, 'https://edited')
    expect(useTabsStore.getState().previewTabId).toBeNull()
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS * 3)
    expect(uiState.saveUIState).toHaveBeenCalledTimes(1)
    expect(uiState.saveUIState).toHaveBeenCalledWith('A', expect.objectContaining({ previewTabId: null }))
    window.dispatchEvent(new Event('beforeunload'))
    expect(uiState.saveUIState).toHaveBeenCalledTimes(2)
    cleanup()
  })

  it('closes the tabs of deleted environments and pins an environment preview once its variables are edited', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    const dev = { id: 'e1', name: 'dev', color: 'none', confirmBeforeSend: false, error: null } as const
    useEnvironmentsStore.setState({ environments: [dev], status: 'ready' })
    useTabsStore.getState().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'preview')
    const tabId = useTabsStore.getState().activeTabId!
    const draft = { base: [], rows: [], dirty: true, confirming: null, failure: null, saving: false }
    useEnvironmentsStore.setState({ drafts: { 'env:e1': draft } })
    expect(useTabsStore.getState().previewTabId).toBeNull()
    useEnvironmentsStore.setState({ environments: [], status: 'loading' })
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(true)
    useEnvironmentsStore.setState({ environments: [], status: 'ready' })
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(false)
    useEnvironmentsStore.getState().reset()
    cleanup()
  })

  it('keeps open environment tabs open and clean when the environments are reordered', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    const dev = { id: 'e1', name: 'dev', color: 'none', confirmBeforeSend: false, error: null } as const
    const prod = { id: 'e2', name: 'prod', color: 'red', confirmBeforeSend: true, error: null } as const
    useEnvironmentsStore.setState({ environments: [dev, prod], activeId: 'e1', status: 'ready' })
    useTabsStore.getState().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    useTabsStore.getState().openEnvironmentTab({ kind: 'environment', environmentId: 'e2' }, 'pinned')
    const tabIds = useTabsStore.getState().tabs.map((tab) => tab.id)
    vi.mocked(environments.reorderEnvironment).mockResolvedValue({ environments: [prod, dev], activeId: 'e1' })
    await expect(useEnvironmentsStore.getState().reorderEnvironment('e2', 0)).resolves.toBeNull()
    expect(useTabsStore.getState().tabs.map((tab) => tab.id)).toEqual(tabIds)
    expect(isScopeDirty(useEnvironmentsStore.getState(), 'env:e1')).toBe(false)
    expect(isScopeDirty(useEnvironmentsStore.getState(), 'env:e2')).toBe(false)
    useEnvironmentsStore.getState().reset()
    cleanup()
  })

  it('follows the tree: closes the tabs of deleted collections, rereads open settings, and pins a collection preview once edited', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', defaultUIState())
    const settings = { id: 'c1', kind: 'collection', name: 'Billing', description: '', auth: { type: 'inherit' }, headers: [], scripts: { preRequest: '', postResponse: '', disabled: [] }, inheritScripts: null } as never
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    const billing = { id: 'c1', kind: 'collection' as const, name: 'Billing', method: null, url: null, error: null, warning: null, children: [] }
    useSidebarStore.setState({ tree: [billing], status: 'ready' })
    useTabsStore.getState().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    const tabId = useTabsStore.getState().activeTabId!
    await useItemSettingsStore.getState().open('c1')
    useItemSettingsStore.getState().edit('c1', { description: 'edited' })
    expect(useTabsStore.getState().previewTabId).toBeNull()
    vi.mocked(tree.getItemSettings).mockClear()
    useSidebarStore.setState({ tree: [{ ...billing, name: 'Payments' }] })
    expect(tree.getItemSettings).toHaveBeenCalledWith('A', 'c1')
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(true)
    useSidebarStore.setState({ tree: [] })
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(false)
    useItemSettingsStore.getState().reset()
    cleanup()
  })

  it('catches the sidebar, request tabs and the collection settings up with a saved documentation edit', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useTabsStore.getState().hydrate('A', {
      ...defaultUIState(),
      tabs: [
        { tabId: 't1', requestId: 'r1', pinned: true, baseRevision: null, draft: null },
        { tabId: 't2', requestId: 'r2', pinned: true, baseRevision: null, draft: null },
      ],
      activeTabId: 't1',
    })
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.getItemSettings).mockResolvedValue({ id: 'c1' } as never)
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: { ...(await import('@/platform/mappers/request')).emptyDraft(), id: 'r1', name: 'Renamed' }, revision: 'rev-2', aliases: [] })
    await useItemSettingsStore.getState().open('c1')
    vi.mocked(tree.getItemSettings).mockClear()
    const item = { folder: '', method: 'GET', path: '' }
    useDocumentationStore.setState({
      saved: {
        localId: 'A',
        itemId: 'c1',
        collectionChanged: true,
        updated: [{ ...item, id: 'r1', name: 'Renamed', revision: 'rev-2' }],
        created: [],
        deleted: [{ ...item, id: 'r2', name: 'Gone', revision: '' }],
      },
    })
    await vi.waitFor(() => expect(tree.getTree).toHaveBeenCalledWith('A'))
    expect(tree.getItemSettings).toHaveBeenCalledWith('A', 'c1')
    expect(useTabsStore.getState().tabs.map((tab) => tab.id)).toEqual(['t1'])
    useDocumentationStore.getState().reset()
    useItemSettingsStore.getState().reset()
    cleanup()
  })

  it('refreshes the shown history when a send finishes', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(history.listHistory).mockResolvedValue([])
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    const viewer = useResponseViewerStore.getState()
    viewer.startSending('t1')
    viewer.setFailure('t1', new (await import('@/platform/errors')).AppError('x', 'y'))
    expect(history.listHistory).not.toHaveBeenCalled()
    useSidebarStore.getState().setView('history')
    expect(history.listHistory).toHaveBeenCalledTimes(1)
    viewer.startSending('t1')
    expect(history.listHistory).toHaveBeenCalledTimes(1)
    viewer.setFailure('t1', new (await import('@/platform/errors')).AppError('x', 'y'))
    expect(history.listHistory).toHaveBeenCalledTimes(2)
    cleanup()
  })

  it('saves the UI state when the layout changes', async () => {
    vi.useFakeTimers()
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = bootstrap()
    await vi.advanceTimersByTimeAsync(0)
    useWorkspacesStore.getState().setLayout({ sidebarWidth: 400 })
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS)
    expect(uiState.saveUIState).not.toHaveBeenCalled()
    useTabsStore.getState().hydrate('A', defaultUIState())
    useWorkspacesStore.getState().setLayout({ sidebarWidth: 420 })
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS)
    expect(uiState.saveUIState).toHaveBeenCalledWith('A', expect.objectContaining({ layout: expect.objectContaining({ sidebarWidth: 420 }) }))
    cleanup()
  })

  it('restores the running indicators on workspace activation, follows run events, and resets the runner on a switch', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const summary = { iterations: 1, totalRequests: 1, testsPassed: 2, testsFailed: 0, testsSkipped: 0, avgResponseMs: 1, durationMs: 1, passed: 1, failed: 0, skipped: 0, errors: 0, cancelled: 0 }
    const progress = { currentRow: 1, currentName: 'Login', iteration: 0, completed: 1, planned: 2, testsPassed: 1, testsFailed: 0, elapsedMs: 10, metrics: null }
    vi.mocked(runner.activeRuns).mockResolvedValue([
      { id: 'R1', itemId: 'c1', targetKind: 'collection', targetName: 'Orders', type: 'functional', status: 'running', environmentName: '', startedAt: '', finishedAt: '', progress, summary },
    ])
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    expect(runner.activeRuns).toHaveBeenCalledWith('A')
    await vi.waitFor(() => expect(Object.keys(useRunnerStore.getState().active)).toEqual(['c1']))
    handlers.runProgress?.({ localId: 'A', runId: 'R2', itemId: 'f1', type: 'functional', status: 'running', progress })
    expect(Object.keys(useRunnerStore.getState().active)).toEqual(['c1', 'f1'])
    handlers.runFinished?.({ localId: 'A', runId: 'R1', itemId: 'c1', targetName: 'Orders', type: 'functional', status: 'completed', summary })
    expect(Object.keys(useRunnerStore.getState().active)).toEqual(['f1'])
    expect(useRunnerStore.getState().notice).toMatchObject({ itemId: 'c1', targetName: 'Orders', passed: 2 })
    useWorkspacesStore.setState({ activeLocalId: null })
    expect(useRunnerStore.getState().active).toEqual({})
    expect(useRunnerStore.getState().notice).toBeNull()
    cleanup()
  })

  it('loads the vault status at start and on workspace activation, and follows vault events', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = await started()
    expect(vault.getVaultStatus).toHaveBeenCalledTimes(1)
    await vi.waitFor(() => expect(useEnvironmentsStore.getState().vault).toEqual(lockedVault))
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    expect(vault.getVaultStatus).toHaveBeenCalledTimes(2)
    await vi.waitFor(() => expect(useEnvironmentsStore.getState().passphrasePrompt).toBe('unlock'))
    handlers.vaultStatusChanged?.({ ...lockedVault, lockedLocalIds: [], brokenLocalIds: ['A'] })
    expect(useEnvironmentsStore.getState().vault.brokenLocalIds).toEqual(['A'])
    cleanup()
  })

  it('refreshes the environments and the open variable tables when variables change', async () => {
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    vi.mocked(variables.getVariableTable).mockResolvedValue({ scope: 'globals', rows: [], warnings: [] })
    const cleanup = await started()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ tables: { globals: { rows: [], warnings: [], status: 'ready', error: null } } })
    handlers.variablesChanged?.({ localId: 'other', scopes: ['globals'], reason: null })
    expect(variables.getVariableTable).not.toHaveBeenCalled()
    handlers.variablesChanged?.({ localId: 'A', scopes: ['env:e1'], reason: null })
    expect(environments.listEnvironments).toHaveBeenCalledWith('A')
    expect(variables.getVariableTable).not.toHaveBeenCalled()
    handlers.variablesChanged?.({ localId: 'A', scopes: ['globals'], reason: null })
    expect(variables.getVariableTable).toHaveBeenCalledWith('A', 'globals')
    useEnvironmentsStore.getState().reset()
    cleanup()
  })

  it('follows the sync status of the active workspace: activation, events, focus, saves and every 5 minutes', async () => {
    vi.useFakeTimers()
    const { STATUS_REFRESH_MS } = await import('@/features/sync/store')
    vi.mocked(services.listWorkspaces).mockResolvedValue({ workspaces: [acme], activeLocalId: null })
    const cleanup = bootstrap()
    await vi.advanceTimersByTimeAsync(0)
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    await vi.advanceTimersByTimeAsync(0)
    expect(useSyncStore.getState().localId).toBe('A')
    expect(sync.getSyncStatus).toHaveBeenCalledTimes(1)
    handlers.gitStatusChanged?.({ localId: 'A', status: { ...syncStatus, state: 'behind', behind: 2 } })
    expect(useSyncStore.getState().status?.behind).toBe(2)
    handlers.gitStatusChanged?.({ localId: 'other', status: syncStatus })
    expect(useSyncStore.getState().status?.state).toBe('behind')
    window.dispatchEvent(new Event('focus'))
    await vi.advanceTimersByTimeAsync(0)
    expect(sync.getSyncStatus).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(STATUS_REFRESH_MS)
    expect(sync.getSyncStatus).toHaveBeenCalledTimes(3)
    useRequestBuilderStore.setState({ tabs: { t1: { ...useRequestBuilderStore.getState().tabs.t1!, dirty: true, baseRevision: 'r1' } } })
    useRequestBuilderStore.setState({ tabs: { t1: { ...useRequestBuilderStore.getState().tabs.t1!, dirty: false, baseRevision: 'r2' } } })
    await vi.advanceTimersByTimeAsync(0)
    expect(sync.getSyncStatus).toHaveBeenCalledTimes(4)
    useSyncStore.setState({ busy: 'pull' })
    handlers.gitProgress?.({ localId: 'A', phase: 'fetch', done: 0, total: 3 })
    expect(useSyncStore.getState().progress?.phase).toBe('fetch')
    useSyncStore.setState({ busy: null })
    cleanup()
    await vi.advanceTimersByTimeAsync(STATUS_REFRESH_MS)
    expect(sync.getSyncStatus).toHaveBeenCalledTimes(4)
  })
})
