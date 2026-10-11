import { useConsoleStore } from '@/features/console/store'
import { useCookiesStore } from '@/features/cookies/store'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useImportStore } from '@/features/import-export/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { ResponseEntry } from '@/features/response-viewer/model'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useTabsStore } from '@/features/tabs/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import {
  onConsoleEntries,
  onCookiesChanged,
  onImportFilesDropped,
  onImportProgress,
  onImportRequested,
  onRequestChangedOnDisk,
  onVariablesChanged,
  onVaultStatusChanged,
  onWorkspaceFolderMissing,
  onWorkspaceTreeChanged,
} from '@/platform/events'

type ResponseEntries = Record<string, ResponseEntry>

export function bootstrap(): () => void {
  const flush = () => void useTabsStore.getState().flushPersist()
  window.addEventListener('beforeunload', flush)
  const unsubscribers = [
    onWorkspaceFolderMissing(({ localId }) => useWorkspacesStore.getState().markMissing(localId)),
    onWorkspaceTreeChanged(({ localId }) => {
      if (useWorkspacesStore.getState().activeLocalId === localId) {
        void useSidebarStore.getState().load()
      }
    }),
    onRequestChangedOnDisk((event) => void useTabsStore.getState().handleChangedOnDisk(event)),
    onVariablesChanged((event) => {
      useEnvironmentsStore.getState().handleVariablesChanged(event)
      if (event.localId === useWorkspacesStore.getState().activeLocalId) {
        useRequestBuilderStore.getState().refreshInspections()
      }
    }),
    onImportProgress((progress) => useImportStore.getState().applyProgress(progress)),
    onConsoleEntries((event) => useConsoleStore.getState().applyEvent(event)),
    onCookiesChanged((localId) => useCookiesStore.getState().jarChanged(localId)),
    onImportRequested(() => openImport()),
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
    }),
    useRequestBuilderStore.subscribe((state, previous) => {
      if (state.tabs !== previous.tabs) {
        useTabsStore.getState().pinIfEdited()
        persistTabs()
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
      if (
        state.expanded !== previous.expanded ||
        state.selectedId !== previous.selectedId ||
        state.query !== previous.query ||
        state.view !== previous.view
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
      }
    }),
    useResponseViewerStore.subscribe((state, previous) => {
      if (state.entries !== previous.entries && finishedSending(state.entries, previous.entries)) {
        refreshHistory()
      }
    }),
    () => window.removeEventListener('beforeunload', flush),
  ]
  void useEnvironmentsStore.getState().loadVault()
  void openLastWorkspace()
  return () => {
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  }
}

/** Opens the Import dialog when a workspace is active (imports need a target); returns whether it did. */
function openImport(): boolean {
  if (useWorkspacesStore.getState().activeLocalId === null) {
    return false
  }
  useImportStore.getState().open()
  return true
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
