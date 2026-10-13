import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/scriptservice', () => ({
  Check: vi.fn(),
  Reference: vi.fn(),
  Snippets: vi.fn(),
}))

const ScriptService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/scriptservice')
const scripts = await import('./scripts')
const { AppError } = await import('../errors')

describe('scripts service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('checks a script and maps its syntax errors', async () => {
    vi.mocked(ScriptService.Check).mockResolvedValueOnce({ errors: [{ message: 'Unexpected token', line: 2, column: 4 }] } as never)
    await expect(scripts.checkScript('pm.')).resolves.toEqual([{ message: 'Unexpected token', line: 2, column: 4 }])
    expect(ScriptService.Check).toHaveBeenCalledWith({ text: 'pm.' })
  })

  it('reads the reference and the snippets, unknown kinds and phases falling back', async () => {
    vi.mocked(ScriptService.Reference).mockResolvedValueOnce({ members: [{ path: 'pm.test', signature: 'pm.test(name, fn)', description: 'A test', kind: 'macro' }] } as never)
    await expect(scripts.scriptReference()).resolves.toEqual([{ path: 'pm.test', signature: 'pm.test(name, fn)', description: 'A test', kind: 'property' }])
    vi.mocked(ScriptService.Snippets).mockResolvedValueOnce(null as never)
    await expect(scripts.scriptSnippets()).resolves.toEqual([])
  })

  it('maps errors to AppError', async () => {
    vi.mocked(ScriptService.Reference).mockRejectedValueOnce(new AppError('SCR-06500', 'unreadable'))
    await expect(scripts.scriptReference()).rejects.toMatchObject({ code: 'SCR-06500' })
  })
})
