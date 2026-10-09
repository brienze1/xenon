import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import { defaultUIState, type UIState } from '@/platform/mappers/ui-state'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
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
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn().mockResolvedValue([]) }))

const requests = await import('@/platform/services/requests')
const uiState = await import('@/platform/services/ui-state')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
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
    vi.resetAllMocks()
    vi.useFakeTimers()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.applyUrl).mockImplementation(async (sent) => sent)
    vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: saved(id, `Request ${id}`), revision: `rev-${id}` }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    builder().reset()
    tabs().reset()
    tabs().hydrate('ws', defaultUIState())
  })

  afterEach(() => vi.useRealTimers())

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
  })

  it('saves a new tab into a chosen collection and points the tab to it', async () => {
    const tabId = tabs().activeTabId!
    builder().setUrl(tabId, 'https://new')
    await tabs().save()
    expect(tabs().saveDialog).toEqual({ tabId, mode: 'save', closeAfter: false })
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: { ...saved('N', 'Users'), url: 'https://new' }, revision: 'rev-N' })
    await tabs().confirmSave('collection-1', 'Users')
    expect(requests.saveRequestAs).toHaveBeenCalledWith('ws', 'collection-1', 'Users', expect.objectContaining({ url: 'https://new' }))
    expect(tabs().saveDialog).toBeNull()
    expect(tabs().tabs[0]).toMatchObject({ id: tabId, requestId: 'N', pinned: true })
    expect(builder().tabs[tabId]!.draft.id).toBe('N')
    expect(builder().tabs[tabId]!.baseRevision).toBe('rev-N')
    expect(dirty(tabId)).toBe(false)
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
      sidebar: { expanded: ['c1'], selectedId: 'A', view: 'collections', search: 'q' },
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
