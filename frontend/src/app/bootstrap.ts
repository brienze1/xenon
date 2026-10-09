import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { onWorkspaceFolderMissing } from '@/platform/events'

export function bootstrap(): () => void {
  const unsubscribers = [
    onWorkspaceFolderMissing(({ localId }) => useWorkspacesStore.getState().markMissing(localId)),
    useWorkspacesStore.subscribe((state, previous) => {
      if (state.activeLocalId === previous.activeLocalId) {
        return
      }
      if (previous.activeLocalId !== null) {
        resetWorkspaceStores()
      }
      if (state.activeLocalId !== null) {
        void useSidebarStore.getState().load()
      }
    }),
  ]
  void openLastWorkspace()
  return () => {
    for (const unsubscribe of unsubscribers) {
      unsubscribe()
    }
  }
}

export async function openLastWorkspace(): Promise<void> {
  await useWorkspacesStore.getState().load()
  const activeLocalId = useWorkspacesStore.getState().activeLocalId
  if (activeLocalId) {
    await useWorkspacesStore.getState().activate(activeLocalId)
  }
}

function resetWorkspaceStores(): void {
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}
