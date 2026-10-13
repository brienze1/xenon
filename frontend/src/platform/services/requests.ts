import * as RequestService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/requestservice'
import { AppError, toAppError } from '../errors'
import {
  toAutoHeader,
  toKeyValueDto,
  toKeyValueRow,
  toRequestDraft,
  toRequestDto,
  toSavedRequest,
  toSendResult,
  type AutoHeader,
  type KeyValueRow,
  type RequestDraft,
  type SavedRequest,
  type SendAcknowledgement,
  type SendResult,
} from '../mappers/request'

const EMPTY_RESPONSE = 'UNK-00204'
export const NO_RESPONSE = 'SND-03404'

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

export const CONFIRMATION_REQUIRED = 'SND-05409'

/** The environment a send must be confirmed for (`SND-05409`). */
export interface SendConfirmation {
  environmentId: string
  name: string
  color: string
}

/** Reads the environment a send preflight asks to confirm (`SND-05409`), or `null` for any other error. */
export function confirmationOf(error: AppError): SendConfirmation | null {
  if (error.code !== CONFIRMATION_REQUIRED) {
    return null
  }
  const first = (attribute: string) => error.details.find((detail) => detail.attribute === attribute)?.messages[0] ?? ''
  return { environmentId: first('environment_id'), name: first('name'), color: first('color') }
}

export const EXPIRED_TOKEN = 'SND-07409'

/** The auth owner whose OAuth 2.0 token expired and could not be refreshed (`SND-07409`), or `null` for any other error. */
export function expiredTokenOwnerOf(error: AppError): string | null {
  if (error.code !== EXPIRED_TOKEN) {
    return null
  }
  return error.details.find((detail) => detail.attribute === 'owner_id')?.messages[0] ?? ''
}

export interface PendingSend {
  result: Promise<SendResult>
  cancel: () => void
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getRequest(localId: string, id: string): Promise<SavedRequest> {
  return call(async () => toSavedRequest(required(await RequestService.Get({ local_id: localId, id }))))
}

export function saveRequest(localId: string, draft: RequestDraft, baseRevision: string, force: boolean): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(
      required(await RequestService.Save({ local_id: localId, base_revision: baseRevision, force, request: toRequestDto(draft) })),
    ),
  )
}

export function saveRequestAs(localId: string, parentId: string, name: string, draft: RequestDraft): Promise<SavedRequest> {
  return call(async () =>
    toSavedRequest(required(await RequestService.SaveAs({ local_id: localId, parent_id: parentId, name, request: toRequestDto(draft) }))),
  )
}

export function applyUrl(draft: RequestDraft): Promise<RequestDraft> {
  return call(async () => toRequestDraft(required(await RequestService.ApplyUrl({ request: toRequestDto(draft) }))))
}

export function applyParams(draft: RequestDraft): Promise<RequestDraft> {
  return call(async () => toRequestDraft(required(await RequestService.ApplyParams({ request: toRequestDto(draft) }))))
}

export function parseBulk(text: string): Promise<KeyValueRow[]> {
  return call(async () => (required(await RequestService.ParseBulk({ text })).rows ?? []).map(toKeyValueRow))
}

/** A request parsed from a pasted cURL command, plus what was ignored (unknown options, extra URLs). */
export interface ParsedCurl {
  draft: RequestDraft
  notices: string[]
}

export const NOT_CURL = 'IMP-06400'
export const CURL_NOT_HTTP = 'IMP-07400'
export const CURL_NO_URL = 'IMP-40400'
export const CURL_UNREADABLE = 'IMP-41400'

const CURL_REFUSALS: ReadonlySet<string> = new Set([NOT_CURL, CURL_NOT_HTTP, CURL_NO_URL, CURL_UNREADABLE])

/** Whether ParseCurl refused a pasted command: not cURL, no URL, unreadable or not an HTTP(S) URL. */
export function isCurlRefusal(error: AppError): boolean {
  return CURL_REFUSALS.has(error.code)
}

export function parseCurl(text: string): Promise<ParsedCurl> {
  return call(async () => {
    const parsed = required(await RequestService.ParseCurl({ text }))
    return { draft: toRequestDraft(parsed.request), notices: parsed.notices ?? [] }
  })
}

export function formatBulk(rows: KeyValueRow[]): Promise<string> {
  return call(async () => required(await RequestService.FormatBulk({ rows: rows.map(toKeyValueDto) })).text)
}

export function autoHeaders(localId: string, draft: RequestDraft): Promise<AutoHeader[]> {
  return call(async () => (await RequestService.AutoHeaders({ local_id: localId, request: toRequestDto(draft) }) ?? []).map(toAutoHeader))
}

export function sendRequest(localId: string, tabId: string, draft: RequestDraft, acknowledgement: SendAcknowledgement): PendingSend {
  const pending = RequestService.Send({
    local_id: localId,
    tab_id: tabId,
    request: toRequestDto(draft),
    acknowledge_missing_secrets: acknowledgement.missingSecrets,
    confirmed_environment_id: acknowledgement.confirmedEnvironmentId,
    send_with_expired_token: acknowledgement.expiredToken,
  })
  return {
    result: call(async () => toSendResult(required(await pending))),
    cancel: () => pending.cancel(),
  }
}

export async function lastResult(tabId: string): Promise<SendResult | null> {
  try {
    return toSendResult(required(await RequestService.LastResult({ tab_id: tabId })))
  } catch (error) {
    const appError = toAppError(error)
    if (appError.code === NO_RESPONSE) {
      return null
    }
    throw appError
  }
}

export function releaseResponse(tabId: string): Promise<void> {
  return call(() => RequestService.ReleaseResponse({ tab_id: tabId }))
}

export function saveResponseBody(tabId: string, path: string): Promise<void> {
  return call(() => RequestService.SaveResponseBody({ tab_id: tabId, path }))
}
