import type { VaultStatus as VaultStatusDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'

export const VAULT_MODES = ['keychain', 'passphrase', 'none'] as const

export type VaultMode = (typeof VAULT_MODES)[number]

export interface VaultStatus {
  mode: VaultMode
  needsPassphraseSetup: boolean
  lockedLocalIds: string[]
  brokenLocalIds: string[]
}

export function emptyVaultStatus(): VaultStatus {
  return { mode: 'none', needsPassphraseSetup: false, lockedLocalIds: [], brokenLocalIds: [] }
}

export function toVaultStatus(dto: VaultStatusDto | null): VaultStatus {
  if (!dto) {
    return emptyVaultStatus()
  }
  return {
    mode: (VAULT_MODES as readonly string[]).includes(dto.mode) ? (dto.mode as VaultMode) : 'none',
    needsPassphraseSetup: dto.needs_passphrase_setup,
    lockedLocalIds: dto.locked_local_ids ?? [],
    brokenLocalIds: dto.broken_local_ids ?? [],
  }
}
