import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('@/platform/services/requests', () => ({ lastResult: vi.fn(), releaseResponse: vi.fn(), saveResponseBody: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickSaveFile: vi.fn() }))
vi.mock('@/platform/response-body', () => ({ fetchResponseText: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const responseBody = await import('@/platform/response-body')
const { useResponseViewerStore } = await import('./store')

const result: SendResult = {
  tabId: 'tab',
  ok: true,
  statusCode: 200,
  statusText: 'OK',
  proto: 'HTTP/1.1',
  headers: [],
  sizeBytes: 8,
  contentType: 'application/json',
  isBinary: false,
  bodyUrl: '/xenon/response/tab',
  inlineBody: null,
  timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 1 },
  error: null,
}

function entry() {
  return useResponseViewerStore.getState().entries.tab
}

describe('response viewer store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useResponseViewerStore.getState().reset()
  })

  it('rehydrates a tab from the last result kept by Go, once', async () => {
    vi.mocked(requests.lastResult).mockResolvedValue({ ...result, inlineBody: '{}' })
    await useResponseViewerStore.getState().rehydrate('tab')
    expect(entry()).toMatchObject({ version: 1, body: '{}', bodyStatus: 'ready' })
    await useResponseViewerStore.getState().rehydrate('tab')
    expect(requests.lastResult).toHaveBeenCalledTimes(1)
  })

  it('leaves the tab empty when Go has no result and reports failures', async () => {
    vi.mocked(requests.lastResult).mockResolvedValueOnce(null).mockRejectedValueOnce(new AppError('SND-01400', 'x'))
    await useResponseViewerStore.getState().rehydrate('tab')
    expect(entry()).toBeUndefined()
    await useResponseViewerStore.getState().rehydrate('tab')
    expect(entry()?.failure?.code).toBe('SND-01400')
  })

  it('loads large bodies from the response route once', async () => {
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('{"a":1}')
    useResponseViewerStore.getState().setResult('tab', result)
    expect(entry()).toMatchObject({ bodyStatus: 'idle', body: null })
    await useResponseViewerStore.getState().loadBody('tab')
    expect(entry()).toMatchObject({ bodyStatus: 'ready', body: '{"a":1}' })
    await useResponseViewerStore.getState().loadBody('tab')
    expect(responseBody.fetchResponseText).toHaveBeenCalledTimes(1)
    await useResponseViewerStore.getState().loadBody('none')
  })

  it('reports a body that can no longer be loaded', async () => {
    vi.mocked(responseBody.fetchResponseText).mockRejectedValue(new AppError('SND-03404', 'gone'))
    useResponseViewerStore.getState().setResult('tab', result)
    await useResponseViewerStore.getState().loadBody('tab')
    expect(entry()).toMatchObject({ bodyStatus: 'error', failure: { code: 'SND-03404' } })
  })

  it('tracks sending, results and failures', () => {
    const store = useResponseViewerStore.getState()
    store.startSending('tab')
    expect(entry()).toMatchObject({ sending: true, result: null })
    store.setFailure('tab', new AppError('SND-00400', 'x'))
    expect(entry()).toMatchObject({ sending: false, failure: { code: 'SND-00400' } })
    store.setResult('tab', { ...result, inlineBody: 'x' })
    store.setResult('tab', { ...result, inlineBody: 'y' })
    expect(entry()).toMatchObject({ version: 2, body: 'y', failure: null })
  })

  it('saves the body where the user chooses', async () => {
    vi.mocked(dialogs.pickSaveFile).mockResolvedValueOnce('/out/response.json').mockResolvedValueOnce(null)
    await useResponseViewerStore.getState().saveBody('tab')
    expect(dialogs.pickSaveFile).not.toHaveBeenCalled()
    useResponseViewerStore.getState().setResult('tab', result)
    await useResponseViewerStore.getState().saveBody('tab')
    expect(dialogs.pickSaveFile).toHaveBeenCalledWith('Save response body', 'response.json')
    expect(requests.saveResponseBody).toHaveBeenCalledWith('tab', '/out/response.json')
    await useResponseViewerStore.getState().saveBody('tab')
    expect(requests.saveResponseBody).toHaveBeenCalledTimes(1)
    vi.mocked(dialogs.pickSaveFile).mockRejectedValueOnce(new AppError('DLG-03500', 'x'))
    await useResponseViewerStore.getState().saveBody('tab')
    expect(entry()?.failure?.code).toBe('DLG-03500')
  })

  it('releases the response of a tab', async () => {
    useResponseViewerStore.getState().setResult('tab', result)
    await useResponseViewerStore.getState().release('tab')
    expect(entry()).toBeUndefined()
    expect(requests.releaseResponse).toHaveBeenCalledWith('tab')
  })
})
