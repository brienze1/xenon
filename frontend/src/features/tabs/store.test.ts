import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import { defaultUIState, type UIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn(), getHistoryEntry: vi.fn(), clearHistory: vi.fn() }))

const requests = await import('@/platform/services/requests')
const uiState = await import('@/platform/services/ui-state')
const history = await import('@/platform/services/history')
const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { PERSIST_DELAY_MS, useTabsStore } = await import('./store')

function tabs() {
  return useTabsStore.getState()
}

function builder() {
  return useRequestBuilderStore.getState()
}

function saved(id: string, name: string): RequestDraft {
  return { ...emptyDraft(), id, name, url: `https://x/${id}` }
}

function dirty(tabId: string): boolean {
  return builder().tabs[tabId]?.dirty ?? false
}

describe('tabs store', () => {
  beforeEach(() => {
    tabs().reset()
    builder().reset()
    useResponseViewerStore.getState().reset()
    useSidebarStore.getState().reset()
    vi.resetAllMocks()
    vi.useFakeTimers()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.applyUrl).mockImplementation(async (sent) => sent)
    vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: saved(id, `Request ${id}`), revision: `rev-${id}` }))
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    vi.mocked(history.listHistory).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabs().hydrate('ws', defaultUIState())
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('opens an empty pinned tab when the workspace has none', () => {
    expect(tabs().tabs).toHaveLength(1)
    expect(tabs().tabs[0]).toMatchObject({ requestId: null, pinned: true })
    expect(tabs().activeTabId).toBe(tabs().tabs[0]!.id)
  })

  it('replaces the preview tab on the next single click', async () => {
    await tabs().openPreview('A')
    const preview = tabs().previewTabId
    expect(preview).not.toBeNull()
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ requestId: 'A', pinned: false })
    expect(builder().tabs[preview!]!.draft.id).toBe('A')
    await tabs().openPreview('B')
    expect(tabs().previewTabId).toBe(preview)
    expect(tabs().tabs).toHaveLength(2)
    expect(tabs().tabs.find((tab) => tab.id === preview)?.requestId).toBe('B')
    expect(builder().tabs[preview!]!.draft.id).toBe('B')
    expect(tabs().activeTabId).toBe(preview)
  })

  it('pins the preview tab on double click and on the first edit', async () => {
    await tabs().openPinned('A')
    expect(tabs().previewTabId).toBeNull()
    expect(tabs().tabs.at(-1)).toMatchObject({ requestId: 'A', pinned: true })
    await tabs().openPreview('B')
    const preview = tabs().previewTabId!
    builder().setUrl(preview, 'https://edited')
    tabs().pinIfEdited()
    expect(tabs().previewTabId).toBeNull()
    expect(tabs().tabs.find((tab) => tab.id === preview)?.pinned).toBe(true)
    await tabs().openPreview('C')
    expect(tabs().tabs).toHaveLength(4)
  })

  it('pins a tab double-clicked in the strip', async () => {
    await tabs().openPreview('A')
    const preview = tabs().previewTabId!
    tabs().pin(preview)
    expect(tabs().previewTabId).toBeNull()
    expect(tabs().tabs.find((tab) => tab.id === preview)?.pinned).toBe(true)
  })

  it('focuses the tab that already shows a request', async () => {
    await tabs().openPinned('A')
    const first = tabs().activeTabId
    tabs().newTab()
    expect(tabs().activeTabId).not.toBe(first)
    await tabs().openPreview('A')
    await tabs().openPinned('A')
    expect(tabs().activeTabId).toBe(first)
    expect(tabs().tabs.filter((tab) => tab.requestId === 'A')).toHaveLength(1)
  })

  it('marks edited tabs dirty and clean again after saving', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    expect(dirty(tabId)).toBe(false)
    builder().setName(tabId, 'Renamed')
    expect(dirty(tabId)).toBe(true)
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...saved('A', 'Renamed') }, revision: 'rev-2' })
    await tabs().save()
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.objectContaining({ name: 'Renamed' }), 'rev-A', false)
    expect(dirty(tabId)).toBe(false)
  })

  it('closes a clean tab at once and activates its neighbour', async () => {
    const first = tabs().activeTabId!
    await tabs().openPinned('A')
    const second = tabs().activeTabId!
    tabs().requestClose(second)
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().tabs.map((tab) => tab.id)).toEqual([first])
    expect(tabs().activeTabId).toBe(first)
    expect(builder().tabs[second]).toBeUndefined()
    expect(requests.releaseResponse).toHaveBeenCalledWith(second)
  })

  it('asks before closing a dirty tab and follows the answer', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://changed')
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    await tabs().resolveClose('cancel')
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)

    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved('A', 'A'), revision: 'rev-2' })
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(requests.saveRequest).toHaveBeenCalledTimes(1)
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)

    await tabs().openPinned('B')
    const other = tabs().activeTabId!
    builder().setUrl(other, 'https://changed')
    tabs().requestClose(other)
    await tabs().resolveClose('discard')
    expect(tabs().tabs.some((tab) => tab.id === other)).toBe(false)
    expect(requests.saveRequest).toHaveBeenCalledTimes(1)
  })

  it('keeps the tab open when saving from the close prompt fails', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://changed')
    vi.mocked(requests.saveRequest).mockRejectedValue(new AppError('REQ-02409', 'changed on disk'))
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)
    expect(builder().tabs[tabId]!.conflict).toBe(true)
    expect(tabs().closePrompt).toBe(tabId)
    expect(tabs().closeError?.code).toBe('REQ-02409')
    await tabs().resolveClose('cancel')
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().closeError).toBeNull()
  })

  it('saves a new tab into a chosen collection and points the tab to it', async () => {
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://new')
    await tabs().save()
    expect(tabs().saveDialog).toEqual({ tabId, mode: 'save', closeAfter: false })
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: { ...saved('N', 'Users'), url: 'https://new' }, revision: 'rev-N' })
    await expect(tabs().confirmSave('collection-1', 'Users')).resolves.toBeNull()
    expect(requests.saveRequestAs).toHaveBeenCalledWith('ws', 'collection-1', 'Users', expect.objectContaining({ url: 'https://new' }))
    expect(tabs().saveDialog).toBeNull()
    expect(tabs().tabs[0]).toMatchObject({ id: tabId, requestId: 'N', pinned: true })
    expect(builder().tabs[tabId]!.draft.id).toBe('N')
    expect(builder().tabs[tabId]!.baseRevision).toBe('rev-N')
    expect(dirty(tabId)).toBe(false)
  })

  it('keeps the save dialog open and returns the error when saving fails', async () => {
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://new')
    await tabs().save()
    vi.mocked(requests.saveRequestAs).mockRejectedValue(new AppError('TRE-01404', 'no parent'))
    await expect(tabs().confirmSave('gone', 'Users')).resolves.toMatchObject({ code: 'TRE-01404' })
    expect(tabs().saveDialog).toEqual({ tabId, mode: 'save', closeAfter: false })
    expect(builder().tabs[tabId]!.error).toBeNull()
  })

  it('save as switches a saved tab to the copy', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    tabs().saveAs()
    expect(tabs().saveDialog).toEqual({ tabId, mode: 'saveAs', closeAfter: false })
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: saved('A2', 'A copy'), revision: 'rev-A2' })
    await tabs().confirmSave('folder-1', 'A copy')
    expect(tabs().tabs.find((tab) => tab.id === tabId)?.requestId).toBe('A2')
    tabs().cancelSave()
    expect(tabs().saveDialog).toBeNull()
  })

  it('closes a new tab after it was saved from the close prompt', async () => {
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://new')
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs().saveDialog).toEqual({ tabId, mode: 'save', closeAfter: true })
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: saved('N', 'New'), revision: 'rev-N' })
    await tabs().confirmSave('c', 'New')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
  })

  it('closes clean tabs of deleted requests and offers to keep dirty ones', async () => {
    await tabs().openPinned('A')
    const clean = tabs().activeTabId!
    await tabs().openPinned('B')
    const edited = tabs().activeTabId!
    builder().setUrl(edited, 'https://mine')
    tabs().handleDeleted(['A', 'B'])
    expect(tabs().tabs.some((tab) => tab.id === clean)).toBe(false)
    expect(tabs().notices[edited]).toBe('deleted')
    tabs().keepAsNew(edited)
    expect(tabs().notices[edited]).toBeUndefined()
    expect(tabs().tabs.find((tab) => tab.id === edited)).toMatchObject({ requestId: null, pinned: true })
    expect(builder().tabs[edited]!.draft).toMatchObject({ id: '', url: 'https://mine' })
    expect(dirty(edited)).toBe(true)

    await tabs().openPinned('C')
    const other = tabs().activeTabId!
    builder().setUrl(other, 'https://x')
    tabs().handleDeleted(['C'])
    tabs().closeTab(other)
    expect(tabs().tabs.some((tab) => tab.id === other)).toBe(false)
  })

  it('reloads clean tabs and flags dirty ones when a file changes on disk', async () => {
    await tabs().openPinned('A')
    const clean = tabs().activeTabId!
    await tabs().openPinned('B')
    const edited = tabs().activeTabId!
    builder().setUrl(edited, 'https://mine')
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved('A', 'From disk'), revision: 'rev-disk' })
    await tabs().handleChangedOnDisk({ localId: 'ws', requestId: 'A', deleted: false, revision: 'rev-disk' })
    expect(builder().tabs[clean]!.draft.name).toBe('From disk')
    expect(builder().tabs[clean]!.baseRevision).toBe('rev-disk')
    await tabs().handleChangedOnDisk({ localId: 'other', requestId: 'B', deleted: false, revision: 'x' })
    expect(tabs().notices[edited]).toBeUndefined()
    await tabs().handleChangedOnDisk({ localId: 'ws', requestId: 'B', deleted: false, revision: 'rev-new' })
    expect(tabs().notices[edited]).toBe('changed')

    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: saved('B', 'B'), revision: 'rev-mine' })
    await tabs().keepMine(edited)
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.objectContaining({ url: 'https://mine' }), 'rev-B', true)
    expect(tabs().notices[edited]).toBeUndefined()

    builder().setUrl(edited, 'https://again')
    await tabs().handleChangedOnDisk({ localId: 'ws', requestId: 'B', deleted: false, revision: 'rev-other' })
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: saved('B', 'Disk B'), revision: 'rev-other' })
    await tabs().loadFromDisk(edited)
    expect(builder().tabs[edited]!.draft.name).toBe('Disk B')
    expect(dirty(edited)).toBe(false)
    expect(tabs().notices[edited]).toBeUndefined()

    builder().setUrl(edited, 'https://again')
    await tabs().handleChangedOnDisk({ localId: 'ws', requestId: 'B', deleted: true, revision: '' })
    expect(tabs().notices[edited]).toBe('deleted')
  })

  it('reorders tabs by dragging', async () => {
    const first = tabs().activeTabId!
    await tabs().openPinned('A')
    await tabs().openPinned('B')
    const [, second, third] = tabs().tabs.map((tab) => tab.id)
    tabs().reorder(third!, first)
    expect(tabs().tabs.map((tab) => tab.id)).toEqual([third, first, second])
  })

  it('persists the tabs, drafts and sidebar 300 ms after the last change', async () => {
    const fresh = tabs().activeTabId!
    builder().setUrl(fresh, 'https://draft')
    await tabs().openPinned('A')
    await tabs().openPreview('B')
    useSidebarStore.setState({ expanded: { c1: true, c2: false }, selectedId: 'B', query: 'us' })
    tabs().schedulePersist()
    tabs().schedulePersist()
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS - 1)
    expect(uiState.saveUIState).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(uiState.saveUIState).toHaveBeenCalledTimes(1)
    const [localId, state] = vi.mocked(uiState.saveUIState).mock.calls[0]!
    expect(localId).toBe('ws')
    expect(state.tabs.map((tab) => [tab.requestId, tab.pinned])).toEqual([
      [null, true],
      ['A', true],
      ['B', false],
    ])
    expect(state.tabs[0]!.draft?.url).toBe('https://draft')
    expect(state.tabs[1]!.draft).toBeNull()
    expect(state.tabs[1]!.baseRevision).toBe('rev-A')
    expect(state.previewTabId).toBe(state.tabs[2]!.tabId)
    expect(state.activeTabId).toBe(state.tabs[2]!.tabId)
    expect(state.sidebar).toMatchObject({ expanded: ['c1'], selectedId: 'B', search: 'us' })
  })

  it('opens a request from history as a new unsaved tab', async () => {
    const before = tabs().tabs.length
    const draft = { ...emptyDraft(), name: 'From history', method: 'PUT', url: 'https://x/h' }
    const id = tabs().openDraft(draft)
    expect(tabs().tabs).toHaveLength(before + 1)
    expect(tabs().tabs.at(-1)).toEqual({ id, requestId: null, pinned: true })
    expect(tabs().activeTabId).toBe(id)
    expect(builder().tabs[id]!.draft).toEqual(draft)
    expect(builder().tabs[id]!.baseRevision).toBeNull()
    expect(dirty(id)).toBe(true)
    await tabs().flushPersist()
    const [, state] = vi.mocked(uiState.saveUIState).mock.calls[0]!
    expect(state.tabs.at(-1)).toMatchObject({ tabId: id, requestId: null, draft: { name: 'From history', url: 'https://x/h' } })
  })

  it('remembers which sidebar view is shown', async () => {
    vi.mocked(history.listHistory).mockResolvedValue([])
    tabs().reset()
    tabs().hydrate('ws', { ...defaultUIState(), sidebar: { expanded: [], selectedId: null, view: 'history', search: '', historySource: 'ai' } })
    await vi.advanceTimersByTimeAsync(0)
    expect(useSidebarStore.getState().view).toBe('history')
    expect(history.listHistory).toHaveBeenCalledWith('ws', { query: '', source: 'ai' })
    await tabs().flushPersist()
    expect(vi.mocked(uiState.saveUIState).mock.calls[0]![1].sidebar).toMatchObject({ view: 'history', historySource: 'ai' })
    useSidebarStore.getState().setView('collections')
    await tabs().flushPersist()
    expect(vi.mocked(uiState.saveUIState).mock.calls[1]![1].sidebar.view).toBe('collections')
  })

  it('flushes a pending save at once', async () => {
    tabs().schedulePersist()
    await tabs().flushPersist()
    expect(uiState.saveUIState).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(PERSIST_DELAY_MS)
    expect(uiState.saveUIState).toHaveBeenCalledTimes(1)
  })

  it('restores order, active and preview tabs, pinned flags and drafts', async () => {
    const restored: UIState = {
      ...defaultUIState(),
      tabs: [
        { tabId: 't-saved', requestId: 'A', pinned: true, baseRevision: 'rev-A', draft: null },
        { tabId: 't-new', requestId: null, pinned: true, baseRevision: null, draft: { ...emptyDraft(), url: 'https://unsaved' } },
        { tabId: 't-edited', requestId: 'B', pinned: true, baseRevision: 'rev-old', draft: { ...saved('B', 'B'), url: 'https://edit' } },
        { tabId: 't-preview', requestId: 'C', pinned: false, baseRevision: 'rev-C', draft: null },
      ],
      activeTabId: 't-new',
      previewTabId: 't-preview',
      sidebar: { expanded: ['c1'], selectedId: 'A', view: 'collections', search: 'q', historySource: 'user' },
    }
    tabs().reset()
    builder().reset()
    tabs().hydrate('ws', restored)
    await vi.advanceTimersByTimeAsync(0)
    expect(tabs().tabs.map((tab) => tab.id)).toEqual(['t-saved', 't-new', 't-edited', 't-preview'])
    expect(tabs().activeTabId).toBe('t-new')
    expect(tabs().previewTabId).toBe('t-preview')
    expect(builder().tabs['t-new']!.draft.url).toBe('https://unsaved')
    expect(dirty('t-new')).toBe(true)
    expect(builder().tabs['t-edited']!.draft.url).toBe('https://edit')
    expect(builder().tabs['t-edited']!.baseRevision).toBe('rev-old')
    expect(dirty('t-edited')).toBe(true)
    expect(builder().tabs['t-saved']!.draft.id).toBe('A')
    expect(dirty('t-saved')).toBe(false)
    expect(useSidebarStore.getState()).toMatchObject({ expanded: { c1: true }, selectedId: 'A', query: 'q' })
  })

  it('forgets everything on reset', async () => {
    await tabs().openPinned('A')
    tabs().reset()
    expect(tabs()).toMatchObject({ tabs: [], activeTabId: null, previewTabId: null, closePrompt: null, saveDialog: null, notices: {} })
  })
})

describe('tabs store: context menu, duplicates, closed tabs and positions', () => {
  beforeEach(() => {
    tabs().reset()
    builder().reset()
    useResponseViewerStore.getState().reset()
    useSidebarStore.getState().reset()
    vi.resetAllMocks()
    vi.useFakeTimers()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.applyUrl).mockImplementation(async (sent) => sent)
    vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: saved(id, `Request ${id}`), revision: `rev-${id}` }))
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    vi.mocked(history.listHistory).mockResolvedValue([])
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabs().hydrate('ws', defaultUIState())
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  function ids(): string[] {
    return tabs().tabs.map((tab) => tab.id)
  }

  async function openThree(): Promise<[string, string, string, string]> {
    const blank = tabs().activeTabId!
    await tabs().openPinned('A')
    const a = tabs().activeTabId!
    await tabs().openPinned('B')
    const b = tabs().activeTabId!
    await tabs().openPinned('C')
    const c = tabs().activeTabId!
    return [blank, a, b, c]
  }

  it('duplicates a tab as a second tab on the same request with its own draft', async () => {
    await tabs().openPinned('A')
    const first = tabs().activeTabId!
    builder().setUrl(first, 'https://edited')
    const copy = tabs().duplicateTab(first)
    expect(copy).not.toBeNull()
    expect(tabs().tabs.map((tab) => tab.requestId)).toEqual([null, 'A', 'A'])
    expect(tabs().tabs.at(-1)).toMatchObject({ id: copy, requestId: 'A', pinned: true })
    expect(tabs().activeTabId).toBe(copy)
    expect(builder().tabs[copy!]!.draft.url).toBe('https://edited')
    expect(builder().tabs[copy!]!.baseRevision).toBe('rev-A')
    expect(dirty(copy!)).toBe(true)
    builder().setUrl(copy!, 'https://other')
    expect(builder().tabs[first]!.draft.url).toBe('https://edited')
  })

  it('duplicates a clean tab as a clean second tab placed right after it', async () => {
    const [, a, b] = await openThree()
    const copy = tabs().duplicateTab(a)!
    expect(ids().indexOf(copy)).toBe(ids().indexOf(a) + 1)
    expect(ids().indexOf(b)).toBe(ids().indexOf(copy) + 1)
    expect(dirty(copy)).toBe(false)
    expect(builder().tabs[copy]!.draft.id).toBe('A')
  })

  it('shows the changed-on-disk banner in the other tab of a request after a save', async () => {
    await tabs().openPinned('A')
    const first = tabs().activeTabId!
    builder().setUrl(first, 'https://mine')
    const copy = tabs().duplicateTab(first)!
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...saved('A', 'Request A'), url: 'https://mine' }, revision: 'rev-2' })
    await tabs().save(first)
    expect(tabs().notices[copy]).toBe('changed')
    expect(tabs().notices[first]).toBeUndefined()
  })

  it('reloads the clean other tab of a request after a save', async () => {
    await tabs().openPinned('A')
    const first = tabs().activeTabId!
    const copy = tabs().duplicateTab(first)!
    builder().setUrl(first, 'https://mine')
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...saved('A', 'Request A'), url: 'https://mine' }, revision: 'rev-2' })
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: { ...saved('A', 'Request A'), url: 'https://mine' }, revision: 'rev-2' })
    await tabs().save(first)
    expect(tabs().notices[copy]).toBeUndefined()
    expect(builder().tabs[copy]!.draft.url).toBe('https://mine')
    expect(builder().tabs[copy]!.baseRevision).toBe('rev-2')
  })

  it('focuses the most recently active tab of a request when it is clicked in the sidebar', async () => {
    await tabs().openPinned('A')
    const first = tabs().activeTabId!
    const copy = tabs().duplicateTab(first)!
    tabs().activate(first)
    tabs().newTab()
    await tabs().openPreview('A')
    expect(tabs().activeTabId).toBe(first)
    tabs().activate(copy)
    tabs().newTab()
    await tabs().openPinned('A')
    expect(tabs().activeTabId).toBe(copy)
  })

  it('force closes a dirty tab without asking', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://changed')
    tabs().forceClose(tabId)
    expect(tabs().closePrompt).toBeNull()
    expect(ids()).not.toContain(tabId)
    expect(requests.saveRequest).not.toHaveBeenCalled()
  })

  it('closes the other tabs, asking about each dirty one in turn', async () => {
    const [blank, a, b, c] = await openThree()
    builder().setUrl(a, 'https://a')
    builder().setUrl(c, 'https://c')
    tabs().closeOthers(b)
    expect(ids()).toEqual([a, b, c])
    expect(tabs().closePrompt).toBe(a)
    await tabs().resolveClose('discard')
    expect(ids()).toEqual([b, c])
    expect(tabs().closePrompt).toBe(c)
    await tabs().resolveClose('discard')
    expect(ids()).toEqual([b])
    expect(tabs().closePrompt).toBeNull()
    expect(ids()).not.toContain(blank)
  })

  it('stops closing the remaining tabs when a prompt is cancelled', async () => {
    const [, a, b, c] = await openThree()
    builder().setUrl(a, 'https://a')
    builder().setUrl(b, 'https://b')
    tabs().closeAll()
    expect(ids()).toEqual([a, b])
    expect(tabs().closePrompt).toBe(a)
    await tabs().resolveClose('cancel')
    expect(tabs().closePrompt).toBeNull()
    expect(ids()).toEqual([a, b])
    expect(ids()).not.toContain(c)
  })

  it('closes the tabs to the left and to the right of a tab', async () => {
    const [blank, a, b, c] = await openThree()
    tabs().closeLeft(b)
    expect(ids()).toEqual([b, c])
    expect(ids()).not.toContain(blank)
    expect(ids()).not.toContain(a)
    tabs().closeRight(b)
    expect(ids()).toEqual([b])
    expect(tabs().activeTabId).toBe(b)
  })

  it('force closes all tabs without asking', async () => {
    const [, a] = await openThree()
    builder().setUrl(a, 'https://a')
    tabs().forceCloseAll()
    expect(ids()).toEqual([])
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().activeTabId).toBeNull()
  })

  it('opens a new request tab from the menu', () => {
    const before = ids().length
    const id = tabs().newTab()
    expect(ids()).toHaveLength(before + 1)
    expect(tabs().activeTabId).toBe(id)
  })

  it('reveals the request of a tab in the sidebar', async () => {
    await tabs().openPinned('A')
    const tabId = tabs().activeTabId!
    const request = { id: 'A', kind: 'request' as const, name: 'A', method: 'GET', url: null, error: null, warning: null, children: [] }
    const folder = { id: 'f1', kind: 'folder' as const, name: 'F', method: null, url: null, error: null, warning: null, children: [request] }
    useSidebarStore.setState({ tree: [{ id: 'c1', kind: 'collection', name: 'C', method: null, url: null, error: null, warning: null, children: [folder] }], view: 'history', expanded: {} })
    useWorkspacesStore.getState().setLayout({ sidebarVisible: false })
    tabs().revealInSidebar(tabId)
    expect(useSidebarStore.getState()).toMatchObject({ selectedId: 'A', view: 'collections', expanded: { c1: true, f1: true } })
    expect(useWorkspacesStore.getState().layout.sidebarVisible).toBe(true)
  })

  it('reopens closed tabs newest first, with their unsaved drafts', async () => {
    const [, a, b] = await openThree()
    builder().setUrl(a, 'https://unsaved-a')
    tabs().forceClose(a)
    tabs().closeTab(b)
    expect(tabs().closedTabs.map((closed) => closed.tab.requestId)).toEqual(['A', 'B'])
    await tabs().reopenClosed()
    expect(tabs().tabs.at(-1)).toMatchObject({ requestId: 'B', pinned: true })
    expect(dirty(tabs().activeTabId!)).toBe(false)
    await tabs().reopenClosed()
    const reopened = tabs().activeTabId!
    expect(tabs().tabs.find((tab) => tab.id === reopened)?.requestId).toBe('A')
    expect(builder().tabs[reopened]!.draft.url).toBe('https://unsaved-a')
    expect(builder().tabs[reopened]!.baseRevision).toBe('rev-A')
    expect(dirty(reopened)).toBe(true)
    expect(tabs().closedTabs).toEqual([])
    await tabs().reopenClosed()
    expect(tabs().tabs.filter((tab) => tab.requestId === 'A')).toHaveLength(1)
  })

  it('keeps the 20 most recently closed tabs and persists them', async () => {
    for (let index = 0; index < 22; index++) {
      const id = tabs().openDraft({ ...emptyDraft(), url: `https://x/${index}` })
      tabs().closeTab(id)
    }
    expect(tabs().closedTabs).toHaveLength(20)
    expect(tabs().closedTabs[0]!.tab.draft?.url).toBe('https://x/2')
    expect(tabs().closedTabs.at(-1)!.tab.draft?.url).toBe('https://x/21')
    await tabs().flushPersist()
    const [, state] = vi.mocked(uiState.saveUIState).mock.calls.at(-1)!
    expect(state.closedTabs).toHaveLength(20)
    expect(state.closedTabs.at(-1)!.closedAt).toMatch(/^\d{4}-\d{2}-\d{2}T/)
  })

  it('does not remember an empty untitled tab or the tabs of deleted requests', async () => {
    const blank = tabs().activeTabId!
    tabs().closeTab(blank)
    await tabs().openPinned('A')
    tabs().handleDeleted(['A'])
    expect(tabs().closedTabs).toEqual([])
  })

  it('reopens a tab closed in an earlier session', async () => {
    tabs().reset()
    builder().reset()
    tabs().hydrate('ws', {
      ...defaultUIState(),
      tabs: [{ tabId: 't1', requestId: 'A', pinned: true, baseRevision: 'rev-A', draft: null }],
      activeTabId: 't1',
      closedTabs: [{ closedAt: '2026-10-11T10:00:00.000Z', tab: { tabId: 'old', requestId: 'B', pinned: true, baseRevision: 'rev-B', draft: null } }],
    })
    await tabs().reopenClosed()
    expect(tabs().tabs.at(-1)).toMatchObject({ requestId: 'B' })
    expect(tabs().activeTabId).not.toBe('t1')
    await vi.advanceTimersByTimeAsync(0)
    expect(builder().tabs[tabs().activeTabId!]!.draft.id).toBe('B')
  })

  it('activates tabs 1 to 8 by position and 9 as the last tab', async () => {
    const [blank, a, , c] = await openThree()
    tabs().activateIndex(1)
    expect(tabs().activeTabId).toBe(blank)
    tabs().activateIndex(2)
    expect(tabs().activeTabId).toBe(a)
    tabs().activateIndex(8)
    expect(tabs().activeTabId).toBe(a)
    tabs().activateIndex(9)
    expect(tabs().activeTabId).toBe(c)
  })

  it('cycles to the next and previous tab, wrapping around', async () => {
    const [blank, , b, c] = await openThree()
    tabs().cycle(1)
    expect(tabs().activeTabId).toBe(blank)
    tabs().cycle(-1)
    expect(tabs().activeTabId).toBe(c)
    tabs().cycle(-1)
    expect(tabs().activeTabId).toBe(b)
  })
})

