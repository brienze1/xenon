import type { AppError } from '@/platform/errors'
import type { ItemSettings } from '@/platform/mappers/tree'

/** The inner tabs of a collection or folder tab; `variables` and `documentation` exist for collections only. */
export type ItemSection = 'overview' | 'auth' | 'headers' | 'scripts' | 'variables' | 'documentation'

export type ItemStatus = 'loading' | 'ready'

export interface ItemFailure {
  error: AppError
  nextStep: string
}

/** One open collection or folder: what is on disk, the edits and where the tab is. */
export interface ItemEntry {
  status: ItemStatus
  loadError: AppError | null
  /** The settings as last read from (or written to) disk. */
  base: ItemSettings | null
  draft: ItemSettings | null
  /** The headers as `key:value` lines while the headers table is in bulk mode; null in table mode. */
  bulkText: string | null
  dirty: boolean
  saving: boolean
  failure: ItemFailure | null
  /** The settings changed on disk while this tab had unsaved edits. */
  stale: boolean
  section: ItemSection
}

/** What the app's scripts editor edits inside a collection or folder tab. */
export interface ItemScriptsSlot {
  itemId: string
  scripts: ItemSettings['scripts']
  inheritScripts: boolean | null
  onChange: (scripts: ItemSettings['scripts']) => void
  onInheritChange: (inheritScripts: boolean) => void
}

/** What the app's auth editor edits inside a collection or folder tab. */
export interface ItemAuthSlot {
  itemId: string
  kind: ItemSettings['kind']
  auth: ItemSettings['auth']
  onChange: (auth: ItemSettings['auth']) => void
}
