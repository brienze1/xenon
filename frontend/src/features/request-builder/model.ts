import type { AppError } from '@/platform/errors'
import type { VariableInspection } from '@/platform/mappers/inspection'
import type { AutoHeader, RequestDraft, SavedAlias, SendAcknowledgement } from '@/platform/mappers/request'
import type { MissingSecrets, PendingSend, SendConfirmation } from '@/platform/services/requests'

export type SubTab = 'params' | 'auth' | 'headers' | 'body' | 'scripts' | 'settings'

export type BulkTable = 'params' | 'headers' | 'urlencoded'

/** `example` edits an example's request snapshot: no Send, no cURL paste, no scripts or settings. */
export type RequestBuilderMode = 'request' | 'example'

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
  /** The auth owner whose OAuth 2.0 token expired and could not be refreshed; asks Get new token / Send anyway. */
  expiredTokenOwner: string | null
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
  /** The request's name on disk, against which a name draft is pending; null when never saved or not known (a restored draft). */
  savedName: string | null
  conflict: boolean
  saving: boolean
  dirty: boolean
  /** The alias the tab shows; null for the Default. `draft`, `savedName` and `dirty` are that version's. */
  aliasId: string | null
  /** The request's aliases as last read from disk, in file order. */
  aliases: SavedAlias[]
  /** The request's Default as last read from disk; null when the tab never read the request. */
  savedDefault: RequestDraft | null
  /** The versions not on screen that the tab touched, keyed by alias id (`''` for the Default). */
  versions: Record<string, VersionState>
}

/** One version of a request kept aside while another is on screen. */
export interface VersionState {
  draft: RequestDraft
  savedName: string | null
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
