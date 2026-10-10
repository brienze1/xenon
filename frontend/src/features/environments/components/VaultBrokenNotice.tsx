import { useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { useEnvironmentsStore } from '@/features/environments/store'
import styles from './VaultBrokenNotice.module.css'

const RESET_NEXT_STEP = 'Try again, or restart the app.'

export function VaultBrokenNotice() {
  const localId = useWorkspacesStore((state) => state.activeLocalId)
  const broken = useEnvironmentsStore((state) => localId !== null && state.vault.brokenLocalIds.includes(localId))
  const resetVault = useEnvironmentsStore((state) => state.resetVault)
  const [confirming, setConfirming] = useState(false)
  const [error, setError] = useState<AppError | null>(null)
  const [resetting, startReset] = useTransition()

  if (!broken) {
    return null
  }

  function reset() {
    startReset(async () => {
      const failure = await resetVault()
      setError(failure)
      setConfirming(false)
    })
  }

  return (
    <Banner.Root tone="error" className={styles.notice}>
      <Banner.Message>
        {confirming ? (
          'Reset the vault? Every stored secret value of this workspace is discarded; the secret variables stay and need new values.'
        ) : (
          <>
            The stored secret values of this workspace cannot be decrypted: the key was lost, the passphrase changed or the vault file is
            damaged. Reset the vault to start again with empty secret values.
          </>
        )}
        {error ? (
          <span className={styles.error}>
            <NoTranslateText text={withNextStep(error.description || error.code, RESET_NEXT_STEP)} terms={[error.code]} />
          </span>
        ) : null}
      </Banner.Message>
      <Banner.Actions>
        {confirming ? (
          <>
            <Button variant="secondary" size="sm" onClick={() => setConfirming(false)}>
              Cancel
            </Button>
            <Button variant="danger" size="sm" disabled={resetting} onClick={reset}>
              Discard Secret Values
            </Button>
          </>
        ) : (
          <Button variant="danger" size="sm" onClick={() => setConfirming(true)}>
            Reset Vault
          </Button>
        )}
      </Banner.Actions>
    </Banner.Root>
  )
}
