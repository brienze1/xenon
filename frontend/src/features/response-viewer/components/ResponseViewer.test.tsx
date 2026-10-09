import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', () => ({ lastResult: vi.fn(), releaseResponse: vi.fn(), saveResponseBody: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))

const requests = await import('@/platform/services/requests')
const dialogs = await import('@/platform/services/dialogs')
const responseBody = await import('@/platform/services/response-body')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { ResponseViewer } = await import('./ResponseViewer')
const { ResponseHeaders } = await import('./ResponseHeaders')

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

  it('shows the status, the size and a timing breakdown on hover', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result({ statusCode: 404, statusText: 'Not Found' }))
    expect(screen.getByText('404 Not Found')).toBeInTheDocument()
    expect(screen.getByText('9 B')).toBeInTheDocument()
    await user.hover(screen.getByRole('button', { name: 'Time 1.52\u00a0s' }))
    const tooltip = screen.getByRole('tooltip')
    for (const phase of ['DNS lookup', 'TCP connection', 'TLS handshake', 'Time to first byte', 'Download', 'Total']) {
      expect(within(tooltip).getByText(phase)).toBeInTheDocument()
    }
    expect(within(tooltip).getByText('40 ms')).toBeInTheDocument()
    show(result({ statusCode: 302, statusText: 'Found' }))
    expect(screen.getByText('302 Found')).toBeInTheDocument()
    show(result())
    expect(screen.getByText('200 OK')).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'Response status' })).toBeInTheDocument()
  })

  it('pretty prints by default and shows the raw body as received', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result())
    expect(screen.getByRole('radio', { name: 'Pretty' })).toBeChecked()
    expect(screen.getByRole('textbox', { name: 'Pretty response body' })).toHaveTextContent('"a": [')
    await user.click(screen.getByRole('radio', { name: 'Raw' }))
    expect(screen.getByRole('textbox', { name: 'Raw response body' })).toHaveTextContent('{"a":[1]}')
  })

  it('moves between body views with the arrow keys', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result())
    expect(screen.getByRole('group', { name: 'Body view' })).toBeInTheDocument()
    screen.getByRole('radio', { name: 'Pretty' }).focus()
    await user.keyboard('{ArrowRight}')
    expect(screen.getByRole('radio', { name: 'Raw' })).toHaveFocus()
    expect(screen.getByRole('radio', { name: 'Raw' })).toBeChecked()
    await user.keyboard('{ArrowDown}')
    expect(screen.getByRole('radio', { name: 'Preview' })).toBeChecked()
    await user.keyboard('{ArrowDown}')
    expect(screen.getByRole('radio', { name: 'Pretty' })).toHaveFocus()
    await user.keyboard('{ArrowUp}')
    expect(screen.getByRole('radio', { name: 'Preview' })).toBeChecked()
  })

  it('opens bodies larger than 5 MB in Raw first and loads them from the route', async () => {
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('big body')
    render(<ResponseViewer tabId="tab" />)
    show(result({ sizeBytes: 6 * 1024 * 1024, inlineBody: null, contentType: 'text/plain' }))
    expect(screen.getByRole('radio', { name: 'Raw' })).toBeChecked()
    expect(await screen.findByRole('textbox', { name: 'Raw response body' })).toHaveTextContent('big body')
    expect(responseBody.fetchResponseText).toHaveBeenCalledWith('/xenon/response/tab')
  })

  it('handles binary bodies with an image preview', async () => {
    const user = userEvent.setup()
    render(<ResponseViewer tabId="tab" />)
    show(result({ isBinary: true, inlineBody: null, contentType: 'image/png', sizeBytes: 70 }))
    expect(screen.getByText('Binary body (70 B). Use Preview or Save to File.')).toBeInTheDocument()
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
    expect(screen.getByText(/No preview available for/)).toHaveTextContent('No preview available for application/json.')
    expect(screen.getByText('application/json')).toHaveAttribute('translate', 'no')
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
    expect(screen.getByRole('status')).toHaveTextContent('Xenon copied the response body to the clipboard.')
    await user.click(screen.getByRole('button', { name: 'Save to File' }))
    expect(requests.saveResponseBody).toHaveBeenCalledWith('tab', '/out/response.json')
  })

  it('tells the user when the clipboard refuses the copy', async () => {
    const user = userEvent.setup()
    vi.spyOn(navigator.clipboard, 'writeText').mockRejectedValue(new Error('denied'))
    render(<ResponseViewer tabId="tab" />)
    show(result())
    await user.click(screen.getByRole('button', { name: 'Copy' }))
    expect(screen.getByRole('button', { name: 'Copy Failed' })).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveTextContent('Xenon couldn’t copy to the clipboard')
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
    ['timeout', 'Timed out after 5000 ms', 'Increase the timeout in the request Settings (0 = no timeout)', 'Request Timed Out', ''],
    ['tls', 'Certificate not trusted', 'Turn off SSL certificate verification in Settings to send anyway', 'SSL/TLS Error', ''],
    ['dns', 'Could not resolve host x.invalid', '', 'Could Not Resolve Host', 'Check the host name in the URL and your network connection, then send again.'],
    ['cancelled', 'Request cancelled', '', 'Request Cancelled', 'Send the request again when you are ready.'],
  ] as const)('explains a %s error with its hint or a next step', (kind, message, hint, title, nextStep) => {
    render(<ResponseViewer tabId="tab" />)
    show(result({ ok: false, statusCode: 0, error: { kind, message, hint } }))
    const alert = screen.getByRole('alert', { name: 'Request error' })
    expect(within(alert).getByRole('heading', { name: title })).toBeInTheDocument()
    expect(alert).toHaveTextContent(message)
    expect(alert).toHaveTextContent(hint || nextStep)
  })
})

describe('ResponseHeaders', () => {
  it('says when the response has no headers', () => {
    render(<ResponseHeaders headers={[]} />)
    expect(screen.getByText('No headers')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })
})
