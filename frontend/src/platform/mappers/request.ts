import type {
  Auth as AuthDto,
  AutoHeader as AutoHeaderDto,
  FormDataRow as FormDataRowDto,
  KeyValue as KeyValueDto,
  PathVariable as PathVariableDto,
  Request as RequestDto,
  SendResult as SendResultDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { newRowId } from '@/shared/lib/rows'
import type { AppError } from '../errors'

export const STANDARD_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const
export const AUTH_TYPES = ['inherit', 'none', 'bearer', 'basic', 'apikey', 'oauth2', 'awsv4', 'digest'] as const
export const BODY_MODES = ['none', 'formdata', 'urlencoded', 'raw', 'binary', 'graphql'] as const
export const RAW_LANGUAGES = ['json', 'xml', 'text', 'html', 'javascript'] as const
export const SEND_ERROR_KINDS = ['dns', 'connection_refused', 'timeout', 'tls', 'cancelled', 'missing_file', 'invalid_url', 'other'] as const

const AUTH_TYPE_SET: ReadonlySet<string> = new Set(AUTH_TYPES)
const BODY_MODE_SET: ReadonlySet<string> = new Set(BODY_MODES)
const RAW_LANGUAGE_SET: ReadonlySet<string> = new Set(RAW_LANGUAGES)
export const SEND_ERROR_KIND_SET: ReadonlySet<string> = new Set(SEND_ERROR_KINDS)

export type AuthType = (typeof AUTH_TYPES)[number]
export type BodyMode = (typeof BODY_MODES)[number]
export type RawLanguage = (typeof RAW_LANGUAGES)[number]
export type SendErrorKind = (typeof SEND_ERROR_KINDS)[number]

export interface KeyValueRow {
  id: string
  key: string
  value: string
  description: string
  enabled: boolean
}

export interface PathVariableRow {
  id: string
  key: string
  value: string
  description: string
}

export interface FormDataRow extends KeyValueRow {
  type: 'text' | 'file'
  src: string
}

export interface RequestAuth {
  type: AuthType
  bearer: { token: string }
  basic: { username: string; password: string }
}

export interface RequestBody {
  mode: BodyMode
  raw: { language: RawLanguage; content: string }
  formdata: FormDataRow[]
  urlencoded: KeyValueRow[]
  binary: { src: string }
  graphql: { query: string; variables: string }
}

export interface RequestSettings {
  timeoutMs: number
  followRedirects: boolean
  verifySsl: boolean
  encodeUrl: boolean
}

export interface RequestDraft {
  id: string
  name: string
  description: string
  method: string
  url: string
  params: KeyValueRow[]
  pathVariables: PathVariableRow[]
  headers: KeyValueRow[]
  disabledAutoHeaders: string[]
  auth: RequestAuth
  body: RequestBody
  settings: RequestSettings
}

export interface AutoHeader {
  key: string
  value: string
  disabled: boolean
}

export interface SendTimings {
  dnsMs: number
  connectMs: number
  tlsMs: number
  firstByteMs: number
  downloadMs: number
  totalMs: number
}

export interface SendFailure {
  kind: SendErrorKind
  message: string
  hint: string
}

export interface SendResult {
  tabId: string
  ok: boolean
  statusCode: number
  statusText: string
  proto: string
  headers: Array<{ key: string; value: string }>
  sizeBytes: number
  contentType: string
  isBinary: boolean
  bodyUrl: string
  inlineBody: string | null
  timings: SendTimings
  error: SendFailure | null
}

export function emptyDraft(): RequestDraft {
  return {
    id: '',
    name: 'Untitled Request',
    description: '',
    method: 'GET',
    url: '',
    params: [],
    pathVariables: [],
    headers: [],
    disabledAutoHeaders: [],
    auth: { type: 'inherit', bearer: { token: '' }, basic: { username: '', password: '' } },
    body: {
      mode: 'none',
      raw: { language: 'json', content: '' },
      formdata: [],
      urlencoded: [],
      binary: { src: '' },
      graphql: { query: '', variables: '' },
    },
    settings: { timeoutMs: 0, followRedirects: true, verifySsl: true, encodeUrl: true },
  }
}

function oneOf<T extends string>(values: ReadonlySet<string>, value: string | undefined, fallback: T): T {
  return values.has(value ?? '') ? (value as T) : fallback
}

export function toKeyValueRow(dto: KeyValueDto): KeyValueRow {
  return { id: newRowId(), key: dto.key, value: dto.value, description: dto.description, enabled: dto.enabled ?? true }
}

export function toKeyValueDto(row: KeyValueRow): KeyValueDto {
  return { key: row.key, value: row.value, description: row.description, enabled: row.enabled }
}

function toPathVariableRow(dto: PathVariableDto): PathVariableRow {
  return { id: newRowId(), key: dto.key, value: dto.value, description: dto.description }
}

function toFormDataRow(dto: FormDataRowDto): FormDataRow {
  return { ...toKeyValueRow(dto), type: dto.type === 'file' ? 'file' : 'text', src: dto.src }
}

export interface SavedRequest {
  draft: RequestDraft
  revision: string
}

export function toAuth(dto: AuthDto): RequestAuth {
  return {
    type: oneOf<AuthType>(AUTH_TYPE_SET, dto.type, 'inherit'),
    bearer: { token: dto.bearer.token },
    basic: { username: dto.basic.username, password: dto.basic.password },
  }
}

export function toAuthDto(auth: RequestAuth): AuthDto {
  return {
    type: auth.type,
    bearer: { token: auth.bearer.token },
    basic: { username: auth.basic.username, password: auth.basic.password },
  }
}

export function toSavedRequest(dto: RequestDto): SavedRequest {
  return { draft: toRequestDraft(dto), revision: dto.revision }
}

export function toRequestDraft(dto: RequestDto): RequestDraft {
  const fallback = emptyDraft()
  return {
    id: dto.id,
    name: dto.name || fallback.name,
    description: dto.description,
    method: dto.method || fallback.method,
    url: dto.url,
    params: (dto.params ?? []).map(toKeyValueRow),
    pathVariables: (dto.path_variables ?? []).map(toPathVariableRow),
    headers: (dto.headers ?? []).map(toKeyValueRow),
    disabledAutoHeaders: dto.auto_headers.disabled ?? [],
    auth: toAuth(dto.auth),
    body: {
      mode: oneOf<BodyMode>(BODY_MODE_SET, dto.body.mode, 'none'),
      raw: { language: oneOf<RawLanguage>(RAW_LANGUAGE_SET, dto.body.raw.language, 'json'), content: dto.body.raw.content },
      formdata: (dto.body.formdata ?? []).map(toFormDataRow),
      urlencoded: (dto.body.urlencoded ?? []).map(toKeyValueRow),
      binary: { src: dto.body.binary.src },
      graphql: { query: dto.body.graphql.query, variables: dto.body.graphql.variables },
    },
    settings: {
      timeoutMs: dto.settings.timeout_ms,
      followRedirects: dto.settings.follow_redirects ?? true,
      verifySsl: dto.settings.verify_ssl ?? true,
      encodeUrl: dto.settings.encode_url ?? true,
    },
  }
}

export function toRequestDto(draft: RequestDraft): RequestDto {
  return {
    id: draft.id,
    name: draft.name,
    description: draft.description,
    method: draft.method,
    url: draft.url,
    revision: '',
    params: draft.params.map(toKeyValueDto),
    path_variables: draft.pathVariables.map(({ key, value, description }) => ({ key, value, description })),
    headers: draft.headers.map(toKeyValueDto),
    auto_headers: { disabled: draft.disabledAutoHeaders },
    auth: toAuthDto(draft.auth),
    body: {
      mode: draft.body.mode,
      raw: { language: draft.body.raw.language, content: draft.body.raw.content },
      formdata: draft.body.formdata.map((row) => ({ ...toKeyValueDto(row), type: row.type, src: row.src })),
      urlencoded: draft.body.urlencoded.map(toKeyValueDto),
      binary: { src: draft.body.binary.src },
      graphql: { query: draft.body.graphql.query, variables: draft.body.graphql.variables },
    },
    settings: {
      timeout_ms: draft.settings.timeoutMs,
      follow_redirects: draft.settings.followRedirects,
      verify_ssl: draft.settings.verifySsl,
      encode_url: draft.settings.encodeUrl,
    },
  }
}

export function toAutoHeader(dto: AutoHeaderDto): AutoHeader {
  return { key: dto.key, value: dto.value, disabled: dto.disabled }
}

export function toSendResult(dto: SendResultDto): SendResult {
  return {
    tabId: dto.tab_id,
    ok: dto.ok,
    statusCode: dto.status_code,
    statusText: dto.status_text,
    proto: dto.proto,
    headers: (dto.headers ?? []).map(({ key, value }) => ({ key, value })),
    sizeBytes: dto.size_bytes,
    contentType: dto.content_type,
    isBinary: dto.is_binary,
    bodyUrl: dto.body_url,
    inlineBody: dto.inline_body ?? null,
    timings: {
      dnsMs: dto.timings.dns_ms,
      connectMs: dto.timings.connect_ms,
      tlsMs: dto.timings.tls_ms,
      firstByteMs: dto.timings.first_byte_ms,
      downloadMs: dto.timings.download_ms,
      totalMs: dto.timings.total_ms,
    },
    error: dto.error ? { kind: oneOf<SendErrorKind>(SEND_ERROR_KIND_SET, dto.error.kind, 'other'), message: dto.error.message, hint: dto.error.hint } : null,
  }
}

export function cancelledResult(tabId: string): SendResult {
  return {
    tabId,
    ok: false,
    statusCode: 0,
    statusText: '',
    proto: '',
    headers: [],
    sizeBytes: 0,
    contentType: '',
    isBinary: false,
    bodyUrl: '',
    inlineBody: null,
    timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 0 },
    error: { kind: 'cancelled', message: 'Request cancelled', hint: '' },
  }
}

export const MISSING_SECRETS = 'SND-06409'

export interface MissingSecrets {
  missing: string[]
  locked: string[]
}

/** Reads the secret names a send preflight refused (`SND-06409`), or `null` for any other error. */
export function missingSecretsOf(error: AppError): MissingSecrets | null {
  if (error.code !== MISSING_SECRETS) {
    return null
  }
  const names = (attribute: string) => error.details.find((detail) => detail.attribute === attribute)?.messages ?? []
  return { missing: names('missing_secrets'), locked: names('locked_secrets') }
}
