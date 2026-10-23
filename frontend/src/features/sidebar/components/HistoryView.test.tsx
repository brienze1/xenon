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
  return { id, sentAt: sentAt.toISOString(), kind: 'send', source: 'user', client: null, activity: null, method, url, outcome: 'sent', statusCode: 200, errorKind: null, totalMs: 42, sizeBytes: 10, sourceRequestId: null, aliasName: null, ...extra }
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

async function pickView(name: string) {
  await userEvent.click(screen.getByRole('button', { name: /^Sidebar view: / }))
  await userEvent.click(screen.getByRole('menuitemradio', { name }))
}

async function renderHistory() {
  const actions = { onOpenRequest: vi.fn(), onPinRequest: vi.fn(), onOpenExample: vi.fn(), onPinExample: vi.fn(), onExampleDeleted: vi.fn(), onRequestsDeleted: vi.fn(), onRenamed: vi.fn(), onOpenHistoryEntry: vi.fn(), onOpenVariables: vi.fn(), selectedVariables: null, onOpenItem: vi.fn(), onRunItem: vi.fn() }
  render(<SidebarPanel {...actions} />)
  await pickView('History')
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
    expect(screen.getByRole('button', { name: 'Sidebar view: History' })).toBeInTheDocument()
    expect(useSidebarStore.getState().view).toBe('history')
    expect(history.listHistory).toHaveBeenCalledWith('ws', { query: '', source: 'all' })
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
    await pickView('Collections')
    expect(useSidebarStore.getState().view).toBe('collections')
    expect(screen.getByRole('button', { name: 'New Collection' })).toBeInTheDocument()
  })

  it('labels sends a pre-request script skipped or stopped', async () => {
    vi.mocked(history.listHistory).mockResolvedValue([
      entry('h5', 'GET', 'https://api.test/skip', today, { outcome: 'skipped_by_script', statusCode: null }),
      entry('h4', 'GET', 'https://api.test/stop', today, { outcome: 'stopped_by_script_error', statusCode: null }),
    ])
    await renderHistory()
    const items = within(screen.getByRole('list', { name: 'History' })).getAllByRole('button')
    expect(items[0]).toHaveAccessibleName(/^GET https:\/\/api\.test\/skip, skipped, /)
    expect(items[1]).toHaveAccessibleName(/^GET https:\/\/api\.test\/stop, script error, /)
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
    expect(history.listHistory).toHaveBeenLastCalledWith('ws', { query: 'post', source: 'all' })
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
    vi.mocked(history.getHistoryEntry).mockResolvedValue({ entry: entries[0]!, request, activity: null })
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
    await waitFor(() => expect(history.clearHistory).toHaveBeenCalledWith('ws', 'all'))
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

  it('filters by source next to Clear History and remembers the choice', async () => {
    await renderHistory()
    const filter = screen.getByRole('radiogroup', { name: 'History source' })
    expect(filter.parentElement).toContainElement(screen.getByRole('button', { name: 'Clear History' }))
    expect(within(filter).getAllByRole('radio').map((radio) => radio.textContent)).toEqual(['All', 'User', 'AI'])
    expect(within(filter).getByRole('radio', { name: 'All' })).toBeChecked()
    vi.mocked(history.listHistory).mockResolvedValue([])
    await userEvent.click(within(filter).getByRole('radio', { name: 'AI' }))
    expect(within(filter).getByRole('radio', { name: 'AI' })).toBeChecked()
    expect(history.listHistory).toHaveBeenLastCalledWith('ws', { query: '', source: 'ai' })
    expect(useSidebarStore.getState().historySource).toBe('ai')
    expect(await screen.findByText('No entries from this source yet.')).toBeInTheDocument()
    await userEvent.keyboard('{ArrowLeft}')
    expect(useSidebarStore.getState().historySource).toBe('user')
    expect(within(filter).getByRole('radio', { name: 'User' })).toHaveFocus()
  })

  it('names the request and the alias of an alias send', async () => {
    const login = { id: 'r1', kind: 'request' as const, name: 'Login', method: 'POST', url: null, error: null, warning: null, children: [], examples: [] }
    vi.mocked(tree.getTree).mockResolvedValue([{ id: 'c1', kind: 'collection', name: 'Billing', method: null, url: null, error: null, warning: null, children: [login], examples: [] }])
    await useSidebarStore.getState().load()
    vi.mocked(history.listHistory).mockResolvedValue([
      entry('h8', 'POST', 'https://api.test/login', today, { sourceRequestId: 'r1', aliasName: 'mx-test-user' }),
      entry('h7', 'POST', 'https://api.test/login', today, { sourceRequestId: 'r1' }),
    ])
    await renderHistory()
    const rows = within(screen.getByRole('list', { name: 'History' })).getAllByRole('button')
    expect(rows[0]).toHaveTextContent('Login · mx-test-user')
    expect(rows[0]).toHaveAccessibleName(/^Login · mx-test-user, POST https:\/\/api\.test\/login, 200, /)
    expect(rows[1]).not.toHaveTextContent('·')
  })

  it('tags AI sends with the client and opens them like any send', async () => {
    vi.mocked(history.listHistory).mockResolvedValue([entry('h9', 'POST', 'https://api.test/orders', today, { source: 'ai', client: 'Claude Code', statusCode: 201 })])
    const request = { ...emptyDraft(), method: 'POST', url: 'https://api.test/orders' }
    vi.mocked(history.getHistoryEntry).mockResolvedValue({ entry: entries[0]!, request, activity: null })
    const actions = await renderHistory()
    const row = within(screen.getByRole('list', { name: 'History' })).getByRole('button')
    expect(row).toHaveTextContent('AI · Claude Code')
    expect(row).toHaveAccessibleName(/^AI · Claude Code — POST https:\/\/api\.test\/orders, 201, /)
    await userEvent.click(row)
    await waitFor(() => expect(actions.onOpenHistoryEntry).toHaveBeenCalledWith({ ...request, id: '' }))
  })

  it('shows AI action rows with tool and action, target, outcome and duration, and expands them to their masked arguments', async () => {
    const activity = { tool: 'tree', action: 'create', target: { id: 'r9', name: 'POST /orders', path: 'Payments / Orders' }, outcome: 'ok' as const, code: null, durationMs: 12, args: null }
    const failed = { tool: 'env_edit', action: 'set_current', target: null, outcome: 'error' as const, code: 'VAR-01400', durationMs: 3, args: null }
    vi.mocked(history.listHistory).mockResolvedValue([
      entry('a2', '', '', today, { kind: 'action', source: 'ai', client: 'Claude Code', activity, totalMs: 12 }),
      entry('a1', '', '', today, { kind: 'action', source: 'ai', client: 'Claude Code', activity: failed, totalMs: 3 }),
    ])
    vi.mocked(history.getHistoryEntry).mockResolvedValue({ entry: entries[0]!, request: emptyDraft(), activity: { ...activity, args: '{"name":"POST /orders","secret_value":"•••••"}' } })
    const actions = await renderHistory()
    const [created, refused] = within(screen.getByRole('list', { name: 'History' })).getAllByRole('button')
    expect(created).toHaveTextContent('AI · Claude Code')
    expect(created).toHaveTextContent('tree.create')
    expect(created).toHaveTextContent('POST /orders')
    expect(created).toHaveTextContent('ok')
    expect(created).toHaveTextContent('12 ms')
    expect(created).toHaveAttribute('aria-expanded', 'false')
    expect(refused).toHaveTextContent('env_edit.set_current')
    expect(refused).toHaveTextContent('VAR-01400')
    await userEvent.click(created!)
    expect(created).toHaveAttribute('aria-expanded', 'true')
    const args = await screen.findByRole('region', { name: 'Arguments (masked)' })
    expect(await within(args).findByText('{"name":"POST /orders","secret_value":"•••••"}')).toBeInTheDocument()
    expect(history.getHistoryEntry).toHaveBeenCalledWith('ws', 'a2')
    expect(actions.onOpenHistoryEntry).not.toHaveBeenCalled()
    await userEvent.click(created!)
    expect(screen.queryByRole('region', { name: 'Arguments (masked)' })).not.toBeInTheDocument()
  })

  it('clears only the source the filter shows', async () => {
    vi.mocked(history.clearHistory).mockResolvedValue()
    await renderHistory()
    await userEvent.click(screen.getByRole('radio', { name: 'AI' }))
    await userEvent.click(screen.getByRole('button', { name: 'Clear History' }))
    const dialog = await screen.findByRole('dialog', { name: 'Clear AI History?' })
    expect(dialog).toHaveTextContent('removes every AI send and action in this workspace')
    expect(dialog).toHaveTextContent('Your own sends stay.')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Clear AI History' }))
    await waitFor(() => expect(history.clearHistory).toHaveBeenCalledWith('ws', 'ai'))
  })

  it('shows the failure when the history cannot be loaded', async () => {
    const { AppError } = await import('@/platform/errors')
    vi.mocked(history.listHistory).mockRejectedValue(new AppError('HIS-04500', 'cannot read history'))
    render(<SidebarPanel onOpenRequest={vi.fn()} onPinRequest={vi.fn()} onOpenExample={vi.fn()} onPinExample={vi.fn()} onExampleDeleted={vi.fn()} onRequestsDeleted={vi.fn()} onRenamed={vi.fn()} onOpenHistoryEntry={vi.fn()} onOpenVariables={vi.fn()} selectedVariables={null} onOpenItem={vi.fn()} onRunItem={vi.fn()} />)
    await pickView('History')
    expect(await screen.findByText('cannot read history. Try again, or check that the workspace folder exists and is writable.')).toBeInTheDocument()
    expect(useSidebarStore.getState().historyStatus).toBe('ready')
  })
})
