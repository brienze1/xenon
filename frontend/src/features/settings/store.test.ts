import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/settings', () => ({
  SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400',
  HISTORY_USER_CAP_OUT_OF_RANGE: 'SET-02400',
  HISTORY_AI_CAP_OUT_OF_RANGE: 'SET-03400',
  AI_PORT_OUT_OF_RANGE: 'SET-04400',
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
}))
vi.mock('@/platform/services/ai-access', () => ({ getAIAccessStatus: vi.fn(), copyText: vi.fn() }))

const service = await import('@/platform/services/settings')
const aiAccess = await import('@/platform/services/ai-access')
const { defaultAppSettings } = await import('@/platform/mappers/settings')
const { isSettingsDirty, useSettingsStore } = await import('./store')

const saved = { scriptTimeoutMs: 2500, historyUserCap: 5000, historyAiCap: 200, aiAccess: { enabled: true, port: 47821, skipApprovals: false }, autoUpdate: true }
const running = { state: 'running' as const, address: 'http://127.0.0.1:47821/mcp', reason: null, code: null, clients: 0, stdioCommand: 'xenon mcp', httpConfig: '{}', stdioConfig: '{}' }

function settings() {
  return useSettingsStore.getState()
}

describe('settings store', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(aiAccess.getAIAccessStatus).mockResolvedValue(running)
    useSettingsStore.setState({ settings: defaultAppSettings(), status: 'idle', loadError: null, fieldErrors: {}, saving: false, saveError: null, aiStatus: null, pageOpen: false })
    settings().discard()
  })

  it('loads the saved settings into the form and the AI access status', async () => {
    vi.mocked(service.getSettings).mockResolvedValue(saved)
    await settings().load()
    expect(settings()).toMatchObject({ status: 'ready', settings: saved, aiStatus: running })
    expect(settings().draft).toEqual({ scriptTimeoutSeconds: '2.5', historyUserCap: '5000', historyAiCap: '200', aiEnabled: true, aiPort: '47821', skipApprovals: false, autoUpdate: true })
    expect(isSettingsDirty(settings())).toBe(false)
  })

  it('keeps unsaved edits when it loads again', async () => {
    vi.mocked(service.getSettings).mockResolvedValue(saved)
    await settings().load()
    settings().edit({ aiPort: '50000' })
    expect(isSettingsDirty(settings())).toBe(true)
    await settings().load()
    expect(settings().draft.aiPort).toBe('50000')
    settings().discard()
    expect(settings().draft.aiPort).toBe('47821')
    expect(isSettingsDirty(settings())).toBe(false)
  })

  it('keeps the load error', async () => {
    vi.mocked(service.getSettings).mockRejectedValue(new AppError('UNK-00500', 'broken'))
    await settings().load()
    expect(settings()).toMatchObject({ status: 'ready', loadError: { code: 'UNK-00500' } })
  })

  it('validates every field before saving and saves nothing when one is invalid', async () => {
    settings().edit({ scriptTimeoutSeconds: '61', historyUserCap: '99', historyAiCap: '20001', aiPort: '80' })
    await expect(settings().save()).resolves.toEqual({ ok: true, value: false })
    expect(service.saveSettings).not.toHaveBeenCalled()
    expect(settings().fieldErrors).toEqual({
      scriptTimeout: 'Enter a whole number of seconds from 1 to 60.',
      historyUserCap: 'Enter a whole number from 100 to 20,000.',
      historyAiCap: 'Enter a whole number from 100 to 20,000.',
      aiPort: 'Enter a port from 1024 to 65535.',
    })
    settings().edit({ aiPort: '1024' })
    expect(settings().fieldErrors.aiPort).toBeUndefined()
    expect(settings().fieldErrors.scriptTimeout).toBeDefined()
  })

  it('saves every field together', async () => {
    vi.mocked(service.saveSettings).mockImplementation(async (value) => value)
    settings().edit({ scriptTimeoutSeconds: '30', historyUserCap: '10000', historyAiCap: '100', aiEnabled: false, aiPort: '50000', skipApprovals: true, autoUpdate: false })
    await expect(settings().save()).resolves.toEqual({ ok: true, value: true })
    const expected = { scriptTimeoutMs: 30000, historyUserCap: 10000, historyAiCap: 100, aiAccess: { enabled: false, port: 50000, skipApprovals: true }, autoUpdate: false }
    expect(service.saveSettings).toHaveBeenCalledWith(expected)
    expect(settings().settings).toEqual(expected)
    expect(isSettingsDirty(settings())).toBe(false)
  })

  it('shows a Go range error inline and other failures as a save error', async () => {
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SET-03400', 'AI cap out of range'))
    await expect(settings().save()).resolves.toEqual({ ok: true, value: false })
    expect(settings().fieldErrors).toEqual({ historyAiCap: 'Enter a whole number from 100 to 20,000.' })
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SCR-03500', 'app.json could not be written'))
    await expect(settings().save()).resolves.toMatchObject({ ok: false, error: { code: 'SCR-03500' } })
    expect(settings()).toMatchObject({ saving: false, saveError: { code: 'SCR-03500' } })
  })

  it('follows the AI access status events, keeping the client configs', async () => {
    vi.mocked(service.getSettings).mockResolvedValue(saved)
    await settings().load()
    settings().applyStatus({ state: 'error', address: null, reason: 'port 47821 is in use', code: 'MCP-01500', clients: 0 })
    expect(settings().aiStatus).toEqual({ ...running, state: 'error', address: null, reason: 'port 47821 is in use', code: 'MCP-01500' })
  })

  it('copies through the clipboard and reports failures as false', async () => {
    vi.mocked(aiAccess.copyText).mockResolvedValueOnce(true)
    await expect(settings().copy('x')).resolves.toBe(true)
    vi.mocked(aiAccess.copyText).mockRejectedValueOnce(new Error('denied'))
    await expect(settings().copy('x')).resolves.toBe(false)
  })

  it('opens and closes the full-page Settings', () => {
    settings().openPage()
    expect(settings().pageOpen).toBe(true)
    settings().closePage()
    expect(settings().pageOpen).toBe(false)
  })
})

describe('git accounts store', () => {
  it('updates an account identity and reloads, keeping failures', async () => {
    const gitAccounts = await import('@/platform/services/git-accounts')
    const { useGitAccountsStore } = await import('./store')
    const updated = vi.spyOn(gitAccounts, 'updateGitAccount')
    const listed = vi.spyOn(gitAccounts, 'listGitAccounts').mockResolvedValue([])
    updated.mockResolvedValueOnce({} as never)
    await expect(useGitAccountsStore.getState().update('a1', 'Work', 'Luis', 'l@x.test')).resolves.toBeNull()
    expect(updated).toHaveBeenCalledWith('a1', 'Work', 'Luis', 'l@x.test')
    expect(listed).toHaveBeenCalledTimes(1)
    updated.mockRejectedValueOnce(new Error('boom'))
    await expect(useGitAccountsStore.getState().update('a1', '', '', '')).resolves.toMatchObject({ description: 'boom' })
    expect(listed).toHaveBeenCalledTimes(1)
    vi.restoreAllMocks()
  })
})
