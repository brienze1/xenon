import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './ConfirmSendDialog.module.css'

interface ConfirmSendDialogProps {
  tabId: string
}

const NEON_COLORS: ReadonlySet<string> = new Set(['green', 'cyan', 'magenta', 'yellow', 'red'])

/** Asks before sending to an environment with "Confirm before send" (Go refused the send with SND-05409). */
export function ConfirmSendDialog({ tabId }: ConfirmSendDialogProps) {
  const confirmation = useRequestBuilderStore((state) => state.tabs[tabId]?.confirmation ?? null)
  const send = useRequestBuilderStore((state) => state.send)
  const dismiss = useRequestBuilderStore((state) => state.dismissConfirmation)
  const close = () => dismiss(tabId)

  function confirm() {
    if (confirmation === null) {
      return
    }
    const missingSecrets = useRequestBuilderStore.getState().tabs[tabId]?.acknowledgement.missingSecrets === true
    dismiss(tabId)
    void send(tabId, { confirmedEnvironmentId: confirmation.environmentId, missingSecrets })
  }

  const accent = confirmation !== null && NEON_COLORS.has(confirmation.color) ? styles[confirmation.color] : undefined

  return (
    <Dialog.Root open={confirmation !== null} onOpenChange={(open) => (open ? undefined : close())} className={cn(styles.dialog, accent)}>
      {confirmation ? (
        <>
          <Dialog.Title>
            Send to <span translate="no">{confirmation.name}</span>?
          </Dialog.Title>
          <Dialog.Body>
            <p>
              The <strong className={styles.environment} translate="no">{confirmation.name}</strong> environment asks for a confirmation before every
              send.
            </p>
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={close}>
              Cancel
            </Button>
            <Button variant="danger" onClick={confirm}>
              Send to <span translate="no">{confirmation.name}</span>
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
