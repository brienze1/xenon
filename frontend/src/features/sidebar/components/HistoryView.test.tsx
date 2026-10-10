import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { HistoryEntry } from '@/platform/mappers/history'
import { emptyDraft } from '@/platform/mappers/request'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn(), getHistoryEntry: vi.fn(), clearHistory: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ parseBulk: vi.fn(), formatBulk: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const history = await import('@/platform/services/history')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { SEARCH_DELAY_MS, useSidebarStore } = await import('@/features/sidebar/store')
const { SidebarPanel } = await import('./SidebarPanel')

const now = new Date()
const today = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 30)
const yesterday = new Date(today.getTime() - 24 * 60 * 60 * 1000)

function entry(id: string, method: string, url: string, sentAt: Date, extra: Partial<HistoryEntry> = {}): HistoryEntry {
  return { id, sentAt: sentAt.toISOString(), method, url, statusCode: 200, errorKind: null, totalMs: 42, sizeBytes: 10, sourceRequestId: null, ...extra }
}

const entries = [
  entry('h3', 'POST', 'https://api.test/orders', today, { statusCode: 201 }),
  entry('h2', 'GET', 'https://api.test/missing', today, { statusCode: 404 }),
  entry('h1', 'GET', '', yesterday, { statusCode: null, errorKind: 'invalid_url', totalMs: 0 }),
]

const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight')

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, get: () => 600 })
})

afterAll(() => {
  if (offsetHeight) {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', offsetHeight)
  }
})

async function renderHistory() {
  const actions = { onOpenRequest: vi.fn(), onPinRequest: vi.fn(), onOpenExample: vi.fn(), onPinExample: vi.fn(), onExampleDeleted: vi.fn(), onRequestsDeleted: vi.fn(), onRenamed: vi.fn(), onOpenHistoryEntry: vi.fn() }
  render(<SidebarPanel {...actions} />)
  await userEvent.click(screen.getByRole('tab', { name: 'History' }))
  await screen.findByRole('list', { name: 'History' })
  return actions
}

describe('HistoryView', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.getState().reset()
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(history.listHistory).mockResolvedValue(entries)
  })

  it('switches between Collections and History and lists entries grouped by day', async () => {
    await renderHistory()
    expect(screen.getByRole('tab', { name: 'History' })).toHaveAttribute('aria-selected', 'true')
    expect(useSidebarStore.getState().view).toBe('history')
    expect(history.listHistory).toHaveBeenCalledWith('ws', { query: '' })
    expect(screen.getByRole('heading', { name: 'Today' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Yesterday' })).toBeInTheDocument()
    const list = screen.getByRole('list', { name: 'History' })
    const items = within(list).getAllByRole('button')
    expect(items).toHaveLength(3)
    expect(items[0]).toHaveTextContent('POST')
    expect(items[0]).toHaveTextContent('https://api.test/orders')
    expect(items[0]).toHaveTextContent('201')
    expect(items[0]).toHaveTextContent('42 ms')
    expect(items[2]).toHaveAccessibleName(/^GET no URL, invalid_url, /)
    expect(screen.getByText('3 entries')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'New Collection' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('tab', { name: 'Collections' }))
    expect(useSidebarStore.getState().view).toBe('collections')
    expect(screen.getByRole('button', { name: 'New Collection' })).toBeInTheDocument()
  })

  it('searches the history by URL or method', async () => {
    await renderHistory()
    vi.mocked(history.listHistory).mockResolvedValue([entries[0]!])
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      await userEvent.type(screen.getByRole('searchbox', { name: 'Search history' }), 'post')
      await act(() => vi.advanceTimersByTimeAsync(SEARCH_DELAY_MS))
    } finally {
      vi.useRealTimers()
    }
    await waitFor(() => expect(within(screen.getByRole('list', { name: 'History' })).getAllByRole('button')).toHaveLength(1))
    expect(history.listHistory).toHaveBeenLastCalledWith('ws', { query: 'post' })
    expect(useSidebarStore.getState().query).toBe('')
    vi.mocked(history.listHistory).mockResolvedValue([])
    await act(async () => {
      useSidebarStore.getState().setHistoryQuery('nothing')
      await useSidebarStore.getState().loadHistory()
    })
    expect(screen.getByText('No history entries match your search.')).toBeInTheDocument()
  })

  it('opens an entry as a new unsaved request', async () => {
    const request = { ...emptyDraft(), id: 'r1', name: 'Create order', method: 'POST', url: 'https://api.test/orders' }
    vi.mocked(history.getHistoryEntry).mockResolvedValue({ entry: entries[0]!, request })
    const actions = await renderHistory()
    await userEvent.click(screen.getByRole('button', { name: /^POST https:\/\/api.test\/orders/ }))
    await waitFor(() => expect(actions.onOpenHistoryEntry).toHaveBeenCalledTimes(1))
    expect(history.getHistoryEntry).toHaveBeenCalledWith('ws', 'h3')
    expect(actions.onOpenHistoryEntry).toHaveBeenCalledWith({ ...request, id: '' })
  })

  it('clears the history only after confirmation', async () => {
    vi.mocked(history.clearHistory).mockResolvedValue()
    await renderHistory()
    await userEvent.click(screen.getByRole('button', { name: 'Clear History' }))
    const dialog = await screen.findByRole('dialog', { name: 'Clear History?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(history.clearHistory).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Clear History' }))
    await userEvent.click(within(await screen.findByRole('dialog', { name: 'Clear History?' })).getByRole('button', { name: 'Clear History' }))
    await waitFor(() => expect(history.clearHistory).toHaveBeenCalledWith('ws'))
    expect(await screen.findByRole('region', { name: 'No history' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Clear History' })).toBeDisabled()
  })

  it('keeps the clear dialog open and shows why clearing failed', async () => {
    const { AppError } = await import('@/platform/errors')
    vi.mocked(history.clearHistory).mockRejectedValue(new AppError('HIS-05500', 'The history file is locked'))
    await renderHistory()
    await userEvent.click(screen.getByRole('button', { name: 'Clear History' }))
    const dialog = await screen.findByRole('dialog', { name: 'Clear History?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Clear History' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('The history file is locked. Check that the workspace folder exists')
    expect(useSidebarStore.getState().error).toBeNull()
    expect(screen.getByRole('list', { name: 'History' })).toBeInTheDocument()
  })

  it('shows the failure when the history cannot be loaded', async () => {
    const { AppError } = await import('@/platform/errors')
    vi.mocked(history.listHistory).mockRejectedValue(new AppError('HIS-04500', 'cannot read history'))
    render(<SidebarPanel onOpenRequest={vi.fn()} onPinRequest={vi.fn()} onOpenExample={vi.fn()} onPinExample={vi.fn()} onExampleDeleted={vi.fn()} onRequestsDeleted={vi.fn()} onRenamed={vi.fn()} onOpenHistoryEntry={vi.fn()} />)
    await userEvent.click(screen.getByRole('tab', { name: 'History' }))
    expect(await screen.findByText('cannot read history. Try again, or check that the workspace folder exists and is writable.')).toBeInTheDocument()
    expect(useSidebarStore.getState().historyStatus).toBe('ready')
  })
})
