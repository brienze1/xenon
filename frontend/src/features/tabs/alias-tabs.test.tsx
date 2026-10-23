import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft, type RequestDraft, type SavedRequest } from '@/platform/mappers/request'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(async (draft) => draft),
  applyParams: vi.fn(async (draft) => draft),
  autoHeaders: vi.fn(async () => []),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  releaseResponse: vi.fn(async () => undefined),
  lastResult: vi.fn(async () => null),
}))
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const requests = await import('@/platform/services/requests')
const aliasService = await import('@/platform/services/aliases')
const uiState = await import('@/platform/services/ui-state')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { TabStrip } = await import('./components/TabStrip')
const { CloseTabPrompt } = await import('./components/CloseTabPrompt')

function draftOf(name: string, url: string): RequestDraft {
  return { ...emptyDraft(), id: 'A', name, url, method: 'POST' }
}

const SAVED: SavedRequest = {
  draft: draftOf('Login', 'https://x/v1'),
  revision: 'rev-1',
  aliases: [{ id: 'a1', name: 'mx-test-user', environmentIds: [], draft: draftOf('mx-test-user', 'https://x/v2') }],
}

function builder() {
  return useRequestBuilderStore.getState()
}

function tabs() {
  return useTabsStore.getState()
}

function reset() {
  tabs().reset()
  builder().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

describe('request tabs with aliases', () => {
  beforeEach(() => {
    reset()
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockResolvedValue(SAVED)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabs().hydrate('ws', defaultUIState())
  })

  afterEach(() => reset())

  it('titles the tab with the alias and shows the dirty dot while a hidden version has edits', async () => {
    await act(() => tabs().openPinned('A'))
    const tabId = tabs().activeTabId!
    render(<TabStrip />)
    act(() => builder().setUrl(tabId, 'https://x/v1?edited'))
    act(() => builder().selectVersion(tabId, 'a1'))
    const tab = within(screen.getByRole('tablist', { name: 'Open requests' })).getAllByRole('tab').at(-1)!
    expect(tab).toHaveTextContent('Login · mx-test-user')
    expect(within(tab).getByRole('img', { name: 'Unsaved changes' })).toBeInTheDocument()
  })

  it('opening the request again focuses its tab on the alias it shows', async () => {
    await act(() => tabs().openPinned('A'))
    const tabId = tabs().activeTabId!
    act(() => builder().selectVersion(tabId, 'a1'))
    await act(() => tabs().openPreview('A'))
    expect(tabs().activeTabId).toBe(tabId)
    expect(builder().tabs[tabId]!.aliasId).toBe('a1')
  })

  it('Save in the close prompt saves every version with edits, then closes the tab', async () => {
    await act(() => tabs().openPinned('A'))
    const tabId = tabs().activeTabId!
    act(() => builder().setUrl(tabId, 'https://x/v1?edited'))
    act(() => builder().selectVersion(tabId, 'a1'))
    act(() => builder().setUrl(tabId, 'https://x/v2?edited'))
    vi.mocked(aliasService.saveAlias).mockResolvedValue({ ...SAVED, revision: 'rev-2', aliases: [{ id: 'a1', name: 'mx-test-user', environmentIds: [], draft: draftOf('mx-test-user', 'https://x/v2?edited') }] })
    vi.mocked(requests.saveRequest).mockResolvedValue({ ...SAVED, draft: draftOf('Login', 'https://x/v1?edited'), revision: 'rev-3' })
    render(
      <>
        <TabStrip />
        <CloseTabPrompt />
      </>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Close Login · mx-test-user' }))
    await userEvent.click(within(screen.getByRole('dialog', { name: 'Save Changes?' })).getByRole('button', { name: 'Save Changes' }))
    expect(aliasService.saveAlias).toHaveBeenCalledWith('ws', 'a1', expect.objectContaining({ url: 'https://x/v2?edited' }), 'rev-1', false)
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.objectContaining({ url: 'https://x/v1?edited' }), 'rev-2', false)
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
  })

  it('keeps the alias shown and the drafts of the other versions in the tab state', async () => {
    await act(() => tabs().openPinned('A'))
    const tabId = tabs().activeTabId!
    act(() => builder().setUrl(tabId, 'https://x/v1?edited'))
    act(() => builder().selectVersion(tabId, 'a1'))
    await tabs().flushPersist()
    const [, state] = vi.mocked(uiState.saveUIState).mock.calls.at(-1)!
    const saved = state.tabs.find((tab) => tab.tabId === tabId)!
    expect(saved.aliasId).toBe('a1')
    expect(saved.versionDrafts?.['']?.url).toBe('https://x/v1?edited')
  })

  it('restores a tab on the alias it showed, with the drafts of the other versions', async () => {
    reset()
    tabs().hydrate('ws', {
      ...defaultUIState(),
      tabs: [{ tabId: 't1', requestId: 'A', pinned: true, baseRevision: 'rev-1', draft: null, aliasId: 'a1', versionDrafts: { '': draftOf('Login', 'https://x/v1?kept') } }],
      activeTabId: 't1',
    })
    await vi.waitFor(() => expect(builder().tabs.t1?.versions['']?.draft.url).toBe('https://x/v1?kept'))
    expect(builder().tabs.t1!.aliasId).toBe('a1')
    expect(builder().tabs.t1!.draft.url).toBe('https://x/v2')
  })
})
