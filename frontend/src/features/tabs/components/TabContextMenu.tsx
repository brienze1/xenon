import type { ReactNode } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { useTabsStore } from '@/features/tabs/store'
import styles from './TabContextMenu.module.css'

interface TabContextMenuProps {
  tabId: string
  name: string
  children: ReactNode
}

/** Postman's tab right-click menu: "close" items ask about unsaved changes, "force" items discard them. */
export function TabContextMenu({ tabId, name, children }: TabContextMenuProps) {
  return (
    <ContextMenu.Root>
      <ContextMenu.Trigger className={styles.trigger}>{children}</ContextMenu.Trigger>
      <TabMenu tabId={tabId} name={name} />
    </ContextMenu.Root>
  )
}

interface TabMenuProps {
  tabId: string
  name: string
}

function TabMenu({ tabId, name }: TabMenuProps) {
  const { position, count, requestId, isExample } = useTabsStore(
    useShallow((state) => {
      const index = state.tabs.findIndex((tab) => tab.id === tabId)
      const tab = state.tabs[index]
      return { position: index, count: state.tabs.length, requestId: tab?.requestId ?? null, isExample: tab?.kind === 'example' }
    }),
  )
  const tabs = useTabsStore.getState
  return (
    <ContextMenu.Menu label={`${name} tab actions`}>
      <ContextMenu.Item onSelect={() => tabs().newTab()}>New Request</ContextMenu.Item>
      <ContextMenu.Item disabled={isExample} onSelect={() => tabs().duplicateTab(tabId)}>
        Duplicate Tab
      </ContextMenu.Item>
      <ContextMenu.Separator />
      <ContextMenu.Item onSelect={() => tabs().requestClose(tabId)}>Close Tab</ContextMenu.Item>
      <ContextMenu.Item onSelect={() => tabs().forceClose(tabId)}>Force Close Tab</ContextMenu.Item>
      <ContextMenu.Item disabled={count < 2} onSelect={() => tabs().closeOthers(tabId)}>
        Close Other Tabs
      </ContextMenu.Item>
      <ContextMenu.Item onSelect={() => tabs().closeAll()}>Close All Tabs</ContextMenu.Item>
      <ContextMenu.Item disabled={position < 1} onSelect={() => tabs().closeLeft(tabId)}>
        Close Tabs to the Left
      </ContextMenu.Item>
      <ContextMenu.Item disabled={position >= count - 1} onSelect={() => tabs().closeRight(tabId)}>
        Close Tabs to the Right
      </ContextMenu.Item>
      <ContextMenu.Item tone="danger" onSelect={() => tabs().forceCloseAll()}>
        Force Close All Tabs
      </ContextMenu.Item>
      <ContextMenu.Separator />
      <ContextMenu.Item disabled={requestId === null} onSelect={() => tabs().revealInSidebar(tabId)}>
        Reveal in Sidebar
      </ContextMenu.Item>
    </ContextMenu.Menu>
  )
}
