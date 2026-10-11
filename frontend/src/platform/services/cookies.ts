import * as CookieService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/cookieservice'
import { AppError, toAppError } from '../errors'
import { toCookie, toCookieDomain, type Cookie, type CookieDomain, type CookieKey } from '../mappers/cookie'

export const INVALID_COOKIE = 'CKE-00400'

const EMPTY_RESPONSE = 'UNK-00204'

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

function required<T>(value: T | null): T {
  if (value === null) {
    throw new AppError(EMPTY_RESPONSE, 'The backend returned no data')
  }
  return value
}

export function listCookies(localId: string): Promise<CookieDomain[]> {
  return call(async () => ((await CookieService.List({ local_id: localId })) ?? []).map(toCookieDomain))
}

/** Adds or replaces the cookie written as a Set-Cookie line for the domain; `replaces` is the cookie being edited. */
export function upsertCookie(localId: string, domain: string, raw: string, replaces?: CookieKey): Promise<Cookie> {
  return call(async () => toCookie(required(await CookieService.Upsert({ local_id: localId, domain, raw, replaces: replaces ?? null }))))
}

export function deleteCookie(localId: string, cookie: CookieKey): Promise<void> {
  return call(() => CookieService.Delete({ local_id: localId, domain: cookie.domain, path: cookie.path, name: cookie.name }))
}

export function deleteCookieDomain(localId: string, domain: string): Promise<void> {
  return call(() => CookieService.DeleteDomain({ local_id: localId, domain }))
}

export function clearCookies(localId: string): Promise<void> {
  return call(() => CookieService.Clear({ local_id: localId }))
}

export function getCookieAllowlist(localId: string): Promise<string[]> {
  return call(async () => required(await CookieService.GetAllowlist({ local_id: localId })).domains ?? [])
}

export function saveCookieAllowlist(localId: string, domains: string[]): Promise<string[]> {
  return call(async () => required(await CookieService.SaveAllowlist({ local_id: localId, domains })).domains ?? [])
}
