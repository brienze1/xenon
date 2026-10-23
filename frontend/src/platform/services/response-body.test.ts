import { afterEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {}, Clipboard: { SetText: vi.fn() } }))

const { Clipboard } = await import('@wailsio/runtime')
const { copyResponseText, fetchResponseText } = await import('./response-body')

describe('response body route', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('fetches the body as text from the in-app route', async () => {
    const fetchMock = vi.fn(async () => new Response('hello', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(fetchResponseText('/xenon/response/tab')).resolves.toBe('hello')
    expect(fetchMock).toHaveBeenCalledWith('/xenon/response/tab')
  })

  it('fails when the body was released', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('', { status: 404 })))
    await expect(fetchResponseText('/xenon/response/tab')).rejects.toMatchObject({ code: 'SND-03404' })
  })

  it('maps a network failure to an AppError', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => Promise.reject(new TypeError('Failed to fetch'))))
    await expect(fetchResponseText('/xenon/response/tab')).rejects.toMatchObject({
      name: 'AppError',
      code: 'UNK-00500',
      description: 'Failed to fetch',
    })
  })

  it('copies the body through the native clipboard', async () => {
    vi.mocked(Clipboard.SetText).mockResolvedValue(undefined as never)
    await expect(copyResponseText('{"a":1}')).resolves.toBe(true)
    expect(Clipboard.SetText).toHaveBeenCalledWith('{"a":1}')
  })

  it('maps a clipboard refusal to an AppError', async () => {
    vi.mocked(Clipboard.SetText).mockRejectedValue(new Error('denied'))
    await expect(copyResponseText('x')).rejects.toMatchObject({ name: 'AppError', description: 'denied' })
  })
})
