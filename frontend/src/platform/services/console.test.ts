import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/consoleservice', () => ({
  List: vi.fn(),
  Get: vi.fn(),
  Clear: vi.fn(),
}))

const ConsoleService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/consoleservice')
const consoleService = await import('./console')
const { AppError } = await import('../errors')

const ENTRY = { id: 'e1', send_id: 's1', kind: 'log', level: 'warn', summary: 'hi', created_at: '2026-10-10T00:00:00.000Z', arguments: ['hi'], masked: false }

describe('console service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists entries with the query in Go field names, an empty answer as none', async () => {
    vi.mocked(ConsoleService.List).mockResolvedValueOnce([ENTRY] as never)
    await expect(consoleService.listConsole({ limit: 10, filter: 'logs' })).resolves.toMatchObject([{ id: 'e1', sendId: 's1', kind: 'log', level: 'warn' }])
    expect(ConsoleService.List).toHaveBeenCalledWith({ after_id: '', limit: 10, filter: 'logs', search: '' })
    vi.mocked(ConsoleService.List).mockResolvedValueOnce(null as never)
    await expect(consoleService.listConsole({ limit: 10 })).resolves.toEqual([])
  })

  it('gets one entry and maps a gone entry to its code', async () => {
    vi.mocked(ConsoleService.Get).mockResolvedValueOnce(ENTRY as never)
    await expect(consoleService.getConsoleEntry('e1')).resolves.toMatchObject({ id: 'e1' })
    expect(ConsoleService.Get).toHaveBeenCalledWith({ id: 'e1' })
    vi.mocked(ConsoleService.Get).mockRejectedValueOnce(new AppError('CON-01404', 'gone'))
    await expect(consoleService.getConsoleEntry('e1')).rejects.toMatchObject({ code: consoleService.CONSOLE_ENTRY_GONE })
  })

  it('clears the console', async () => {
    vi.mocked(ConsoleService.Clear).mockResolvedValueOnce(undefined as never)
    await expect(consoleService.clearConsole()).resolves.toBeUndefined()
    expect(ConsoleService.Clear).toHaveBeenCalledOnce()
  })
})
