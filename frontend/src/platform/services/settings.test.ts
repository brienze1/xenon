import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/settingsservice', () => ({
  Get: vi.fn(),
  Save: vi.fn(),
}))

const SettingsService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/settingsservice')
const settings = await import('./settings')
const { AppError } = await import('../errors')

describe('settings service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('reads the settings, falling back to the defaults on an empty answer', async () => {
    vi.mocked(SettingsService.Get).mockResolvedValueOnce({ script_timeout_ms: 2500 } as never)
    await expect(settings.getSettings()).resolves.toEqual({ scriptTimeoutMs: 2500 })
    vi.mocked(SettingsService.Get).mockResolvedValueOnce(null as never)
    await expect(settings.getSettings()).resolves.toEqual({ scriptTimeoutMs: 10000 })
  })

  it('saves the settings and maps errors', async () => {
    vi.mocked(SettingsService.Save).mockResolvedValueOnce({ script_timeout_ms: 30000 } as never)
    await expect(settings.saveSettings({ scriptTimeoutMs: 30000 })).resolves.toEqual({ scriptTimeoutMs: 30000 })
    expect(SettingsService.Save).toHaveBeenCalledWith({ script_timeout_ms: 30000 })
    vi.mocked(SettingsService.Save).mockRejectedValueOnce(new AppError('SCR-01400', 'out of range'))
    await expect(settings.saveSettings({ scriptTimeoutMs: 5 })).rejects.toMatchObject({ code: settings.SCRIPT_TIMEOUT_OUT_OF_RANGE })
  })
})
