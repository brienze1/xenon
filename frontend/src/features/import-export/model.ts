import type { ImportClash, ImportPreview, ImportResult } from '@/platform/mappers/import'

export type ImportStep = 'closed' | 'choose' | 'detecting' | 'preview' | 'clashes' | 'importing' | 'summary'

/** Where one detected item goes: a workspace and, optionally, an existing collection or folder in it. */
export interface ImportTarget {
  include: boolean
  localId: string
  parentId: string
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

export interface ImportSession {
  preview: ImportPreview | null
  targets: Record<string, ImportTarget>
  clashes: ImportClash[]
  result: ImportResult | null
}
