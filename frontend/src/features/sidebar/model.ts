import type { SidebarView } from '@/platform/mappers/ui-state'

export type SidebarStatus = 'idle' | 'loading' | 'ready'

export interface SidebarSnapshot {
  expanded: Record<string, boolean>
  selectedId: string | null
  query: string
  view: SidebarView
}

/** An entry of the Environments view: one environment or the workspace globals. */
export type VariablesItem = { kind: 'environment'; environmentId: string } | { kind: 'globals' }

/** A single click previews an entry in a tab; a double click (or Open) pins the tab. */
export type OpenMode = 'preview' | 'pinned'
