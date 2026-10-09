import type { AppError } from '@/platform/errors'
import type { AutoHeader, RequestDraft } from '@/platform/mappers/request'
import type { PendingSend } from '@/platform/services/requests'

export type SubTab = 'params' | 'auth' | 'headers' | 'body' | 'scripts' | 'settings'

export type BulkTable = 'params' | 'headers' | 'urlencoded'

export interface TabDraft {
  draft: RequestDraft
  autoHeaders: AutoHeader[]
  subTab: SubTab
  bulk: Partial<Record<BulkTable, string>>
  pending: PendingSend | null
  error: AppError | null
  baseRevision: string | null
  conflict: boolean
  saving: boolean
  dirty: boolean
}
