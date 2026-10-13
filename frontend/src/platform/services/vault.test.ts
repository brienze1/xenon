import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/vaultservice', () => ({
  Status: vi.fn(),
  SetupPassphrase: vi.fn(),
  Unlock: vi.fn(),
  Reveal: vi.fn(),
  CopySecret: vi.fn(),
  Reset: vi.fn(),
}))

const VaultService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/vaultservice')
const vault = await import('./vault')

const lockedDto = { mode: 'passphrase', needs_passphrase_setup: false, locked_local_ids: ['L1'], broken_local_ids: null }
const locked = { mode: 'passphrase', needsPassphraseSetup: false, lockedLocalIds: ['L1'], brokenLocalIds: [] }

describe('vault service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('reads the status, mapping unknown modes and empty answers', async () => {
    vi.mocked(VaultService.Status).mockResolvedValueOnce(lockedDto as never)
    await expect(vault.getVaultStatus()).resolves.toEqual(locked)
    vi.mocked(VaultService.Status).mockResolvedValueOnce({ ...lockedDto, mode: 'weird', locked_local_ids: null } as never)
    await expect(vault.getVaultStatus()).resolves.toMatchObject({ mode: 'none', lockedLocalIds: [] })
    vi.mocked(VaultService.Status).mockResolvedValueOnce(null as never)
    await expect(vault.getVaultStatus()).resolves.toEqual({ mode: 'none', needsPassphraseSetup: false, lockedLocalIds: [], brokenLocalIds: [] })
  })

  it('sets up and unlocks with a passphrase', async () => {
    vi.mocked(VaultService.SetupPassphrase).mockResolvedValue({ ...lockedDto, locked_local_ids: [] } as never)
    await expect(vault.setupPassphrase('long enough', 'long enough')).resolves.toMatchObject({ lockedLocalIds: [] })
    expect(VaultService.SetupPassphrase).toHaveBeenCalledWith({ passphrase: 'long enough', confirmation: 'long enough' })
    vi.mocked(VaultService.Unlock).mockResolvedValue({ ...lockedDto, locked_local_ids: [] } as never)
    await expect(vault.unlockVault('long enough')).resolves.toMatchObject({ lockedLocalIds: [] })
    expect(VaultService.Unlock).toHaveBeenCalledWith({ passphrase: 'long enough' })
  })

  it('reveals and copies one secret', async () => {
    vi.mocked(VaultService.Reveal).mockResolvedValueOnce({ value: 's3cr3t' } as never).mockResolvedValueOnce(null as never)
    await expect(vault.revealSecret('L1', 'env:e1', 'api_key')).resolves.toBe('s3cr3t')
    expect(VaultService.Reveal).toHaveBeenCalledWith({ local_id: 'L1', scope: 'env:e1', key: 'api_key' })
    await expect(vault.revealSecret('L1', 'env:e1', 'api_key')).resolves.toBe('')
    vi.mocked(VaultService.CopySecret).mockResolvedValue(undefined as never)
    await vault.copySecret('L1', 'globals', 'token')
    expect(VaultService.CopySecret).toHaveBeenCalledWith({ local_id: 'L1', scope: 'globals', key: 'token' })
  })

  it('resets a workspace vault with confirmation', async () => {
    vi.mocked(VaultService.Reset).mockResolvedValue({ ...lockedDto, locked_local_ids: [] } as never)
    await expect(vault.resetVault('L1')).resolves.toMatchObject({ brokenLocalIds: [] })
    expect(VaultService.Reset).toHaveBeenCalledWith({ local_id: 'L1', confirm: true })
  })

  it('maps vault errors', async () => {
    vi.mocked(VaultService.Unlock).mockRejectedValue({ cause: { error: { code: 'VLT-02401', description: 'wrong passphrase' } } })
    await expect(vault.unlockVault('nope nope')).rejects.toMatchObject({ code: vault.WRONG_PASSPHRASE })
    vi.mocked(VaultService.Reveal).mockRejectedValue({ cause: { error: { code: 'VLT-04409', description: 'locked' } } })
    await expect(vault.revealSecret('L1', 'globals', 'k')).rejects.toMatchObject({ code: vault.VAULT_LOCKED })
  })
})
