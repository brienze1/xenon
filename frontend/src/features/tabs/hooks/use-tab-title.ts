import { useEnvironmentsStore } from '@/features/environments/store'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { selectLocation, useSidebarStore } from '@/features/sidebar/store'
import { treeExampleName, treeNodeName } from '@/features/tabs/lib/tab-name'
import { useTabsStore } from '@/features/tabs/store'

export const UNTITLED_REQUEST = 'Untitled Request'
export const GLOBALS_TITLE = 'Globals'
export const SETTINGS_TITLE = 'Settings'
const UNKNOWN_ENVIRONMENT = 'Environment'
const UNKNOWN_COLLECTION = 'Collection'
const UNKNOWN_FOLDER = 'Folder'

/**
 * The name a tab shows: the saved name of its request, example, environment, collection or folder (a name being
 * edited shows only once it is saved), "Globals" or "Settings"; a request never saved shows its draft name.
 */
export function useTabTitle(tabId: string | null): string {
  const kind = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.kind)
  const environmentId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.environmentId ?? null)
  const draftName = useRequestBuilderStore((state) => (tabId ? state.tabs[tabId]?.draft.name : null) || UNTITLED_REQUEST)
  const loadedName = useRequestBuilderStore((state) => (tabId ? (state.tabs[tabId]?.savedName ?? null) : null))
  const environmentName = useEnvironmentsStore((state) =>
    environmentId === null ? null : (state.environments.find((environment) => environment.id === environmentId)?.name ?? null),
  )
  const itemId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.itemId ?? null)
  const itemName = useSidebarStore((state) => (itemId === null ? null : (selectLocation(state.tree, itemId).at(-1) ?? null)))
  const requestId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.requestId ?? null)
  const exampleId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.exampleId ?? null)
  const savedRequestName = useSidebarStore((state) => (requestId === null ? null : treeNodeName(state.tree, requestId)))
  const savedExampleName = useSidebarStore((state) => (exampleId === null ? null : treeExampleName(state.tree, exampleId)))
  const loadedItemName = useItemSettingsStore((state) => (itemId === null ? null : (state.items[itemId]?.base?.name ?? null)))
  if (kind === 'collection' || kind === 'folder') {
    return itemName ?? loadedItemName ?? (kind === 'collection' ? UNKNOWN_COLLECTION : UNKNOWN_FOLDER)
  }
  if (kind === 'globals') {
    return GLOBALS_TITLE
  }
  if (kind === 'settings') {
    return SETTINGS_TITLE
  }
  if (kind === 'environment') {
    return environmentName ?? UNKNOWN_ENVIRONMENT
  }
  if (kind === 'example') {
    return savedExampleName ?? draftName
  }
  return savedRequestName ?? loadedName ?? draftName
}
