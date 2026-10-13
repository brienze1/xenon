import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth } from '@/platform/mappers/request'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/oauthservice', () => ({
  GetNewToken: vi.fn(),
  Refresh: vi.fn(),
  Get: vi.fn(),
  Reveal: vi.fn(),
  Copy: vi.fn(),
  Use: vi.fn(),
  Clear: vi.fn(),
}))

const OAuthService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/oauthservice')
const oauth = await import('./oauth')

const infoDto = { owner_id: 'c1', token_type: 'Bearer', scope: 'api', expires_at: '2026-01-01T00:00:00Z', obtained_at: '2025-12-31T23:00:00Z', use_token: 'access_token', has_refresh_token: true, has_id_token: false, expired: false }
const info = { ownerId: 'c1', tokenType: 'Bearer', scope: 'api', expiresAt: '2026-01-01T00:00:00Z', obtainedAt: '2025-12-31T23:00:00Z', useToken: 'access_token', hasRefreshToken: true, hasIdToken: false, expired: false }

describe('oauth service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('gets a new token with the draft configuration and can cancel the wait', async () => {
    const cancel = vi.fn()
    vi.mocked(OAuthService.GetNewToken).mockReturnValue(Object.assign(Promise.resolve(infoDto), { cancel }) as never)
    const oauth2 = { ...emptyAuth().oauth2, clientId: 'xenon' }
    const pending = oauth.getNewToken('ws', 'c1', oauth2)
    await expect(pending.result).resolves.toEqual(info)
    pending.cancel()
    expect(cancel).toHaveBeenCalled()
    expect(OAuthService.GetNewToken).toHaveBeenCalledWith(expect.objectContaining({ local_id: 'ws', owner_id: 'c1', auth: expect.objectContaining({ client_id: 'xenon', grant_type: 'authorization_code' }) }))
    vi.mocked(OAuthService.GetNewToken).mockReturnValue(Object.assign(Promise.resolve(null), { cancel }) as never)
    await expect(oauth.getNewToken('ws', 'c1', oauth2).result).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('reads, refreshes, chooses and clears tokens', async () => {
    vi.mocked(OAuthService.Get).mockResolvedValueOnce({ ...infoDto, expires_at: null, use_token: 'id_token' } as never)
    await expect(oauth.getToken('ws', 'c1')).resolves.toEqual({ ...info, expiresAt: null, useToken: 'id_token' })
    vi.mocked(OAuthService.Get).mockResolvedValueOnce(null as never)
    await expect(oauth.getToken('ws', 'c1')).resolves.toBeNull()
    vi.mocked(OAuthService.Refresh).mockResolvedValue(infoDto as never)
    await expect(oauth.refreshToken('ws', 'c1', emptyAuth().oauth2)).resolves.toEqual(info)
    vi.mocked(OAuthService.Use).mockResolvedValue(infoDto as never)
    await expect(oauth.chooseToken('ws', 'c1', 'id_token')).resolves.toEqual(info)
    expect(OAuthService.Use).toHaveBeenCalledWith({ local_id: 'ws', owner_id: 'c1', use_token: 'id_token' })
    vi.mocked(OAuthService.Clear).mockResolvedValue(undefined as never)
    await oauth.clearToken('ws', 'c1')
    expect(OAuthService.Clear).toHaveBeenCalledWith({ local_id: 'ws', owner_id: 'c1' })
  })

  it('reveals and copies a token part, mapping failures', async () => {
    vi.mocked(OAuthService.Reveal).mockResolvedValueOnce({ value: 'access-1' } as never)
    await expect(oauth.revealToken('ws', 'c1', 'access')).resolves.toBe('access-1')
    vi.mocked(OAuthService.Reveal).mockResolvedValueOnce(null as never)
    await expect(oauth.revealToken('ws', 'c1', 'refresh')).resolves.toBe('')
    vi.mocked(OAuthService.Copy).mockRejectedValue(Object.assign(new Error('x'), { cause: { error: { code: 'OAU-10404', description: 'none' } } }))
    await expect(oauth.copyToken('ws', 'c1', 'id')).rejects.toMatchObject({ code: 'OAU-10404' })
  })
})
