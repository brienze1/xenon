import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth, type OAuth2Auth } from '@/platform/mappers/request'

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
const { OAuth2Form } = await import('./OAuth2Form')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

let latest: OAuth2Auth = emptyAuth().oauth2

function Harness({ initial }: { initial: OAuth2Auth }) {
  const [oauth2, setOAuth2] = useState(initial)
  return (
    <OAuth2Form
      oauth2={oauth2}
      ownerId="c1"
      onChange={(next) => {
        latest = next
        setOAuth2(next)
      }}
    />
  )
}

describe('OAuth2Form', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(oauth.getToken).mockResolvedValue(null)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  it('shows the fields of each grant and flags the legacy ones', async () => {
    const user = userEvent.setup()
    render(<Harness initial={emptyAuth().oauth2} />)
    expect(screen.getByRole('textbox', { name: 'Callback URL' })).toHaveValue('http://127.0.0.1:53682/callback')
    expect(screen.getByRole('textbox', { name: 'Auth URL' })).toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: 'Code verifier' })).not.toBeInTheDocument()
    const grant = screen.getByRole('combobox', { name: /Grant type/ })
    await user.selectOptions(grant, 'authorization_code_with_pkce')
    expect(screen.getByRole('textbox', { name: 'Code verifier' })).toBeInTheDocument()
    await user.selectOptions(screen.getByRole('combobox', { name: 'Code challenge method' }), 'plain')
    await user.selectOptions(grant, 'implicit')
    expect(screen.getByText('Legacy')).toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: 'Access token URL' })).not.toBeInTheDocument()
    await user.selectOptions(grant, 'password_credentials')
    expect(screen.getByText('Legacy')).toBeInTheDocument()
    await user.type(screen.getByRole('textbox', { name: 'Username' }), 'ana')
    await user.selectOptions(grant, 'client_credentials')
    expect(screen.queryByText('Legacy')).not.toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: 'Callback URL' })).not.toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: 'State' })).not.toBeInTheDocument()
    expect(latest).toMatchObject({ grantType: 'client_credentials', challengeAlgorithm: 'plain', username: 'ana' })
  })

  it('edits the token request settings and parameters', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ ...emptyAuth().oauth2, grantType: 'client_credentials' }} />)
    await user.type(screen.getByRole('textbox', { name: 'Access token URL' }), 'https://sso/token')
    await user.type(screen.getByRole('textbox', { name: 'Client ID' }), 'xenon')
    await user.type(screen.getByLabelText('Client secret'), '{{{{secret}}')
    await user.type(screen.getByRole('textbox', { name: 'Scope' }), 'api')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Client authentication' }), 'body')
    await user.clear(screen.getByRole('textbox', { name: 'Header prefix' }))
    await user.type(screen.getByRole('textbox', { name: 'Header prefix' }), 'Token')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Add auth data to' }), 'queryParams')
    expect(screen.queryByRole('textbox', { name: 'Header prefix' })).not.toBeInTheDocument()
    await user.click(screen.getByRole('checkbox', { name: 'Refresh the token automatically when it expires' }))
    await user.click(screen.getByText('Advanced'))
    await user.type(screen.getByRole('textbox', { name: 'Refresh token URL' }), 'https://sso/refresh')
    await user.type(screen.getByRole('textbox', { name: 'Audience' }), 'aud')
    await user.type(screen.getByRole('textbox', { name: 'Resource' }), 'res')
    const token = screen.getByRole('group', { name: 'Token request' })
    await user.click(within(token).getByRole('button', { name: 'Add Parameter' }))
    await user.type(within(token).getByRole('textbox', { name: 'Token request key 1' }), 'tenant')
    await user.type(within(token).getByRole('textbox', { name: 'Token request value 1' }), 'acme')
    await user.selectOptions(within(token).getByRole('combobox', { name: 'Token request send as 1' }), 'request_header')
    await user.click(within(token).getByRole('checkbox', { name: 'Enable Token request row 1' }))
    const refresh = screen.getByRole('group', { name: 'Refresh request' })
    await user.click(within(refresh).getByRole('button', { name: 'Add Parameter' }))
    await user.click(within(refresh).getByRole('button', { name: 'Remove Refresh request row 1' }))
    expect(latest).toMatchObject({
      accessTokenUrl: 'https://sso/token',
      clientId: 'xenon',
      clientSecret: '{{secret}}',
      scope: 'api',
      clientAuthentication: 'body',
      headerPrefix: 'Token',
      addTokenTo: 'queryParams',
      autoRefresh: false,
      refreshTokenUrl: 'https://sso/refresh',
      audience: 'aud',
      resource: 'res',
      tokenRequestParams: [{ key: 'tenant', value: 'acme', sendAs: 'request_header', enabled: false }],
      refreshRequestParams: [],
    })
    expect(screen.queryByRole('group', { name: 'Auth request' })).not.toBeInTheDocument()
  })

  it('explains Postman\'s callback and switches to the loopback URL', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ ...emptyAuth().oauth2, redirectUri: 'https://oauth.pstmn.io/v1/callback', usesPostmanCallback: true, suggestedCallback: 'http://127.0.0.1:53682/callback' }} />)
    expect(screen.getByRole('status')).toHaveTextContent("Postman's web callback cannot reach Xenon. Register this callback URL with your provider: http://127.0.0.1:53682/callback")
    await user.click(screen.getByRole('button', { name: 'Use This Callback' }))
    expect(screen.getByRole('textbox', { name: 'Callback URL' })).toHaveValue('http://127.0.0.1:53682/callback')
    expect(screen.queryByText(/Postman's web callback/)).not.toBeInTheDocument()
    await user.type(screen.getByRole('textbox', { name: 'State' }), 's')
    const auth = screen.getByRole('group', { name: 'Auth request' })
    await user.click(within(auth).getByRole('button', { name: 'Add Parameter' }))
    expect(latest.authRequestParams).toEqual([{ key: '', value: '', enabled: true, sendAs: 'request_url' }])
    expect(latest.state).toBe('s')
  })
})
