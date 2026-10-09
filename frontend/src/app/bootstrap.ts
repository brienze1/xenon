import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore, type ResponseEntry } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useTabsStore } from '@/features/tabs/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { onRequestChangedOnDisk, onWorkspaceFolderMissing, onWorkspaceTreeChanged } from '@/platform/events'

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
    useRequestBuilderStore.subscribe((state, previous) => {
      if (state.tabs !== previous.tabs) {
        useTabsStore.getState().pinIfEdited()
        persistTabs()
      }
    }),
    useTabsStore.subscribe((state, previous) => {
      if (state.tabs !== previous.tabs || state.activeTabId !== previous.activeTabId || state.previewTabId !== previous.previewTabId) {
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
    }),
    useResponseViewerStore.subscribe((state, previous) => {
      if (state.entries !== previous.entries && finishedSending(state.entries, previous.entries)) {
        refreshHistory()
      }
    }),
    () => window.removeEventListener('beforeunload', flush),
  ]
  void openLastWorkspace()
  return () => {
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  }
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
