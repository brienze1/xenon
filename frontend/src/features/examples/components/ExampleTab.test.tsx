import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Example } from '@/platform/mappers/example'
import { emptyDraft } from '@/platform/mappers/request'
import { injectStylesheets } from '@/test/stylesheets'

vi.mock('@/platform/services/examples', () => ({ getExample: vi.fn(), saveExample: vi.fn(), createExampleFromResponse: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(async () => []), getRequest: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => null), suggestVariables: vi.fn() }))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const examples = await import('@/platform/services/examples')
const responseBody = await import('@/platform/services/response-body')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useExamplesStore } = await import('@/features/examples/store')
const { ExampleTab } = await import('./ExampleTab')
await import('@/shared/ui/code-editor/CodeEditor')

function example(overrides: Partial<Example['response']> = {}): Example {
  return {
    id: 'e1',
    requestId: 'r1',
    name: 'Created',
    revision: 'rev1',
    request: { ...emptyDraft(), name: 'Created', method: 'POST' },
    response: {
      code: 201,
      status: 'Created',
      contentType: 'application/json',
      headers: [{ id: 'h1', key: 'X-Trace', value: 't-1', description: '', enabled: true }],
      body: '{"id":1}',
      bodyEncoding: 'text',
      bodyUrl: '',
      sizeBytes: 8,
      isBinary: false,
      ...overrides,
    },
  }
}

async function renderTab(value: Example) {
  vi.mocked(examples.getExample).mockResolvedValue(value)
  await act(() => useExamplesStore.getState().open('tab-1', 'r1', 'e1'))
  render(<ExampleTab tabId="tab-1" requestEditor={<p>request editors</p>} />)
}

describe('ExampleTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useExamplesStore.getState().reset()
  })

  it('shows the request slot and an editable response', async () => {
    await renderTab(example())
    expect(screen.getByText('request editors')).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Example Response' })).toBeInTheDocument()
    const code = screen.getByRole('spinbutton', { name: 'Status code' })
    expect(code).toHaveValue(201)
    await userEvent.clear(code)
    await userEvent.type(code, '404')
    await userEvent.clear(screen.getByRole('textbox', { name: 'Status text' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Status text' }), 'Not Found')
    expect(useExamplesStore.getState().tabs['tab-1']?.response).toMatchObject({ code: 404, status: 'Not Found' })
    expect(useExamplesStore.getState().tabs['tab-1']?.dirty).toBe(true)
    expect(await screen.findByRole('textbox', { name: 'Pretty example body' })).toHaveTextContent('"id": 1')
    await userEvent.click(screen.getByRole('radio', { name: 'Raw' }))
    expect(await screen.findByRole('textbox', { name: 'Raw example body' })).toHaveTextContent('{"id":1}')
    await userEvent.click(screen.getByRole('tab', { name: 'Headers (1)' }))
    expect(await screen.findByText('X-Trace')).toBeInTheDocument()
  })

  it('scrolls the example response instead of clipping it', async () => {
    injectStylesheets('src/shared/ui/tabs/Tabs.module.css', 'src/features/examples/components/ExampleTab.module.css')
    await renderTab(example())
    const response = getComputedStyle(screen.getByRole('region', { name: 'Example Response' }))
    expect(response.overflowY).toBe('auto')
    expect(response.overflowX).toBe('hidden')
    expect(getComputedStyle(screen.getByRole('tablist', { name: 'Example response' }).closest('.tabs')!).minHeight).toBe('auto')
    expect(getComputedStyle(screen.getByRole('tabpanel')).minHeight).toBe('160px')
  })

  it('opens a large routed body read-only in Raw first', async () => {
    vi.mocked(responseBody.fetchResponseText).mockResolvedValue('a,b')
    await renderTab(example({ body: null, bodyUrl: '/xenon/response/example:e1', sizeBytes: 8 * 1024 * 1024, contentType: 'text/csv' }))
    expect(screen.getByRole('radio', { name: 'Raw' })).toBeChecked()
    expect(screen.getByText(/shown read-only/)).toBeInTheDocument()
    expect(await screen.findByRole('textbox', { name: 'Raw example body' })).toHaveTextContent('a,b')
    expect(responseBody.fetchResponseText).toHaveBeenCalledWith('/xenon/response/example:e1')
  })

  it('previews a binary image through the response route', async () => {
    await renderTab(example({ body: null, bodyUrl: '/xenon/response/example:e1', contentType: 'image/png', isBinary: true, bodyEncoding: 'base64' }))
    expect(screen.getByText(/Binary body/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('radio', { name: 'Preview' }))
    expect(screen.getByRole('img', { name: 'Example response preview' })).toHaveAttribute('src', '/xenon/response/example:e1')
  })

  it('offers keep mine or load from disk on a revision conflict', async () => {
    await renderTab(example())
    vi.mocked(examples.saveExample).mockRejectedValueOnce(new AppError('REQ-02409', 'changed')).mockResolvedValueOnce(example())
    await act(() => useExamplesStore.getState().save('tab-1'))
    await userEvent.click(screen.getByRole('button', { name: 'Keep Your Version' }))
    expect(vi.mocked(examples.saveExample).mock.calls[1]?.[4]).toBe(true)
    expect(screen.queryByRole('button', { name: 'Keep Your Version' })).toBeNull()
  })

  it('shows a load failure', async () => {
    vi.mocked(examples.getExample).mockRejectedValue(new AppError('EXM-01404', 'the example does not exist'))
    await act(() => useExamplesStore.getState().open('tab-1', 'r1', 'e1'))
    render(<ExampleTab tabId="tab-1" requestEditor={<p>request editors</p>} />)
    expect(screen.getByRole('alert')).toHaveTextContent('the example does not exist. Close this tab and open the example again from the sidebar.')
    expect(screen.queryByText('request editors')).toBeNull()
  })
})
