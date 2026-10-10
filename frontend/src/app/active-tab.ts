import type { TabMeta } from '@/features/tabs/model'

interface OpenTabs {
  tabs: readonly TabMeta[]
  activeTabId: string | null
}

/** Whether the active tab shows a saved example rather than a request. */
export function isActiveTabExample({ tabs, activeTabId }: OpenTabs): boolean {
  return activeTabId !== null && tabs.some((tab) => tab.id === activeTabId && tab.kind === 'example')
}
