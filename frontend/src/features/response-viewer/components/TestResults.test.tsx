import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { SendResult, TestResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', () => ({ lastResult: vi.fn(), releaseResponse: vi.fn(), saveResponseBody: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))

const requests = await import('@/platform/services/requests')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { ResponseViewer } = await import('./ResponseViewer')
const { TestResultsTab } = await import('./TestResultsTab')

const tests: TestResult[] = [
  { index: 0, name: 'status 200', status: 'failed', errorMessage: 'expected response to have status code 200 but got 404', scriptLabel: 'Billing · Post-response', async: false },
  { index: 1, name: 'has id', status: 'passed', errorMessage: '', scriptLabel: 'Create · Post-response', async: false },
  { index: 2, name: 'later', status: 'skipped', errorMessage: '', scriptLabel: 'Create · Post-response', async: false },
  { index: 3, name: 'token', status: 'passed', errorMessage: '', scriptLabel: 'Create · Pre-request', async: true },
]
const summary = { total: 4, passed: 2, failed: 1, skipped: 1 }

function result(overrides: Partial<SendResult> = {}): SendResult {
  return {
    tabId: 'tab',
    sendId: 's',
    outcome: 'sent',
    tests,
    testSummary: summary,
    scriptErrors: [],
    ok: true,
    statusCode: 404,
    statusText: 'Not Found',
    proto: 'HTTP/1.1',
    headers: [],
    cookiesSet: [],
    sizeBytes: 2,
    contentType: 'application/json',
    isBinary: false,
    bodyUrl: '/xenon/response/tab',
    inlineBody: '{}',
    timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 3 },
    error: null,
    ...overrides,
  }
}

function show(value: SendResult) {
  act(() => useResponseViewerStore.getState().setResult('tab', value))
}

describe('Test results', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    useResponseViewerStore.getState().reset()
  })

  it('counts passed out of all tests in the tab title and lists each test with its script and message', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result())
    const tab = screen.getByRole('tab', { name: 'Test Results (2/4)' })
    await user.click(tab)
    const list = screen.getByRole('list', { name: 'Test results' })
    const rows = within(list).getAllByRole('listitem')
    expect(rows).toHaveLength(4)
    expect(rows[0]).toHaveTextContent('Fail')
    expect(rows[0]).toHaveTextContent('status 200')
    expect(rows[0]).toHaveTextContent('expected response to have status code 200 but got 404')
    expect(rows[0]).toHaveTextContent('Billing · Post-response')
    expect(rows[2]).toHaveTextContent('Skip')
  })

  it('filters by status', async () => {
    const user = userEvent.setup()
    render(<TestResultsTab tests={tests} summary={summary} />)
    expect(screen.getByRole('button', { name: 'All (4)' })).toHaveAttribute('aria-pressed', 'true')
    await user.click(screen.getByRole('button', { name: 'Passed (2)' }))
    expect(screen.getAllByRole('listitem').map((row) => row.textContent)).toEqual([
      expect.stringContaining('has id'),
      expect.stringContaining('token'),
    ])
    await user.click(screen.getByRole('button', { name: 'Failed (1)' }))
    expect(screen.getAllByRole('listitem')).toHaveLength(1)
    await user.click(screen.getByRole('button', { name: 'Skipped (1)' }))
    expect(screen.getByRole('listitem')).toHaveTextContent('later')
  })

  it('says when a filter matches nothing and when there are no tests', async () => {
    const user = userEvent.setup()
    const { rerender } = render(<TestResultsTab tests={tests.slice(1, 2)} summary={{ total: 1, passed: 1, failed: 0, skipped: 0 }} />)
    await user.click(screen.getByRole('button', { name: 'Failed (0)' }))
    expect(screen.getByText('No failed tests.')).toBeInTheDocument()
    rerender(<TestResultsTab tests={[]} summary={{ total: 0, passed: 0, failed: 0, skipped: 0 }} />)
    expect(screen.getByRole('region', { name: 'No tests' })).toBeInTheDocument()
  })

  it('titles the tab without a count when no test ran', () => {
    render(<ResponseViewer tabId="tab" />)
    show(result({ tests: [], testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 } }))
    expect(screen.getByRole('tab', { name: 'Test Results' })).toBeInTheDocument()
  })
})

describe('Script errors in the response area', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    useResponseViewerStore.getState().reset()
  })

  it('explains a send stopped by a pre-request script with the script, location and pre-request tests', () => {
    render(<ResponseViewer tabId="tab" />)
    show(
      result({
        outcome: 'stopped_by_script_error',
        ok: false,
        statusCode: 0,
        statusText: '',
        tests: tests.slice(3),
        testSummary: { total: 1, passed: 1, failed: 0, skipped: 0 },
        scriptErrors: [{ kind: 'script_runtime', message: 'Error: boom', line: 2, column: 7, scriptLabel: 'Create · Pre-request' }],
      }),
    )
    const panel = screen.getByRole('alert', { name: 'Script outcome' })
    expect(panel).toHaveTextContent('Request Not Sent: a Pre-request Script Failed')
    expect(panel).toHaveTextContent('Create · Pre-request (line 2, column 7)')
    expect(panel).toHaveTextContent('Error: boom')
    expect(screen.getByRole('list', { name: 'Test results' })).toHaveTextContent('token')
    expect(screen.queryByRole('tab', { name: /Body/ })).not.toBeInTheDocument()
  })

  it('says a script skipped the request', () => {
    render(<ResponseViewer tabId="tab" />)
    show(result({ outcome: 'skipped_by_script', ok: false, statusCode: 0, tests: [], testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 } }))
    const panel = screen.getByRole('alert', { name: 'Script outcome' })
    expect(panel).toHaveTextContent('Request Skipped by a Script')
    expect(panel).toHaveTextContent('pm.execution.skipRequest()')
    expect(within(panel).queryByRole('list')).not.toBeInTheDocument()
    expect(screen.queryByRole('list', { name: 'Test results' })).not.toBeInTheDocument()
  })

  it('keeps the response and lists post-response script errors above it', () => {
    render(<ResponseViewer tabId="tab" />)
    show(
      result({
        scriptErrors: [
          { kind: 'script_timeout', message: 'Script execution timed out', line: null, column: null, scriptLabel: 'Billing · Post-response' },
          { kind: 'script_runtime', message: 'TypeError: x', line: 1, column: 1, scriptLabel: 'Create · Post-response' },
        ],
      }),
    )
    const panel = screen.getByRole('alert', { name: 'Script outcome' })
    expect(panel).toHaveTextContent('Scripts Failed')
    expect(within(panel).getAllByRole('listitem')[0]).toHaveTextContent(/^Billing · Post-responseScript execution timed out$/)
    expect(screen.getByText('404 Not Found')).toBeInTheDocument()
    show(result({ scriptErrors: [{ kind: 'script_runtime', message: 'TypeError: x', line: 1, column: 1, scriptLabel: 'Create · Post-response' }] }))
    expect(screen.getByRole('alert', { name: 'Script outcome' })).toHaveTextContent('A Script Failed')
  })
})
