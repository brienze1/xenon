import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
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
const { TokenPanel } = await import('./TokenPanel')
const { useAuthStore } = await import('@/features/auth/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

const info: TokenInfo = { ownerId: 'c1', tokenType: 'Bearer', scope: 'openid api', expiresAt: '2026-01-01T10:00:00Z', obtainedAt: '', useToken: 'access_token', hasRefreshToken: true, hasIdToken: true, expired: true }
const oauth2 = emptyAuth().oauth2

describe('TokenPanel', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useAuthStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  it('shows the masked token with its expiry and scopes, and acts on it', async () => {
    const user = userEvent.setup()
    vi.mocked(oauth.getToken).mockResolvedValue(info)
    render(<TokenPanel ownerId="c1" oauth2={oauth2} />)
    expect(await screen.findByText('•••••')).toBeInTheDocument()
    expect(oauth.getToken).toHaveBeenCalledWith('ws', 'c1')
    expect(screen.getByText('openid api')).toBeInTheDocument()
    expect(screen.getByText('Expired')).toBeInTheDocument()
    vi.mocked(oauth.revealToken).mockResolvedValue('access-1')
    await user.click(screen.getByRole('button', { name: 'Reveal' }))
    expect(await screen.findByText('access-1')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Hide' }))
    expect(screen.getByText('•••••')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Copy' }))
    expect(await screen.findByRole('button', { name: 'Copied' })).toBeInTheDocument()
    vi.mocked(oauth.refreshToken).mockResolvedValue({ ...info, expired: false })
    await user.click(screen.getByRole('button', { name: 'Refresh' }))
    expect(oauth.refreshToken).toHaveBeenCalledWith('ws', 'c1', oauth2)
    await vi.waitFor(() => expect(screen.queryByText('Expired')).not.toBeInTheDocument())
    vi.mocked(oauth.chooseToken).mockResolvedValue({ ...info, useToken: 'id_token' })
    await user.selectOptions(screen.getByRole('combobox', { name: 'Use' }), 'id_token')
    expect(oauth.chooseToken).toHaveBeenCalledWith('ws', 'c1', 'id_token')
    await user.click(screen.getByRole('button', { name: 'Clear' }))
    expect(await screen.findByText(/No token yet/)).toBeInTheDocument()
  })

  it('waits for the browser sign-in with Cancel, then shows provider errors', async () => {
    const user = userEvent.setup()
    vi.mocked(oauth.getToken).mockResolvedValue(null)
    let fail: (error: unknown) => void = () => undefined
    const cancel = vi.fn()
    vi.mocked(oauth.getNewToken).mockReturnValue({ result: new Promise((_, reject) => (fail = reject)), cancel })
    render(<TokenPanel ownerId="c1" oauth2={oauth2} />)
    await user.click(await screen.findByRole('button', { name: 'Get New Access Token' }))
    expect(screen.getByText('Waiting for the sign-in in your browser…')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(cancel).toHaveBeenCalled()
    await act(async () => fail(new AppError('OAU-04401', 'the provider refused', [{ attribute: 'error', messages: ['access_denied'] }])))
    expect(await screen.findByRole('alert')).toHaveTextContent('the provider refused')
    expect(screen.getByRole('alert')).toHaveTextContent('error: access_denied')
    expect(screen.getByRole('button', { name: 'Get New Access Token' })).toBeEnabled()
  })

  it('explains that a mac token is not sent and disables Refresh without a refresh token', async () => {
    vi.mocked(oauth.getToken).mockResolvedValue({ ...info, tokenType: 'mac', hasRefreshToken: false, hasIdToken: false, expiresAt: null, scope: '', expired: false })
    render(<TokenPanel ownerId="c1" oauth2={oauth2} />)
    expect(await screen.findByText('A mac token is not sent with requests, as in Postman.')).toBeInTheDocument()
    expect(screen.getByText('No expiry')).toBeInTheDocument()
    expect(screen.getByText('none')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Refresh' })).toBeDisabled()
    expect(screen.queryByRole('combobox', { name: 'Use' })).not.toBeInTheDocument()
  })
})
