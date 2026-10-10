import { useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { MIN_PASSPHRASE_LENGTH, PASSPHRASE_MISMATCH, PASSPHRASE_TOO_SHORT, WRONG_PASSPHRASE } from '@/platform/services/vault'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useEnvironmentsStore } from '@/features/environments/store'
import type { PassphrasePrompt } from '@/features/environments/model'
import styles from './PassphraseDialog.module.css'

const MESSAGES: Record<string, string> = {
  [PASSPHRASE_MISMATCH]: 'The two passphrases do not match. Type the same passphrase twice.',
  [PASSPHRASE_TOO_SHORT]: `Use at least ${MIN_PASSPHRASE_LENGTH} characters.`,
  [WRONG_PASSPHRASE]: 'That passphrase is not right. Try again.',
}

function messageOf(error: AppError): string {
  return MESSAGES[error.code] ?? withNextStep(error.description || error.code, 'Try again.')
}

const SETUP_INTRO = (
  <p>
    The system keychain is not available, so secret values are encrypted with a passphrase. Choose one now; it is never stored, and{' '}
    <span translate="no">Xenon</span> asks for it once each time it starts. If you forget it, the stored secret values cannot be
    recovered.
  </p>
)

const UNLOCK_INTRO = <p>Secret values are locked. Type your vault passphrase to use and reveal them.</p>

export function PassphraseDialog() {
  const prompt = useEnvironmentsStore((state) => state.passphrasePrompt)
  const dismiss = useEnvironmentsStore((state) => state.dismissPassphrase)
  return (
    <Dialog.Root open={prompt !== null} onOpenChange={(open) => (open ? undefined : dismiss())}>
      {prompt ? <PassphraseForm key={prompt} prompt={prompt} onCancel={dismiss} /> : null}
    </Dialog.Root>
  )
}

interface PassphraseFormProps {
  prompt: PassphrasePrompt
  onCancel: () => void
}

function PassphraseForm({ prompt, onCancel }: PassphraseFormProps) {
  const setupPassphrase = useEnvironmentsStore((state) => state.setupPassphrase)
  const unlockVault = useEnvironmentsStore((state) => state.unlockVault)
  const [passphrase, setPassphrase] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, startSubmit] = useTransition()
  const setup = prompt === 'setup'

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (setup && passphrase.length < MIN_PASSPHRASE_LENGTH) {
      setError(MESSAGES[PASSPHRASE_TOO_SHORT] ?? null)
      return
    }
    if (setup && passphrase !== confirmation) {
      setError(MESSAGES[PASSPHRASE_MISMATCH] ?? null)
      return
    }
    startSubmit(async () => {
      const failure = setup ? await setupPassphrase(passphrase, confirmation) : await unlockVault(passphrase)
      setError(failure ? messageOf(failure) : null)
    })
  }

  return (
    <form onSubmit={submit}>
      <Dialog.Title>{setup ? 'Choose a Vault Passphrase' : 'Unlock the Vault'}</Dialog.Title>
      <Dialog.Body>
        {setup ? SETUP_INTRO : UNLOCK_INTRO}
        <div className={styles.fields}>
          <label className={styles.field}>
            Passphrase
            <input
              name="passphrase"
              type="password"
              autoComplete={setup ? 'new-password' : 'current-password'}
              className={styles.input}
              aria-invalid={error ? true : undefined}
              value={passphrase}
              onChange={(event) => setPassphrase(event.target.value)}
            />
          </label>
          {setup ? (
            <label className={styles.field}>
              Confirm Passphrase
              <input
                name="confirmation"
                type="password"
                autoComplete="new-password"
                className={styles.input}
                aria-invalid={error ? true : undefined}
                value={confirmation}
                onChange={(event) => setConfirmation(event.target.value)}
              />
            </label>
          ) : null}
        </div>
        {error ? (
          <Banner.Root tone="error" className={styles.error}>
            <Banner.Message>{error}</Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          {setup ? 'Not Now' : 'Stay Locked'}
        </Button>
        <Button type="submit" variant="primary" disabled={submitting || passphrase === ''}>
          {setup ? 'Save Passphrase' : 'Unlock'}
        </Button>
      </Dialog.Actions>
    </form>
  )
}
