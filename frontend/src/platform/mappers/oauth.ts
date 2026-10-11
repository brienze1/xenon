import type { TokenInfo as TokenInfoDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import type { TokenUse } from './request'

/** What the UI may know about an OAuth 2.0 token: never its value (Reveal and Copy only). */
export interface TokenInfo {
  ownerId: string
  tokenType: string
  scope: string
  expiresAt: string | null
  obtainedAt: string
  useToken: TokenUse
  hasRefreshToken: boolean
  hasIdToken: boolean
  expired: boolean
}

export type TokenPart = 'access' | 'refresh' | 'id'

export function toTokenInfo(dto: TokenInfoDto): TokenInfo {
  return {
    ownerId: dto.owner_id,
    tokenType: dto.token_type,
    scope: dto.scope,
    expiresAt: dto.expires_at ?? null,
    obtainedAt: dto.obtained_at,
    useToken: dto.use_token === 'id_token' ? 'id_token' : 'access_token',
    hasRefreshToken: dto.has_refresh_token,
    hasIdToken: dto.has_id_token,
    expired: dto.expired,
  }
}
