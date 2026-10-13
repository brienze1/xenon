import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return {
    confirmationOf,
    expiredTokenOwnerOf,
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
    saveRequestAs: vi.fn(),
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
const { useSidebarStore } = await import('@/features/sidebar/store')
const { INSPECT_DELAY_MS, SYNC_DELAY_MS, useRequestBuilderStore } = await import('./store')

function row(key: string, value: string, enabled = true) {
  return { id: `id-${key}`, key, value, description: '', enabled }
}

function state() {
  return useRequestBuilderStore.getState()
}

function resetStores() {
  state().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

function mockDefaults() {
  vi.mocked(requests.autoHeaders).mockResolvedValue([])
  vi.mocked(requests.applyUrl).mockImplementation(async (sent) => sent)
  vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
  vi.mocked(requests.parseBulk).mockResolvedValue([])
  vi.mocked(requests.releaseResponse).mockResolvedValue()
  vi.mocked(tree.getTree).mockResolvedValue([])
}

function draft(tabId: string): RequestDraft {
  return state().tabs[tabId]!.draft
}

const okResult: SendResult = {
  cookiesSet: [],
  tabId: 'tab',
  sendId: 's', outcome: 'sent', tests: [], testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 }, scriptErrors: [],
  ok: true,
  statusCode: 200,
  statusText: 'OK',
  proto: 'HTTP/1.1',
  headers: [],
  sizeBytes: 2,
  contentType: 'text/plain',
  isBinary: false,
  bodyUrl: '/xenon/response/tab',
  inlineBody: 'hi',
  timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 1 },
  error: null,
}

describe('request builder store', () => {
  let tabId: string

  beforeEach(() => {
    resetStores()
    vi.resetAllMocks()
    vi.useFakeTimers()
    mockDefaults()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabId = 'tab-1'
    state().ensureTab(tabId)
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('reports a sync result it cannot read instead of throwing in the timer', async () => {
    vi.mocked(requests.applyUrl).mockResolvedValue(undefined as unknown as RequestDraft)
    state().setUrl(tabId, 'https://h?a=1')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(state().tabs[tabId]!.error).not.toBeNull()
    expect(draft(tabId).url).toBe('https://h?a=1')
  })

  it('drops every pending sync on reset', async () => {
    state().setUrl(tabId, 'https://h?a=1')
    state().setBulkText(tabId, 'headers', 'A:1')
    state().reset()
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.applyUrl).not.toHaveBeenCalled()
    expect(requests.parseBulk).not.toHaveBeenCalled()
    expect(requests.autoHeaders).not.toHaveBeenCalled()
  })

  it('ignores a name edit when there is no tab', () => {
    state().setName('', 'Orphan')
    expect(state().tabs['']).toBeUndefined()
  })

  it('keeps one clean draft per tab id and marks edits dirty', () => {
    expect(draft(tabId)).toEqual(emptyDraft())
    expect(state().tabs[tabId]!.dirty).toBe(false)
    state().ensureTab('tab-other')
    state().ensureTab(tabId)
    expect(Object.keys(state().tabs)).toHaveLength(2)
    state().setName(tabId, 'x')
    expect(state().tabs[tabId]!.dirty).toBe(true)
    expect(state().tabs['tab-other']!.dirty).toBe(false)
  })

  it('loads a draft and drops a tab with its pending work', async () => {
    const cancel = vi.fn()
    state().loadDraft(tabId, { ...emptyDraft(), url: 'https://d' }, 'rev', true)
    expect(state().tabs[tabId]).toMatchObject({ baseRevision: 'rev', dirty: true, draft: { url: 'https://d' } })
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel })
    void state().send(tabId)
    state().setUrl(tabId, 'https://e')
    state().dropTab(tabId)
    expect(cancel).toHaveBeenCalled()
    expect(state().tabs[tabId]).toBeUndefined()
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.applyUrl).not.toHaveBeenCalled()
  })

  it('re-derives the params from the URL after 150 ms, keeping row ids', async () => {
    state().updateParams(tabId, [row('page', '1')])
    vi.mocked(requests.applyUrl).mockImplementation(async (sent) => ({ ...sent, params: [row('page', '2'), row('limit', '10')].map((r) => ({ ...r, id: `new-${r.key}` })) }))
    vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
    state().setUrl(tabId, 'https://h?page=2')
    state().setUrl(tabId, 'https://h?page=2&limit=10')
    expect(draft(tabId).url).toBe('https://h?page=2&limit=10')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS - 1)
    expect(requests.applyUrl).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(requests.applyUrl).toHaveBeenCalledTimes(1)
    expect(draft(tabId).params.map((r) => [r.id, r.key, r.value])).toEqual([
      ['id-page', 'page', '2'],
      ['new-limit', 'limit', '10'],
    ])
  })

  it('ignores a derived result once the URL changed again', async () => {
    let resolve: (value: RequestDraft) => void = () => undefined
    vi.mocked(requests.applyUrl).mockImplementation((sent) => new Promise((done) => (resolve = () => done({ ...sent, params: [row('stale', '1')] }))))
    state().setUrl(tabId, 'https://h?stale=1')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    state().setUrl(tabId, 'https://h')
    resolve(draft(tabId))
    await vi.advanceTimersByTimeAsync(0)
    expect(draft(tabId).params).toEqual([])
  })

  it('rebuilds the URL from the params and path variables', async () => {
    vi.mocked(requests.applyParams).mockResolvedValue({ ...emptyDraft(), url: 'https://h/:userId?a=1' })
    state().updateParams(tabId, [row('a', '1')])
    state().updatePathVariables(tabId, [{ id: 'p', key: 'userId', value: '', description: '' }])
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.applyParams).toHaveBeenCalledTimes(1)
    expect(draft(tabId).url).toBe('https://h/:userId?a=1')
    expect(draft(tabId).params[0]?.key).toBe('a')
  })

  it('shows sync errors on the tab', async () => {
    vi.mocked(requests.applyUrl).mockRejectedValue(new AppError('REQ-03400', 'bad'))
    state().setUrl(tabId, 'x')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(state().tabs[tabId]!.error?.code).toBe('REQ-03400')
  })

  it('switches tables to bulk text and parses edits back into rows', async () => {
    state().updateHeaders(tabId, [row('Accept', '*/*')])
    vi.mocked(requests.formatBulk).mockResolvedValue('Accept:*/*')
    await state().setBulkMode(tabId, 'headers', true)
    expect(state().tabs[tabId]!.bulk.headers).toBe('Accept:*/*')
    vi.mocked(requests.parseBulk).mockResolvedValue([row('Accept', 'json'), row('X-New', '1', false)])
    state().setBulkText(tabId, 'headers', 'Accept:json\n//X-New:1')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(draft(tabId).headers.map((r) => [r.key, r.value, r.enabled])).toEqual([
      ['Accept', 'json', true],
      ['X-New', '1', false],
    ])
    await state().setBulkMode(tabId, 'headers', false)
    expect(state().tabs[tabId]!.bulk.headers).toBeUndefined()
  })

  it('parses bulk params and url-encoded rows into their tables', async () => {
    vi.mocked(requests.parseBulk).mockResolvedValue([row('a', '1')])
    vi.mocked(requests.applyParams).mockResolvedValue({ ...emptyDraft(), url: 'https://h?a=1' })
    state().setBulkText(tabId, 'params', 'a:1')
    state().setBulkText(tabId, 'urlencoded', 'a:1')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS * 2)
    expect(draft(tabId).params[0]?.key).toBe('a')
    expect(draft(tabId).body.urlencoded[0]?.key).toBe('a')
    expect(draft(tabId).url).toBe('https://h?a=1')
  })

  it('keeps editing errors from bulk conversion', async () => {
    vi.mocked(requests.formatBulk).mockRejectedValue(new AppError('REQ-06400', 'x'))
    await state().setBulkMode(tabId, 'params', true)
    expect(state().tabs[tabId]!.error?.code).toBe('REQ-06400')
  })

  it('edits method, auth, body, settings and auto headers', async () => {
    state().setMethod(tabId, ' purge ')
    state().updateAuth(tabId, { type: 'bearer', bearer: { token: 'abc' } })
    state().updateBody(tabId, { mode: 'raw' })
    state().updateSettings(tabId, { timeoutMs: 500 })
    state().setAutoHeaderEnabled(tabId, 'User-Agent', false)
    state().setAutoHeaderEnabled(tabId, 'Accept', false)
    state().setAutoHeaderEnabled(tabId, 'user-agent', true)
    state().setSubTab(tabId, 'headers')
    expect(draft(tabId)).toMatchObject({ method: 'PURGE', auth: { type: 'bearer', bearer: { token: 'abc' } }, body: { mode: 'raw' }, settings: { timeoutMs: 500, verifySsl: true } })
    expect(draft(tabId).disabledAutoHeaders).toEqual(['Accept'])
    expect(state().tabs[tabId]!.subTab).toBe('headers')
    vi.mocked(requests.autoHeaders).mockResolvedValue([{ key: 'Accept', value: '*/*', disabled: true }])
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.autoHeaders).toHaveBeenCalledWith('ws', draft(tabId))
    expect(state().tabs[tabId]!.autoHeaders).toEqual([{ key: 'Accept', value: '*/*', disabled: true }])
  })

  it('does not ask for auto headers without a workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().refreshAutoHeaders(tabId)
    expect(requests.autoHeaders).not.toHaveBeenCalled()
  })

  it('chooses body files with the native picker', async () => {
    vi.mocked(dialogs.pickFile).mockResolvedValueOnce('/a.bin').mockResolvedValueOnce('/c.txt').mockResolvedValueOnce(null)
    await state().chooseBodyFile(tabId)
    expect(draft(tabId).body.binary.src).toBe('/a.bin')
    state().updateBody(tabId, { formdata: [{ ...row('f', ''), type: 'file', src: '', srcState: 'repick' }, { ...row('g', ''), type: 'file', src: '/keep', srcState: 'ok' }] })
    await state().chooseBodyFile(tabId, 'id-f')
    expect(draft(tabId).body.formdata.map((r) => r.src)).toEqual(['/c.txt', '/keep'])
    expect(draft(tabId).body.formdata.map((r) => r.srcState)).toEqual(['ok', 'ok'])
    await state().chooseBodyFile(tabId)
    expect(draft(tabId).body.binary.src).toBe('/a.bin')
    vi.mocked(dialogs.pickFile).mockRejectedValueOnce(new AppError('DLG-02500', 'x'))
    await state().chooseBodyFile(tabId)
    expect(state().tabs[tabId]!.error?.code).toBe('DLG-02500')
  })

  it('sends the draft and stores the result for the response viewer', async () => {
    vi.mocked(requests.sendRequest).mockReturnValue({ result: Promise.resolve({ ...okResult, tabId }), cancel: vi.fn() })
    const sending = state().send(tabId)
    expect(state().tabs[tabId]!.pending).not.toBeNull()
    expect(useResponseViewerStore.getState().entries[tabId]?.sending).toBe(true)
    await sending
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', tabId, draft(tabId), { confirmedEnvironmentId: '', missingSecrets: false, expiredToken: false })
    expect(state().tabs[tabId]!.pending).toBeNull()
    expect(useResponseViewerStore.getState().entries[tabId]?.result?.statusCode).toBe(200)
  })

  it('cancels an in-flight send and shows it as cancelled', async () => {
    let reject: (error: unknown) => void = () => undefined
    const cancel = vi.fn(() => reject(new AppError('cancelled', 'You cancelled the request')))
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise((_, fail) => (reject = fail)), cancel })
    const sending = state().send(tabId)
    await state().send(tabId)
    expect(requests.sendRequest).toHaveBeenCalledTimes(1)
    state().cancel(tabId)
    await sending
    expect(cancel).toHaveBeenCalled()
    expect(useResponseViewerStore.getState().entries[tabId]?.result?.error?.kind).toBe('cancelled')
    expect(state().tabs[tabId]!.pending).toBeNull()
  })

  it('keeps binding failures apart from send results', async () => {
    vi.mocked(requests.sendRequest).mockReturnValue({ result: Promise.reject(new AppError('SND-00400', 'invalid')), cancel: vi.fn() })
    await state().send(tabId)
    expect(useResponseViewerStore.getState().entries[tabId]?.failure?.code).toBe('SND-00400')
  })

  it('asks about missing or locked secrets instead of showing a failure, then sends anyway or cancels', async () => {
    useResponseViewerStore.getState().setResult(tabId, { ...okResult, tabId })
    const refused = new AppError('SND-06409', 'secrets missing', [
      { attribute: 'missing_secrets', messages: ['api_key'] },
      { attribute: 'locked_secrets', messages: ['token'] },
    ])
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.reject(refused), cancel: vi.fn() })
    await state().send(tabId)
    expect(state().tabs[tabId]!.missingSecrets).toEqual({ missing: ['api_key'], locked: ['token'] })
    expect(useResponseViewerStore.getState().entries[tabId]).toMatchObject({ sending: false, failure: null, result: { statusCode: 200 } })
    state().dismissMissingSecrets(tabId)
    expect(state().tabs[tabId]!.missingSecrets).toBeNull()
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.reject(refused), cancel: vi.fn() })
    await state().send(tabId)
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.resolve({ ...okResult, tabId }), cancel: vi.fn() })
    await state().send(tabId, { missingSecrets: true })
    expect(requests.sendRequest).toHaveBeenLastCalledWith('ws', tabId, draft(tabId), { confirmedEnvironmentId: '', missingSecrets: true, expiredToken: false })
    expect(state().tabs[tabId]!.missingSecrets).toBeNull()
  })

  it('does not send without a workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().send(tabId)
    expect(requests.sendRequest).not.toHaveBeenCalled()
    state().cancel(tabId)
  })
})

describe('request builder store - saved requests', () => {
  const savedDraft: RequestDraft = { ...emptyDraft(), id: 'r1', name: 'Ping', url: 'https://x/ping' }

  beforeEach(() => {
    resetStores()
    vi.resetAllMocks()
    mockDefaults()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().ensureTab(TAB)
  })

  afterEach(() => resetStores())

  const TAB = 'tab-1'

  function activeTab() {
    return useRequestBuilderStore.getState().tabs[TAB]!
  }

  it('opens a saved request into the tab with its revision', async () => {
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-1' })
    const tabId = TAB
    await state().openSaved(tabId, 'r1')
    expect(requests.getRequest).toHaveBeenCalledWith('ws', 'r1')
    expect(activeTab().draft).toEqual(savedDraft)
    expect(activeTab().baseRevision).toBe('rev-1')
  })

  it('reports a request that cannot be opened', async () => {
    vi.mocked(requests.getRequest).mockRejectedValue(new AppError('REQ-01404', 'gone'))
    await state().openSaved(TAB, 'r1')
    expect(activeTab().error?.code).toBe('REQ-01404')
  })

  it('saves with the base revision and keeps the new one', async () => {
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...savedDraft, url: 'https://x/v2' }, revision: 'rev-2' })
    const tabId = TAB
    await state().openSaved(tabId, 'r1')
    await state().save(tabId)
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', savedDraft, 'rev-1', false)
    expect(activeTab().baseRevision).toBe('rev-2')
    expect(activeTab().draft.url).toBe('https://x/v2')
    expect(activeTab().saving).toBe(false)
  })

  it('does not save an unsaved draft', async () => {
    await expect(state().save(TAB)).resolves.toEqual({ ok: true, value: false })
    expect(requests.saveRequest).not.toHaveBeenCalled()
  })

  it('saves a draft as a new request at a chosen place', async () => {
    state().setUrl(TAB, 'https://new')
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: { ...savedDraft, id: 'n1', url: 'https://new' }, revision: 'rev-n' })
    await expect(state().saveAs(TAB, 'c1', 'New')).resolves.toEqual({ ok: true, value: 'n1' })
    expect(requests.saveRequestAs).toHaveBeenCalledWith('ws', 'c1', 'New', expect.objectContaining({ url: 'https://new' }))
    expect(activeTab()).toMatchObject({ baseRevision: 'rev-n', dirty: false, saving: false, draft: { id: 'n1' } })
    vi.mocked(requests.saveRequestAs).mockRejectedValue(new AppError('TRE-01404', 'no parent'))
    await expect(state().saveAs(TAB, 'gone', 'New')).resolves.toMatchObject({ ok: false, error: { code: 'TRE-01404' } })
    expect(activeTab().error).toBeNull()
    useWorkspacesStore.setState({ activeLocalId: null })
    await expect(state().saveAs(TAB, 'c1', 'New')).resolves.toEqual({ ok: true, value: null })
    await expect(state().openSaved(TAB, 'r1')).resolves.toBe(false)
  })

  it('offers keep mine when the file changed on disk', async () => {
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-1' })
    vi.mocked(requests.saveRequest).mockRejectedValueOnce(new AppError('REQ-02409', 'changed'))
    const tabId = TAB
    await state().openSaved(tabId, 'r1')
    await state().save(tabId)
    expect(activeTab().conflict).toBe(true)
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-3' })
    await state().save(tabId, true)
    expect(requests.saveRequest).toHaveBeenLastCalledWith('ws', savedDraft, 'rev-1', true)
    expect(activeTab().conflict).toBe(false)
    vi.mocked(requests.saveRequest).mockRejectedValue(new AppError('REQ-13500', 'disk full'))
    await expect(state().save(tabId)).resolves.toMatchObject({ ok: false, error: { code: 'REQ-13500' } })
    expect(activeTab().error?.code).toBe('REQ-13500')
  })

  it('turns a deleted request into an unsaved draft', async () => {
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-1' })
    await state().openSaved(TAB, 'r1')
    expect(activeTab().dirty).toBe(false)
    state().detach(TAB)
    expect(activeTab().draft.id).toBe('')
    expect(activeTab().dirty).toBe(true)
    expect(activeTab().baseRevision).toBeNull()
    expect(activeTab().draft.url).toBe('https://x/ping')
  })

  it('adopts the new name and revision after a rename', async () => {
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: savedDraft, revision: 'rev-1' })
    const tabId = TAB
    await state().openSaved(tabId, 'r1')
    state().setUrl(tabId, 'https://edited')
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: { ...savedDraft, name: 'Pong' }, revision: 'rev-9' })
    await state().adoptRename('r1')
    expect(activeTab().draft.name).toBe('Pong')
    expect(activeTab().draft.url).toBe('https://edited')
    expect(activeTab().baseRevision).toBe('rev-9')
    await state().adoptRename('unknown')
    expect(requests.getRequest).toHaveBeenCalledTimes(2)
    vi.mocked(requests.getRequest).mockRejectedValue(new AppError('REQ-01404', 'gone'))
    await state().adoptRename('r1')
    expect(activeTab().error?.code).toBe('REQ-01404')
  })

  it('edits the name and the description', () => {
    const tabId = TAB
    state().setName(tabId, 'Ping')
    state().setDescription(tabId, 'docs')
    expect(activeTab().draft).toMatchObject({ name: 'Ping', description: 'docs' })
  })
})

describe('request builder store - variables', () => {
  const tabId = 'tab-v'
  const inspection = {
    references: [{ field: 'url', name: 'base', start: 0, end: 8, status: 'resolved' as const, scope: 'environment' as const, displayValue: 'x', secret: false }],
    variables: [],
    scopes: { globals: 'globals' },
  }

  beforeEach(() => {
    resetStores()
    vi.resetAllMocks()
    vi.useFakeTimers()
    mockDefaults()
    vi.mocked(variables.inspectVariables).mockResolvedValue(inspection)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    state().ensureTab(tabId)
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('inspects the draft once typing pauses', async () => {
    state().setUrl(tabId, '{{b')
    state().setUrl(tabId, '{{base}}')
    await vi.advanceTimersByTimeAsync(INSPECT_DELAY_MS - 1)
    expect(variables.inspectVariables).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(variables.inspectVariables).toHaveBeenCalledTimes(1)
    expect(variables.inspectVariables).toHaveBeenCalledWith('ws', expect.objectContaining({ url: '{{base}}' }))
    expect(state().tabs[tabId]!.inspection).toEqual(inspection)
  })

  it('drops answers for a draft that changed meanwhile and keeps the last one on failure', async () => {
    let answer: (value: typeof inspection) => void = () => undefined
    vi.mocked(variables.inspectVariables).mockReturnValueOnce(new Promise((resolve) => (answer = resolve)))
    const pending = state().inspect(tabId)
    state().setName(tabId, 'changed')
    answer(inspection)
    await pending
    expect(state().tabs[tabId]!.inspection).toBeNull()
    await state().inspect(tabId)
    expect(state().tabs[tabId]!.inspection).toEqual(inspection)
    vi.mocked(variables.inspectVariables).mockRejectedValueOnce(new AppError('VAR-15400', 'invalid'))
    await state().inspect(tabId)
    expect(state().tabs[tabId]!.inspection).toEqual(inspection)
  })

  it('refreshes every open tab and skips without a workspace', async () => {
    state().ensureTab('tab-w')
    state().refreshInspections()
    await vi.runOnlyPendingTimersAsync()
    expect(variables.inspectVariables).toHaveBeenCalledTimes(2)
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().inspect(tabId)
    expect(variables.inspectVariables).toHaveBeenCalledTimes(2)
    await expect(state().suggest(tabId, 'a')).resolves.toEqual([])
  })

  it('suggests with the saved request id and swallows failures', async () => {
    useRequestBuilderStore.setState((current) => ({ tabs: { ...current.tabs, [tabId]: { ...current.tabs[tabId]!, draft: { ...draft(tabId), id: 'r1' } } } }))
    vi.mocked(variables.suggestVariables).mockResolvedValueOnce([{ name: 'base', scope: 'globals', description: '', secret: false, dynamic: false }])
    await expect(state().suggest(tabId, 'ba')).resolves.toHaveLength(1)
    expect(variables.suggestVariables).toHaveBeenCalledWith('ws', 'r1', 'ba')
    vi.mocked(variables.suggestVariables).mockRejectedValueOnce(new AppError('VAR-16400', 'invalid'))
    await expect(state().suggest(tabId, 'ba')).resolves.toEqual([])
  })

  it('replaces the selected text of a field with a reference', async () => {
    state().updateHeaders(tabId, [row('Authorization', 'Bearer abc')])
    state().setSelection(tabId, { field: 'headers.0.value', from: 7, to: 10, text: 'abc' })
    expect(state().replaceSelection(tabId, '{{token}}')).toBe(true)
    expect(draft(tabId).headers[0]!.value).toBe('Bearer {{token}}')
    expect(state().tabs[tabId]!.selection).toBeNull()
    expect(state().replaceSelection(tabId, '{{token}}')).toBe(false)
    state().setSelection(tabId, { field: 'headers.0.value', from: 0, to: 3, text: 'xyz' })
    expect(state().replaceSelection(tabId, '{{t}}')).toBe(false)
    state().setSelection(tabId, { field: 'auth.bearer.token', from: 0, to: 1, text: 'a' })
    expect(state().replaceSelection(tabId, '{{t}}')).toBe(false)
    state().setSelection('missing-tab', { field: 'url', from: 0, to: 1, text: 'a' })
    expect(state().tabs['missing-tab']).toBeUndefined()
  })

  it('syncs the URL or the params after replacing in them', async () => {
    state().setUrl(tabId, 'https://h/abc')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    vi.mocked(requests.applyUrl).mockClear()
    state().setSelection(tabId, { field: 'url', from: 10, to: 13, text: 'abc' })
    state().replaceSelection(tabId, '{{id}}')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.applyUrl).toHaveBeenCalledWith(expect.objectContaining({ url: 'https://h/{{id}}' }))
    state().updateParams(tabId, [row('key', 'abc')])
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    vi.mocked(requests.applyParams).mockClear()
    state().setSelection(tabId, { field: 'params.0.value', from: 0, to: 3, text: 'abc' })
    state().replaceSelection(tabId, '{{k}}')
    await vi.advanceTimersByTimeAsync(SYNC_DELAY_MS)
    expect(requests.applyParams).toHaveBeenCalledTimes(1)
  })

  it('asks to confirm the environment, then sends with the confirmation and keeps it for a secrets prompt', async () => {
    const confirm = new AppError('SND-05409', 'confirm', [
      { attribute: 'environment_id', messages: ['01P'] },
      { attribute: 'name', messages: ['prod'] },
      { attribute: 'color', messages: ['red'] },
    ])
    const secrets = new AppError('SND-06409', 'secrets', [{ attribute: 'missing_secrets', messages: ['api_key'] }])
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.reject(confirm), cancel: vi.fn() })
    await state().send(tabId)
    expect(state().tabs[tabId]!.confirmation).toEqual({ environmentId: '01P', name: 'prod', color: 'red' })
    expect(useResponseViewerStore.getState().entries[tabId]?.sending).toBe(false)
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.reject(secrets), cancel: vi.fn() })
    await state().send(tabId, { confirmedEnvironmentId: '01P' })
    expect(requests.sendRequest).toHaveBeenLastCalledWith('ws', tabId, draft(tabId), { confirmedEnvironmentId: '01P', missingSecrets: false, expiredToken: false })
    expect(state().tabs[tabId]!.confirmation).toBeNull()
    expect(state().tabs[tabId]!.missingSecrets).toEqual({ missing: ['api_key'], locked: [] })
    expect(state().tabs[tabId]!.acknowledgement).toEqual({ confirmedEnvironmentId: '01P', missingSecrets: false, expiredToken: false })
    state().dismissConfirmation(tabId)
    expect(state().tabs[tabId]!.confirmation).toBeNull()
  })

  it('replaces the draft with a pasted cURL command and restores it on undo', async () => {
    vi.useFakeTimers()
    state().ensureTab('tab')
    state().setName('tab', 'Kept')
    state().setUrl('tab', 'https://old.test')
    const parsed = { ...emptyDraft(), name: 'ignored', method: 'PUT', url: 'https://new.test', headers: [row('A', '1')], settings: { ...emptyDraft().settings, verifySsl: false } }
    vi.mocked(requests.parseCurl).mockResolvedValue({ draft: parsed, notices: ['n'] })
    await state().pasteCurl('tab', 'curl -X PUT https://new.test')
    expect(draft('tab')).toMatchObject({ name: 'Kept', method: 'PUT', url: 'https://new.test', headers: [row('A', '1')] })
    expect(draft('tab').settings.verifySsl).toBe(false)
    expect(state().tabs.tab!.curlPaste).toMatchObject({ notices: ['n'], wasDirty: true })
    expect(state().tabs.tab!.dirty).toBe(true)
    state().undoCurlPaste('tab')
    expect(draft('tab')).toMatchObject({ name: 'Kept', method: 'GET', url: 'https://old.test', headers: [] })
    expect(state().tabs.tab!.curlPaste).toBeNull()
    state().undoCurlPaste('tab')
    expect(draft('tab').url).toBe('https://old.test')
  })

  it('forgets the undo after the next edit or a dismiss, and keeps a clean tab clean on undo', async () => {
    vi.useFakeTimers()
    state().ensureTab('tab')
    vi.mocked(requests.parseCurl).mockResolvedValue({ draft: { ...emptyDraft(), url: 'https://new.test' }, notices: [] })
    await state().pasteCurl('tab', 'curl https://new.test')
    state().setUrl('tab', 'https://edited.test')
    expect(state().tabs.tab!.curlPaste).toBeNull()
    await state().pasteCurl('tab', 'curl https://new.test')
    state().dismissCurlPaste('tab')
    expect(state().tabs.tab!.curlPaste).toBeNull()
    state().loadDraft('clean', emptyDraft(), 'rev', false)
    await state().pasteCurl('clean', 'curl https://new.test')
    state().undoCurlPaste('clean')
    expect(state().tabs.clean!.dirty).toBe(false)
  })

  it('keeps the draft and shows the error when a pasted cURL command is refused', async () => {
    state().ensureTab('tab')
    vi.mocked(requests.parseCurl).mockRejectedValue(new AppError('IMP-41400', 'The cURL command could not be read'))
    await state().pasteCurl('tab', "curl 'x")
    expect(state().tabs.tab!.error?.code).toBe('IMP-41400')
    expect(state().tabs.tab!.curlPaste).toBeNull()
  })
})
