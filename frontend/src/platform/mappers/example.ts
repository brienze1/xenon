import type {
  Example as ExampleDto,
  ExampleEdit as ExampleEditDto,
  ExampleRequest as ExampleRequestDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { toKeyValueDto, toKeyValueRow, toRequestDraft, toRequestDto, type KeyValueRow, type RequestDraft } from './request'

export type BodyEncoding = 'text' | 'base64'

/** The response half of a saved example; `body` is `null` when it is served by `bodyUrl` instead of inline. */
export interface ExampleResponse {
  code: number
  status: string
  contentType: string
  headers: KeyValueRow[]
  body: string | null
  bodyEncoding: BodyEncoding
  bodyUrl: string
  sizeBytes: number
  isBinary: boolean
}

export interface Example {
  id: string
  requestId: string
  name: string
  revision: string
  /** The request snapshot as a draft (id `''`), so the request editors can edit it. */
  request: RequestDraft
  response: ExampleResponse
}

function toExampleRequestDraft(name: string, dto: ExampleRequestDto): RequestDraft {
  return toRequestDraft({
    id: '',
    name,
    description: '',
    method: dto.method,
    url: dto.url,
    revision: '',
    params: dto.params,
    path_variables: dto.path_variables,
    headers: dto.headers,
    auto_headers: { disabled: [] },
    auth: dto.auth,
    body: dto.body,
    settings: { timeout_ms: 0, disable_cookie_jar: false },
  })
}

export function toExample(dto: ExampleDto): Example {
  return {
    id: dto.id,
    requestId: dto.request_id,
    name: dto.name,
    revision: dto.revision,
    request: toExampleRequestDraft(dto.name, dto.request),
    response: {
      code: dto.response.code,
      status: dto.response.status,
      contentType: dto.response.content_type,
      headers: (dto.response.headers ?? []).map(toKeyValueRow),
      body: dto.response.body ?? null,
      bodyEncoding: dto.response.body_encoding === 'base64' ? 'base64' : 'text',
      bodyUrl: dto.response.body_url,
      sizeBytes: dto.response.size_bytes,
      isBinary: dto.response.is_binary,
    },
  }
}

export function toExampleRequestDto(draft: RequestDraft): ExampleRequestDto {
  const request = toRequestDto(draft)
  return {
    method: request.method,
    url: request.url,
    params: request.params,
    path_variables: request.path_variables,
    headers: request.headers,
    body: request.body,
    auth: request.auth,
  }
}

/** A `null` body keeps the stored one (it was never loaded inline, e.g. large or binary). */
export function toExampleEditDto(example: Example): ExampleEditDto {
  return {
    id: example.id,
    name: example.name,
    request: toExampleRequestDto(example.request),
    response: {
      code: example.response.code,
      status: example.response.status,
      content_type: example.response.contentType,
      headers: example.response.headers.map(toKeyValueDto),
      body: example.response.body,
      body_encoding: example.response.bodyEncoding,
    },
  }
}
