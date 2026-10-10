import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { VaultStatus } from '@/platform/mappers/vault'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({
  MIN_PASSPHRASE_LENGTH: 8,
  PASSPHRASE_MISMATCH: 'VLT-00400',
  PASSPHRASE_TOO_SHORT: 'VLT-01400',
  WRONG_PASSPHRASE: 'VLT-02401',
  PASSPHRASE_SETUP_REQUIRED: 'VLT-05409',
  VAULT_LOCKED: 'VLT-04409',
  setupPassphrase: vi.fn(),
  unlockVault: vi.fn(),
}))

const vault = await import('@/platform/services/vault')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { PassphraseDialog } = await import('./PassphraseDialog')

const unlocked: VaultStatus = { mode: 'passphrase', needsPassphraseSetup: false, lockedLocalIds: [], brokenLocalIds: [] }

describe('PassphraseDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ passphrasePrompt: null, dismissedPrompts: [] })
  })

  afterEach(() => {
    useEnvironmentsStore.setState({ passphrasePrompt: null, dismissedPrompts: [] })
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('stays closed without a prompt', () => {
    render(<PassphraseDialog />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('asks for a new passphrase twice and checks length and match before saving', async () => {
    render(<PassphraseDialog />)
    act(() => useEnvironmentsStore.getState().openPassphrase('setup'))
    expect(screen.getByRole('dialog', { name: 'Choose a Vault Passphrase' })).toBeInTheDocument()
    const passphrase = screen.getByLabelText('Passphrase')
    const confirmation = screen.getByLabelText('Confirm Passphrase')
    expect(passphrase).toHaveFocus()
    await userEvent.type(passphrase, 'short')
    await userEvent.click(screen.getByRole('button', { name: 'Save Passphrase' }))
    expect(screen.getByRole('alert')).toHaveTextContent('Use at least 8 characters.')
    await userEvent.type(passphrase, ' but long')
    await userEvent.type(confirmation, 'something else')
    await userEvent.click(screen.getByRole('button', { name: 'Save Passphrase' }))
    expect(screen.getByRole('alert')).toHaveTextContent('The two passphrases do not match.')
    expect(vault.setupPassphrase).not.toHaveBeenCalled()
    await userEvent.clear(confirmation)
    await userEvent.type(confirmation, 'short but long')
    vi.mocked(vault.setupPassphrase).mockResolvedValue(unlocked)
    await userEvent.click(screen.getByRole('button', { name: 'Save Passphrase' }))
    expect(vault.setupPassphrase).toHaveBeenCalledWith('short but long', 'short but long')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('asks for a passphrase before unlocking with an empty one', async () => {
    render(<PassphraseDialog />)
    act(() => useEnvironmentsStore.getState().openPassphrase('unlock'))
    await userEvent.click(screen.getByRole('button', { name: 'Unlock' }))
    const passphrase = screen.getByLabelText('Passphrase')
    expect(screen.getByRole('alert')).toHaveTextContent('Type the passphrase.')
    expect(passphrase).toHaveFocus()
    expect(passphrase).toHaveAccessibleDescription('Type the passphrase.')
    expect(vault.unlockVault).not.toHaveBeenCalled()
  })

  it('shows the backend validation of a new passphrase', async () => {
    render(<PassphraseDialog />)
    act(() => useEnvironmentsStore.getState().openPassphrase('setup'))
    await userEvent.type(screen.getByLabelText('Passphrase'), 'long enough')
    await userEvent.type(screen.getByLabelText('Confirm Passphrase'), 'long enough')
    vi.mocked(vault.setupPassphrase).mockRejectedValue(new AppError('VLT-09409', 'a passphrase was already chosen'))
    await userEvent.click(screen.getByRole('button', { name: 'Save Passphrase' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('a passphrase was already chosen')
  })

  it('unlocks the vault and reports a wrong passphrase', async () => {
    render(<PassphraseDialog />)
    act(() => useEnvironmentsStore.getState().applyVaultStatus({ ...unlocked, lockedLocalIds: ['A'] }))
    expect(screen.getByRole('dialog', { name: 'Unlock the Vault' })).toBeInTheDocument()
    expect(screen.queryByLabelText('Confirm Passphrase')).not.toBeInTheDocument()
    vi.mocked(vault.unlockVault).mockRejectedValueOnce(new AppError('VLT-02401', 'wrong passphrase'))
    await userEvent.type(screen.getByLabelText('Passphrase'), 'nope nope{Enter}')
    expect(await screen.findByRole('alert')).toHaveTextContent('That passphrase is not right. Try again.')
    vi.mocked(vault.unlockVault).mockResolvedValueOnce(unlocked)
    await userEvent.clear(screen.getByLabelText('Passphrase'))
    await userEvent.type(screen.getByLabelText('Passphrase'), 'right one{Enter}')
    expect(vault.unlockVault).toHaveBeenLastCalledWith('right one')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('can be dismissed and does not come back for the same status', async () => {
    render(<PassphraseDialog />)
    act(() => useEnvironmentsStore.getState().applyVaultStatus({ ...unlocked, lockedLocalIds: ['A'] }))
    await userEvent.click(screen.getByRole('button', { name: 'Stay Locked' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    act(() => useEnvironmentsStore.getState().applyVaultStatus({ ...unlocked, lockedLocalIds: ['A'] }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
