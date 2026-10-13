import type {
  HistoryEntry as HistoryEntryDto,
  HistoryEntryDetail as HistoryEntryDetailDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { SEND_ERROR_KIND_SET, SEND_OUTCOME_SET, toRequestDraft, type RequestDraft, type SendErrorKind, type SendOutcome } from './request'

export interface HistoryEntry {
  id: string
  sentAt: string
  method: string
  url: string
  outcome: SendOutcome
  statusCode: number | null
  errorKind: SendErrorKind | null
  totalMs: number
  sizeBytes: number
  sourceRequestId: string | null
}

export interface HistoryEntryDetail {
  entry: HistoryEntry
  request: RequestDraft
}

export interface HistoryQuery {
  query?: string
  limit?: number
  beforeId?: string
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

export function toHistoryEntry(dto: HistoryEntryDto): HistoryEntry {
  return {
    id: dto.id,
    sentAt: dto.sent_at,
    method: dto.method,
    url: dto.url,
    outcome: toOutcome(dto.outcome),
    statusCode: dto.status_code ?? null,
    errorKind: toErrorKind(dto.error_kind),
    totalMs: dto.total_ms,
    sizeBytes: dto.size_bytes,
    sourceRequestId: dto.source_request_id === '' ? null : dto.source_request_id,
  }
}

export function toHistoryEntryDetail(dto: HistoryEntryDetailDto): HistoryEntryDetail {
  return { entry: toHistoryEntry(dto.entry), request: toRequestDraft(dto.request) }
}
