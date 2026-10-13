import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '../mappers/request'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/treeservice', () => ({
  GetTree: vi.fn(),
  Search: vi.fn(),
  QuickSearch: vi.fn(),
  CreateCollection: vi.fn(),
  CreateFolder: vi.fn(),
  CreateRequest: vi.fn(),
  Rename: vi.fn(),
  Duplicate: vi.fn(),
  Delete: vi.fn(),
  GetSettings: vi.fn(),
  SaveSettings: vi.fn(),
  GetDocumentation: vi.fn(),
  PreviewDocumentation: vi.fn(),
  PlanDocumentation: vi.fn(),
  ApplyDocumentation: vi.fn(),
}))

const TreeService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/treeservice')
const tree = await import('./tree')
const { toRequestDto } = await import('../mappers/request')

const requestNode = { id: 'r1', kind: 'request', name: 'Ping', method: 'GET', url: 'https://x/ping' }
const collectionNode = { id: 'c1', kind: 'collection', name: 'Billing', children: [requestNode, { id: 'error:x', kind: 'error', name: 'x.yaml', error: 'cannot be read' }] }
const settings = {
  id: 'c1',
  kind: 'collection',
  name: 'Billing',
  description: 'docs',
  auth: { type: 'bearer', bearer: { token: 't' }, basic: { username: '', password: '' } },
  headers: [{ key: 'X-Tenant', value: 'acme', description: '', enabled: true }],
}

describe('tree service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('maps the tree with error and warning nodes', async () => {
    vi.mocked(TreeService.GetTree).mockResolvedValue([collectionNode] as never)
    const nodes = await tree.getTree('ws')
    expect(TreeService.GetTree).toHaveBeenCalledWith({ local_id: 'ws' })
    expect(nodes).toEqual([
      {
        id: 'c1',
        kind: 'collection',
        name: 'Billing',
        method: null,
        url: null,
        error: null,
        warning: null,
        children: [
          { id: 'r1', kind: 'request', name: 'Ping', method: 'GET', url: 'https://x/ping', error: null, warning: null, children: [], examples: [] },
          { id: 'error:x', kind: 'error', name: 'x.yaml', method: null, url: null, error: 'cannot be read', warning: null, children: [], examples: [] },
        ],
        examples: [],
      },
    ])
  })

  it('maps the examples of a request', async () => {
    vi.mocked(TreeService.GetTree).mockResolvedValue([{ ...requestNode, examples: [{ id: 'e1', kind: 'example', name: 'Created', code: 201 }] }] as never)
    const nodes = await tree.getTree('ws')
    expect(nodes[0]?.examples).toEqual([{ id: 'e1', name: 'Created', code: 201 }])
  })

  it('treats an empty tree and unknown kinds safely', async () => {
    vi.mocked(TreeService.GetTree).mockResolvedValue(null as never)
    await expect(tree.getTree('ws')).resolves.toEqual([])
    vi.mocked(TreeService.GetTree).mockResolvedValue([{ id: 'z', kind: 'weird', name: 'z' }] as never)
    await expect(tree.getTree('ws')).resolves.toMatchObject([{ kind: 'error' }])
  })

  it('searches the workspace', async () => {
    vi.mocked(TreeService.Search).mockResolvedValue({ match_ids: ['r1'], ancestor_ids: null } as never)
    await expect(tree.searchTree('ws', 'pi')).resolves.toEqual({ matchIds: ['r1'], ancestorIds: [] })
    expect(TreeService.Search).toHaveBeenCalledWith({ local_id: 'ws', query: 'pi' })
  })

  it('creates, renames and duplicates items', async () => {
    vi.mocked(TreeService.CreateCollection).mockResolvedValue(collectionNode as never)
    vi.mocked(TreeService.CreateFolder).mockResolvedValue({ id: 'f1', kind: 'folder', name: 'F' } as never)
    vi.mocked(TreeService.CreateRequest).mockResolvedValue(requestNode as never)
    vi.mocked(TreeService.Rename).mockResolvedValue({ ...requestNode, name: 'Pong' } as never)
    vi.mocked(TreeService.Duplicate).mockResolvedValue({ ...requestNode, id: 'r2', name: 'Ping Copy' } as never)
    await expect(tree.createCollection('ws', 'Billing')).resolves.toMatchObject({ id: 'c1' })
    await expect(tree.createFolder('ws', 'c1', 'F')).resolves.toMatchObject({ id: 'f1', kind: 'folder' })
    await expect(tree.createRequest('ws', 'c1', 'Ping')).resolves.toMatchObject({ id: 'r1' })
    const draft = { ...emptyDraft(), url: 'https://x' }
    await tree.createRequest('ws', 'c1', 'Ping', draft)
    await expect(tree.renameItem('ws', 'r1', 'Pong')).resolves.toMatchObject({ name: 'Pong' })
    await expect(tree.duplicateItem('ws', 'r1')).resolves.toMatchObject({ id: 'r2' })
    expect(TreeService.CreateCollection).toHaveBeenCalledWith({ local_id: 'ws', name: 'Billing' })
    expect(TreeService.CreateFolder).toHaveBeenCalledWith({ local_id: 'ws', parent_id: 'c1', name: 'F' })
    expect(TreeService.CreateRequest).toHaveBeenNthCalledWith(1, { local_id: 'ws', parent_id: 'c1', name: 'Ping', request: null })
    expect(TreeService.CreateRequest).toHaveBeenNthCalledWith(2, { local_id: 'ws', parent_id: 'c1', name: 'Ping', request: toRequestDto(draft) })
    expect(TreeService.Rename).toHaveBeenCalledWith({ local_id: 'ws', id: 'r1', name: 'Pong' })
    expect(TreeService.Duplicate).toHaveBeenCalledWith({ local_id: 'ws', id: 'r1' })
  })

  it('deletes and returns the deleted requests', async () => {
    vi.mocked(TreeService.Delete).mockResolvedValue({ deleted_request_ids: ['r1'] } as never)
    await expect(tree.deleteItem('ws', 'c1')).resolves.toEqual(['r1'])
    vi.mocked(TreeService.Delete).mockResolvedValue({ deleted_request_ids: null } as never)
    await expect(tree.deleteItem('ws', 'c1')).resolves.toEqual([])
  })

  it('reads and writes item settings', async () => {
    vi.mocked(TreeService.GetSettings).mockResolvedValue(settings as never)
    vi.mocked(TreeService.SaveSettings).mockResolvedValue({ ...settings, kind: 'folder', headers: null } as never)
    const loaded = await tree.getItemSettings('ws', 'c1')
    expect(loaded).toMatchObject({ id: 'c1', kind: 'collection', description: 'docs', auth: { type: 'bearer', bearer: { token: 't' } } })
    expect(loaded.headers).toMatchObject([{ key: 'X-Tenant', value: 'acme', enabled: true }])
    await expect(tree.saveItemSettings('ws', loaded)).resolves.toMatchObject({ kind: 'folder', headers: [] })
    expect(TreeService.SaveSettings).toHaveBeenCalledWith({
      local_id: 'ws',
      id: 'c1',
      description: 'docs',
      auth: expect.objectContaining(settings.auth),
      headers: [{ key: 'X-Tenant', value: 'acme', description: '', enabled: true }],
      scripts: { pre_request: '', post_response: '', disabled: [] },
    })
    await tree.saveItemSettings('ws', { ...loaded, kind: 'folder', inheritScripts: false })
    expect(TreeService.SaveSettings).toHaveBeenLastCalledWith(expect.objectContaining({ inherit_scripts: false }))
  })

  it('turns failures and empty answers into app errors', async () => {
    vi.mocked(TreeService.Rename).mockRejectedValue(Object.assign(new Error('x'), { cause: { error: { code: 'TRE-02404', description: 'gone' } } }))
    await expect(tree.renameItem('ws', 'r1', 'x')).rejects.toMatchObject({ code: 'TRE-02404', description: 'gone' })
    vi.mocked(TreeService.Duplicate).mockResolvedValue(null as never)
    await expect(tree.duplicateItem('ws', 'r1')).rejects.toMatchObject({ code: 'UNK-00204' })
  })

  it('runs a quick search and maps the hits', async () => {
    vi.mocked(TreeService.QuickSearch).mockResolvedValue([{ id: 'r1', name: 'Ping', method: 'GET', url: 'https://x/ping', path: 'Billing / Health' }] as never)
    await expect(tree.quickSearch('ws', 'pi', 20)).resolves.toEqual([{ id: 'r1', name: 'Ping', method: 'GET', url: 'https://x/ping', path: 'Billing / Health' }])
    expect(TreeService.QuickSearch).toHaveBeenCalledWith({ local_id: 'ws', query: 'pi', limit: 20 })
    vi.mocked(TreeService.QuickSearch).mockResolvedValue(null as never)
    await expect(tree.quickSearch('ws', 'zz', 20)).resolves.toEqual([])
    vi.mocked(TreeService.QuickSearch).mockRejectedValue(Object.assign(new Error('x'), { cause: { error: { code: 'TRE-40400', description: 'bad' } } }))
    await expect(tree.quickSearch('', 'x', 20)).rejects.toMatchObject({ code: 'TRE-40400' })
  })

  describe('documentation', () => {
    const documentDto = {
      title: 'Billing',
      description: '',
      version: '1.0.0',
      servers: ['{{host}}'],
      tags: null,
      operations: [
        {
          request_body: { media_type: 'application/json', example: '{}', fields: null, binary: false },
          request_id: 'r1',
          tag: 'Invoices',
          method: 'POST',
          path: '/invoices/{id}',
          server: '{{host}}',
          summary: 'Create',
          description: '',
          parameters: [{ name: 'id', in: 'path', description: '', example: '7', required: true, variable: false }],
          responses: [{ code: 201, description: 'Created', examples: [{ id: 'e1', name: 'Created', media_type: 'application/json', value: '{}', binary: false }] }],
          security: null,
          variant: false,
        },
      ],
      security_schemes: [{ name: 'bearerAuth', type: 'http', scheme: 'bearer', in: '', param_name: '', flow: '', authorization_url: '', token_url: '', description: '', scopes: null }],
    }
    const item = { id: 'r1', name: 'Create', folder: '', method: 'POST', path: '', revision: 'rev' }

    it('generates the documentation and maps the rendered view', async () => {
      vi.mocked(TreeService.GetDocumentation).mockResolvedValue({ yaml: 'openapi: 3.1.0', document: documentDto } as never)
      const documentation = await tree.getDocumentation('ws', 'c1')
      expect(TreeService.GetDocumentation).toHaveBeenCalledWith({ local_id: 'ws', id: 'c1' })
      expect(documentation.yaml).toBe('openapi: 3.1.0')
      expect(documentation.document.tags).toEqual([])
      expect(documentation.document.operations[0]).toMatchObject({
        requestId: 'r1',
        body: { mediaType: 'application/json', example: '{}', fields: [], binary: false },
        parameters: [{ name: 'id', in: 'path', example: '7', required: true }],
        responses: [{ code: 201, examples: [{ id: 'e1', mediaType: 'application/json' }] }],
        security: [],
      })
      expect(documentation.document.securitySchemes[0]).toMatchObject({ name: 'bearerAuth', paramName: '', scopes: [] })
    })

    it('previews, plans and applies edited YAML', async () => {
      vi.mocked(TreeService.PreviewDocumentation).mockResolvedValue({ yaml: 'x', document: documentDto } as never)
      vi.mocked(TreeService.PlanDocumentation).mockResolvedValue({ name: 'Billing', collection_changed: false, updated: [item], created: null, deleted: null } as never)
      vi.mocked(TreeService.ApplyDocumentation).mockResolvedValue({
        name: 'Billing',
        collection_changed: true,
        updated: [item],
        created: [],
        deleted: [],
        yaml: 'y',
        document: documentDto,
      } as never)
      await expect(tree.previewDocumentation('ws', 'c1', 'x')).resolves.toMatchObject({ yaml: 'x' })
      expect(TreeService.PreviewDocumentation).toHaveBeenCalledWith({ local_id: 'ws', id: 'c1', yaml: 'x', confirm_deletions: false })
      await expect(tree.planDocumentation('ws', 'c1', 'x')).resolves.toEqual({ name: 'Billing', collectionChanged: false, updated: [item], created: [], deleted: [] })
      const applied = await tree.applyDocumentation('ws', 'c1', 'x', true)
      expect(TreeService.ApplyDocumentation).toHaveBeenCalledWith({ local_id: 'ws', id: 'c1', yaml: 'x', confirm_deletions: true })
      expect(applied).toMatchObject({ collectionChanged: true, updated: [item], documentation: { yaml: 'y' } })
    })

    it('turns a document problem into an AppError with its line and column', async () => {
      vi.mocked(TreeService.PlanDocumentation).mockRejectedValue({
        cause: { error: { code: 'DOC-01400', description: 'line 4, column 4: bad', error_details: [{ attribute: 'line', messages: ['4'] }] } },
      })
      await expect(tree.planDocumentation('ws', 'c1', 'x')).rejects.toMatchObject({ code: 'DOC-01400', details: [{ attribute: 'line', messages: ['4'] }] })
    })
  })
})
