import { useExamplesStore } from '@/features/examples/store'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import type { RunFinishedNotice } from '@/features/runner/model'
import { useRunnerStore } from '@/features/runner/store'
import { useSidebarStore } from '@/features/sidebar/store'
import type { OpenMode } from '@/features/tabs/model'
import { useTabsStore } from '@/features/tabs/store'
import type { ContainerKind, TreeNode } from '@/platform/mappers/tree'

/** Opens a collection's or folder's settings tab, as a sidebar row does. */
export function openItemTab(itemId: string, kind: ContainerKind, mode: OpenMode) {
  useTabsStore.getState().openItemTab({ kind, itemId }, mode)
}

/** Run…: focuses (or opens) the item's tab pinned on its Runner, on Results while a run is active there, else on Setup. */
export function runItemTab(itemId: string, kind: ContainerKind) {
  openItemTab(itemId, kind, 'pinned')
  useItemSettingsStore.getState().setSection(itemId, 'runner')
  useRunnerStore.getState().showRunner(itemId)
}

function containerKindOf(nodes: readonly TreeNode[], id: string): ContainerKind | null {
  for (const node of nodes) {
    if (node.id === id) {
      return node.kind === 'collection' || node.kind === 'folder' ? node.kind : null
    }
    const found = containerKindOf(node.children, id)
    if (found) {
      return found
    }
  }
  return null
}

/** Open results (the finish toast): focuses (or opens) the target's tab on Runner › Results for that run. */
export function openRunResults(notice: RunFinishedNotice) {
  const kind = notice.kind ?? containerKindOf(useSidebarStore.getState().tree, notice.itemId)
  if (kind === null) {
    return
  }
  openItemTab(notice.itemId, kind, 'pinned')
  useItemSettingsStore.getState().setSection(notice.itemId, 'runner')
  void useRunnerStore.getState().openRun(notice.itemId, notice.runId)
}

/** Opens a saved request in the preview tab, as a sidebar row does. */
export function openRequestPreview(requestId: string) {
  void useTabsStore.getState().openPreview(requestId)
}

/** A request or example was renamed in the tree: its tabs take the new name, unless they hold a name of their own. */
export function adoptTreeRename(id: string, previousName: string) {
  void useRequestBuilderStore.getState().adoptRename(id, previousName)
  void useExamplesStore.getState().adoptRename(id, previousName)
}
