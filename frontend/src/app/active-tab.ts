import type { EnvironmentTabTarget, TabMeta } from '@/features/tabs/model'
import type { TabKind } from '@/platform/mappers/ui-state'

interface OpenTabs {
  tabs: readonly TabMeta[]
  activeTabId: string | null
}

/** What the active tab edits; `request` when no tab is open. */
export function activeTabKind({ tabs, activeTabId }: OpenTabs): TabKind {
  return tabs.find((tab) => tab.id === activeTabId)?.kind ?? 'request'
}

/** Whether the active tab shows a saved example rather than a request. */
export function isActiveTabExample(open: OpenTabs): boolean {
  return open.activeTabId !== null && activeTabKind(open) === 'example'
}

/** Whether the active tab is a request tab (the only kind that sends and has request details). */
export function isActiveTabRequest(open: OpenTabs): boolean {
  return open.activeTabId !== null && activeTabKind(open) === 'request'
}

/** The environment (or the globals) the active tab edits; null for request and example tabs. */
export function activeEnvironmentTarget({ tabs, activeTabId }: OpenTabs): EnvironmentTabTarget | null {
  const active = tabs.find((tab) => tab.id === activeTabId)
  if (active?.kind === 'globals') {
    return { kind: 'globals' }
  }
  return active?.kind === 'environment' && active.environmentId ? { kind: 'environment', environmentId: active.environmentId } : null
}
