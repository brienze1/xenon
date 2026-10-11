import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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
const { useCookiesStore } = await import('@/features/cookies/store')
const { CookiesDialog } = await import('./CookiesDialog')
const { CookiesButton } = await import('./CookiesButton')

function cookie(name: string, domain = 'api.test'): Cookie {
  return { name, value: `${name}-value`, domain, path: '/', expires: null, sameSite: '', secure: false, httpOnly: false, hostOnly: true, raw: `${name}=${name}-value; Path=/` }
}

const jar: CookieDomain[] = [
  { domain: 'api.test', cookies: [cookie('session'), cookie('theme')] },
  { domain: 'other.test', cookies: [cookie('a', 'other.test')] },
]

async function openDialog() {
  render(
    <>
      <CookiesButton />
      <CookiesDialog />
    </>,
  )
  await userEvent.click(screen.getByRole('button', { name: 'Cookies' }))
  return screen.findByRole('dialog', { name: 'Cookies' })
}

describe('CookiesDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useCookiesStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
    vi.mocked(cookies.listCookies).mockResolvedValue(jar)
    vi.mocked(cookies.getCookieAllowlist).mockResolvedValue(['api.test'])
  })

  it('lists the domains with their cookie count and the first domain’s cookies', async () => {
    const dialog = await openDialog()
    const domains = within(dialog).getByRole('list', { name: 'Domains' })
    expect(within(domains).getByRole('button', { name: 'api.test 2 cookies' })).toHaveAttribute('aria-pressed', 'true')
    expect(within(domains).getByRole('button', { name: 'other.test 1 cookie' })).toHaveAttribute('aria-pressed', 'false')
    const shown = within(dialog).getByRole('region', { name: 'Cookies of api.test' })
    expect(within(shown).getByRole('button', { name: 'session' })).toBeInTheDocument()
    await userEvent.click(within(domains).getByRole('button', { name: 'other.test 1 cookie' }))
    expect(within(dialog).getByRole('region', { name: 'Cookies of other.test' })).toBeInTheDocument()
  })

  it('edits a cookie as a Set-Cookie line and shows refusals', async () => {
    const dialog = await openDialog()
    await userEvent.click(within(dialog).getByRole('button', { name: 'session' }))
    const field = within(dialog).getByRole('textbox', { name: 'Cookie' })
    expect(field).toHaveValue('session=session-value; Path=/')
    vi.mocked(cookies.upsertCookie).mockRejectedValueOnce(new AppError('CKE-00400', 'the cookie is not valid: http: blank cookie'))
    await userEvent.clear(field)
    await userEvent.type(field, 'oops')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('the cookie is not valid')
    vi.mocked(cookies.upsertCookie).mockResolvedValueOnce(cookie('session'))
    await userEvent.clear(field)
    await userEvent.type(field, 'session=2')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(cookies.upsertCookie).toHaveBeenLastCalledWith('L1', 'api.test', 'session=2', cookie('session'))
    expect(within(dialog).queryByRole('textbox', { name: 'Cookie' })).not.toBeInTheDocument()
  })

  it('adds a domain and a new cookie for it, or cancels the editor', async () => {
    const dialog = await openDialog()
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Domain' }), 'new.test')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Add Domain' }))
    expect(within(dialog).getByRole('button', { name: 'new.test 0 cookies' })).toHaveAttribute('aria-pressed', 'true')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Add Cookie' }))
    expect(within(dialog).getByRole('textbox', { name: 'Cookie' })).toHaveValue('Cookie_1=value; Path=/')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(within(dialog).queryByRole('textbox', { name: 'Cookie' })).not.toBeInTheDocument()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Add Cookie' }))
    vi.mocked(cookies.upsertCookie).mockResolvedValue(cookie('Cookie_1', 'new.test'))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(cookies.upsertCookie).toHaveBeenCalledWith('L1', 'new.test', 'Cookie_1=value; Path=/', undefined)
  })

  it('deletes a cookie, a domain and everything', async () => {
    const dialog = await openDialog()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Delete cookie theme' }))
    expect(cookies.deleteCookie).toHaveBeenCalledWith('L1', cookie('theme'))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Delete domain other.test' }))
    expect(cookies.deleteCookieDomain).toHaveBeenCalledWith('L1', 'other.test')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Clear All' }))
    expect(cookies.clearCookies).toHaveBeenCalledWith('L1')
  })

  it('says when the jar is empty and keeps Clear All disabled', async () => {
    vi.mocked(cookies.listCookies).mockResolvedValue([])
    const dialog = await openDialog()
    expect(within(dialog).getByText(/No cookies yet/)).toBeInTheDocument()
    expect(within(dialog).getByRole('button', { name: 'Clear All' })).toBeDisabled()
  })

  it('reports a jar that cannot be read', async () => {
    vi.mocked(cookies.listCookies).mockRejectedValue(new AppError('WKS-07404', 'the workspace is not registered'))
    const dialog = await openDialog()
    expect(within(dialog).getByRole('alert')).toHaveTextContent('the workspace is not registered')
  })

  it('manages the script allowlist', async () => {
    const dialog = await openDialog()
    await userEvent.click(within(dialog).getByRole('tab', { name: 'Allowlist' }))
    expect(within(dialog).queryByRole('button', { name: 'Clear All' })).not.toBeInTheDocument()
    const allowed = within(dialog).getByRole('list', { name: 'Allowed domains' })
    expect(within(allowed).getByText('api.test')).toBeInTheDocument()
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Allowed domain' }), 'Example.TEST')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Add' }))
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Allowed domain' }), 'example.test')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Add' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove api.test' }))
    vi.mocked(cookies.saveCookieAllowlist).mockResolvedValueOnce(['example.test'])
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Allowlist' }))
    expect(cookies.saveCookieAllowlist).toHaveBeenCalledWith('L1', ['example.test'])
    await userEvent.click(within(dialog).getByRole('button', { name: 'Remove example.test' }))
    expect(within(dialog).getByText('No domain is allowed yet.')).toBeInTheDocument()
    vi.mocked(cookies.saveCookieAllowlist).mockRejectedValueOnce(new AppError('UIS-11500', 'the state could not be written'))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Allowlist' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('the state could not be written')
  })

  it('closes with the Close button', async () => {
    const dialog = await openDialog()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Close' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    act(() => useCookiesStore.setState({ open: true, loading: true }))
    expect(screen.getAllByRole('status')[0]).toHaveTextContent('Loading cookies…')
  })
})
