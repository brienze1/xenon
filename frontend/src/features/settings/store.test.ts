import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/settings', () => ({ getSettings: vi.fn(), saveSettings: vi.fn() }))

const service = await import('@/platform/services/settings')
const { useSettingsStore } = await import('./store')

describe('settings store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useSettingsStore.setState({ open: false, settings: { scriptTimeoutMs: 10000 }, loading: false, error: null })
  })

  it('opens with the saved settings', async () => {
    vi.mocked(service.getSettings).mockResolvedValue({ scriptTimeoutMs: 2500 })
    await useSettingsStore.getState().openSettings()
    expect(useSettingsStore.getState()).toMatchObject({ open: true, loading: false, settings: { scriptTimeoutMs: 2500 }, error: null })
  })

  it('keeps the dialog open with the error when the settings cannot be read', async () => {
    vi.mocked(service.getSettings).mockRejectedValue(new AppError('UNK-00500', 'broken'))
    await useSettingsStore.getState().openSettings()
    expect(useSettingsStore.getState()).toMatchObject({ open: true, loading: false, error: { code: 'UNK-00500' } })
    useSettingsStore.getState().close()
    expect(useSettingsStore.getState()).toMatchObject({ open: false, error: null })
  })

  it('saves and closes, or returns the error and stays open', async () => {
    useSettingsStore.setState({ open: true })
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SCR-01400', 'out of range'))
    await expect(useSettingsStore.getState().save({ scriptTimeoutMs: 5 })).resolves.toMatchObject({ code: 'SCR-01400' })
    expect(useSettingsStore.getState().open).toBe(true)
    vi.mocked(service.saveSettings).mockResolvedValueOnce({ scriptTimeoutMs: 30000 })
    await expect(useSettingsStore.getState().save({ scriptTimeoutMs: 30000 })).resolves.toBeNull()
    expect(useSettingsStore.getState()).toMatchObject({ open: false, settings: { scriptTimeoutMs: 30000 } })
  })
})
