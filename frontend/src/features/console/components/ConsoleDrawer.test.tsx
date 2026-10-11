import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ConsoleEntry } from '@/platform/mappers/console'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/console', () => ({ listConsole: vi.fn(), getConsoleEntry: vi.fn(), clearConsole: vi.fn() }))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))

const consoleService = await import('@/platform/services/console')
const responseBody = await import('@/platform/services/response-body')
const { useConsoleStore } = await import('@/features/console/store')
const { ConsoleDrawer } = await import('./ConsoleDrawer')

const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight')
const offsetWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth')

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, get: () => 600 })
  Object.defineProperty(HTMLElement.prototype, 'offsetWidth', { configurable: true, get: () => 800 })
  Element.prototype.scrollTo = () => undefined
})

afterAll(() => {
  if (offsetHeight) {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', offsetHeight)
  }
  if (offsetWidth) {
    Object.defineProperty(HTMLElement.prototype, 'offsetWidth', offsetWidth)
  }
})

function entry(overrides: Partial<ConsoleEntry>): ConsoleEntry {
  return {
    id: 'e',
    sendId: 's',
    kind: 'log',
    level: 'log',
    summary: '',
    createdAt: '2026-01-01T10:00:00.000Z',
    arguments: [],
    masked: false,
    scriptLabel: null,
    line: null,
    column: null,
    request: null,
    response: null,
    values: null,
    ...overrides,
  }
}

const request = entry({
  id: 'r1',
  kind: 'request',
  level: 'info',
  summary: 'GET https://api.test/invoices',
  request: { method: 'GET', url: 'https://api.test/invoices', nested: false, headers: null, body: null },
  response: { statusCode: 200, statusText: 'OK', timeMs: 12, sizeBytes: 14, error: null, headers: null, body: null, bodyUrl: null, contentType: '', isBinary: false },
})
const log = entry({ id: 'l1', summary: 'token t-1', arguments: ['token', 't-1'], scriptLabel: 'Billing · Pre-request' })
const warning = entry({ id: 'w1', level: 'warn', summary: 'slow', arguments: ['slow'], scriptLabel: 'Create · Post-response' })
const failure = entry({ id: 'f1', kind: 'test_failure', level: 'error', summary: 'is 200: expected 404', scriptLabel: 'Create · Post-response' })

async function renderWith(entries: ConsoleEntry[]) {
  vi.mocked(consoleService.listConsole).mockResolvedValue(entries)
  render(<ConsoleDrawer />)
  if (entries.length > 0) {
    await screen.findByRole('list', { name: 'Console entries' })
  }
}

describe('ConsoleDrawer', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useConsoleStore.setState({ entries: [], status: 'idle', filter: 'all', search: '', selectedId: null, detail: null, failure: null })
  })

  it('loads the session console once and shows an empty state', async () => {
    await renderWith([])
    expect(await screen.findByRole('region', { name: 'Console is empty' })).toBeInTheDocument()
    expect(consoleService.listConsole).toHaveBeenCalledWith({ limit: 5000 })
    expect(screen.getByRole('button', { name: 'Clear' })).toBeDisabled()
  })

  it('lists requests with method, URL, status, time and size and logs with their level and script', async () => {
    await renderWith([log, request, warning])
    const rows = within(screen.getByRole('list', { name: 'Console entries' })).getAllByRole('button')
    expect(rows).toHaveLength(3)
    expect(rows[0]).toHaveAccessibleName('log: token t-1')
    expect(rows[0]).toHaveTextContent('Billing · Pre-request')
    expect(rows[1]).toHaveTextContent('GET')
    expect(rows[1]).toHaveTextContent('https://api.test/invoices')
    expect(rows[1]).toHaveTextContent('200 OK')
    expect(rows[1]).toHaveTextContent(/12.ms/)
    expect(rows[1]).toHaveTextContent(/14.B/)
    expect(rows[2]).toHaveAccessibleName('warn: slow')
  })

  it('appends batched events, empties on a cleared event and skips entries it already has', async () => {
    await renderWith([log])
    act(() => useConsoleStore.getState().applyEvent({ entries: [log, warning], cleared: false }))
    expect(within(screen.getByRole('list', { name: 'Console entries' })).getAllByRole('listitem')).toHaveLength(2)
    act(() => useConsoleStore.getState().applyEvent({ entries: [failure], cleared: true }))
    expect(within(screen.getByRole('list', { name: 'Console entries' })).getAllByRole('listitem')).toHaveLength(1)
    expect(screen.getByText('1 entries')).toBeInTheDocument()
  })

  it('filters by kind and searches text, URL and script names', async () => {
    const user = userEvent.setup()
    await renderWith([log, request, warning, failure])
    await user.click(screen.getByRole('button', { name: 'Requests' }))
    expect(screen.getAllByRole('listitem')).toHaveLength(1)
    await user.click(screen.getByRole('button', { name: 'Logs' }))
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
    await user.click(screen.getByRole('button', { name: 'Errors' }))
    expect(screen.getByRole('listitem')).toHaveTextContent('is 200: expected 404')
    await user.click(screen.getByRole('button', { name: 'All' }))
    await user.type(screen.getByRole('searchbox', { name: 'Search console' }), 'BILLING')
    expect(screen.getByRole('listitem')).toHaveTextContent('token t-1')
    expect(screen.getByText('1 of 4')).toBeInTheDocument()
    await user.clear(screen.getByRole('searchbox', { name: 'Search console' }))
    await user.type(screen.getByRole('searchbox', { name: 'Search console' }), 'nothing here')
    expect(screen.getByText('No entries match.')).toBeInTheDocument()
  })

  it('clears through Go', async () => {
    const user = userEvent.setup()
    vi.mocked(consoleService.clearConsole).mockResolvedValue()
    await renderWith([log])
    await user.click(screen.getByRole('button', { name: 'Clear' }))
    expect(consoleService.clearConsole).toHaveBeenCalledOnce()
    expect(screen.getByRole('region', { name: 'Console is empty' })).toBeInTheDocument()
  })

  it('opens a request with what was sent and received, loading large bodies on demand', async () => {
    const user = userEvent.setup()
    vi.mocked(consoleService.getConsoleEntry).mockResolvedValue({
      ...request,
      masked: true,
      request: { method: 'GET', url: 'https://api.test/invoices', nested: false, headers: [{ key: 'X-Key', value: '•••••' }], body: 'a=1' },
      response: { ...request.response!, headers: [{ key: 'Content-Type', value: 'text/plain' }], body: null, bodyUrl: '/xenon/response/console:r1' },
    })
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('big body')
    await renderWith([request])
    await user.click(screen.getByRole('button', { name: 'GET https://api.test/invoices' }))
    const detail = await screen.findByRole('region', { name: 'Console entry' })
    expect(consoleService.getConsoleEntry).toHaveBeenCalledWith('r1')
    expect(await within(detail).findByText('Secrets masked')).toBeInTheDocument()
    expect(within(detail).getByRole('definition')).toHaveTextContent('•••••')
    expect(within(detail).getByLabelText('Request body')).toHaveTextContent('a=1')
    await user.click(within(detail).getByRole('tab', { name: 'Response' }))
    await user.click(within(detail).getByRole('button', { name: 'Load Body' }))
    expect(await within(detail).findByLabelText('Response body')).toHaveTextContent('big body')
    expect(responseBody.fetchResponseText).toHaveBeenCalledWith('/xenon/response/console:r1')
    await user.click(within(detail).getByRole('button', { name: 'Close entry' }))
    expect(screen.queryByRole('region', { name: 'Console entry' })).not.toBeInTheDocument()
  })

  it('shows a logged object as an expandable tree and errors with their location', async () => {
    const user = userEvent.setup()
    const logged = entry({ id: 'o1', summary: '{…}', scriptLabel: 'Create · Post-response', values: [{ id: 'inv-1', tags: ['a', 'b'], missing: { $type: 'undefined' } }] })
    vi.mocked(consoleService.getConsoleEntry).mockResolvedValueOnce(logged)
    await renderWith([logged, failure])
    await user.click(screen.getByRole('button', { name: 'log: {…}' }))
    const values = await screen.findByRole('group', { name: 'Logged values' })
    expect(values).toHaveTextContent('{…} 3 keys')
    await user.click(within(values).getByText('{…} 3 keys'))
    expect(within(values).getByText('"inv-1"')).toBeVisible()
    expect(within(values).getByText('Array(2)')).toBeInTheDocument()
    expect(within(values).getByText('undefined')).toBeInTheDocument()
    vi.mocked(consoleService.getConsoleEntry).mockResolvedValueOnce({ ...failure, kind: 'error', line: 3, column: 9, arguments: ['TypeError: x'] })
    await user.click(screen.getByRole('button', { name: 'error: is 200: expected 404' }))
    expect(await screen.findByText('Line 3, column 9')).toBeInTheDocument()
  })

  it('reports an entry that is no longer kept', async () => {
    const user = userEvent.setup()
    vi.mocked(consoleService.getConsoleEntry).mockRejectedValue(new AppError('CON-01404', 'the console no longer keeps this entry'))
    await renderWith([log])
    await user.click(screen.getByRole('button', { name: 'log: token t-1' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('the console no longer keeps this entry')
  })
})
