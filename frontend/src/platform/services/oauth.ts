import * as OAuthService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/oauthservice'
import { AppError, toAppError } from '../errors'
import { toTokenInfo, type TokenInfo, type TokenPart } from '../mappers/oauth'
import { toOAuth2Dto, type OAuth2Auth, type TokenUse } from '../mappers/request'

export const INVALID_OAUTH_CONFIG = 'OAU-01400'
export const CALLBACK_PORT_BUSY = 'OAU-02409'
export const SIGN_IN_TIMED_OUT = 'OAU-03408'
export const PROVIDER_ERROR = 'OAU-04401'
export const STATE_MISMATCH = 'OAU-05400'
export const NO_REFRESH_TOKEN = 'OAU-06404'

const EMPTY_TOKEN = 'UNK-00204'

/** A "Get new access token" that may wait for the browser; `cancel` stops waiting at once. */
export interface PendingToken {
  result: Promise<TokenInfo>
  cancel: () => void
}

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_TOKEN, 'The backend returned no token')
  }
  return value
}

export function getNewToken(localId: string, ownerId: string, oauth2: OAuth2Auth): PendingToken {
  const pending = OAuthService.GetNewToken({ local_id: localId, owner_id: ownerId, auth: toOAuth2Dto(oauth2) })
  return { result: call(async () => toTokenInfo(required(await pending))), cancel: () => pending.cancel() }
}

export function refreshToken(localId: string, ownerId: string, oauth2: OAuth2Auth): Promise<TokenInfo> {
  return call(async () => toTokenInfo(required(await OAuthService.Refresh({ local_id: localId, owner_id: ownerId, auth: toOAuth2Dto(oauth2) }))))
}

export function getToken(localId: string, ownerId: string): Promise<TokenInfo | null> {
  return call(async () => {
    const info = await OAuthService.Get({ local_id: localId, owner_id: ownerId })
    return info ? toTokenInfo(info) : null
  })
}

/** The only call that returns a token value: keep it out of stores' persisted state and logs. */
export function revealToken(localId: string, ownerId: string, which: TokenPart): Promise<string> {
  return call(async () => (await OAuthService.Reveal({ local_id: localId, owner_id: ownerId, which }))?.value ?? '')
}

/** Go writes the token to the clipboard itself, so it never crosses the bridge. */
export function copyToken(localId: string, ownerId: string, which: TokenPart): Promise<void> {
  return call(() => OAuthService.Copy({ local_id: localId, owner_id: ownerId, which }))
}

export function chooseToken(localId: string, ownerId: string, useTokenKind: TokenUse): Promise<TokenInfo> {
  return call(async () => toTokenInfo(required(await OAuthService.Use({ local_id: localId, owner_id: ownerId, use_token: useTokenKind }))))
}

export function clearToken(localId: string, ownerId: string): Promise<void> {
  return call(() => OAuthService.Clear({ local_id: localId, owner_id: ownerId }))
}
