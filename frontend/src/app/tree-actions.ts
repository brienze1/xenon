import { useExamplesStore } from '@/features/examples/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { OpenMode } from '@/features/tabs/model'
import { useTabsStore } from '@/features/tabs/store'
import type { ContainerKind } from '@/platform/mappers/tree'

/** Opens a collection's or folder's settings tab, as a sidebar row does. */
export function openItemTab(itemId: string, kind: ContainerKind, mode: OpenMode) {
  useTabsStore.getState().openItemTab({ kind, itemId }, mode)
}

/** Opens a saved request in the preview tab, as a sidebar row does. */
export function openRequestPreview(requestId: string) {
  void useTabsStore.getState().openPreview(requestId)
}

/** A request or example was renamed in the tree: the tabs showing it take the new name. */
export function adoptTreeRename(id: string) {
  void useRequestBuilderStore.getState().adoptRename(id)
  void useExamplesStore.getState().adoptRename(id)
}
