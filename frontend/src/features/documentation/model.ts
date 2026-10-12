import type { AppError } from '@/platform/errors'
import type { ApiDocument, Documentation, DocumentationItem, DocumentationPlan } from '@/platform/mappers/documentation'

/** The rendered documentation or the YAML it is written in. */
export type DocumentationView = 'preview' | 'yaml'

export type DocumentationStatus = 'loading' | 'ready'

/** One collection's documentation: what Go generated, the YAML being edited and the save waiting for confirmation. */
export interface DocumentationEntry {
  status: DocumentationStatus
  loadError: AppError | null
  /** The document generated from the collection as last read. */
  base: Documentation | null
  /** The YAML in the editor; the base YAML until edited. */
  draft: string
  dirty: boolean
  view: DocumentationView
  /** The rendered view of the edited YAML; null until it has been read. */
  preview: ApiDocument | null
  /** Why the edited YAML cannot be read (with its line and column). */
  problem: AppError | null
  /** The changes a save would make, waiting for the user to confirm them. */
  plan: DocumentationPlan | null
  saving: boolean
  failure: AppError | null
  /** The collection changed on disk while the YAML had unsaved edits. */
  stale: boolean
}

/** What the last confirmed save wrote, for the other open views to catch up. */
export interface DocumentationSaved {
  localId: string
  collectionId: string
  collectionChanged: boolean
  updated: DocumentationItem[]
  created: DocumentationItem[]
  deleted: DocumentationItem[]
}
