import type { TabKind } from '@/platform/mappers/ui-state'

export interface TabMeta {
  id: string
  requestId: string | null
  pinned: boolean
  /**
   * Absent means `request`; `example` tabs edit one saved example of `requestId`; `environment`
   * and `globals` tabs edit a variable table; `collection` and `folder` tabs edit the settings of
   * `itemId` (`requestId` is null for all three).
   */
  kind?: TabKind
  exampleId?: string | null
  environmentId?: string | null
  itemId?: string | null
}

/** What a collection or folder tab edits. */
export interface ItemTabTarget {
  kind: 'collection' | 'folder'
  itemId: string
}

/** What an environment tab edits: one environment's variables or the workspace globals. */
export type EnvironmentTabTarget = { kind: 'environment'; environmentId: string } | { kind: 'globals' }

/** A single click opens a preview tab; a double click (or an explicit action) opens a pinned one. */
export type OpenMode = 'preview' | 'pinned'

export type CloseChoice = 'save' | 'discard' | 'cancel'

export type TabNotice = 'changed' | 'deleted'

export interface SaveDialogState {
  tabId: string
  mode: 'save' | 'saveAs'
  closeAfter: boolean
}
