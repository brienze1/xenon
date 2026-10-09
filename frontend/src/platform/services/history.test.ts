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

const okEntry = { id: 'h2', sent_at: '2026-10-09T10:00:00Z', method: 'POST', url: 'https://x/users', status_code: 201, source_request_id: 'r1', total_ms: 12, size_bytes: 40 }
const failedEntry = { id: 'h1', sent_at: '2026-10-09T09:00:00Z', method: 'GET', url: '', error_kind: 'invalid_url', source_request_id: '', total_ms: 0, size_bytes: 0 }

describe('history service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists entries newest first with query, limit and cursor', async () => {
    vi.mocked(HistoryService.List).mockResolvedValue([okEntry, failedEntry, { ...failedEntry, id: 'h0', error_kind: 'weird' }] as never)
    const entries = await listHistory('L1', { query: 'post', limit: 50, beforeId: 'h9' })
    expect(HistoryService.List).toHaveBeenCalledWith({ local_id: 'L1', query: 'post', limit: 50, before_id: 'h9' })
    expect(entries).toEqual([
      { id: 'h2', sentAt: '2026-10-09T10:00:00Z', method: 'POST', url: 'https://x/users', statusCode: 201, errorKind: null, totalMs: 12, sizeBytes: 40, sourceRequestId: 'r1' },
      { id: 'h1', sentAt: '2026-10-09T09:00:00Z', method: 'GET', url: '', statusCode: null, errorKind: 'invalid_url', totalMs: 0, sizeBytes: 0, sourceRequestId: null },
      expect.objectContaining({ id: 'h0', errorKind: 'other' }),
    ])
  })

  it('lists everything by default and tolerates an empty answer', async () => {
    vi.mocked(HistoryService.List).mockResolvedValue(null as never)
    await expect(listHistory('L1')).resolves.toEqual([])
    expect(HistoryService.List).toHaveBeenCalledWith({ local_id: 'L1', query: '', limit: 0 })
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
    expect(HistoryService.Clear).toHaveBeenCalledWith({ local_id: 'L1' })
  })
})
