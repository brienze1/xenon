import type { TabKind } from '@/platform/mappers/ui-state'

export interface TabMeta {
  id: string
  requestId: string | null
  pinned: boolean
  /** Absent means `request`; `example` tabs edit one saved example of `requestId`. */
  kind?: TabKind
  exampleId?: string | null
}

export type CloseChoice = 'save' | 'discard' | 'cancel'

export type TabNotice = 'changed' | 'deleted'

export interface SaveDialogState {
  tabId: string
  mode: 'save' | 'saveAs'
  closeAfter: boolean
}
