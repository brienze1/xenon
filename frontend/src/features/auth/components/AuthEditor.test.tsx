import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth, type RequestAuth } from '@/platform/mappers/request'

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
const { AuthEditor } = await import('./AuthEditor')
const { useWorkspacesStore } = await import('@/features/workspaces/store')

let latest: RequestAuth = emptyAuth()

interface HarnessProps {
  initial: RequestAuth
  kind?: 'request' | 'collection' | 'folder'
}

function Harness({ initial, kind = 'request' }: HarnessProps) {
  const [auth, setAuth] = useState(initial)
  return (
    <AuthEditor
      auth={auth}
      ownerId="owner-1"
      kind={kind}
      onChange={(next) => {
        latest = next
        setAuth(next)
      }}
    />
  )
}

describe('AuthEditor', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(oauth.getToken).mockResolvedValue(null)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    latest = emptyAuth()
  })

  it('offers every auth type and explains inherit per owner kind', async () => {
    const user = userEvent.setup()
    const { rerender } = render(<Harness initial={emptyAuth()} />)
    expect(screen.getByText('The request uses the auth of its folder or collection. An unsaved request has none.')).toBeInTheDocument()
    for (const name of ['API key', 'OAuth 2.0', 'AWS signature', 'Digest auth']) {
      expect(screen.getByRole('option', { name })).toBeEnabled()
    }
    await user.selectOptions(screen.getByRole('combobox', { name: 'Auth type' }), 'none')
    expect(screen.getByText('No authorization is sent.')).toBeInTheDocument()
    rerender(<Harness key="c" initial={emptyAuth()} kind="collection" />)
    expect(screen.getByText('A collection has no parent, so its requests use no auth.')).toBeInTheDocument()
    rerender(<Harness key="f" initial={emptyAuth()} kind="folder" />)
    expect(screen.getByText('Uses the auth of the folder or collection above.')).toBeInTheDocument()
  })

  it('edits bearer, basic and API key auth', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ ...emptyAuth(), type: 'bearer' }} />)
    await user.type(screen.getByRole('textbox', { name: 'Token' }), 'abc')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Auth type' }), 'basic')
    await user.type(screen.getByRole('textbox', { name: 'Username' }), 'ana')
    await user.type(screen.getByLabelText('Password'), 'pw')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Auth type' }), 'apikey')
    await user.type(screen.getByRole('textbox', { name: 'Key' }), 'X-Key')
    await user.type(screen.getByRole('textbox', { name: 'Value' }), 'k3y')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Add to' }), 'query')
    expect(latest).toMatchObject({ type: 'apikey', bearer: { token: 'abc' }, basic: { username: 'ana', password: 'pw' }, apikey: { key: 'X-Key', value: 'k3y', in: 'query' } })
  })

  it('edits AWS signature with its advanced fields', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ ...emptyAuth(), type: 'awsv4' }} />)
    await user.type(screen.getByRole('textbox', { name: 'Access key' }), 'AKID')
    await user.type(screen.getByLabelText('Secret key'), 's')
    await user.click(screen.getByRole('checkbox', { name: 'Add auth data to the query (presigned URL)' }))
    await user.click(screen.getByText('Advanced'))
    await user.type(screen.getByRole('textbox', { name: 'AWS region' }), 'eu-west-1')
    await user.type(screen.getByRole('textbox', { name: 'Service name' }), 's3')
    await user.type(screen.getByLabelText('Session token'), 't')
    expect(latest.awsv4).toEqual({ accessKey: 'AKID', secretKey: 's', sessionToken: 't', region: 'eu-west-1', service: 's3', addAuthDataToQuery: true })
    expect(screen.getByText(/Form-data bodies are signed as an empty payload/)).toBeInTheDocument()
  })

  it('edits Digest auth with the challenge fields under Advanced', async () => {
    const user = userEvent.setup()
    render(<Harness initial={{ ...emptyAuth(), type: 'digest' }} />)
    await user.type(screen.getByRole('textbox', { name: 'Username' }), 'ana')
    await user.type(screen.getByLabelText('Password'), 'pw')
    await user.click(screen.getByRole('checkbox', { name: /Disable retrying the request/ }))
    expect(screen.getByRole('textbox', { name: 'Realm' })).not.toBeVisible()
    await user.click(screen.getByText('Advanced'))
    await user.type(screen.getByRole('textbox', { name: 'Realm' }), 'r')
    await user.type(screen.getByRole('textbox', { name: 'Nonce' }), 'n')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Algorithm' }), 'SHA-256-sess')
    await user.selectOptions(screen.getByRole('combobox', { name: 'qop' }), 'auth-int')
    await user.type(screen.getByRole('textbox', { name: 'Nonce count' }), '00000001')
    await user.type(screen.getByRole('textbox', { name: 'Client nonce' }), 'c')
    await user.type(screen.getByRole('textbox', { name: 'Opaque' }), 'o')
    expect(latest.digest).toEqual({ username: 'ana', password: 'pw', algorithm: 'SHA-256-sess', qop: 'auth-int', realm: 'r', nonce: 'n', opaque: 'o', nonceCount: '00000001', clientNonce: 'c', disableRetryRequest: true })
  })

  it('keeps an unsupported type read-only', () => {
    render(<Harness initial={{ ...emptyAuth(), type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } }} />)
    expect(screen.getByRole('combobox', { name: 'Auth type' })).toHaveValue('unsupported')
    expect(screen.getByRole('option', { name: 'Unsupported (kept)' })).toBeDisabled()
    expect(screen.getByText(/Unsupported auth type \(kept\):/)).toHaveTextContent('Unsupported auth type (kept): hawk.')
  })
})
