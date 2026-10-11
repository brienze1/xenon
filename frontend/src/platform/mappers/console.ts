import type { ConsoleEntry as ConsoleEntryDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import type {
  ConsoleEntries as ConsoleEntriesDto,
  ConsoleEntry as ConsoleEventEntryDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/publisher/dto/models'

export type ConsoleEntryKind = 'request' | 'log' | 'error' | 'test_failure' | 'info'
export type ConsoleLevel = 'log' | 'info' | 'warn' | 'error' | 'debug'
export type ConsoleFilter = 'all' | 'requests' | 'logs' | 'errors'

const KINDS: ReadonlySet<string> = new Set<ConsoleEntryKind>(['request', 'log', 'error', 'test_failure', 'info'])
const LEVELS: ReadonlySet<string> = new Set<ConsoleLevel>(['log', 'info', 'warn', 'error', 'debug'])

export interface ConsoleHeader {
  key: string
  value: string
}

export interface ConsoleRequest {
  method: string
  url: string
  nested: boolean
  headers: ConsoleHeader[] | null
  body: string | null
}

export interface ConsoleResponse {
  statusCode: number
  statusText: string
  timeMs: number
  sizeBytes: number
  error: string | null
  headers: ConsoleHeader[] | null
  body: string | null
  bodyUrl: string | null
  contentType: string
  isBinary: boolean
}

/** One console line; a detail (from Get) adds headers, bodies and logged values. */
export interface ConsoleEntry {
  id: string
  sendId: string
  kind: ConsoleEntryKind
  level: ConsoleLevel
  summary: string
  createdAt: string
  arguments: string[]
  masked: boolean
  scriptLabel: string | null
  line: number | null
  column: number | null
  request: ConsoleRequest | null
  response: ConsoleResponse | null
  values: unknown[] | null
}

export interface ConsoleEntriesEvent {
  entries: ConsoleEntry[]
  cleared: boolean
}

export interface ConsoleQuery {
  afterId?: string
  limit: number
  filter?: ConsoleFilter
  search?: string
}

function kindOf(raw: string): ConsoleEntryKind {
  return KINDS.has(raw) ? (raw as ConsoleEntryKind) : 'info'
}

function levelOf(raw: string): ConsoleLevel {
  return LEVELS.has(raw) ? (raw as ConsoleLevel) : 'log'
}

export function toConsoleEntry(dto: ConsoleEntryDto | ConsoleEventEntryDto): ConsoleEntry {
  const detail = dto as ConsoleEntryDto
  const response = detail.response
  return {
    id: dto.id,
    sendId: dto.send_id,
    kind: kindOf(dto.kind),
    level: levelOf(dto.level),
    summary: dto.summary,
    createdAt: dto.created_at,
    arguments: dto.arguments ?? [],
    masked: dto.masked,
    scriptLabel: dto.script?.label ?? null,
    line: dto.line ?? null,
    column: dto.column ?? null,
    request: detail.request
      ? { method: detail.request.method, url: detail.request.url, nested: detail.request.nested, headers: detail.request.headers ?? null, body: detail.request.body ?? null }
      : null,
    response: response
      ? {
          statusCode: response.status_code,
          statusText: response.status_text,
          timeMs: response.time_ms,
          sizeBytes: response.size_bytes,
          error: typeof response.error === 'string' ? response.error || null : (response.error?.message ?? null),
          headers: response.headers ?? null,
          body: response.body ?? null,
          bodyUrl: response.body_url || null,
          contentType: response.content_type ?? '',
          isBinary: response.is_binary ?? false,
        }
      : null,
    values: detail.values ?? null,
  }
}

export function toConsoleEntriesEvent(dto: ConsoleEntriesDto): ConsoleEntriesEvent {
  return { entries: (dto.entries ?? []).map(toConsoleEntry), cleared: dto.cleared }
}
