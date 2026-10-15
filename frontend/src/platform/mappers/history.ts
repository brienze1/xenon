import type {
  AIActivity as AIActivityDto,
  HistoryEntry as HistoryEntryDto,
  HistoryEntryDetail as HistoryEntryDetailDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { SEND_ERROR_KIND_SET, SEND_OUTCOME_SET, toRequestDraft, type RequestDraft, type SendErrorKind, type SendOutcome } from './request'

/** Who made a history entry: the human in the UI or an AI client through MCP. */
export type HistorySource = 'user' | 'ai'

/** The History view's source filter. */
export type HistorySourceFilter = 'all' | HistorySource

/** A `send` entry is a request that went out; an `action` entry is any other AI tool call. */
export type HistoryKind = 'send' | 'action'

/** `ok`, an error (with its code), denied by the human or timed out waiting for approval. */
export type ActivityOutcome = 'ok' | 'error' | 'denied' | 'timed_out'

export interface ActivityTarget {
  id: string | null
  name: string | null
  path: string | null
}

/** What an AI action row did; `args` (masked, at most 2 KB) only comes with one entry's detail. */
export interface HistoryActivity {
  tool: string
  action: string | null
  target: ActivityTarget | null
  outcome: ActivityOutcome
  code: string | null
  durationMs: number
  args: string | null
}

export interface HistoryEntry {
  id: string
  sentAt: string
  kind: HistoryKind
  source: HistorySource
  /** The AI client's name; null for the human's sends. */
  client: string | null
  method: string
  url: string
  outcome: SendOutcome
  statusCode: number | null
  errorKind: SendErrorKind | null
  totalMs: number
  sizeBytes: number
  sourceRequestId: string | null
  /** The alias the send used; null for the Default or an unsaved request. */
  aliasName: string | null
  activity: HistoryActivity | null
}

export interface HistoryEntryDetail {
  entry: HistoryEntry
  request: RequestDraft
  activity: HistoryActivity | null
}

export interface HistoryQuery {
  query?: string
  limit?: number
  beforeId?: string
  source?: HistorySourceFilter
}

export const HISTORY_SOURCE_FILTERS: readonly HistorySourceFilter[] = ['all', 'user', 'ai']

const ACTIVITY_OUTCOMES: readonly ActivityOutcome[] = ['ok', 'error', 'denied', 'timed_out']

export function toHistorySourceFilter(value: string | undefined): HistorySourceFilter {
  return HISTORY_SOURCE_FILTERS.find((known) => known === value) ?? 'all'
}

function toErrorKind(kind: string | undefined): SendErrorKind | null {
  if (!kind) {
    return null
  }
  return SEND_ERROR_KIND_SET.has(kind) ? (kind as SendErrorKind) : 'other'
}

function toOutcome(outcome: string | undefined): SendOutcome {
  return outcome && SEND_OUTCOME_SET.has(outcome) ? (outcome as SendOutcome) : 'sent'
}

function orNull(value: string | undefined): string | null {
  return value ? value : null
}

function toActivity(dto: AIActivityDto): HistoryActivity {
  return {
    tool: dto.tool,
    action: orNull(dto.action),
    target: dto.target ? { id: orNull(dto.target.id), name: orNull(dto.target.name), path: orNull(dto.target.path) } : null,
    outcome: ACTIVITY_OUTCOMES.find((known) => known === dto.outcome) ?? 'error',
    code: orNull(dto.code),
    durationMs: dto.duration_ms,
    args: orNull(dto.args),
  }
}

export function toHistoryEntry(dto: HistoryEntryDto): HistoryEntry {
  return {
    id: dto.id,
    sentAt: dto.sent_at,
    kind: dto.kind === 'action' ? 'action' : 'send',
    source: dto.source === 'ai' ? 'ai' : 'user',
    client: orNull(dto.client),
    method: dto.method ?? '',
    url: dto.url ?? '',
    outcome: toOutcome(dto.outcome),
    statusCode: dto.status_code ?? null,
    errorKind: toErrorKind(dto.error_kind),
    totalMs: dto.total_ms,
    sizeBytes: dto.size_bytes,
    sourceRequestId: dto.source_request_id ? dto.source_request_id : null,
    aliasName: orNull(dto.alias_name ?? ''),
    activity: dto.activity ? toActivity(dto.activity) : null,
  }
}

export function toHistoryEntryDetail(dto: HistoryEntryDetailDto): HistoryEntryDetail {
  return { entry: toHistoryEntry(dto.entry), request: toRequestDraft(dto.request), activity: dto.activity ? toActivity(dto.activity) : null }
}
