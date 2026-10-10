import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '../mappers/request'
import { AppError } from '../errors'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/requestservice', () => ({
  Get: vi.fn(),
  Save: vi.fn(),
  SaveAs: vi.fn(),
  ApplyUrl: vi.fn(),
  ApplyParams: vi.fn(),
  ParseBulk: vi.fn(),
  FormatBulk: vi.fn(),
  AutoHeaders: vi.fn(),
  Send: vi.fn(),
  LastResult: vi.fn(),
  ReleaseResponse: vi.fn(),
  SaveResponseBody: vi.fn(),
}))

const RequestService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/requestservice')
const requests = await import('./requests')
const { toRequestDto } = await import('../mappers/request')

const timings = { dns_ms: 0, connect_ms: 0, tls_ms: 0, first_byte_ms: 0, download_ms: 0, total_ms: 3 }
const sendResult = {
  tab_id: 'tab',
  ok: true,
  status_code: 200,
  status_text: 'OK',
  proto: 'HTTP/1.1',
  headers: [],
  size_bytes: 2,
  content_type: 'text/plain',
  is_binary: false,
  body_url: '/xenon/response/tab',
  inline_body: 'hi',
  timings,
}

function failure(code: string) {
  return Object.assign(new Error('failed'), { cause: { error: { code, description: 'd' } } })
}

describe('request service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('applies the URL and the params through Go', async () => {
    const draft = { ...emptyDraft(), url: 'https://h?a=1' }
    vi.mocked(RequestService.ApplyUrl).mockResolvedValue({ ...toRequestDto(draft), params: [{ key: 'a', value: '1', description: '' }] } as never)
    vi.mocked(RequestService.ApplyParams).mockResolvedValue({ ...toRequestDto(draft), url: 'https://h?b=2' } as never)
    await expect(requests.applyUrl(draft)).resolves.toMatchObject({ params: [{ key: 'a', value: '1', enabled: true }] })
    await expect(requests.applyParams(draft)).resolves.toMatchObject({ url: 'https://h?b=2' })
    expect(RequestService.ApplyUrl).toHaveBeenCalledWith({ request: toRequestDto(draft) })
  })

  it('opens and saves requests with their revision', async () => {
    const draft = { ...emptyDraft(), id: 'r1', name: 'Ping', url: 'https://h' }
    const dto = { ...toRequestDto(draft), revision: 'rev-1' }
    vi.mocked(RequestService.Get).mockResolvedValue(dto as never)
    vi.mocked(RequestService.Save).mockResolvedValue({ ...dto, revision: 'rev-2' } as never)
    vi.mocked(RequestService.SaveAs).mockResolvedValue({ ...dto, id: 'r2', revision: 'rev-3' } as never)
    await expect(requests.getRequest('ws', 'r1')).resolves.toMatchObject({ revision: 'rev-1', draft: { id: 'r1', url: 'https://h' } })
    await expect(requests.saveRequest('ws', draft, 'rev-1', false)).resolves.toMatchObject({ revision: 'rev-2' })
    await expect(requests.saveRequestAs('ws', 'c1', 'Copy', draft)).resolves.toMatchObject({ revision: 'rev-3', draft: { id: 'r2' } })
    expect(RequestService.Get).toHaveBeenCalledWith({ local_id: 'ws', id: 'r1' })
    expect(RequestService.Save).toHaveBeenCalledWith({ local_id: 'ws', base_revision: 'rev-1', force: false, request: toRequestDto(draft) })
    expect(RequestService.SaveAs).toHaveBeenCalledWith({ local_id: 'ws', parent_id: 'c1', name: 'Copy', request: toRequestDto(draft) })
  })

  it('rejects an empty answer', async () => {
    vi.mocked(RequestService.ApplyUrl).mockResolvedValue(null as never)
    await expect(requests.applyUrl(emptyDraft())).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('parses and formats bulk text', async () => {
    vi.mocked(RequestService.ParseBulk).mockResolvedValue({ rows: [{ key: 'a', value: '1', description: '', enabled: false }] } as never)
    vi.mocked(RequestService.FormatBulk).mockResolvedValue({ text: 'a:1' } as never)
    await expect(requests.parseBulk('//a:1')).resolves.toMatchObject([{ key: 'a', enabled: false }])
    await expect(requests.formatBulk([{ id: 'r', key: 'a', value: '1', description: '', enabled: true }])).resolves.toBe('a:1')
    vi.mocked(RequestService.ParseBulk).mockResolvedValue({ rows: null } as never)
    await expect(requests.parseBulk('')).resolves.toEqual([])
  })

  it('lists auto headers', async () => {
    vi.mocked(RequestService.AutoHeaders).mockResolvedValue([{ key: 'Accept', value: '*/*', disabled: false }] as never)
    await expect(requests.autoHeaders('ws', emptyDraft())).resolves.toEqual([{ key: 'Accept', value: '*/*', disabled: false }])
    vi.mocked(RequestService.AutoHeaders).mockResolvedValue(null as never)
    await expect(requests.autoHeaders('ws', emptyDraft())).resolves.toEqual([])
  })

  it('sends and keeps the cancellable call', async () => {
    const cancel = vi.fn()
    vi.mocked(RequestService.Send).mockReturnValue(Object.assign(Promise.resolve(sendResult), { cancel }) as never)
    const pending = requests.sendRequest('ws', 'tab', emptyDraft(), { confirmedEnvironmentId: '', missingSecrets: false })
    await expect(pending.result).resolves.toMatchObject({ statusCode: 200, inlineBody: 'hi' })
    pending.cancel()
    expect(cancel).toHaveBeenCalled()
    expect(RequestService.Send).toHaveBeenCalledWith({
      local_id: 'ws',
      tab_id: 'tab',
      request: toRequestDto(emptyDraft()),
      acknowledge_missing_secrets: false,
      confirmed_environment_id: '',
    })
    requests.sendRequest('ws', 'tab', emptyDraft(), { confirmedEnvironmentId: 'env-1', missingSecrets: true })
    expect(RequestService.Send).toHaveBeenLastCalledWith(
      expect.objectContaining({ acknowledge_missing_secrets: true, confirmed_environment_id: 'env-1' }),
    )
  })

  it('maps send failures to AppError', async () => {
    vi.mocked(RequestService.Send).mockReturnValue(Object.assign(Promise.reject(failure('SND-00400')), { cancel: vi.fn() }) as never)
    await expect(requests.sendRequest('ws', 'tab', emptyDraft(), { confirmedEnvironmentId: '', missingSecrets: false }).result).rejects.toMatchObject({ code: 'SND-00400' })
  })

  it('returns the last result or null when the tab has none', async () => {
    vi.mocked(RequestService.LastResult).mockResolvedValueOnce(sendResult as never)
    await expect(requests.lastResult('tab')).resolves.toMatchObject({ tabId: 'tab' })
    vi.mocked(RequestService.LastResult).mockRejectedValueOnce(failure('SND-03404'))
    await expect(requests.lastResult('tab')).resolves.toBeNull()
    vi.mocked(RequestService.LastResult).mockRejectedValueOnce(failure('SND-01400'))
    await expect(requests.lastResult('')).rejects.toMatchObject({ code: 'SND-01400' })
  })

  it('releases and saves response bodies', async () => {
    vi.mocked(RequestService.ReleaseResponse).mockResolvedValue(undefined as never)
    vi.mocked(RequestService.SaveResponseBody).mockRejectedValue(failure('SND-04400'))
    await requests.releaseResponse('tab')
    expect(RequestService.ReleaseResponse).toHaveBeenCalledWith({ tab_id: 'tab' })
    await expect(requests.saveResponseBody('tab', 'x')).rejects.toMatchObject({ code: 'SND-04400' })
    expect(RequestService.SaveResponseBody).toHaveBeenCalledWith({ tab_id: 'tab', path: 'x' })
  })

  it('reads the missing and locked secrets of a refused send', () => {
    const refused = new AppError('SND-06409', 'secrets missing', [
      { attribute: 'missing_secrets', messages: ['api_key'] },
      { attribute: 'locked_secrets', messages: ['token', 'password'] },
    ])
    expect(requests.missingSecretsOf(refused)).toEqual({ missing: ['api_key'], locked: ['token', 'password'] })
    expect(requests.missingSecretsOf(new AppError('SND-06409', 'only locked', [{ attribute: 'locked_secrets', messages: ['t'] }]))).toEqual({
      missing: [],
      locked: ['t'],
    })
    expect(requests.missingSecretsOf(new AppError('SND-00400', 'invalid'))).toBeNull()
  })

  it('reads the environment a send must be confirmed for', () => {
    const refused = new AppError('SND-05409', 'confirm', [
      { attribute: 'environment_id', messages: ['01P'] },
      { attribute: 'name', messages: ['prod'] },
      { attribute: 'color', messages: ['red'] },
    ])
    expect(requests.confirmationOf(refused)).toEqual({ environmentId: '01P', name: 'prod', color: 'red' })
    expect(requests.confirmationOf(new AppError('SND-05409', 'no details'))).toEqual({ environmentId: '', name: '', color: '' })
    expect(requests.confirmationOf(new AppError('SND-06409', 'secrets'))).toBeNull()
  })
})
