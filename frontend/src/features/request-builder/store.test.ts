import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(),
  sendRequest: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { SYNC_DELAY_MS, useRequestBuilderStore } = await import('./store')

function row(key: string, value: string, enabled = true) {
  return { id: `id-${key}`, key, value, description: '', enabled }
}

function state() {
  return useRequestBuilderStore.getState()
}

function draft(tabId: string): RequestDraft {
  return state().tabs[tabId]!.draft
}

const okResult: SendResult = {
  tabId: 'tab',
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
    vi.resetAllMocks()
    vi.useFakeTimers()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useResponseViewerStore.getState().reset()
    state().reset()
    tabId = state().activeTabId
  })

  afterEach(() => vi.useRealTimers())

  it('starts with one unsaved tab keyed by its id', () => {
    expect(tabId).toMatch(/^tab-/)
    expect(draft(tabId)).toEqual(emptyDraft())
    const other = state().openTab('tab-other')
    expect(other).toBe('tab-other')
    expect(state().activeTabId).toBe('tab-other')
    expect(Object.keys(state().tabs)).toHaveLength(2)
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
    state().updateBody(tabId, { formdata: [{ ...row('f', ''), type: 'file', src: '' }, { ...row('g', ''), type: 'file', src: '/keep' }] })
    await state().chooseBodyFile(tabId, 'id-f')
    expect(draft(tabId).body.formdata.map((r) => r.src)).toEqual(['/c.txt', '/keep'])
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
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', tabId, draft(tabId))
    expect(state().tabs[tabId]!.pending).toBeNull()
    expect(useResponseViewerStore.getState().entries[tabId]?.result?.statusCode).toBe(200)
  })

  it('cancels an in-flight send and shows it as cancelled', async () => {
    let reject: (error: unknown) => void = () => undefined
    const cancel = vi.fn(() => reject(new AppError('cancelled', 'The request was cancelled')))
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

  it('does not send without a workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    await state().send(tabId)
    expect(requests.sendRequest).not.toHaveBeenCalled()
    state().cancel(tabId)
  })
})
