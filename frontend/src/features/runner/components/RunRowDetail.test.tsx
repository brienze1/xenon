import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { SendResult } from '@/platform/mappers/request'

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
})
