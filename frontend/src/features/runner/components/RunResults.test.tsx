import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { SendResult } from '@/platform/mappers/request'
import type { RunnerSlots } from '@/features/runner/model'
import { injectStylesheets } from '@/test/stylesheets'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  startRun: vi.fn(),
  getRun: vi.fn(),
  exportRun: vi.fn(),
  runResponse: vi.fn(),
}))
vi.mock('@/platform/services/dialogs', () => ({ pickFile: vi.fn(), pickSaveFile: vi.fn(), pickFolder: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const dialogs = await import('@/platform/services/dialogs')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { readyTarget, run, runConfig, runPlan, runRow, runSummaryView } = await import('@/features/runner/test-fixtures')
const { RunResults } = await import('./RunResults')

const slots: RunnerSlots = {
  tests: (tests) => <p>Tests: {tests.map((test) => `${test.name}=${test.status}`).join(', ')}</p>,
  response: (key, result: SendResult) => (
    <p>
      Response {key} {result.statusCode}
    </p>
  ),
}

const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight')

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, get: () => 600 })
})

afterAll(() => {
  if (offsetHeight) {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', offsetHeight)
  }
})

function seed(shown = run()) {
  useRunnerStore.setState({ targets: { c1: readyTarget(runPlan(), { view: 'results', runId: shown.id, run: shown }) } })
  return shown
}

function results() {
  return screen.getByRole('list', { name: 'Run results' })
}

function consoleEntry(id: string, summary: string) {
  return {
    id,
    sendId: 's',
    kind: 'log' as const,
    level: 'log' as const,
    summary,
    createdAt: '',
    arguments: [],
    masked: false,
    scriptLabel: null,
    line: null,
    column: null,
    request: null,
    response: null,
    values: null,
  }
}

describe('RunResults', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows the summary of a finished run', () => {
    const shown = seed()
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const summary = screen.getByLabelText('Run summary')
    const text = summary.textContent
    expect(text).toContain('Completed')
    expect(text).toContain('Iterations1')
    expect(text).toContain('Requests2')
    expect(text).toContain('3 passed · 1 failed · 0 skipped')
    expect(text).toContain('120')
  })

  it('groups rows by iteration and filters them per FR-034', async () => {
    const shown = seed(
      run({
        rows: [
          runRow(0, { name: 'Login', outcome: 'passed' }),
          runRow(1, { name: 'Get order', outcome: 'failed' }),
          runRow(2, { name: 'Login', iteration: 1, outcome: 'skipped' }),
          runRow(3, { name: 'Get order', iteration: 1, outcome: 'cancelled' }),
        ],
      }),
    )
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const lines = within(results())
      .getAllByRole('listitem')
      .map((item) => (item.querySelector('p') ? item.textContent : within(item).getAllByRole('button')[1]!.textContent))
    expect(lines).toEqual(['Iteration 1 2', 'Login', 'Get order', 'Iteration 2 2', 'Login', 'Get order'])
    const filters = screen.getByRole('group', { name: 'Filter results' })
    expect(within(filters).getAllByRole('button').map((chip) => chip.textContent)).toEqual(['All 4', 'Passed 1', 'Failed 2', 'Skipped 1'])
    await userEvent.click(within(filters).getByRole('button', { name: 'Failed 2' }))
    expect(within(filters).getByRole('button', { name: 'Failed 2' })).toHaveAttribute('aria-pressed', 'true')
    expect(within(results()).getAllByRole('img').map((icon) => icon.getAttribute('aria-label'))).toEqual(['Failed', 'Cancelled'])
    await userEvent.click(within(filters).getByRole('button', { name: 'Skipped 1' }))
    expect(within(results()).getAllByRole('img').map((icon) => icon.getAttribute('aria-label'))).toEqual(['Skipped'])
  })

  it('shows a row’s alias, URL, status, time and size, and expands it to its tests, script errors and console', async () => {
    const shown = seed(
      run({
        rows: [
          runRow(0, {
            name: 'Login',
            aliasName: 'mx-test-user',
            url: 'https://api.test/login',
            status: 201,
            tests: [{ index: 0, name: 'is created', status: 'passed', errorMessage: '', scriptLabel: '', async: false }],
            scriptErrors: [{ kind: 'error', message: 'boom', line: null, column: null, scriptLabel: 'Login · Post-response' }],
            console: [
              {
                id: 'k1',
                sendId: 's',
                kind: 'log',
                level: 'log',
                summary: 'hello from the script',
                createdAt: '',
                arguments: [],
                masked: false,
                scriptLabel: null,
                line: null,
                column: null,
                request: null,
                response: null,
                values: null,
              },
            ],
          }),
        ],
      }),
    )
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const row = within(results()).getAllByRole('listitem')[1]!
    expect(row).toHaveTextContent('Login · mx-test-user')
    expect(row).toHaveTextContent('https://api.test/login')
    expect(row).toHaveTextContent('201')
    const toggle = within(row).getByRole('button', { name: 'Details of Login · mx-test-user' })
    await userEvent.click(toggle)
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByText('Tests: is created=passed')).toBeInTheDocument()
    expect(within(screen.getByRole('list', { name: 'Script errors' })).getByRole('listitem')).toHaveTextContent('Login · Post-response: boom')
    expect(within(screen.getByRole('list', { name: 'Console' })).getByRole('listitem')).toHaveTextContent('hello from the script')
  })

  it('opens a row in a side panel with the request as sent and notes responses that were not persisted', async () => {
    const shown = seed(
      run({
        rows: [
          runRow(0, {
            name: 'Login',
            method: 'POST',
            sentRequest: { method: 'POST', url: 'https://api.test/login', headers: [{ key: 'Authorization', value: '•••••' }], body: '{"user":"luis"}' },
          }),
        ],
      }),
    )
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    await userEvent.click(within(results()).getByRole('button', { name: 'Login' }))
    const detail = screen.getByRole('complementary', { name: 'Login details' })
    expect(within(detail).getByRole('region', { name: 'Request' })).toHaveTextContent('https://api.test/login')
    expect(within(detail).getByRole('rowheader', { name: 'Authorization' }).nextSibling).toHaveTextContent('•••••')
    expect(within(detail).getByLabelText('Request body')).toHaveTextContent('{"user":"luis"}')
    expect(within(detail).getByText('Responses were not persisted for this run.')).toBeInTheDocument()
    expect(service.runResponse).not.toHaveBeenCalled()
    await userEvent.click(within(detail).getByRole('button', { name: 'Close details' }))
    expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
  })

  it('shows a persisted response through the response viewer, keyed by the run response key', async () => {
    vi.mocked(service.runResponse).mockResolvedValue({ statusCode: 200, bodyUrl: '/body' } as SendResult)
    const shown = seed(run({ rows: [runRow(0, { name: 'Login', responseBodyUrl: '/body/run:R1:0' })] }))
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    await userEvent.click(within(results()).getByRole('button', { name: 'Login' }))
    expect(await screen.findByText('Response run:R1:0 200')).toBeInTheDocument()
    expect(service.runResponse).toHaveBeenCalledWith('L1', 'R1', 0)
  })

  it('notes a persisted response this session no longer holds', async () => {
    vi.mocked(service.runResponse).mockRejectedValue({ cause: { error: { code: 'RUN-01404', description: 'gone' } } })
    const shown = seed(run({ rows: [runRow(0, { name: 'Login', responseBodyUrl: '/body/run:R1:0' })] }))
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    await userEvent.click(within(results()).getByRole('button', { name: 'Login' }))
    expect(await screen.findByText('Responses were not persisted for this run.')).toBeInTheDocument()
  })

  it('renders only the visible rows of 10,000 results', () => {
    const rows = Array.from({ length: 10_000 }, (_, index) => runRow(index, { iteration: Math.floor(index / 100) }))
    const shown = seed(run({ rows }))
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const mounted = within(results()).getAllByRole('listitem')
    expect(mounted.length).toBeGreaterThan(0)
    expect(mounted.length).toBeLessThan(60)
  })

  it('runs again with the run’s configuration and exports JSON or JUnit XML', async () => {
    const stored = run({ config: runConfig({ environmentId: 'none' }) })
    seed(stored)
    vi.mocked(service.startRun).mockResolvedValue(runSummaryView({ id: 'R2' }))
    vi.mocked(service.getRun).mockResolvedValue(run({ id: 'R2', status: 'running' }))
    vi.mocked(dialogs.pickSaveFile).mockResolvedValue('/out/run.xml')
    render(<RunResults itemId="c1" run={stored} slots={slots} />)
    await userEvent.click(screen.getByRole('button', { name: 'Export' }))
    expect(within(screen.getByRole('menu', { name: 'Export run' })).getAllByRole('menuitem').map((item) => item.textContent)).toEqual(['JSON', 'JUnit XML'])
    await userEvent.click(screen.getByRole('menuitem', { name: 'JUnit XML' }))
    expect(service.exportRun).toHaveBeenCalledWith('L1', 'R1', 'junit', '/out/run.xml')
    await userEvent.click(screen.getByRole('button', { name: 'Run again' }))
    expect(vi.mocked(service.startRun).mock.calls[0]![2]).toEqual(stored.config)
  })

  it('shows the statistics of a performance run and exports it as JSON only', async () => {
    const shown = seed(
      run({
        config: runConfig({ type: 'performance' }),
        rows: [],
        performance: {
          overall: { requestId: '', name: 'All', count: 10, avgMs: 1, p90Ms: 1, p95Ms: 1, p99Ms: 1, minMs: 1, maxMs: 1, throughput: 1, errorRate: 0, statusCounts: [] },
          perRequest: [],
          series: [],
        },
      }),
    )
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    expect(screen.getByRole('table', { name: 'Performance statistics' })).toBeInTheDocument()
    expect(screen.queryByRole('list', { name: 'Run results' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Export' }))
    expect(within(screen.getByRole('menu', { name: 'Export run' })).getAllByRole('menuitem').map((item) => item.textContent)).toEqual(['JSON'])
  })

  it('disables Run again while a run is active on the target', () => {
    const shown = seed()
    useRunnerStore.setState({ active: { c1: runSummaryView({ id: 'R9' }) } })
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    expect(screen.getByRole('button', { name: 'Run again' })).toBeDisabled()
  })

  it('stacks an expanded row’s parts in normal flow and caps the panel at 320 px', async () => {
    injectStylesheets('src/features/runner/components/RunResults.module.css')
    const shown = seed(
      run({
        rows: [
          runRow(0, {
            name: 'Login',
            error: 'timeout',
            tests: [{ index: 0, name: 'is ok', status: 'passed', errorMessage: '', scriptLabel: '', async: false }],
            scriptErrors: [{ kind: 'error', message: 'boom', line: null, column: null, scriptLabel: 'Login · Post-response' }],
            console: [consoleEntry('k1', 'GET https://api.test/health')],
          }),
        ],
      }),
    )
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const toggle = screen.getByRole('button', { name: 'Details of Login' })
    await userEvent.click(toggle)
    const panel = toggle.closest('[role="listitem"]')!.querySelector<HTMLElement>('.expansion')!
    const style = getComputedStyle(panel)
    expect(style.display).toBe('block')
    expect(style.maxHeight).toBe('320px')
    expect(style.overflow).toBe('auto')
    expect(panel.children).toHaveLength(4)
  })

  it('lets the keyboard reach and scroll an expanded row’s details', async () => {
    const shown = seed(run({ rows: [runRow(0, { name: 'Login', aliasName: 'mx-test-user' })] }))
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    const row = within(results()).getAllByRole('listitem')[1]!
    await userEvent.click(within(row).getByRole('button', { name: 'Details of Login · mx-test-user' }))
    const details = screen.getByRole('group', { name: 'Details of Login · mx-test-user' })
    expect(details).toHaveAttribute('tabindex', '0')
    expect(details).toHaveTextContent('Tests:')
    within(row).getByRole('button', { name: 'Login · mx-test-user' }).focus()
    await userEvent.tab()
    expect(details).toHaveFocus()
  })

  it('measures every row instead of giving it a fixed height', async () => {
    injectStylesheets('src/features/runner/components/RunResults.module.css')
    const shown = seed(run({ rows: [runRow(0, { name: 'Login' }), runRow(1, { name: 'Get order' }), runRow(2, { name: 'Login', iteration: 1 })] }))
    render(<RunResults itemId="c1" run={shown} slots={slots} />)
    await userEvent.click(screen.getAllByRole('button', { name: 'Details of Login' })[0]!)
    await userEvent.click(screen.getByRole('button', { name: 'Details of Get order' }))
    const items = within(results()).getAllByRole('listitem')
    items.forEach((item, position) => {
      expect(item.dataset.index).toBe(String(position))
      expect(item.style.height).toBe('')
      expect(item.style.transform).toMatch(/^translateY\(/)
    })
    expect(getComputedStyle(items[0]!.querySelector('p')!).height).toBe('28px')
    for (const item of items) {
      const row = item.querySelector('.row')
      if (row) {
        expect(getComputedStyle(row).height).not.toBe('100%')
      }
    }
  })

  it('keeps an expanded row expanded and measured while a live run refreshes', async () => {
    const first = seed(run({ status: 'running', rows: [runRow(0, { name: 'Login', console: [consoleEntry('k1', 'one')] })] }))
    const { rerender } = render(<RunResults itemId="c1" run={first} slots={slots} />)
    await userEvent.click(screen.getByRole('button', { name: 'Details of Login' }))
    const refreshed = run({
      status: 'running',
      rows: [runRow(0, { name: 'Login', console: [consoleEntry('k1', 'one'), consoleEntry('k2', 'two')] }), runRow(1, { name: 'Get order' })],
    })
    rerender(<RunResults itemId="c1" run={refreshed} slots={slots} />)
    const details = screen.getByRole('group', { name: 'Details of Login' })
    expect(within(within(details).getByRole('list', { name: 'Console' })).getAllByRole('listitem')).toHaveLength(2)
    for (const item of Array.from(results().children) as HTMLElement[]) {
      expect(item).toHaveAttribute('data-index')
      expect(item.style.height).toBe('')
    }
  })
})
