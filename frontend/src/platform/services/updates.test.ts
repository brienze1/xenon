import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/updateservice', () => ({
  GetStatus: vi.fn(),
  Check: vi.fn(),
  CheckOnStartup: vi.fn(),
  Update: vi.fn(),
  Restart: vi.fn(),
  OpenReleasePage: vi.fn(),
}))

const UpdateService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/updateservice')
const updates = await import('./updates')

const dto = {
  state: 'available',
  reason: '',
  current_version: '0.1.4',
  latest_version: '0.1.7',
  release_url: 'https://github.com/brienze1/xenon/releases/tag/v0.1.7',
  folder: '',
  received: 0,
  total: 0,
  automatic: false,
}

const CALLS = [
  ['getUpdateStatus', 'GetStatus'],
  ['checkForUpdates', 'Check'],
  ['checkOnStartup', 'CheckOnStartup'],
  ['startUpdate', 'Update'],
  ['restartToUpdate', 'Restart'],
  ['openReleasePage', 'OpenReleasePage'],
] as const

describe('updates service', () => {
  beforeEach(() => vi.resetAllMocks())

  it.each(CALLS)('%s calls UpdateService.%s and maps the status', async (name, binding) => {
    vi.mocked(UpdateService[binding]).mockResolvedValueOnce(dto as never)
    await expect(updates[name]()).resolves.toMatchObject({ state: 'available', currentVersion: '0.1.4', latestVersion: '0.1.7', reason: null })
    expect(UpdateService[binding]).toHaveBeenCalledTimes(1)
  })

  it('maps binding failures to AppError', async () => {
    vi.mocked(UpdateService.Check).mockRejectedValueOnce(new Error('call failed', { cause: JSON.stringify({ error: { code: 'UNK-00500', description: 'boom' } }) }))
    await expect(updates.checkForUpdates()).rejects.toMatchObject({ name: 'AppError', code: 'UNK-00500' })
  })
})
