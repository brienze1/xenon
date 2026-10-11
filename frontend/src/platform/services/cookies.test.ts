import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/cookieservice', () => ({
  List: vi.fn(),
  Upsert: vi.fn(),
  Delete: vi.fn(),
  DeleteDomain: vi.fn(),
  Clear: vi.fn(),
  GetAllowlist: vi.fn(),
  SaveAllowlist: vi.fn(),
}))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/treeservice', () => ({ Move: vi.fn() }))

const CookieService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/cookieservice')
const TreeService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/treeservice')
const cookies = await import('./cookies')
const tree = await import('./tree')

const sessionDto = { name: 'session', value: 'abc', domain: 'api.test', path: '/', expires: '2037-10-21T07:28:00Z', same_site: 'Lax', secure: true, http_only: true, host_only: false, raw: 'session=abc' }
const session = { name: 'session', value: 'abc', domain: 'api.test', path: '/', expires: '2037-10-21T07:28:00Z', sameSite: 'Lax', secure: true, httpOnly: true, hostOnly: false, raw: 'session=abc' }

describe('cookies service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists the jar by domain', async () => {
    vi.mocked(CookieService.List).mockResolvedValueOnce([{ domain: 'api.test', cookies: [sessionDto] }, { domain: 'x.test', cookies: null }] as never)
    await expect(cookies.listCookies('L1')).resolves.toEqual([{ domain: 'api.test', cookies: [session] }, { domain: 'x.test', cookies: [] }])
    expect(CookieService.List).toHaveBeenCalledWith({ local_id: 'L1' })
    vi.mocked(CookieService.List).mockResolvedValueOnce(null as never)
    await expect(cookies.listCookies('L1')).resolves.toEqual([])
  })

  it('upserts a Set-Cookie line, replacing the edited cookie', async () => {
    vi.mocked(CookieService.Upsert).mockResolvedValue({ ...sessionDto, expires: '' } as never)
    await expect(cookies.upsertCookie('L1', 'api.test', 'session=abc')).resolves.toEqual({ ...session, expires: null })
    expect(CookieService.Upsert).toHaveBeenLastCalledWith({ local_id: 'L1', domain: 'api.test', raw: 'session=abc', replaces: null })
    await cookies.upsertCookie('L1', 'api.test', 'other=1', { domain: 'api.test', path: '/', name: 'session' })
    expect(CookieService.Upsert).toHaveBeenLastCalledWith({ local_id: 'L1', domain: 'api.test', raw: 'other=1', replaces: { domain: 'api.test', path: '/', name: 'session' } })
    vi.mocked(CookieService.Upsert).mockResolvedValue(null as never)
    await expect(cookies.upsertCookie('L1', 'api.test', 'a=1')).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('deletes cookies, domains and the jar and maps errors', async () => {
    await cookies.deleteCookie('L1', { domain: 'api.test', path: '/', name: 'session' })
    expect(CookieService.Delete).toHaveBeenCalledWith({ local_id: 'L1', domain: 'api.test', path: '/', name: 'session' })
    await cookies.deleteCookieDomain('L1', 'api.test')
    expect(CookieService.DeleteDomain).toHaveBeenCalledWith({ local_id: 'L1', domain: 'api.test' })
    vi.mocked(CookieService.Clear).mockRejectedValue({ cause: { error: { code: 'CKE-05400', description: 'invalid' } } })
    await expect(cookies.clearCookies('L1')).rejects.toMatchObject({ code: 'CKE-05400' })
  })

  it('reads and saves the allowlist', async () => {
    vi.mocked(CookieService.GetAllowlist).mockResolvedValue({ domains: null } as never)
    await expect(cookies.getCookieAllowlist('L1')).resolves.toEqual([])
    vi.mocked(CookieService.SaveAllowlist).mockResolvedValue({ domains: ['api.test'] } as never)
    await expect(cookies.saveCookieAllowlist('L1', ['API.test'])).resolves.toEqual(['api.test'])
    expect(CookieService.SaveAllowlist).toHaveBeenCalledWith({ local_id: 'L1', domains: ['API.test'] })
  })
})

describe('tree moves', () => {
  it('moves an item and returns the new tree', async () => {
    vi.mocked(TreeService.Move).mockResolvedValueOnce([{ id: 'c1', kind: 'collection', name: 'C', method: '', url: '', error: '', warning: '', children: null, examples: null }] as never)
    await expect(tree.moveItem('L1', 'r1', 'c1', 2)).resolves.toMatchObject([{ id: 'c1', kind: 'collection', children: [] }])
    expect(TreeService.Move).toHaveBeenCalledWith({ local_id: 'L1', id: 'r1', new_parent_id: 'c1', index: 2 })
    vi.mocked(TreeService.Move).mockResolvedValueOnce(null as never)
    await expect(tree.moveItem('L1', 'r1', 'c1', 2)).resolves.toEqual([])
  })
})
