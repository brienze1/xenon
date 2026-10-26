import { useAIAccessStore } from '@/features/ai-access/store'
import { useConsoleStore } from '@/features/console/store'
import { useCookiesStore } from '@/features/cookies/store'
import type { DocumentationSaved } from '@/features/documentation/model'
import { useDocumentationStore } from '@/features/documentation/store'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useExportStore } from '@/features/export/store'
import { useImportStore } from '@/features/import-export/store'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useGitAccountsStore, useSettingsStore } from '@/features/settings/store'
import type { ResponseEntry } from '@/features/response-viewer/model'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useRunnerStore } from '@/features/runner/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { STATUS_REFRESH_MS, useSyncStore } from '@/features/sync/store'
import { useTabsStore } from '@/features/tabs/store'
import { useUpdatesStore } from '@/features/updates/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import {
  onAIAccessStatusChanged,
  onApprovalRequested,
  onApprovalResolved,
  onConsoleEntries,
  onHistoryChanged,
  onCookiesChanged,
  onGitHostKeyUnknown,
  onGitProgress,
  onGitStatusChanged,
  onExportProgress,
  onExportRequested,
  onImportFilesDropped,
  onImportProgress,
  onImportRequested,
  onUpdateStatus,
  onRequestChangedOnDisk,
  onRunFinished,
  onRunProgress,
  onVariablesChanged,
  onVaultStatusChanged,
  onWorkspaceFolderMissing,
  onWorkspaceTreeChanged,
  onWorkspacesChanged,
} from '@/platform/events'

type ResponseEntries = Record<string, ResponseEntry>

/** Brings the sidebar, open request tabs and the collection or folder tab up to date with a saved documentation edit. */
function catchUpWithDocumentation(saved: DocumentationSaved): void {
  if (saved.localId !== useWorkspacesStore.getState().activeLocalId) {
    return
  }
  const tabs = useTabsStore.getState()
  for (const request of saved.updated) {
    void tabs.handleChangedOnDisk({ localId: saved.localId, requestId: request.id, deleted: false, revision: request.revision })
  }
  if (saved.deleted.length > 0) {
    tabs.handleDeleted(saved.deleted.map((request) => request.id))
  }
  if (saved.collectionChanged) {
    void useItemSettingsStore.getState().load(saved.itemId)
  }
  void useSidebarStore.getState().load()
}
type BuilderTabs = ReturnType<typeof useRequestBuilderStore.getState>['tabs']

export function bootstrap(): () => void {
  const flush = () => void useTabsStore.getState().flushPersist()
  const refreshSync = () => void useSyncStore.getState().refresh()
  window.addEventListener('beforeunload', flush)
  window.addEventListener('focus', refreshSync)
  const syncTimer = window.setInterval(refreshSync, STATUS_REFRESH_MS)
  const unsubscribers = [
    onWorkspaceFolderMissing(({ localId }) => useWorkspacesStore.getState().markMissing(localId)),
    onWorkspaceTreeChanged(({ localId }) => {
      if (useWorkspacesStore.getState().activeLocalId === localId) {
        void useSidebarStore.getState().load()
      }
    }),
    onRequestChangedOnDisk((event) => {
      void useTabsStore.getState().handleChangedOnDisk(event)
      void useDocumentationStore.getState().refreshAll()
    }),
    onVariablesChanged((event) => {
      useEnvironmentsStore.getState().handleVariablesChanged(event)
      if (event.localId === useWorkspacesStore.getState().activeLocalId) {
        useRequestBuilderStore.getState().refreshInspections()
      }
    }),
    onHistoryChanged((localId) => {
      if (useWorkspacesStore.getState().activeLocalId === localId) {
        refreshHistory()
      }
    }),
    onWorkspacesChanged((activeLocalId) => void useWorkspacesStore.getState().follow(activeLocalId)),
    onApprovalRequested((approval) => useAIAccessStore.getState().add(approval)),
    onApprovalResolved(({ id }) => useAIAccessStore.getState().resolve(id)),
    onAIAccessStatusChanged((status) => useSettingsStore.getState().applyStatus(status)),
    onImportProgress((progress) => useImportStore.getState().applyProgress(progress)),
    onConsoleEntries((event) => useConsoleStore.getState().applyEvent(event)),
    onCookiesChanged((localId) => useCookiesStore.getState().jarChanged(localId)),
    onGitHostKeyUnknown((prompt) => useGitAccountsStore.getState().promptHostKey(prompt)),
    onGitStatusChanged(({ localId, status }) => useSyncStore.getState().applyStatus(localId, status)),
    onGitProgress((progress) => useSyncStore.getState().applyProgress(progress)),
    onImportRequested(() => openImport()),
    onExportProgress((progress) => useExportStore.getState().applyProgress(progress)),
    onExportRequested(() => void useExportStore.getState().openWorkspace()),
    onUpdateStatus((status) => useUpdatesStore.getState().applyEvent(status)),
    onRunProgress((event) => useRunnerStore.getState().applyProgress(event)),
    onRunFinished((event) => useRunnerStore.getState().applyFinished(event)),
    onImportFilesDropped((paths) => {
      if (openImport()) {
        void useImportStore.getState().detect({ paths })
      }
    }),
    onVaultStatusChanged((status) => {
      useEnvironmentsStore.getState().applyVaultStatus(status)
      useRequestBuilderStore.getState().refreshInspections()
    }),
    useEnvironmentsStore.subscribe((state, previous) => {
      if (state.activeId !== previous.activeId) {
        useRequestBuilderStore.getState().refreshInspections()
      }
      if (state.environments !== previous.environments && state.status === 'ready' && state.error === null) {
        useTabsStore.getState().handleEnvironmentsChanged(state.environments.map((environment) => environment.id))
      }
      if (state.drafts !== previous.drafts) {
        useTabsStore.getState().pinIfEdited()
      }
    }),
    useItemSettingsStore.subscribe((state, previous) => {
      if (state.items !== previous.items) {
        useTabsStore.getState().pinIfEdited()
      }
    }),
    useDocumentationStore.subscribe((state, previous) => {
      if (state.entries !== previous.entries) {
        useTabsStore.getState().pinIfEdited()
      }
      if (state.saved !== previous.saved && state.saved !== null) {
        catchUpWithDocumentation(state.saved)
      }
    }),
    useRequestBuilderStore.subscribe((state, previous) => {
      if (state.tabs !== previous.tabs) {
        useTabsStore.getState().pinIfEdited()
        persistTabs()
        if (savedATab(state.tabs, previous.tabs)) {
          refreshSync()
        }
      }
    }),
    useTabsStore.subscribe((state, previous) => {
      if (
        state.tabs !== previous.tabs ||
        state.activeTabId !== previous.activeTabId ||
        state.previewTabId !== previous.previewTabId ||
        state.closedTabs !== previous.closedTabs ||
        state.snippetOptions !== previous.snippetOptions
      ) {
        persistTabs()
      }
    }),
    useSidebarStore.subscribe((state, previous) => {
      if (state.tree !== previous.tree && state.status === 'ready') {
        useTabsStore.getState().handleTreeChanged(state.tree)
        void useItemSettingsStore.getState().refreshAll()
        void useDocumentationStore.getState().refreshAll()
        void useRunnerStore.getState().refreshAll()
      }
      if (
        state.expanded !== previous.expanded ||
        state.selectedId !== previous.selectedId ||
        state.query !== previous.query ||
        state.view !== previous.view ||
        state.historySource !== previous.historySource
      ) {
        persistTabs()
      }
    }),
    useWorkspacesStore.subscribe((state, previous) => {
      if (state.layout !== previous.layout) {
        persistTabs()
      }
      if (state.activeLocalId !== previous.activeLocalId && state.activeLocalId !== null) {
        void useEnvironmentsStore.getState().loadVault()
        useSettingsStore.getState().closePage()
      }
      if (state.activeLocalId !== previous.activeLocalId) {
        useSyncStore.getState().reset(state.activeLocalId)
        refreshSync()
        useRunnerStore.getState().reset()
        if (state.activeLocalId !== null) {
          void useRunnerStore.getState().restoreActive()
        }
      }
    }),
    useResponseViewerStore.subscribe((state, previous) => {
      if (state.entries !== previous.entries && finishedSending(state.entries, previous.entries)) {
        refreshHistory()
      }
    }),
    () => window.removeEventListener('beforeunload', flush),
    () => window.removeEventListener('focus', refreshSync),
    () => window.clearInterval(syncTimer),
  ]
  void useEnvironmentsStore.getState().loadVault()
  void useAIAccessStore.getState().load()
  void openLastWorkspace()
  void checkForUpdatesOnStartup()
  return () => {
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  }
}

/** Reads the update status (the running version), then runs the once-per-run startup check. */
async function checkForUpdatesOnStartup(): Promise<void> {
  await useUpdatesStore.getState().load()
  await useUpdatesStore.getState().checkOnStartup()
}

/** Opens the Import dialog when a workspace is active (imports need a target); returns whether it did. */
function openImport(): boolean {
  if (useWorkspacesStore.getState().activeLocalId === null) {
    return false
  }
  useImportStore.getState().open()
  return true
}

/** A save clears the dirty flag and moves the tab to the revision it wrote. */
function savedATab(tabs: BuilderTabs, previous: BuilderTabs): boolean {
  return Object.entries(tabs).some(([tabId, tab]) => {
    const before = previous[tabId]
    return before !== undefined && before.dirty && !tab.dirty && before.baseRevision !== tab.baseRevision
  })
}

function finishedSending(entries: ResponseEntries, previous: ResponseEntries): boolean {
  return Object.entries(entries).some(([tabId, entry]) => !entry.sending && previous[tabId]?.sending === true)
}

function refreshHistory(): void {
  if (useSidebarStore.getState().view === 'history') {
    void useSidebarStore.getState().loadHistory()
  }
}

function persistTabs(): void {
  if (useTabsStore.getState().localId !== null) {
    useTabsStore.getState().schedulePersist()
  }
}

export async function openLastWorkspace(): Promise<void> {
  await useWorkspacesStore.getState().load()
  const activeLocalId = useWorkspacesStore.getState().activeLocalId
  if (activeLocalId) {
    await useWorkspacesStore.getState().activate(activeLocalId)
  }
}
