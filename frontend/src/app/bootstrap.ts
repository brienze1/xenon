import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { onWorkspaceFolderMissing } from '@/platform/events'

export function bootstrap(): () => void {
  const unsubscribers = [
    onWorkspaceFolderMissing(({ localId }) => useWorkspacesStore.getState().markMissing(localId)),
    useWorkspacesStore.subscribe((state, previous) => {
      if (previous.activeLocalId !== null && state.activeLocalId !== previous.activeLocalId) {
        resetWorkspaceStores()
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
}
