import { render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { SendResult } from '@/platform/mappers/request'
import { injectStylesheets } from '@/test/stylesheets'

vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  runResponse: vi.fn(),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const service = await import('@/platform/services/runner')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { runRow } = await import('@/features/runner/test-fixtures')
const { RunRowDetail } = await import('./RunRowDetail')

const SENT_HEADERS = [
  { key: 'Content-Type', value: 'application/json' },
  { key: 'User-Agent', value: 'Xenon/0.1.14' },
  { key: 'Accept', value: '*/*' },
  { key: 'Accept-Encoding', value: 'gzip, deflate, br' },
  { key: 'Connection', value: 'keep-alive' },
  { key: 'Authorization', value: '•••••' },
]

const LONG_NAME = 'X-Very-Long-Custom-Correlation-Header-Name-For-Tracing-Abcdef'
const LONG_VALUE = 'a'.repeat(400)

function sentRow(headers: Array<{ key: string; value: string }>) {
  return runRow(0, {
    name: 'Guideline',
    method: 'POST',
    sentRequest: { method: 'POST', url: 'https://api.test/campaigns/510876/briefings/7/guidelines', headers, body: '{}' },
  })
}

function headerList(): HTMLElement {
  return within(screen.getByRole('region', { name: 'Request' })).getByLabelText('Request headers')
}

const response = (key: string, result: SendResult) => (
  <p>
    Response {key} {result.statusCode}
  </p>
)

describe('RunRowDetail', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRunnerStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows the status, time, size and error of a row whose response was not persisted', () => {
    render(<RunRowDetail runId="R1" row={runRow(0, { name: 'Login', status: 0, sizeBytes: 2048, error: 'connection refused' })} response={response} onClose={vi.fn()} />)
    expect(screen.getByRole('region', { name: 'Response' })).toHaveTextContent('—')
    expect(screen.getByRole('region', { name: 'Response' })).toHaveTextContent('2.00')
    expect(screen.getByRole('alert')).toHaveTextContent('connection refused')
    expect(screen.getByText('Responses were not persisted for this run.')).toBeInTheDocument()
  })

  it('loads a persisted response once and renders it with the run response key', async () => {
    vi.mocked(service.runResponse).mockResolvedValue({ statusCode: 404 } as SendResult)
    render(<RunRowDetail runId="R1" row={runRow(3, { responseBodyUrl: '/body' })} response={response} onClose={vi.fn()} />)
    expect(screen.getByRole('status')).toHaveTextContent('Loading response…')
    expect(await screen.findByText('Response run:R1:3 404')).toBeInTheDocument()
    expect(service.runResponse).toHaveBeenCalledTimes(1)
  })

  it('shows the sent headers as a labelled list of names and values', () => {
    const { container } = render(<RunRowDetail runId="R1" row={sentRow(SENT_HEADERS)} response={response} onClose={vi.fn()} />)
    const list = headerList()
    expect(list.tagName).toBe('DL')
    const terms = within(list).getAllByRole('term')
    expect(terms.map((term) => term.textContent)).toEqual(SENT_HEADERS.map((header) => header.key))
    terms.forEach((term, index) => {
      expect(term.nextElementSibling?.tagName).toBe('DD')
      expect(term.nextElementSibling).toHaveTextContent(SENT_HEADERS[index]!.value)
    })
    expect(container.querySelector('table')).toBeNull()
  })

  it('lays the headers out as a capped name column and a value column', () => {
    injectStylesheets('src/features/runner/components/RunRowDetail.module.css')
    render(<RunRowDetail runId="R1" row={sentRow(SENT_HEADERS)} response={response} onClose={vi.fn()} />)
    const list = getComputedStyle(headerList())
    expect(list.display).toBe('grid')
    expect(list.getPropertyValue('grid-template-columns').replace(/\s+/g, '')).toBe('fit-content(max(40%,16ch))minmax(0,1fr)')
    const name = getComputedStyle(within(headerList()).getAllByRole('term')[0]!)
    const value = getComputedStyle(within(headerList()).getAllByRole('definition')[0]!)
    expect(name.overflowWrap).toBe('anywhere')
    expect(value.overflowWrap).toBe('anywhere')
    expect(['0', '0px']).toContain(name.minWidth)
    expect(['0', '0px']).toContain(value.marginLeft)
  })

  it('shows long header names and values in full, wrapping instead of truncating', () => {
    injectStylesheets('src/features/runner/components/RunRowDetail.module.css')
    render(<RunRowDetail runId="R1" row={sentRow([{ key: LONG_NAME, value: LONG_VALUE }])} response={response} onClose={vi.fn()} />)
    const [term] = within(headerList()).getAllByRole('term')
    const [definition] = within(headerList()).getAllByRole('definition')
    expect(term).toHaveTextContent(LONG_NAME)
    expect(definition!.textContent).toBe(LONG_VALUE)
    for (const element of [term!, definition!]) {
      const style = getComputedStyle(element)
      expect(style.whiteSpace).not.toBe('nowrap')
      expect(style.textOverflow).not.toBe('ellipsis')
      expect(style.overflowWrap).toBe('anywhere')
      expect(element).not.toHaveAttribute('title')
    }
  })
})
