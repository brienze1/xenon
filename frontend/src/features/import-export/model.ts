/** Marks a new workspace created for every included item at once rather than for one item. */
export const ALL_INCLUDED = '*'

export type ImportStep = 'closed' | 'choose' | 'detecting' | 'preview' | 'clashes' | 'importing' | 'summary'

/** Where one detected item goes: a workspace and, optionally, an existing collection or folder in it. */
export interface ImportTarget {
  include: boolean
  localId: string
  parentId: string
  /** A new name for the imported item (a request from a cURL command); empty keeps the detected one. */
  name?: string
}

export interface ImportLocation {
  id: string
  label: string
}

export interface ImportProgressState {
  done: number
  total: number
  currentItem: string
}
