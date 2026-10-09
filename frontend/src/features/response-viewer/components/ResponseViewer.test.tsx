import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', () => ({ lastResult: vi.fn(), releaseResponse: vi.fn(), saveResponseBody: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickSaveFile: vi.fn() }))
vi.mock('@/platform/response-body', () => ({ fetchResponseText: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const responseBody = await import('@/platform/response-body')
const { useResponseViewerStore } = await import('../store')
const { ResponseViewer } = await import('./ResponseViewer')

function result(overrides: Partial<SendResult> = {}): SendResult {
  return {
    tabId: 'tab',
    ok: true,
    statusCode: 200,
    statusText: 'OK',
    proto: 'HTTP/1.1',
    headers: [
      { key: 'Content-Type', value: 'application/json' },
      { key: 'Set-Cookie', value: 'a=1' },
      { key: 'Set-Cookie', value: 'b=2' },
    ],
    sizeBytes: 9,
    contentType: 'application/json',
    isBinary: false,
    bodyUrl: '/xenon/response/tab',
    inlineBody: '{"a":[1]}',
    timings: { dnsMs: 1, connectMs: 2, tlsMs: 3, firstByteMs: 40, downloadMs: 5, totalMs: 1520 },
    error: null,
    ...overrides,
  }
}

function show(value: SendResult) {
  act(() => useResponseViewerStore.getState().setResult('tab', value))
}

describe('ResponseViewer', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(requests.lastResult).mockResolvedValue(null)
    useResponseViewerStore.getState().reset()
  })

  it('shows the empty state until a response arrives', async () => {
    render(<ResponseViewer tabId="tab" />)
    expect(await screen.findByRole('region', { name: 'No response yet' })).toBeInTheDocument()
    expect(requests.lastResult).toHaveBeenCalledWith('tab')
  })

  it('rehydrates the last result kept by Go when the tab is shown again', async () => {
    vi.mocked(requests.lastResult).mockResolvedValue(result())
    render(<ResponseViewer tabId="tab" />)
    expect(await screen.findByText('200 OK')).toBeInTheDocument()
  })

  it('shows the sending state and binding failures', () => {
    render(<ResponseViewer tabId="tab" />)
    act(() => useResponseViewerStore.getState().startSending('tab'))
    expect(screen.getByRole('status')).toHaveTextContent('Sending request…')
    act(() => useResponseViewerStore.getState().setFailure('tab', new AppError('SND-00400', 'invalid request')))
    expect(screen.getByRole('alert')).toHaveTextContent('invalid request')
  })

  it('shows the status coloured by class, the size and a timing breakdown on hover', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result({ statusCode: 404, statusText: 'Not Found' }))
    expect(screen.getByText('404 Not Found')).toHaveClass('error')
    expect(screen.getByText('9 B')).toBeInTheDocument()
    await user.hover(screen.getByRole('button', { name: 'Time 1.52 s' }))
    const tooltip = screen.getByRole('tooltip')
    for (const phase of ['DNS Lookup', 'TCP Connection', 'TLS Handshake', 'Time to First Byte', 'Download', 'Total']) {
      expect(within(tooltip).getByText(phase)).toBeInTheDocument()
    }
    expect(within(tooltip).getByText('40 ms')).toBeInTheDocument()
    show(result({ statusCode: 302, statusText: 'Found' }))
    expect(screen.getByText('302 Found')).toHaveClass('redirect')
    show(result())
    expect(screen.getByText('200 OK')).toHaveClass('success')
  })

  it('pretty prints by default and shows the raw body as received', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result())
    expect(screen.getByRole('radio', { name: 'Pretty' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByRole('textbox', { name: 'Pretty response body' })).toHaveTextContent('"a": [')
    await user.click(screen.getByRole('radio', { name: 'Raw' }))
    expect(screen.getByRole('textbox', { name: 'Raw response body' })).toHaveTextContent('{"a":[1]}')
  })

  it('opens bodies larger than 5 MB in Raw first and loads them from the route', async () => {
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('big body')
    render(<ResponseViewer tabId="tab" />)
    show(result({ sizeBytes: 6 * 1024 * 1024, inlineBody: null, contentType: 'text/plain' }))
    expect(screen.getByRole('radio', { name: 'Raw' })).toHaveAttribute('aria-checked', 'true')
    expect(await screen.findByRole('textbox', { name: 'Raw response body' })).toHaveTextContent('big body')
    expect(responseBody.fetchResponseText).toHaveBeenCalledWith('/xenon/response/tab')
  })

  it('handles binary bodies with an image preview', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result({ isBinary: true, inlineBody: null, contentType: 'image/png', sizeBytes: 70 }))
    expect(screen.getByText('Binary body (70 B). Use Preview or Save to file.')).toBeInTheDocument()
    expect(responseBody.fetchResponseText).not.toHaveBeenCalled()
    await user.click(screen.getByRole('radio', { name: 'Preview' }))
    expect(screen.getByRole('img', { name: 'Response preview' })).toHaveAttribute('src', '/xenon/response/tab')
  })

  it('previews HTML in a sandboxed frame and explains when there is no preview', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result({ contentType: 'text/html', inlineBody: '<p>hi</p>' }))
    await user.click(screen.getByRole('radio', { name: 'Preview' }))
    expect(screen.getByTitle('Response preview')).toHaveAttribute('sandbox', '')
    show(result())
    await user.click(screen.getByRole('radio', { name: 'Preview' }))
    expect(screen.getByText('No preview available for application/json.')).toBeInTheDocument()
  })

  it('copies the exact body and saves it to a file', async () => {
    const user = userEvent.setup()
    const writeText = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue()
    vi.mocked(dialogs.pickSaveFile).mockResolvedValue('/out/response.json')
    render(<ResponseViewer tabId="tab" />)
    show(result())
    await user.click(screen.getByRole('button', { name: 'Copy' }))
    expect(writeText).toHaveBeenCalledWith('{"a":[1]}')
    expect(screen.getByRole('button', { name: 'Copied' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Save to file' }))
    expect(requests.saveResponseBody).toHaveBeenCalledWith('tab', '/out/response.json')
  })

  it('lists every response header, repeated ones included', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result())
    await user.click(screen.getByRole('tab', { name: 'Headers (3)' }))
    const table = screen.getByRole('table', { name: 'Response headers' })
    expect(within(table).getAllByText('Set-Cookie')).toHaveLength(2)
    expect(within(table).getByText('b=2')).toBeInTheDocument()
  })

  it.each([
    ['timeout', 'Timed out after 5000 ms', 'Increase the timeout in the request Settings (0 = no timeout)', 'Request timed out'],
    ['tls', 'Certificate not trusted', 'Turn off SSL certificate verification in Settings to send anyway', 'SSL/TLS error'],
    ['dns', 'Could not resolve host x.invalid', '', 'Could not resolve host'],
    ['cancelled', 'Request cancelled', '', 'Request cancelled'],
  ] as const)('explains a %s error with its hint', (kind, message, hint, title) => {
    render(<ResponseViewer tabId="tab" />)
    show(result({ ok: false, statusCode: 0, error: { kind, message, hint } }))
    const alert = screen.getByRole('alert', { name: 'Request error' })
    expect(within(alert).getByRole('heading', { name: title })).toBeInTheDocument()
    expect(alert).toHaveTextContent(message)
    if (hint) {
      expect(alert).toHaveTextContent(hint)
    } else {
      expect(alert.querySelectorAll('p')).toHaveLength(1)
    }
  })
})
