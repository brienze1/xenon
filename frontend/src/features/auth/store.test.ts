import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { TokenInfo } from '@/platform/mappers/oauth'
import { emptyAuth } from '@/platform/mappers/request'

vi.mock('@/platform/services/oauth', () => ({
  getNewToken: vi.fn(),
  refreshToken: vi.fn(),
  getToken: vi.fn(),
  revealToken: vi.fn(),
  copyToken: vi.fn(),
  chooseToken: vi.fn(),
  clearToken: vi.fn(),
}))

const oauth = await import('@/platform/services/oauth')
const { useAuthStore } = await import('./store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

const info: TokenInfo = { ownerId: 'c1', tokenType: 'Bearer', scope: 'api', expiresAt: null, obtainedAt: '', useToken: 'access_token', hasRefreshToken: true, hasIdToken: true, expired: false }
const oauth2 = emptyAuth().oauth2

function token() {
  return useAuthStore.getState().tokens.c1!
}

describe('auth store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useAuthStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => useWorkspacesStore.setState({ activeLocalId: null }))

  it('loads, refreshes, chooses and clears the owner token', async () => {
    vi.mocked(oauth.getToken).mockResolvedValue(info)
    await useAuthStore.getState().load('c1')
    expect(token()).toMatchObject({ info, busy: false, error: null })
    vi.mocked(oauth.refreshToken).mockResolvedValue({ ...info, scope: 'new' })
    await useAuthStore.getState().refresh('c1', oauth2)
    expect(token().info?.scope).toBe('new')
    vi.mocked(oauth.chooseToken).mockResolvedValue({ ...info, useToken: 'id_token' })
    await useAuthStore.getState().use('c1', 'id_token')
    expect(oauth.chooseToken).toHaveBeenCalledWith('ws', 'c1', 'id_token')
    vi.mocked(oauth.revealToken).mockResolvedValue('id-1')
    await useAuthStore.getState().toggleReveal('c1')
    expect(oauth.revealToken).toHaveBeenCalledWith('ws', 'c1', 'id')
    expect(token().revealed).toBe('id-1')
    await useAuthStore.getState().toggleReveal('c1')
    expect(token().revealed).toBeNull()
    await useAuthStore.getState().copy('c1')
    expect(oauth.copyToken).toHaveBeenCalledWith('ws', 'c1', 'id')
    expect(token().copied).toBe(true)
    await useAuthStore.getState().clear('c1')
    expect(oauth.clearToken).toHaveBeenCalledWith('ws', 'c1')
    expect(token().info).toBeNull()
  })

  it('reveals the access token by default and keeps errors per owner', async () => {
    vi.mocked(oauth.revealToken).mockResolvedValue('access-1')
    await useAuthStore.getState().toggleReveal('c1')
    expect(oauth.revealToken).toHaveBeenCalledWith('ws', 'c1', 'access')
    vi.mocked(oauth.refreshToken).mockRejectedValue(new AppError('OAU-06404', 'no refresh token'))
    await useAuthStore.getState().refresh('c1', oauth2)
    expect(token()).toMatchObject({ busy: false, error: { code: 'OAU-06404' } })
    expect(useAuthStore.getState().tokens.other).toBeUndefined()
  })

  it('waits for the browser, can cancel, and ignores a cancelled wait', async () => {
    let finish: (value: TokenInfo) => void = () => undefined
    const cancel = vi.fn()
    vi.mocked(oauth.getNewToken).mockReturnValue({ result: new Promise((resolve) => (finish = resolve)), cancel })
    const getting = useAuthStore.getState().getNew('c1', oauth2)
    expect(token().waiting).toBe(true)
    await useAuthStore.getState().getNew('c1', oauth2)
    expect(oauth.getNewToken).toHaveBeenCalledTimes(1)
    useAuthStore.getState().cancel('c1')
    expect(cancel).toHaveBeenCalled()
    finish(info)
    await getting
    expect(token()).toMatchObject({ waiting: false, info })
    vi.mocked(oauth.getNewToken).mockReturnValue({ result: Promise.reject(new AppError('OAU-08499', 'cancelled')), cancel })
    await useAuthStore.getState().getNew('c1', oauth2)
    expect(token().error).toBeNull()
    vi.mocked(oauth.getNewToken).mockReturnValue({ result: Promise.reject(new AppError('OAU-04401', 'denied')), cancel })
    await useAuthStore.getState().getNew('c1', oauth2)
    expect(token().error?.code).toBe('OAU-04401')
  })

  it('does nothing without an active workspace and cancels waits on reset', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await useAuthStore.getState().load('c1')
    await useAuthStore.getState().getNew('c1', oauth2)
    expect(oauth.getToken).not.toHaveBeenCalled()
    expect(oauth.getNewToken).not.toHaveBeenCalled()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    const cancel = vi.fn()
    vi.mocked(oauth.getNewToken).mockReturnValue({ result: new Promise(() => undefined), cancel })
    void useAuthStore.getState().getNew('c1', oauth2)
    useAuthStore.getState().reset()
    expect(cancel).toHaveBeenCalled()
    expect(useAuthStore.getState().tokens).toEqual({})
  })
})
