import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Cookie, CookieDomain } from '@/platform/mappers/cookie'

vi.mock('@/platform/services/cookies', () => ({
  listCookies: vi.fn(),
  upsertCookie: vi.fn(),
  deleteCookie: vi.fn(),
  deleteCookieDomain: vi.fn(),
  clearCookies: vi.fn(),
  getCookieAllowlist: vi.fn(),
  saveCookieAllowlist: vi.fn(),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const cookies = await import('@/platform/services/cookies')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useCookiesStore } = await import('./store')

function cookie(name: string, domain = 'api.test', path = '/'): Cookie {
  return { name, value: `${name}-value`, domain, path, expires: null, sameSite: '', secure: false, httpOnly: false, hostOnly: true, raw: `${name}=${name}-value; Path=${path}` }
}

const jar: CookieDomain[] = [
  { domain: 'api.test', cookies: [cookie('session'), cookie('theme')] },
  { domain: 'other.test', cookies: [cookie('a', 'other.test')] },
]

function state() {
  return useCookiesStore.getState()
}

describe('cookies store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useCookiesStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
    vi.mocked(cookies.listCookies).mockResolvedValue(jar)
    vi.mocked(cookies.getCookieAllowlist).mockResolvedValue(['api.test'])
  })

  it('opens on the first domain with the jar and the allowlist loaded', async () => {
    await state().openDialog()
    expect(state()).toMatchObject({ open: true, loading: false, view: 'cookies', domains: jar, allowlist: ['api.test'], selectedDomain: 'api.test', error: null })
    expect(cookies.listCookies).toHaveBeenCalledWith('L1')
    state().setView('allowlist')
    expect(state().view).toBe('allowlist')
    state().close()
    expect(state()).toMatchObject({ open: false, domains: [], view: 'cookies' })
  })

  it('reports a jar that cannot be loaded', async () => {
    vi.mocked(cookies.listCookies).mockRejectedValue(new AppError('WKS-07404', 'unknown workspace'))
    await state().openDialog()
    expect(state()).toMatchObject({ loading: false, error: { code: 'WKS-07404' } })
  })

  it('does nothing without an active workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().openDialog()
    await expect(state().saveCookie('api.test', 'a=1')).resolves.toBeNull()
    expect(cookies.listCookies).not.toHaveBeenCalled()
    expect(cookies.upsertCookie).not.toHaveBeenCalled()
  })

  it('adds domains until a cookie is saved for them, normalised and once', async () => {
    await state().openDialog()
    state().addDomain('  .Example.TEST ')
    state().addDomain('example.test')
    state().addDomain('api.test')
    state().addDomain('   ')
    expect(state()).toMatchObject({ addedDomains: ['example.test'], selectedDomain: 'api.test' })
    state().selectDomain('example.test')
    vi.mocked(cookies.upsertCookie).mockResolvedValue(cookie('fresh', 'example.test'))
    vi.mocked(cookies.listCookies).mockResolvedValue([...jar, { domain: 'example.test', cookies: [cookie('fresh', 'example.test')] }])
    await expect(state().saveCookie('example.test', 'fresh=1')).resolves.toBeNull()
    expect(cookies.upsertCookie).toHaveBeenCalledWith('L1', 'example.test', 'fresh=1', undefined)
    expect(state()).toMatchObject({ addedDomains: [], selectedDomain: 'example.test' })
    expect(state().domains).toHaveLength(3)
  })

  it('saves an edited cookie in place of the old one and returns refusals', async () => {
    await state().openDialog()
    const edited = cookie('session')
    await state().saveCookie('api.test', 'session=2', edited)
    expect(cookies.upsertCookie).toHaveBeenCalledWith('L1', 'api.test', 'session=2', edited)
    vi.mocked(cookies.upsertCookie).mockRejectedValue(new AppError('CKE-00400', 'the cookie is not valid'))
    await expect(state().saveCookie('api.test', 'junk')).resolves.toMatchObject({ code: 'CKE-00400' })
  })

  it('deletes cookies, domains and the whole jar', async () => {
    await state().openDialog()
    await state().removeCookie(cookie('theme'))
    expect(cookies.deleteCookie).toHaveBeenCalledWith('L1', cookie('theme'))
    await state().removeDomain('api.test')
    expect(cookies.deleteCookieDomain).toHaveBeenCalledWith('L1', 'api.test')
    expect(state().selectedDomain).toBe('api.test')
    state().addDomain('empty.test')
    await state().removeDomain('empty.test')
    expect(cookies.deleteCookieDomain).toHaveBeenCalledTimes(1)
    expect(state()).toMatchObject({ addedDomains: [], selectedDomain: null })
    vi.mocked(cookies.clearCookies).mockRejectedValue(new AppError('CKE-05400', 'nope'))
    await state().clearAll()
    expect(cookies.clearCookies).toHaveBeenCalledWith('L1')
    expect(state().error).toMatchObject({ code: 'CKE-05400' })
  })

  it('saves the allowlist as Go normalised it', async () => {
    await state().openDialog()
    vi.mocked(cookies.saveCookieAllowlist).mockResolvedValue(['example.test'])
    await expect(state().saveAllowlist([' Example.test '])).resolves.toBeNull()
    expect(state().allowlist).toEqual(['example.test'])
    vi.mocked(cookies.saveCookieAllowlist).mockRejectedValue(new AppError('UIS-11500', 'disk full'))
    await expect(state().saveAllowlist([])).resolves.toMatchObject({ code: 'UIS-11500' })
  })

  it('reloads on jar changes of the shown workspace only', async () => {
    state().jarChanged('L1')
    expect(cookies.listCookies).not.toHaveBeenCalled()
    await state().openDialog()
    vi.mocked(cookies.listCookies).mockClear()
    state().jarChanged('L2')
    expect(cookies.listCookies).not.toHaveBeenCalled()
    state().jarChanged('L1')
    expect(cookies.listCookies).toHaveBeenCalledWith('L1')
  })

  it('ignores answers for a workspace that is no longer active', async () => {
    vi.mocked(cookies.listCookies).mockImplementation(async () => {
      useWorkspacesStore.setState({ activeLocalId: 'L2' })
      return jar
    })
    await state().openDialog()
    expect(state().domains).toEqual([])
  })
})
