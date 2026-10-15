import { useEnvironmentsStore } from '@/features/environments/store'
import { useExamplesStore } from '@/features/examples/store'
import { useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { isNamePending } from '@/features/tabs/lib/name-draft'
import { namedTreeId, tabNameOf, treeExampleName, treeNodeName } from '@/features/tabs/lib/tab-name'
import { useTabsStore } from '@/features/tabs/store'

/** Whether the tab's name was edited and not saved yet: it marks the tab dirty like any other unsaved edit. */
export function useNamePending(tabId: string): boolean {
  const tab = useTabsStore((state) => state.tabs.find((candidate) => candidate.id === tabId))
  const treeId = tab ? namedTreeId(tab) : null
  const isExample = tab?.kind === 'example'
  const itemId = tab?.itemId ?? null
  const environmentId = tab?.kind === 'environment' ? (tab.environmentId ?? null) : null
  const builderName = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name)
  const builderSavedName = useRequestBuilderStore((state) => state.tabs[tabId]?.savedName ?? null)
  const treeName = useSidebarStore((state) => (treeId === null ? null : isExample ? treeExampleName(state.tree, treeId) : treeNodeName(state.tree, treeId)))
  const itemDraft = useItemSettingsStore((state) => (itemId === null ? undefined : state.items[itemId]?.draft?.name))
  const itemBase = useItemSettingsStore((state) => (itemId === null ? null : (state.items[itemId]?.base?.name ?? null)))
  const exampleName = useExamplesStore((state) => state.tabs[tabId]?.example?.name ?? null)
  const environmentDraft = useEnvironmentsStore((state) => (environmentId === null ? undefined : state.nameDrafts[environmentId]))
  const environmentName = useEnvironmentsStore((state) =>
    environmentId === null ? null : (state.environments.find((environment) => environment.id === environmentId)?.name ?? null),
  )
  if (!tab) {
    return false
  }
  const name = tabNameOf(tab, { builderName, builderSavedName, treeName, itemDraft, itemBase, exampleName, environmentDraft, environmentName })
  return name !== null && isNamePending(name.draft, name.saved)
}
