import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { VaultStatus } from '@/platform/mappers/vault'

vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ getVariableTable: vi.fn() }))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409', resetVault: vi.fn() }))

const vault = await import('@/platform/services/vault')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { VaultBrokenNotice } = await import('./VaultBrokenNotice')

const healthy: VaultStatus = { mode: 'keychain', needsPassphraseSetup: false, lockedLocalIds: [], brokenLocalIds: [] }

describe('VaultBrokenNotice', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'A' })
    useEnvironmentsStore.setState({ vault: { ...healthy, brokenLocalIds: ['A'] } })
  })

  afterEach(() => {
    useEnvironmentsStore.setState({ vault: healthy })
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('shows nothing when the active workspace vault is fine', () => {
    useEnvironmentsStore.setState({ vault: { ...healthy, brokenLocalIds: ['G'] } })
    render(<VaultBrokenNotice />)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('explains the broken vault and resets it after confirmation', async () => {
    render(<VaultBrokenNotice />)
    expect(screen.getByRole('alert')).toHaveTextContent('cannot be decrypted')
    await userEvent.click(screen.getByRole('button', { name: 'Reset Vault' }))
    expect(screen.getByRole('alert')).toHaveTextContent('Every stored secret value of this workspace is discarded')
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(vault.resetVault).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Reset Vault' }))
    vi.mocked(vault.resetVault).mockResolvedValue(healthy)
    await userEvent.click(screen.getByRole('button', { name: 'Discard Secret Values' }))
    expect(vault.resetVault).toHaveBeenCalledWith('A')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('keeps the notice with the error when the reset fails', async () => {
    render(<VaultBrokenNotice />)
    await userEvent.click(screen.getByRole('button', { name: 'Reset Vault' }))
    vi.mocked(vault.resetVault).mockRejectedValue(new AppError('VLT-07500', 'the vault file cannot be written'))
    await userEvent.click(screen.getByRole('button', { name: 'Discard Secret Values' }))
    expect(await screen.findByText(/the vault file cannot be written/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Reset Vault' })).toBeInTheDocument()
  })
})
