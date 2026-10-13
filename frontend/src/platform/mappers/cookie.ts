import type {
  Cookie as CookieDto,
  CookieDomain as CookieDomainDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

/** A cookie in the workspace jar or set by a response; values are shown in clear, as in Postman's Cookies dialog. */
export interface Cookie {
  name: string
  value: string
  domain: string
  path: string
  /** RFC 3339, or null for a session cookie. */
  expires: string | null
  sameSite: string
  secure: boolean
  httpOnly: boolean
  hostOnly: boolean
  /** The cookie as a Set-Cookie line, for the editor. */
  raw: string
}

export interface CookieDomain {
  domain: string
  cookies: Cookie[]
}

export interface CookieKey {
  domain: string
  path: string
  name: string
}

export function toCookie(dto: CookieDto): Cookie {
  return {
    name: dto.name,
    value: dto.value,
    domain: dto.domain,
    path: dto.path,
    expires: dto.expires ? dto.expires : null,
    sameSite: dto.same_site,
    secure: dto.secure,
    httpOnly: dto.http_only,
    hostOnly: dto.host_only,
    raw: dto.raw,
  }
}

export function toCookieDomain(dto: CookieDomainDto): CookieDomain {
  return { domain: dto.domain, cookies: (dto.cookies ?? []).map(toCookie) }
}
