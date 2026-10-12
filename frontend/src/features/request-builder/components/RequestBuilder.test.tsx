import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, isCurlRefusal, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return {
    confirmationOf,
    expiredTokenOwnerOf,
    isCurlRefusal,
    missingSecretsOf,
    applyUrl: vi.fn(),
    applyParams: vi.fn(),
    parseBulk: vi.fn(),
    parseCurl: vi.fn(),
    formatBulk: vi.fn(),
    autoHeaders: vi.fn(),
    sendRequest: vi.fn(),
    getRequest: vi.fn(),
    saveRequest: vi.fn(),
    lastResult: vi.fn(),
    releaseResponse: vi.fn(),
    saveResponseBody: vi.fn(),
  }
})
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const tree = await import('@/platform/services/tree')
const variables = await import('@/platform/services/variables')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { RequestBuilder } = await import('./RequestBuilder')

function resetStores() {
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

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
    resetStores()
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
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(variables.inspectVariables).mockResolvedValue({ references: [], variables: [], scopes: {} })
    vi.mocked(variables.suggestVariables).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabId = 'tab-1'
    useRequestBuilderStore.getState().ensureTab(tabId)
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
    resetStores()
  })

  it('shows the sub-tabs in Postman order with Params first', () => {
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    const tabs = within(screen.getByRole('tablist', { name: 'Request' })).getAllByRole('tab')
    expect(tabs.map((tab) => tab.textContent)).toEqual(['Params', 'Authorization', 'Headers', 'Body', 'Scripts', 'Settings'])
    expect(screen.getByRole('tab', { name: 'Params' })).toHaveAttribute('aria-selected', 'true')
  })

  it('keeps the params table and the URL in sync both ways', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test/users?page=2&limit=10')
    const table = await screen.findByRole('table', { name: 'Query params' })
    await vi.waitFor(() =>
      expect(within(table).getAllByRole('textbox', { name: 'Key' }).map((cell) => cell.textContent)).toEqual(['page', 'limit']),
    )
    await user.click(within(table).getByRole('checkbox', { name: 'Enable limit' }))
    await vi.waitFor(() => expect(screen.getByRole('textbox', { name: 'Request URL' })).toHaveTextContent('https://api.test/users?page=2'))
  })

  it('highlights the variables Go inspected in the URL and the tables', async () => {
    vi.mocked(variables.inspectVariables).mockResolvedValue({
      references: [
        { field: 'url', name: 'base_url', start: 0, end: 12, status: 'resolved', scope: 'globals', displayValue: 'https://api', secret: false },
        { field: 'url', name: 'id', start: 13, end: 19, status: 'unresolved', scope: 'none', displayValue: '', secret: false },
        { field: 'params.0.value', name: 'page', start: 0, end: 8, status: 'dynamic', scope: 'dynamic', displayValue: '', secret: false },
      ],
      variables: [],
      scopes: {},
    })
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), '{{{{base_url}}/{{{{id}}?p={{{{page}}')
    await vi.waitFor(() => expect(variables.inspectVariables).toHaveBeenCalled())
    expect(await screen.findByTitle('Unresolved variable — sent as written')).toHaveTextContent('{{id}}')
    expect(screen.getByTitle('https://api — Globals')).toHaveTextContent('{{base_url}}')
    expect(await screen.findByTitle('Generated at send — Dynamic variable')).toHaveTextContent('{{page}}')
  })

  it('suggests variable names after {{ in the URL', async () => {
    vi.mocked(variables.suggestVariables).mockResolvedValue([{ name: 'base_url', scope: 'globals', description: '', secret: false, dynamic: false }])
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), '{{{{ba')
    expect(await screen.findByRole('option', { name: /base_url/ })).toHaveTextContent('Globals')
    expect(variables.suggestVariables).toHaveBeenLastCalledWith('ws', '', 'ba')
  })

  it('remembers a selection in the URL for "Move to secret variable"', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    const url = screen.getByRole('textbox', { name: 'Request URL' })
    await user.type(url, 'https://x')
    await user.keyboard('{Control>}a{/Control}')
    await vi.waitFor(() =>
      expect(useRequestBuilderStore.getState().tabs[tabId]!.selection).toMatchObject({ field: 'url', from: 0, to: 9, text: 'https://x' }),
    )
  })

  it('lists path variables and stores their values', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test/users/:id')
    const value = await screen.findByRole('textbox', { name: 'Value of id' })
    await user.type(value, '42')
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.pathVariables[0]?.value).toBe('42')
  })

  it('edits an example request without Send, scripts or settings', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    useRequestBuilderStore.getState().setSubTab(tabId, 'settings')
    render(<RequestBuilder tabId={tabId} onSave={onSave} mode="example" />)
    expect(screen.queryByRole('button', { name: 'Send' })).toBeNull()
    expect(screen.queryByRole('tab', { name: 'Scripts' })).toBeNull()
    expect(screen.queryByRole('tab', { name: 'Settings' })).toBeNull()
    expect(screen.getByRole('tab', { name: 'Params' })).toHaveAttribute('aria-selected', 'true')
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://x{Enter}')
    expect(requests.sendRequest).not.toHaveBeenCalled()
    await user.click(screen.getByRole('button', { name: 'Save' }))
    expect(onSave).toHaveBeenCalled()
  })

  it('turns Send into Cancel while the request is in flight', async () => {
    const user = userEvent.setup()
    const cancel = vi.fn()
    let reject: (error: unknown) => void = () => undefined
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise((_, fail) => (reject = fail)), cancel })
    cancel.mockImplementation(() => reject(new AppError('cancelled', 'You cancelled the request')))
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('button', { name: 'Send' }))
    const cancelButton = await screen.findByRole('button', { name: 'Cancel Request' })
    await user.click(cancelButton)
    expect(cancel).toHaveBeenCalled()
    expect(await screen.findByRole('button', { name: 'Send' })).toBeInTheDocument()
    expect(useResponseViewerStore.getState().entries[tabId]?.result?.error?.kind).toBe('cancelled')
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled()
  })

  it('sends with Enter in the URL field', async () => {
    const user = userEvent.setup()
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.type(screen.getByRole('textbox', { name: 'Request URL' }), 'https://api.test{Enter}')
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', tabId, expect.objectContaining({ url: 'https://api.test' }), { confirmedEnvironmentId: '', missingSecrets: false, expiredToken: false })
  })

  it('replaces the request with a pasted cURL command as one change that can be undone', async () => {
    vi.mocked(requests.parseCurl).mockResolvedValue({
      draft: {
        ...emptyDraft(),
        method: 'POST',
        url: 'https://api.test/x',
        headers: [{ id: 'h', key: 'Content-Type', value: 'application/json', description: '', enabled: true }],
      },
      notices: ['Ignored unknown option --frobnicate'],
    })
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    const url = screen.getByRole('textbox', { name: 'Request URL' })
    await user.type(url, 'https://old.test')
    await user.click(url)
    await user.paste("curl -X POST 'https://api.test/x' -H 'Content-Type: application/json' --frobnicate")
    expect(requests.parseCurl).toHaveBeenCalledWith("curl -X POST 'https://api.test/x' -H 'Content-Type: application/json' --frobnicate")
    expect(await screen.findByText('Replaced the request with the pasted cURL command.')).toBeInTheDocument()
    expect(screen.getByRole('listitem')).toHaveTextContent('Ignored unknown option --frobnicate')
    expect(screen.getByRole('button', { name: 'Method POST' })).toBeInTheDocument()
    await vi.waitFor(() => expect(url).toHaveTextContent('https://api.test/x'))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.headers).toHaveLength(1)
    await user.click(screen.getByRole('button', { name: 'Undo Paste' }))
    expect(screen.queryByText('Replaced the request with the pasted cURL command.')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Method GET' })).toBeInTheDocument()
    await vi.waitFor(() => expect(url).toHaveTextContent('https://old.test'))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.headers).toHaveLength(0)
  })

  it('pastes text that is not a cURL command into the URL as it is', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    const url = screen.getByRole('textbox', { name: 'Request URL' })
    await user.click(url)
    await user.paste('https://api.test/curl me')
    await vi.waitFor(() => expect(url).toHaveTextContent('https://api.test/curl me'))
    expect(requests.parseCurl).not.toHaveBeenCalled()
  })

  it('shows why a pasted cURL command could not be read', async () => {
    vi.mocked(requests.parseCurl).mockRejectedValue(new AppError('IMP-07400', 'Only HTTP and HTTPS URLs can be imported, not ftp://x'))
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('textbox', { name: 'Request URL' }))
    await user.paste('curl ftp://x')
    expect(await screen.findByRole('alert')).toHaveTextContent('Only HTTP and HTTPS URLs can be imported, not ftp://x. Check the cURL command and paste it again.')
  })

  it('picks standard and custom methods', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
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
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
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
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Headers' }))
    await user.click(screen.getByRole('button', { name: 'Bulk Edit' }))
    await user.type(await screen.findByRole('textbox', { name: 'Headers bulk edit' }), 'X-Token:{{{{token}}')
    await vi.waitFor(() => expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.headers[0]?.key).toBe('X-Token'))
    expect(screen.getByRole('tab', { name: 'Headers (1)' })).toBeInTheDocument()
  })

  it('shows the auth editor the app composes in the Authorization tab', async () => {
    const user = userEvent.setup()
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} auth={<p>auth editor</p>} />)
    await user.click(screen.getByRole('tab', { name: 'Authorization' }))
    expect(screen.getByText('auth editor')).toBeInTheDocument()
  })

  it('asks to pick imported files again', async () => {
    const user = userEvent.setup()
    act(() => {
      useRequestBuilderStore.getState().updateAuth(tabId, { type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } })
      useRequestBuilderStore.getState().updateBody(tabId, {
        mode: 'formdata',
        formdata: [{ id: 'f', key: 'key', value: '', description: '', enabled: true, type: 'file', src: '/home/ana/.ssh/id_rsa', srcState: 'repick' }],
      })
    })
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Body' }))
    expect(screen.getByText('Imported from another machine — select the file again before sending')).toBeInTheDocument()
    vi.mocked(dialogs.pickFile).mockResolvedValue('/home/me/file.txt')
    await user.click(screen.getByRole('button', { name: 'Select File' }))
    await vi.waitFor(() => expect(screen.queryByText('Imported from another machine — select the file again before sending')).not.toBeInTheDocument())
    act(() => useRequestBuilderStore.getState().updateBody(tabId, { mode: 'binary', binary: { src: '/tmp/payload.bin', srcState: 'repick' } }))
    expect(screen.getByText('Imported from another machine — select the file again before sending')).toBeInTheDocument()
  })

  it('shows where a raw body breaks', async () => {
    const user = userEvent.setup()
    act(() => {
      useRequestBuilderStore.getState().updateBody(tabId, { mode: 'raw', raw: { language: 'json', content: '{\n  "ids": [\n    500947,\n  ]\n}' } })
    })
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Body' }))
    expect(await screen.findByRole('list', { name: 'Raw body problems' })).toHaveTextContent('Line 3, column 11: Trailing comma before "]"')
    act(() => useRequestBuilderStore.getState().updateBody(tabId, { raw: { language: 'json', content: '{"ids": [500947]}' } }))
    await vi.waitFor(() => expect(screen.queryByRole('list', { name: 'Raw body problems' })).not.toBeInTheDocument())
  })

  it('beautifies raw bodies in every language except plain text', async () => {
    const user = userEvent.setup()
    act(() => {
      useRequestBuilderStore.getState().updateBody(tabId, { mode: 'raw', raw: { language: 'json', content: '{\n\n "test":"lkalala"}' } })
    })
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Body' }))
    await user.click(screen.getByRole('button', { name: 'Beautify' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]?.draft.body.raw.content).toBe('{\n  "test": "lkalala"\n}')
    act(() => useRequestBuilderStore.getState().updateBody(tabId, { raw: { language: 'json', content: '{"broken": ' } }))
    expect(screen.getByRole('button', { name: 'Beautify' })).toBeDisabled()
    await user.selectOptions(screen.getByRole('combobox', { name: 'Raw language' }), 'xml')
    act(() => useRequestBuilderStore.getState().updateBody(tabId, { raw: { language: 'xml', content: '<a><b>1</b></a>' } }))
    await user.click(screen.getByRole('button', { name: 'Beautify' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]?.draft.body.raw.content).toBe('<a>\n  <b>1</b>\n</a>')
    await user.selectOptions(screen.getByRole('combobox', { name: 'Raw language' }), 'text')
    expect(screen.queryByRole('button', { name: 'Beautify' })).not.toBeInTheDocument()
  })

  it('edits every body mode and marks a missing file', async () => {
    const user = userEvent.setup()
    vi.mocked(dialogs.pickFile).mockResolvedValue('/files/payload.bin')
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Body' }))
    expect(screen.getByText('This request does not have a body.')).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'Body type' })).toContainElement(screen.getByRole('radio', { name: 'Raw' }))
    await user.click(screen.getByRole('radio', { name: 'Raw' }))
    await user.selectOptions(screen.getByRole('combobox', { name: 'Raw language' }), 'xml')
    expect(await screen.findByRole('textbox', { name: 'Raw body' })).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'Binary' }))
    await user.click(screen.getByRole('button', { name: 'Select File' }))
    expect(await screen.findByText('/files/payload.bin')).toBeInTheDocument()
    act(() =>
      useResponseViewerStore.getState().setResult(tabId, {
        ...emptyResult(tabId),
        error: { kind: 'missing_file', message: 'File not found: /files/payload.bin', hint: 'Choose the file again in the Body tab' },
      }),
    )
    expect(screen.getByRole('status')).toHaveTextContent('Xenon can’t find this file — select it again')
    await user.click(screen.getByRole('radio', { name: 'form-data' }))
    await user.click(screen.getByRole('button', { name: 'Add Field' }))
    await user.selectOptions(screen.getByRole('combobox', { name: 'Type of row' }), 'file')
    await user.click(screen.getByRole('button', { name: 'Select File' }))
    expect(await screen.findByRole('status')).toHaveTextContent('Xenon can’t find this file — select it again')
    await user.click(screen.getByRole('radio', { name: 'x-www-form-urlencoded' }))
    expect(screen.getByRole('table', { name: 'Url-encoded body' })).toBeInTheDocument()
    await user.click(screen.getByRole('radio', { name: 'GraphQL' }))
    expect(await screen.findByRole('textbox', { name: 'GraphQL variables' })).toBeInTheDocument()
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.body.binary.src).toBe('/files/payload.bin')
  })

  it('shows the scripts editor the app composes and edits the settings', async () => {
    const user = userEvent.setup()
    const { rerender } = render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    await user.click(screen.getByRole('tab', { name: 'Scripts' }))
    expect(screen.getByText('Scripts can be edited in the request tab.')).toBeInTheDocument()
    rerender(<RequestBuilder tabId={tabId} onSave={vi.fn()} scripts={<p>scripts editor</p>} />)
    expect(screen.getByText('scripts editor')).toBeInTheDocument()
    await user.click(screen.getByRole('tab', { name: 'Settings' }))
    const timeout = screen.getByRole('spinbutton', { name: 'Request timeout in ms (0 = no timeout)' })
    await user.clear(timeout)
    await user.type(timeout, '5000')
    await user.click(screen.getByRole('checkbox', { name: 'Automatically follow redirects' }))
    await user.click(screen.getByRole('checkbox', { name: 'Enable SSL certificate verification' }))
    await user.click(screen.getByRole('checkbox', { name: 'Encode URL automatically' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.draft.settings).toEqual({ timeoutMs: 5000, followRedirects: false, verifySsl: false, encodeUrl: false, disableCookieJar: false })
  })

  it('shows editing errors in a banner', () => {
    useRequestBuilderStore.setState((state) => ({
      tabs: { ...state.tabs, [tabId]: { ...state.tabs[tabId]!, error: { code: 'REQ-03400', description: 'invalid request' } as never } },
    }))
    render(<RequestBuilder tabId={tabId} onSave={vi.fn()} />)
    expect(screen.getByRole('alert')).toHaveTextContent('invalid request')
    expect(emptyDraft().method).toBe('GET')
  })
})

function emptyResult(tabId: string): SendResult {
  return {
    cookiesSet: [],
    tabId,
    sendId: '',
    outcome: 'error',
    tests: [],
    testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 },
    scriptErrors: [],
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
    resetStores()
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(tree.getTree).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().ensureTab('tab-1')
  })

  afterEach(() => resetStores())

  function saveTab(tabId: string) {
    return () => void useRequestBuilderStore.getState().save(tabId)
  }

  it('hands the Save button to the caller and saves requests opened from the sidebar', async () => {
    const tabId = 'tab-1'
    const onSave = vi.fn(saveTab(tabId))
    render(<RequestBuilder tabId={tabId} onSave={onSave} />)
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(onSave).toHaveBeenCalledOnce()
    expect(requests.saveRequest).not.toHaveBeenCalled()
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved, revision: 'rev-2' })
    await act(() => useRequestBuilderStore.getState().openSaved(tabId, 'r1'))
    expect(screen.getByRole('textbox', { name: 'Request URL' })).toHaveTextContent('https://x/ping')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', saved, 'rev-1', false)
  })

  it('shows the save in progress on the Save button', async () => {
    const tabId = 'tab-1'
    let finish: (value: { draft: RequestDraft; revision: string }) => void = () => undefined
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockImplementation(() => new Promise((resolve) => (finish = resolve)))
    render(<RequestBuilder tabId={tabId} onSave={saveTab(tabId)} />)
    await act(() => useRequestBuilderStore.getState().openSaved(tabId, 'r1'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(screen.getByRole('button', { name: 'Saving…' })).toBeDisabled()
    await act(async () => finish({ draft: saved, revision: 'rev-2' }))
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled()
  })

  it('asks whether to keep my version when the file changed on disk', async () => {
    const tabId = 'tab-1'
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockRejectedValueOnce(new AppError('REQ-02409', 'changed'))
    render(<RequestBuilder tabId={tabId} onSave={saveTab(tabId)} />)
    await act(() => useRequestBuilderStore.getState().openSaved(tabId, 'r1'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(screen.getByRole('status')).toHaveTextContent('changed on disk')
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved, revision: 'rev-2' })
    await userEvent.click(screen.getByRole('button', { name: 'Keep Your Version' }))
    expect(requests.saveRequest).toHaveBeenLastCalledWith('ws', saved, 'rev-1', true)
    expect(screen.queryByText(/changed on disk/)).not.toBeInTheDocument()
    vi.mocked(requests.saveRequest).mockRejectedValueOnce(new AppError('REQ-02409', 'changed'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    await userEvent.click(screen.getByRole('button', { name: 'Load from Disk' }))
    expect(requests.getRequest).toHaveBeenLastCalledWith('ws', 'r1')
  })
})
