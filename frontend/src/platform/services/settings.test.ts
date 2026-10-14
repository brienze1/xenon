import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/settingsservice', () => ({
  Get: vi.fn(),
  Save: vi.fn(),
}))

const SettingsService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/settingsservice')
const settings = await import('./settings')
const { AppError } = await import('../errors')

const saved = { script_timeout_ms: 2500, history_user_cap: 5000, history_ai_cap: 200, ai_access: { enabled: false, port: 50000, skip_approvals: true } }
const model = { scriptTimeoutMs: 2500, historyUserCap: 5000, historyAiCap: 200, aiAccess: { enabled: false, port: 50000, skipApprovals: true } }

describe('settings service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('reads every field, falling back to the defaults for missing ones', async () => {
    vi.mocked(SettingsService.Get).mockResolvedValueOnce(saved as never)
    await expect(settings.getSettings()).resolves.toEqual(model)
    vi.mocked(SettingsService.Get).mockResolvedValueOnce({ script_timeout_ms: 2500 } as never)
    await expect(settings.getSettings()).resolves.toEqual({ scriptTimeoutMs: 2500, historyUserCap: 1000, historyAiCap: 1000, aiAccess: { enabled: true, port: 47821, skipApprovals: false } })
    vi.mocked(SettingsService.Get).mockResolvedValueOnce(null as never)
    await expect(settings.getSettings()).resolves.toMatchObject({ scriptTimeoutMs: 10000 })
  })

  it('saves every field together and maps errors', async () => {
    vi.mocked(SettingsService.Save).mockResolvedValueOnce(saved as never)
    await expect(settings.saveSettings(model)).resolves.toEqual(model)
    expect(SettingsService.Save).toHaveBeenCalledWith(saved)
    vi.mocked(SettingsService.Save).mockRejectedValueOnce(new AppError('SET-04400', 'port out of range'))
    await expect(settings.saveSettings(model)).rejects.toMatchObject({ code: settings.AI_PORT_OUT_OF_RANGE })
  })
})
