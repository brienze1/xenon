import { useState } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useGitAccountsStore } from '@/features/settings/store'
import styles from './GitAccounts.module.css'

export function HostKeyPrompt() {
  const prompt = useGitAccountsStore((state) => state.hostKeyPrompt)
  const answer = useGitAccountsStore((state) => state.answerHostKey)
  const [error, setError] = useState<string | null>(null)

  async function respond(accept: boolean) {
    const failure = await answer(accept)
    setError(failure ? withNextStep(failure.description || failure.code, 'Close and try the connection again.') : null)
  }

  return (
    <Dialog.Root open={prompt !== null} onOpenChange={(next) => (next ? undefined : void respond(false))}>
      {prompt ? (
        <>
          <Dialog.Title>Trust {prompt.host}?</Dialog.Title>
          <Dialog.Body>
            <p className={styles.hint}>
              This is the first connection to {prompt.host}. Compare the {prompt.keyType} key fingerprint with the one your Git host publishes before trusting it.
            </p>
            <p className={styles.publicKey}>{prompt.fingerprint}</p>
            {error ? (
              <Banner.Root tone="error">
                <Banner.Message>{error}</Banner.Message>
              </Banner.Root>
            ) : null}
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={() => void respond(false)}>
              Cancel
            </Button>
            <Button variant="primary" onClick={() => void respond(true)}>
              Trust and remember
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
