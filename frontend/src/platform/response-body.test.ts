import { afterEach, describe, expect, it, vi } from 'vitest'
import { fetchResponseBlob, fetchResponseText } from './response-body'

describe('response body route', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('fetches the body as text or blob from the in-app route', async () => {
    const fetchMock = vi.fn(async () => new Response('hello', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(fetchResponseText('/xenon/response/tab')).resolves.toBe('hello')
    expect(fetchMock).toHaveBeenCalledWith('/xenon/response/tab')
    await expect((await fetchResponseBlob('/xenon/response/tab')).text()).resolves.toBe('hello')
  })

  it('fails when the body was released', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('', { status: 404 })))
    await expect(fetchResponseText('/xenon/response/tab')).rejects.toMatchObject({ code: 'SND-03404' })
  })
})
