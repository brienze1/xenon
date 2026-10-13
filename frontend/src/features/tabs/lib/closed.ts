import type { ClosedTabState } from '@/platform/mappers/ui-state'

export const MAX_CLOSED_TABS = 20

/** Keeps the most recently closed tabs, newest last. */
export function latestClosedTabs(closed: readonly ClosedTabState[]): ClosedTabState[] {
  return closed.slice(Math.max(closed.length - MAX_CLOSED_TABS, 0))
}
