import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {}, Clipboard: { SetText: vi.fn() } }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aiaccessservice', () => ({ Status: vi.fn(), PendingApprovals: vi.fn(), AnswerApproval: vi.fn() }))

const runtime = await import('@wailsio/runtime')
const AIAccessService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aiaccessservice')
const service = await import('./ai-access')

describe('ai-access service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('reads the status with the client configs', async () => {
    vi.mocked(AIAccessService.Status).mockResolvedValueOnce({
      state: 'running',
      address: 'http://127.0.0.1:47821/mcp',
      clients: 1,
      stdio_command: '/Applications/Xenon mcp',
      http_config: '{"http":1}',
      stdio_config: '{"stdio":1}',
    } as never)
    await expect(service.getAIAccessStatus()).resolves.toEqual({
      state: 'running',
      address: 'http://127.0.0.1:47821/mcp',
      reason: null,
      code: null,
      clients: 1,
      stdioCommand: '/Applications/Xenon mcp',
      httpConfig: '{"http":1}',
      stdioConfig: '{"stdio":1}',
    })
    vi.mocked(AIAccessService.Status).mockResolvedValueOnce(null as never)
    await expect(service.getAIAccessStatus()).resolves.toMatchObject({ state: 'off', stdioCommand: '' })
  })

  it('lists and answers pending approvals', async () => {
    vi.mocked(AIAccessService.PendingApprovals).mockResolvedValueOnce([
      { id: 'a1', client: 'Claude Code', method: 'POST', url: '{{base}}/orders', environment: { id: 'e1', name: 'Production', color: 'red' }, created_at: 't' },
    ] as never)
    await expect(service.listPendingApprovals()).resolves.toEqual([
      { id: 'a1', client: 'Claude Code', method: 'POST', url: '{{base}}/orders', environment: { id: 'e1', name: 'Production', color: 'red' }, createdAt: 't' },
    ])
    vi.mocked(AIAccessService.PendingApprovals).mockResolvedValueOnce(null as never)
    await expect(service.listPendingApprovals()).resolves.toEqual([])
    vi.mocked(AIAccessService.AnswerApproval).mockResolvedValueOnce(undefined as never)
    await service.answerApproval('a1', true)
    expect(AIAccessService.AnswerApproval).toHaveBeenLastCalledWith({ id: 'a1', approved: true })
    vi.mocked(AIAccessService.AnswerApproval).mockRejectedValueOnce({ cause: { error: { code: 'MCP-05409', description: 'already resolved' } } })
    await expect(service.answerApproval('a1', false)).rejects.toMatchObject({ code: service.APPROVAL_ALREADY_RESOLVED })
  })

  it('copies text to the clipboard', async () => {
    vi.mocked(runtime.Clipboard.SetText).mockResolvedValueOnce(undefined as never)
    await expect(service.copyText('{"x":1}')).resolves.toBe(true)
    expect(runtime.Clipboard.SetText).toHaveBeenCalledWith('{"x":1}')
  })
})
