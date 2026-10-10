import type { AppError } from '@/platform/errors'
import type { VariableInspection } from '@/platform/mappers/inspection'
import type { AutoHeader, RequestDraft, SendAcknowledgement } from '@/platform/mappers/request'
import type { MissingSecrets, PendingSend, SendConfirmation } from '@/platform/services/requests'

export type SubTab = 'params' | 'auth' | 'headers' | 'body' | 'scripts' | 'settings'

export type BulkTable = 'params' | 'headers' | 'urlencoded'

export interface TabDraft {
  draft: RequestDraft
  autoHeaders: AutoHeader[]
  subTab: SubTab
  bulk: Partial<Record<BulkTable, string>>
  pending: PendingSend | null
  /** Set when Go refused a send because secrets have no value or are locked; asks Send anyway / Cancel. */
  missingSecrets: MissingSecrets | null
  /** Set when Go refused a send because the active environment asks for a confirmation first. */
  confirmation: SendConfirmation | null
  /** What the last send already acknowledged, kept so the next prompt adds to it. */
  acknowledgement: SendAcknowledgement
  /** The latest VariableService.Inspect answer for the draft (highlighting and Variables panel). */
  inspection: VariableInspection | null
  /** The last text selection in a request field, for "Move to secret variable". */
  selection: FieldSelection | null
  error: AppError | null
  /** Set after a cURL command replaced the request from the URL bar, so that one change can be undone. */
  curlPaste: CurlPaste | null
  baseRevision: string | null
  conflict: boolean
  saving: boolean
  dirty: boolean
}

/** A selection inside a request field, the field named like the Inspect references (`url`, `headers.0.value`, …). */
export interface FieldSelection {
  field: string
  from: number
  to: number
  text: string
}

/** The request as it was before a pasted cURL command replaced it, and what the parser ignored. */
export interface CurlPaste {
  previous: RequestDraft
  wasDirty: boolean
  notices: string[]
}
