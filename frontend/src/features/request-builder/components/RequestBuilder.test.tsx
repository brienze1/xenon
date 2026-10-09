import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { RequestBuilder } = await import('./RequestBuilder')

function queryOf(url: string) {
  const query = url.split('?')[1] ?? ''
  return query
    .split('&')
    .filter(Boolean)
    .map((pair, index) => {
      const [key = '', value = ''] = pair.split('=')
      return { id: `derived-${index}`, key, value, description: '', enabled: true }
    })
}

function pathVariablesOf(url: string) {
  return [...url.split('?')[0]!.matchAll(/\/:([^/]+)/g)].map((match, index) => ({ id: `var-${index}`, key: match[1]!, value: '', description: '' }))
}

describe('RequestBuilder', () => {
  let tabId: string

  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([
      { key: 'Content-Type', value: 'application/json', disabled: false },
      { key: 'User-Agent', value: 'Xenon/0.1.0', disabled: false },
    ])
    vi.mocked(requests.applyUrl).mockImplementation(async (draft: RequestDraft) => ({ ...draft, params: queryOf(draft.url), pathVariables: pathVariablesOf(draft.url) }))
    vi.mocked(requests.applyParams).mockImplementation(async (draft: RequestDraft) => {
      const query = draft.params.filter((row) => row.enabled && row.key).map((row) => (row.value ? `${row.key}=${row.value}` : row.key)).join('&')
      return { ...draft, url: draft.url.split('?')[0] + (query ? `?${query}` : '') }
    })
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useResponseViewerStore.getState().reset()
    useRequestBuilderStore.getState().reset()
    tabId = 'tab-1'
    useRequestBuilderStore.getState().ensureTab(tabId)
  })

  afterEach(() => vi.useRealTimers())

  it('shows the sub-tabs in Postman order with Params first', () => {
    render(<RequestBuilder tabId={tabId} />)
    const tabs = within(screen.getByRole('tablist', { name: 'Request' })).getAllByRole('tab')
    expect(tabs.map((tab) => tab.textContent)).toEqual(['Params', 'Authorization', 'Headers', 'Body', 'Scripts', 'Settings'])
    expect(screen.getByRole('tab', { name: 'Params' })).toHaveAttribute('aria-selected', 'true')
  })

  it('keeps the params table and the URL in sync both ways', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test/users?page=2&limit=10')
    const table = await screen.findByRole('table', { name: 'Query params' })
    await within(table).findByDisplayValue('limit')
    expect(within(table).getByDisplayValue('page')).toBeInTheDocument()
    await user.click(within(table).getByRole('checkbox', { name: 'Enable limit' }))
    await vi.waitFor(() => expect(screen.getByRole('textbox', { name: 'Request URL' })).toHaveValue('https://api.test/users?page=2'))
  })

  it('lists path variables and stores their values', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test/users/:id')
    const value = await screen.findByRole('textbox', { name: 'Value of id' })
    await user.type(value, '42')
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.pathVariables[0]?.value).toBe('42')
  })

  it('turns Send into Cancel while the request is in flight', async () => {
    const user = userEvent.setup()
    const cancel = vi.fn()
    let reject: (error: unknown) => void = () => undefined
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise((_, fail) => (reject = fail)), cancel })
    cancel.mockImplementation(() => reject(new AppError('cancelled', 'The request was cancelled')))
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('button', { name: 'Send' }))
    const cancelButton = await screen.findByRole('button', { name: 'Cancel' })
    await user.click(cancelButton)
    expect(cancel).toHaveBeenCalled()
    expect(await screen.findByRole('button', { name: 'Send' })).toBeInTheDocument()
    expect(useResponseViewerStore.getState().entries[tabId]?.result?.error?.kind).toBe('cancelled')
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled()
  })

  it('sends with Enter in the URL field', async () => {
    const user = userEvent.setup()
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    render(<RequestBuilder tabId={tabId} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test{Enter}')
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', tabId, expect.objectContaining({ url: 'https://api.test' }))
  })

  it('picks standard and custom methods', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('button', { name: 'Method GET' }))
    await user.click(screen.getByRole('menuitemradio', { name: 'POST' }))
    expect(screen.getByRole('button', { name: 'Method POST' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Method POST' }))
    await user.type(screen.getByRole('textbox', { name: 'Custom method' }), 'purge{Enter}')
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.method).toBe('PURGE')
    await user.type(screen.getByRole('textbox', { name: 'Custom method' }), 'bad method{Enter}')
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.method).toBe('PURGE')
  })

  it('lists auto headers greyed with a switch each', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('tab', { name: 'Headers' }))
    await act(() => useRequestBuilderStore.getState().refreshAutoHeaders(tabId))
    const auto = screen.getByRole('table', { name: 'Auto-generated headers' })
    expect(within(auto).getByText('Xenon/0.1.0')).toBeInTheDocument()
    await user.click(within(auto).getByRole('checkbox', { name: 'Send User-Agent' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.disabledAutoHeaders).toEqual(['User-Agent'])
  })

  it('edits headers in bulk mode', async () => {
    const user = userEvent.setup()
    vi.mocked(requests.formatBulk).mockResolvedValue('')
    vi.mocked(requests.parseBulk).mockResolvedValue([{ id: 'x', key: 'X-Token', value: '{{token}}', description: '', enabled: true }])
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('tab', { name: 'Headers' }))
    await user.click(screen.getByRole('button', { name: 'Bulk Edit' }))
    await user.type(await screen.findByRole('textbox', { name: 'Headers bulk edit' }), 'X-Token:{{{{token}}')
    await vi.waitFor(() => expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.headers[0]?.key).toBe('X-Token'))
    expect(screen.getByRole('tab', { name: 'Headers (1)' })).toBeInTheDocument()
  })

  it('offers inherit, none, bearer and basic auth, with later types disabled', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('tab', { name: 'Authorization' }))
    const type = screen.getByRole('combobox', { name: 'Auth Type' })
    expect(screen.getByRole('option', { name: 'API Key (later stage)' })).toBeDisabled()
    await user.selectOptions(type, 'bearer')
    await user.type(screen.getByRole('textbox', { name: 'Token' }), 'abc')
    await user.selectOptions(type, 'basic')
    await user.type(screen.getByRole('textbox', { name: 'Username' }), 'user')
    await user.type(screen.getByLabelText('Password'), 'pass')
    await user.selectOptions(type, 'none')
    expect(screen.getByText('This request does not use any authorization.')).toBeInTheDocument()
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.auth).toEqual({ type: 'none', bearer: { token: 'abc' }, basic: { username: 'user', password: 'pass' } })
  })

  it('edits every body mode and marks a missing file', async () => {
    const user = userEvent.setup()
    vi.mocked(dialogs.pickFile).mockResolvedValue('/files/payload.bin')
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('tab', { name: 'Body' }))
    expect(screen.getByText('This request does not have a body.')).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'raw' }))
    await user.selectOptions(screen.getByRole('combobox', { name: 'Raw language' }), 'xml')
    expect(screen.getByRole('textbox', { name: 'Raw body' })).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'binary' }))
    await user.click(screen.getByRole('button', { name: 'Select File' }))
    expect(await screen.findByText('/files/payload.bin')).toBeInTheDocument()
    act(() =>
      useResponseViewerStore.getState().setResult(tabId, {
        ...emptyResult(tabId),
        error: { kind: 'missing_file', message: 'File not found: /files/payload.bin', hint: 'Choose the file again in the Body tab' },
      }),
    )
    expect(screen.getByRole('status')).toHaveTextContent('File missing')
    await user.click(screen.getByRole('radio', { name: 'form-data' }))
    await user.click(screen.getByRole('button', { name: 'Add Field' }))
    await user.selectOptions(screen.getByRole('combobox', { name: 'Type of row' }), 'file')
    await user.click(screen.getByRole('button', { name: 'Select File' }))
    expect(await screen.findByText('missing')).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'x-www-form-urlencoded' }))
    expect(screen.getByRole('table', { name: 'Url-encoded body' })).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'GraphQL' }))
    expect(screen.getByRole('textbox', { name: 'GraphQL variables' })).toBeInTheDocument()
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.body.binary.src).toBe('/files/payload.bin')
  })

  it('shows the scripts placeholder and edits the settings', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} />)
    await user.click(screen.getByRole('tab', { name: 'Scripts' }))
    expect(screen.getByText('Scripts arrive in a later stage of Xenon.')).toBeInTheDocument()
    await user.click(screen.getByRole('tab', { name: 'Settings' }))
    const timeout = screen.getByRole('spinbutton', { name: 'Request timeout in ms (0 = no timeout)' })
    await user.clear(timeout)
    await user.type(timeout, '5000')
    await user.click(screen.getByRole('checkbox', { name: 'Automatically follow redirects' }))
    await user.click(screen.getByRole('checkbox', { name: 'Enable SSL certificate verification' }))
    await user.click(screen.getByRole('checkbox', { name: 'Encode URL automatically' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.settings).toEqual({ timeoutMs: 5000, followRedirects: false, verifySsl: false, encodeUrl: false })
  })

  it('shows editing errors in a banner', () => {
    useRequestBuilderStore.setState((state) => ({
      tabs: { ...state.tabs, [tabId]: { ...state.tabs[tabId]!, error: { code: 'REQ-03400', description: 'invalid request' } as never } },
    }))
    render(<RequestBuilder tabId={tabId} />)
    expect(screen.getByRole('alert')).toHaveTextContent('invalid request')
    expect(emptyDraft().method).toBe('GET')
  })
})

function emptyResult(tabId: string) {
  return {
    tabId,
    ok: false,
    statusCode: 0,
    statusText: '',
    proto: '',
    headers: [],
    sizeBytes: 0,
    contentType: '',
    isBinary: false,
    bodyUrl: '',
    inlineBody: null,
    timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 0 },
    error: null,
  }
}

describe('RequestBuilder saving', () => {
  const saved: RequestDraft = { ...emptyDraft(), id: 'r1', name: 'Ping', url: 'https://x/ping' }

  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().reset()
    useRequestBuilderStore.getState().ensureTab('tab-1')
  })

  it('only saves requests opened from the sidebar', async () => {
    const tabId = 'tab-1'
    render(<RequestBuilder tabId={tabId} />)
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled()
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved, revision: 'rev-2' })
    await act(() => useRequestBuilderStore.getState().openSaved(tabId, 'r1'))
    expect(screen.getByRole('textbox', { name: 'Request URL' })).toHaveValue('https://x/ping')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', saved, 'rev-1', false)
  })

  it('asks whether to keep my version when the file changed on disk', async () => {
    const tabId = 'tab-1'
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockRejectedValueOnce(new AppError('REQ-02409', 'changed'))
    render(<RequestBuilder tabId={tabId} />)
    await act(() => useRequestBuilderStore.getState().openSaved(tabId, 'r1'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(screen.getByRole('status')).toHaveTextContent('changed on disk')
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved, revision: 'rev-2' })
    await userEvent.click(screen.getByRole('button', { name: 'Keep Mine' }))
    expect(requests.saveRequest).toHaveBeenLastCalledWith('ws', saved, 'rev-1', true)
    expect(screen.queryByText(/changed on disk/)).not.toBeInTheDocument()
    vi.mocked(requests.saveRequest).mockRejectedValueOnce(new AppError('REQ-02409', 'changed'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    await userEvent.click(screen.getByRole('button', { name: 'Load from Disk' }))
    expect(requests.getRequest).toHaveBeenLastCalledWith('ws', 'r1')
  })
})
