import { ItemBreadcrumb } from '@/features/sidebar/components/ItemBreadcrumb'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabTitle } from '@/features/tabs/hooks/use-tab-title'
import { selectNameErrorText, useTabsStore } from '@/features/tabs/store'
import { EditableName } from '@/shared/ui/editable-name/EditableName'
import { openItemTab, openRequestPreview } from '@/app/tree-actions'

interface TabBreadcrumbProps {
  tabId: string
}

/** The breadcrumb above a request or example tab's URL bar; its last crumb is the tab's editable name. */
export function TabBreadcrumb({ tabId }: TabBreadcrumbProps) {
  const isExample = useTabsStore((state) => state.tabs.find((candidate) => candidate.id === tabId)?.kind === 'example')
  const itemId = useTabsStore((state) => {
    const tab = state.tabs.find((candidate) => candidate.id === tabId)
    return (tab?.kind === 'example' ? tab.exampleId : tab?.requestId) ?? null
  })
  const title = useTabTitle(tabId)
  const draftName = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name ?? '')
  const nameError = useTabsStore((state) => selectNameErrorText(state, tabId))
  const editTabName = useTabsStore((state) => state.editTabName)
  const cancelTabName = useTabsStore((state) => state.cancelTabName)
  return (
    <ItemBreadcrumb
      itemId={itemId}
      name={title}
      nameEditor={
        <EditableName
          value={draftName}
          placeholder={itemId === null ? title : undefined}
          label={isExample ? 'Rename example' : 'Rename request'}
          error={nameError}
          onChange={(name) => editTabName(tabId, name)}
          onCancel={() => cancelTabName(tabId)}
        />
      }
      onOpenItem={openItemTab}
      onOpenRequest={openRequestPreview}
    />
  )
}
