import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/historyservice', () => ({
  List: vi.fn(),
  Get: vi.fn(),
  Clear: vi.fn(),
}))

const HistoryService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/historyservice')
const { clearHistory, getHistoryEntry, listHistory } = await import('./history')
const { toRequestDto, emptyDraft } = await import('../mappers/request')

const okEntry = { id: 'h2', kind: 'send', source: 'user', sent_at: '2026-10-09T10:00:00Z', method: 'POST', url: 'https://x/users', status_code: 201, source_request_id: 'r1', total_ms: 12, size_bytes: 40 }
const failedEntry = { id: 'h1', kind: 'send', source: 'user', outcome: 'error', sent_at: '2026-10-09T09:00:00Z', method: 'GET', url: '', error_kind: 'invalid_url', source_request_id: '', total_ms: 0, size_bytes: 0 }

describe('history service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists entries newest first with query, limit and cursor', async () => {
    vi.mocked(HistoryService.List).mockResolvedValue([okEntry, failedEntry, { ...failedEntry, id: 'h0', error_kind: 'weird', outcome: 'stopped_by_script_error' }] as never)
    const entries = await listHistory('L1', { query: 'post', limit: 50, beforeId: 'h9' })
    expect(HistoryService.List).toHaveBeenCalledWith({ local_id: 'L1', source: 'all', query: 'post', limit: 50, before_id: 'h9' })
    expect(entries).toEqual([
      { id: 'h2', sentAt: '2026-10-09T10:00:00Z', kind: 'send', source: 'user', client: null, method: 'POST', url: 'https://x/users', outcome: 'sent', statusCode: 201, errorKind: null, totalMs: 12, sizeBytes: 40, sourceRequestId: 'r1', aliasName: null, activity: null },
      { id: 'h1', sentAt: '2026-10-09T09:00:00Z', kind: 'send', source: 'user', client: null, method: 'GET', url: '', outcome: 'error', statusCode: null, errorKind: 'invalid_url', totalMs: 0, sizeBytes: 0, sourceRequestId: null, aliasName: null, activity: null },
      expect.objectContaining({ id: 'h0', errorKind: 'other', outcome: 'stopped_by_script_error' }),
    ])
  })

  it('lists everything by default and tolerates an empty answer', async () => {
    vi.mocked(HistoryService.List).mockResolvedValue(null as never)
    await expect(listHistory('L1')).resolves.toEqual([])
    expect(HistoryService.List).toHaveBeenCalledWith({ local_id: 'L1', source: 'all', query: '', limit: 0 })
  })

  it('lists one source and maps AI action rows with their activity', async () => {
    const action = {
      id: 'h7',
      kind: 'action',
      source: 'ai',
      client: 'Claude Code',
      sent_at: '2026-10-09T11:00:00Z',
      method: '',
      url: '',
      source_request_id: '',
      outcome: '',
      total_ms: 12,
      size_bytes: 0,
      activity: { tool: 'tree', action: 'create', target: { id: 'r9', name: 'POST /orders', path: 'Payments / Orders' }, outcome: 'ok', duration_ms: 12 },
    }
    vi.mocked(HistoryService.List).mockResolvedValue([action] as never)
    const [entry] = await listHistory('L1', { source: 'ai' })
    expect(HistoryService.List).toHaveBeenCalledWith({ local_id: 'L1', source: 'ai', query: '', limit: 0 })
    expect(entry).toMatchObject({
      kind: 'action',
      source: 'ai',
      client: 'Claude Code',
      activity: { tool: 'tree', action: 'create', target: { id: 'r9', name: 'POST /orders', path: 'Payments / Orders' }, outcome: 'ok', code: null, durationMs: 12, args: null },
    })
  })

  it('gets an action entry with its masked arguments', async () => {
    const activity = { tool: 'env_edit', action: 'set_current', outcome: 'error', code: 'VAR-01400', args: '{"value":"•••••"}', duration_ms: 3 }
    vi.mocked(HistoryService.Get).mockResolvedValue({ entry: { ...okEntry, kind: 'action', source: 'ai', activity }, request: toRequestDto(emptyDraft()), activity } as never)
    const detail = await getHistoryEntry('L1', 'h2')
    expect(detail.activity).toEqual({ tool: 'env_edit', action: 'set_current', target: null, outcome: 'error', code: 'VAR-01400', durationMs: 3, args: '{"value":"•••••"}' })
  })

  it('clears only the chosen source', async () => {
    vi.mocked(HistoryService.Clear).mockResolvedValue()
    await clearHistory('L1', 'ai')
    expect(HistoryService.Clear).toHaveBeenCalledWith({ local_id: 'L1', source: 'ai' })
  })

  it('gets an entry with the request as it was sent', async () => {
    const request = toRequestDto({ ...emptyDraft(), id: 'r1', name: 'Create user', method: 'POST', url: 'https://x/users' })
    vi.mocked(HistoryService.Get).mockResolvedValue({ entry: okEntry, request } as never)
    const detail = await getHistoryEntry('L1', 'h2')
    expect(HistoryService.Get).toHaveBeenCalledWith({ local_id: 'L1', id: 'h2' })
    expect(detail.entry.statusCode).toBe(201)
    expect(detail.request).toMatchObject({ id: 'r1', name: 'Create user', method: 'POST', url: 'https://x/users' })
  })

  it('rejects an empty entry and maps backend errors', async () => {
    vi.mocked(HistoryService.Get).mockResolvedValue(null as never)
    await expect(getHistoryEntry('L1', 'h2')).rejects.toMatchObject({ code: 'UNK-00204' })
    vi.mocked(HistoryService.Clear).mockRejectedValue({ cause: { error: { code: 'HIS-08500', description: 'disk' } } })
    await expect(clearHistory('L1')).rejects.toMatchObject({ code: 'HIS-08500' })
    expect(HistoryService.Clear).toHaveBeenCalledWith({ local_id: 'L1', source: 'all' })
  })
})
