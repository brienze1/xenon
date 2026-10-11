import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/tree', () => ({ quickSearch: vi.fn(), getTree: vi.fn(async () => []), searchTree: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { QUICK_SEARCH_DELAY_MS, QUICK_SEARCH_LIMIT, useSearchStore } = await import('./store')

const hits = [
  { id: 'r1', name: 'Ping', method: 'GET', url: 'https://x/ping', path: 'Billing' },
  { id: 'r2', name: 'Pay', method: 'POST', url: 'https://x/pay', path: 'Billing' },
]

function search() {
  return useSearchStore.getState()
}

describe('quick search store', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.resetAllMocks()
    search().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('searches the workspace shortly after typing stops and highlights the first result', async () => {
    vi.mocked(tree.quickSearch).mockResolvedValue(hits)
    search().openSearch()
    expect(search().open).toBe(true)
    search().setQuery('p')
    search().setQuery('pa')
    expect(search().status).toBe('loading')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    expect(tree.quickSearch).toHaveBeenCalledOnce()
    expect(tree.quickSearch).toHaveBeenCalledWith('ws', 'pa', QUICK_SEARCH_LIMIT)
    expect(search()).toMatchObject({ results: hits, status: 'ready', activeIndex: 0 })
  })

  it('shows nothing for a blank query and keeps only the latest answer', async () => {
    search().openSearch()
    search().setQuery('  ')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    expect(tree.quickSearch).not.toHaveBeenCalled()
    expect(search().status).toBe('idle')
    let resolveFirst: (value: typeof hits) => void = () => undefined
    vi.mocked(tree.quickSearch).mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)))
    search().setQuery('p')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    vi.mocked(tree.quickSearch).mockResolvedValueOnce([hits[1]!])
    search().setQuery('pay')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    resolveFirst(hits)
    await vi.advanceTimersByTimeAsync(0)
    expect(search().results).toEqual([hits[1]])
  })

  it('reports a failed search', async () => {
    vi.mocked(tree.quickSearch).mockRejectedValue(new AppError('TRE-40400', 'bad'))
    search().openSearch()
    search().setQuery('x')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    expect(search()).toMatchObject({ status: 'error', results: [] })
    expect(search().error?.code).toBe('TRE-40400')
  })

  it('moves the highlight with wrap-around and opens the chosen request in a pinned tab', async () => {
    vi.mocked(tree.quickSearch).mockResolvedValue(hits)
    const openPinned = vi.spyOn(useTabsStore.getState(), 'openPinned').mockResolvedValue()
    search().openSearch()
    search().move(1)
    expect(search().activeIndex).toBe(0)
    search().setQuery('p')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    search().move(-1)
    expect(search().activeIndex).toBe(1)
    search().move(1)
    expect(search().activeIndex).toBe(0)
    search().move(1)
    await search().choose()
    expect(openPinned).toHaveBeenCalledWith('r2')
    expect(search().open).toBe(false)
    await search().choose(5)
    expect(openPinned).toHaveBeenCalledOnce()
    openPinned.mockRestore()
  })

  it('does not search without an active workspace', async () => {
    useWorkspacesStore.setState({ activeLocalId: null })
    search().openSearch()
    search().setQuery('p')
    await vi.advanceTimersByTimeAsync(QUICK_SEARCH_DELAY_MS)
    expect(tree.quickSearch).not.toHaveBeenCalled()
    expect(search().results).toEqual([])
  })
})
