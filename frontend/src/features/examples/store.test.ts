import { act } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Example } from '@/platform/mappers/example'
import { emptyDraft } from '@/platform/mappers/request'

vi.mock('@/platform/services/examples', () => ({
  getExample: vi.fn(),
  saveExample: vi.fn(),
  createExampleFromResponse: vi.fn(),
  renameExample: vi.fn(),
  duplicateExample: vi.fn(),
  deleteExample: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(async () => []),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => null), suggestVariables: vi.fn() }))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const examples = await import('@/platform/services/examples')
const responseBody = await import('@/platform/services/response-body')
const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useExamplesStore } = await import('./store')

function example(overrides: Partial<Example> = {}): Example {
  return {
    id: 'e1',
    requestId: 'r1',
    name: 'Created',
    revision: 'rev1',
    request: { ...emptyDraft(), id: '', name: 'Created', method: 'POST', url: '{{base_url}}/invoices' },
    response: {
      code: 201,
      status: 'Created',
      contentType: 'application/json',
      headers: [{ id: 'h1', key: 'X-Trace', value: 't', description: '', enabled: true }],
      body: '{"id":1}',
      bodyEncoding: 'text',
      bodyUrl: '',
      sizeBytes: 8,
      isBinary: false,
    },
    ...overrides,
  }
}

function store() {
  return useExamplesStore.getState()
}

describe('examples store', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useExamplesStore.getState().reset()
    useTabsStore.getState().reset()
    useRequestBuilderStore.getState().reset()
    vi.mocked(tree.getTree).mockResolvedValue([])
  })

  it('opens an example into the request editors and its response draft', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    await act(() => store().open('tab-1', 'r1', 'e1'))
    expect(examples.getExample).toHaveBeenCalledWith('ws', 'r1', 'e1')
    expect(useRequestBuilderStore.getState().tabs['tab-1']?.draft).toMatchObject({ id: '', name: 'Created', method: 'POST' })
    expect(store().tabs['tab-1']).toMatchObject({ status: 'ready', dirty: false, response: { code: 201 } })
    expect(store().isDirty('tab-1')).toBe(false)
  })

  it('reports a failed open', async () => {
    vi.mocked(examples.getExample).mockRejectedValue(new AppError('EXM-01404', 'gone'))
    await act(() => store().open('tab-1', 'r1', 'e1'))
    expect(store().tabs['tab-1']).toMatchObject({ status: 'error', error: { code: 'EXM-01404' } })
  })

  it('edits the response and saves the request draft and response with the revision', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    vi.mocked(examples.saveExample).mockResolvedValue(example({ revision: 'rev2', response: { ...example().response, code: 202 } }))
    await act(() => store().open('tab-1', 'r1', 'e1'))
    act(() => {
      store().setCode('tab-1', 202)
      store().setStatusText('tab-1', 'Accepted')
      store().setBody('tab-1', 'queued')
      useRequestBuilderStore.getState().setMethod('tab-1', 'put')
    })
    expect(store().isDirty('tab-1')).toBe(true)
    const saved = await act(() => store().save('tab-1'))
    expect(saved).toEqual({ ok: true, value: true })
    const [localId, requestId, sent, revision, force] = vi.mocked(examples.saveExample).mock.calls[0]!
    expect([localId, requestId, revision, force]).toEqual(['ws', 'r1', 'rev1', false])
    expect(sent.request.method).toBe('PUT')
    expect(sent.response).toMatchObject({ code: 202, status: 'Accepted', body: 'queued' })
    expect(store().tabs['tab-1']).toMatchObject({ dirty: false, example: { revision: 'rev2' } })
    expect(store().isDirty('tab-1')).toBe(false)
  })

  it('flags a revision conflict and keeps the edits', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    vi.mocked(examples.saveExample).mockRejectedValue(new AppError('REQ-02409', 'changed'))
    await act(() => store().open('tab-1', 'r1', 'e1'))
    act(() => store().setCode('tab-1', 500))
    const saved = await act(() => store().save('tab-1'))
    expect(saved.ok).toBe(false)
    expect(store().tabs['tab-1']).toMatchObject({ conflict: true, dirty: true })
    vi.mocked(examples.saveExample).mockResolvedValue(example())
    await act(() => store().save('tab-1', true))
    expect(vi.mocked(examples.saveExample).mock.calls[1]?.[4]).toBe(true)
    expect(store().tabs['tab-1']?.conflict).toBe(false)
  })

  it('edits headers in table and bulk mode', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    await act(() => store().open('tab-1', 'r1', 'e1'))
    act(() => store().setBulkHeaders('tab-1', true))
    expect(store().tabs['tab-1']?.bulkHeaders).toBe('X-Trace: t')
    act(() => store().setBulkHeadersText('tab-1', 'X-Trace: t\nRetry-After: 5'))
    act(() => store().setBulkHeaders('tab-1', false))
    expect(store().tabs['tab-1']?.bulkHeaders).toBeNull()
    expect(store().tabs['tab-1']?.response?.headers.map((row) => row.key)).toEqual(['X-Trace', 'Retry-After'])
  })

  it('loads a routed body read-only and keeps it out of the save', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example({ response: { ...example().response, body: null, bodyUrl: '/xenon/response/example:e1', sizeBytes: 8_000_000 } }))
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('big')
    vi.mocked(examples.saveExample).mockResolvedValue(example())
    await act(() => store().open('tab-1', 'r1', 'e1'))
    await act(() => store().loadRoutedBody('tab-1'))
    expect(responseBody.fetchResponseText).toHaveBeenCalledWith('/xenon/response/example:e1')
    expect(store().tabs['tab-1']).toMatchObject({ routedBody: 'big', bodyStatus: 'ready' })
    act(() => store().setBody('tab-1', 'ignored'))
    act(() => store().setCode('tab-1', 200))
    await act(() => store().save('tab-1'))
    expect(vi.mocked(examples.saveExample).mock.calls[0]?.[2].response.body).toBeNull()
  })

  it('creates an example from the dialog tab response and expands its request', async () => {
    vi.mocked(examples.createExampleFromResponse).mockResolvedValue(example({ id: 'e9' }))
    act(() => store().openSaveDialog('tab-1', 'r1'))
    expect(store().saveDialog).toEqual({ tabId: 'tab-1', requestId: 'r1' })
    const created = await act(() => store().createFromResponse('Created'))
    expect(created).toEqual({ ok: true, value: { requestId: 'r1', exampleId: 'e9' } })
    expect(examples.createExampleFromResponse).toHaveBeenCalledWith('ws', 'r1', 'tab-1', 'Created')
    expect(store().saveDialog).toBeNull()
    expect(useSidebarStore.getState().expanded.r1).toBe(true)
  })

  it('keeps the dialog open when creating fails', async () => {
    vi.mocked(examples.createExampleFromResponse).mockRejectedValue(new AppError('EXM-02404', 'no response'))
    act(() => store().openSaveDialog('tab-1', 'r1'))
    const created = await act(() => store().createFromResponse('Created'))
    expect(created).toMatchObject({ ok: false, error: { code: 'EXM-02404' } })
    expect(store().saveDialog).not.toBeNull()
  })

  it('adopts a rename of an open example', async () => {
    vi.mocked(examples.getExample).mockResolvedValueOnce(example()).mockResolvedValueOnce(example({ name: 'Renamed', revision: 'rev3' }))
    await act(() => store().open('tab-1', 'r1', 'e1'))
    await act(() => store().adoptRename('e1'))
    expect(useRequestBuilderStore.getState().tabs['tab-1']?.draft.name).toBe('Renamed')
    expect(store().tabs['tab-1']?.example?.revision).toBe('rev3')
  })

  it('opens example tabs with preview and pin rules and routes save and close', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    vi.mocked(examples.saveExample).mockResolvedValue(example())
    await act(() => useTabsStore.getState().openExamplePreview('r1', 'e1'))
    const [tab] = useTabsStore.getState().tabs
    expect(tab).toMatchObject({ kind: 'example', requestId: 'r1', exampleId: 'e1', pinned: false })
    expect(useTabsStore.getState().previewTabId).toBe(tab?.id)
    await act(() => useTabsStore.getState().openExamplePinned('r1', 'e1'))
    expect(useTabsStore.getState().tabs).toHaveLength(1)
    expect(useTabsStore.getState().tabs[0]?.pinned).toBe(true)
    const snapshot = useTabsStore.getState().snapshot()
    expect(snapshot.tabs[0]).toMatchObject({ kind: 'example', requestId: 'r1', exampleId: 'e1', draft: null })
    act(() => store().setCode(tab!.id, 299))
    await act(() => useTabsStore.getState().save())
    expect(examples.saveExample).toHaveBeenCalledTimes(1)
    act(() => useTabsStore.getState().handleExampleDeleted('e1'))
    expect(useTabsStore.getState().tabs).toHaveLength(0)
    expect(store().tabs[tab!.id]).toBeUndefined()
  })

  it('restores example tabs from the saved UI state', async () => {
    vi.mocked(examples.getExample).mockResolvedValue(example())
    await act(async () => {
      useTabsStore.getState().hydrate('ws', {
        tabs: [{ tabId: 'tab-x', kind: 'example', requestId: 'r1', exampleId: 'e1', pinned: true, baseRevision: null, draft: null }],
        activeTabId: 'tab-x',
        previewTabId: null,
        sidebar: { expanded: [], selectedId: null, view: 'collections', search: '' },
        layout: useWorkspacesStore.getState().layout,
      })
    })
    expect(useTabsStore.getState().tabs[0]).toMatchObject({ id: 'tab-x', kind: 'example', exampleId: 'e1' })
    expect(examples.getExample).toHaveBeenCalledWith('ws', 'r1', 'e1')
  })
})
