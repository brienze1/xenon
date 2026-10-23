import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useRunnerStore } from '@/features/runner/store'
import styles from './ConfirmRunDialog.module.css'

const NEON_COLORS: ReadonlySet<string> = new Set(['green', 'cyan', 'magenta', 'yellow', 'red'])

interface ConfirmRunDialogProps {
  itemId: string
}

/** One confirmation for the whole run when its environment asks before every send (SND-05409); Cancel starts nothing. */
export function ConfirmRunDialog({ itemId }: ConfirmRunDialogProps) {
  const confirmation = useRunnerStore((state) => state.targets[itemId]?.confirmation ?? null)
  const confirmStart = useRunnerStore((state) => state.confirmStart)
  const dismiss = useRunnerStore((state) => state.dismissConfirmation)
  const close = () => dismiss(itemId)
  const accent = confirmation !== null && NEON_COLORS.has(confirmation.color) ? styles[confirmation.color] : undefined

  return (
    <Dialog.Root open={confirmation !== null} onOpenChange={(open) => (open ? undefined : close())} className={cn(styles.dialog, accent)}>
      {confirmation ? (
        <>
          <Dialog.Title>
            Run with <span translate="no">{confirmation.name}</span>?
          </Dialog.Title>
          <Dialog.Body>
            <p>
              The <strong className={styles.environment} translate="no">{confirmation.name}</strong> environment asks for a confirmation before every
              send. Confirming once covers every request of this run.
            </p>
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={close}>
              Cancel
            </Button>
            <Button variant="danger" onClick={() => void confirmStart(itemId)}>
              Run with <span translate="no">{confirmation.name}</span>
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
