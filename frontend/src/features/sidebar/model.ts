import type { SidebarView } from '@/platform/mappers/ui-state'

export type SidebarStatus = 'idle' | 'loading' | 'ready'

export interface SidebarSnapshot {
  expanded: Record<string, boolean>
  selectedId: string | null
  query: string
  view: SidebarView
}
