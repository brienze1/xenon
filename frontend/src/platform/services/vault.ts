import * as VaultService from '../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/vaultservice'
import { toAppError } from '../errors'
import { toVaultStatus, type VaultStatus } from '../mappers/vault'

export const PASSPHRASE_MISMATCH = 'VLT-00400'
export const PASSPHRASE_TOO_SHORT = 'VLT-01400'
export const WRONG_PASSPHRASE = 'VLT-02401'
export const SECRET_WITHOUT_VALUE = 'VLT-03404'
export const VAULT_LOCKED = 'VLT-04409'
export const PASSPHRASE_SETUP_REQUIRED = 'VLT-05409'
export const VAULT_BROKEN = 'VLT-06409'
export const MIN_PASSPHRASE_LENGTH = 8

async function call<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation()
  } catch (error) {
    throw toAppError(error)
  }
}

export function getVaultStatus(): Promise<VaultStatus> {
  return call(async () => toVaultStatus(await VaultService.Status()))
}

export function setupPassphrase(passphrase: string, confirmation: string): Promise<VaultStatus> {
  return call(async () => toVaultStatus(await VaultService.SetupPassphrase({ passphrase, confirmation })))
}

export function unlockVault(passphrase: string): Promise<VaultStatus> {
  return call(async () => toVaultStatus(await VaultService.Unlock({ passphrase })))
}

/** The only call that returns a secret value: keep it out of stores and logs. */
export function revealSecret(localId: string, scope: string, key: string): Promise<string> {
  return call(async () => (await VaultService.Reveal({ local_id: localId, scope, key }))?.value ?? '')
}

/** Go writes the value to the clipboard itself, so it never crosses the bridge. */
export function copySecret(localId: string, scope: string, key: string): Promise<void> {
  return call(() => VaultService.CopySecret({ local_id: localId, scope, key }))
}

/** Discards every stored secret value of the workspace. */
export function resetVault(localId: string): Promise<VaultStatus> {
  return call(async () => toVaultStatus(await VaultService.Reset({ local_id: localId, confirm: true })))
}
