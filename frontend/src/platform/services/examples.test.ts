import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exampleservice', () => ({
  Get: vi.fn(),
  Create: vi.fn(),
  CreateFromResponse: vi.fn(),
  Save: vi.fn(),
  Rename: vi.fn(),
  Duplicate: vi.fn(),
  Delete: vi.fn(),
}))

const ExampleService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/exampleservice')
const examples = await import('./examples')

const body = {
  mode: 'raw',
  raw: { language: 'json', content: '{"a":1}' },
  formdata: [],
  urlencoded: [],
  binary: { src: '', src_state: 'ok' },
  graphql: { query: '', variables: '' },
}
const auth = { type: 'inherit', bearer: { token: '' }, basic: { username: '', password: '' } }
const exampleDto = {
  id: 'e1',
  request_id: 'r1',
  name: 'Created',
  revision: 'rev1',
  request: {
    method: 'POST',
    url: '{{base_url}}/invoices',
    params: [{ key: 'dry', value: '1', description: '', enabled: true }],
    path_variables: [],
    headers: [{ key: 'Content-Type', value: 'application/json', description: '' }],
    body,
    auth,
  },
  response: {
    code: 201,
    status: 'Created',
    content_type: 'application/json',
    headers: [{ key: 'X-Trace', value: 't', description: '', enabled: true }],
    body: '{"id":1}',
    body_encoding: 'text',
    body_url: '',
    size_bytes: 8,
    is_binary: false,
  },
}

describe('examples service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('maps an example with its request snapshot as a draft', async () => {
    vi.mocked(ExampleService.Get).mockResolvedValue(exampleDto as never)
    const example = await examples.getExample('ws', 'r1', 'e1')
    expect(ExampleService.Get).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', example_id: 'e1' })
    expect(example.name).toBe('Created')
    expect(example.revision).toBe('rev1')
    expect(example.request.id).toBe('')
    expect(example.request.name).toBe('Created')
    expect(example.request.method).toBe('POST')
    expect(example.request.params[0]).toMatchObject({ key: 'dry', value: '1', enabled: true })
    expect(example.request.headers[0]?.enabled).toBe(true)
    expect(example.request.body.raw.content).toBe('{"a":1}')
    expect(example.response).toMatchObject({ code: 201, status: 'Created', contentType: 'application/json', body: '{"id":1}', bodyEncoding: 'text', bodyUrl: '', sizeBytes: 8, isBinary: false })
    expect(example.response.headers[0]).toMatchObject({ key: 'X-Trace', value: 't' })
  })

  it('keeps a body served by the route as null', async () => {
    vi.mocked(ExampleService.Get).mockResolvedValue({
      ...exampleDto,
      response: { ...exampleDto.response, body: undefined, body_url: '/xenon/response/example:e1', body_encoding: 'base64', is_binary: true },
    } as never)
    const example = await examples.getExample('ws', 'r1', 'e1')
    expect(example.response.body).toBeNull()
    expect(example.response.bodyUrl).toBe('/xenon/response/example:e1')
    expect(example.response.bodyEncoding).toBe('base64')
  })

  it('saves the example with its revision and a null body when it was not loaded', async () => {
    vi.mocked(ExampleService.Get).mockResolvedValue(exampleDto as never)
    vi.mocked(ExampleService.Save).mockResolvedValue(exampleDto as never)
    const example = await examples.getExample('ws', 'r1', 'e1')
    await examples.saveExample('ws', 'r1', { ...example, response: { ...example.response, body: null } }, 'rev1', true)
    const request = vi.mocked(ExampleService.Save).mock.calls[0]?.[0]
    expect(request).toMatchObject({ local_id: 'ws', request_id: 'r1', base_revision: 'rev1', force: true })
    expect(request?.example).toMatchObject({ id: 'e1', name: 'Created', response: { code: 201, status: 'Created', body: null, body_encoding: 'text' } })
    expect(request?.example.request).toMatchObject({ method: 'POST', url: '{{base_url}}/invoices' })
  })

  it('creates, renames, duplicates and deletes', async () => {
    vi.mocked(ExampleService.Create).mockResolvedValue(exampleDto as never)
    vi.mocked(ExampleService.CreateFromResponse).mockResolvedValue(exampleDto as never)
    vi.mocked(ExampleService.Rename).mockResolvedValue(exampleDto as never)
    vi.mocked(ExampleService.Duplicate).mockResolvedValue(exampleDto as never)
    vi.mocked(ExampleService.Delete).mockResolvedValue(undefined as never)
    await examples.createExample('ws', 'r1', 'Draft')
    await examples.createExampleFromResponse('ws', 'r1', 'tab-1', 'Created')
    await examples.renameExample('ws', 'r1', 'e1', 'Renamed')
    await examples.duplicateExample('ws', 'r1', 'e1')
    await examples.deleteExample('ws', 'r1', 'e1')
    expect(ExampleService.Create).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', name: 'Draft' })
    expect(ExampleService.CreateFromResponse).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', tab_id: 'tab-1', name: 'Created' })
    expect(ExampleService.Rename).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', example_id: 'e1', name: 'Renamed' })
    expect(ExampleService.Duplicate).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', example_id: 'e1' })
    expect(ExampleService.Delete).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', example_id: 'e1' })
  })

  it('maps Go errors and empty responses to AppError', async () => {
    vi.mocked(ExampleService.Get).mockRejectedValue({ cause: { error: { code: 'EXM-01404', description: 'gone' } } })
    await expect(examples.getExample('ws', 'r1', 'e1')).rejects.toMatchObject({ code: 'EXM-01404' })
    vi.mocked(ExampleService.Create).mockResolvedValue(null as never)
    await expect(examples.createExample('ws', 'r1', 'x')).rejects.toMatchObject({ code: 'UNK-00204' })
  })
})
