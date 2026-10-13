import { ItemBreadcrumb } from '@/features/sidebar/components/ItemBreadcrumb'
import { useTabTitle } from '@/features/tabs/hooks/use-tab-title'
import { useTabsStore } from '@/features/tabs/store'
import { adoptTreeRename, openItemTab, openRequestPreview } from '@/app/tree-actions'

interface TabBreadcrumbProps {
  tabId: string
}

/** The breadcrumb above a request or example tab's URL bar. */
export function TabBreadcrumb({ tabId }: TabBreadcrumbProps) {
  const itemId = useTabsStore((state) => {
    const tab = state.tabs.find((candidate) => candidate.id === tabId)
    return (tab?.kind === 'example' ? tab.exampleId : tab?.requestId) ?? null
  })
  const name = useTabTitle(tabId)
  return <ItemBreadcrumb itemId={itemId} name={name} onOpenItem={openItemTab} onOpenRequest={openRequestPreview} onRenamed={adoptTreeRename} />
}
