import { GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useDiscardGuard } from '@/features/environments/hooks/use-discard-guard'
import { useEnvironmentsStore } from '@/features/environments/store'
import { DiscardChangesConfirm } from './DiscardChangesConfirm'
import { VariableTable } from './VariableTable'
import styles from './GlobalsDialog.module.css'

const GLOBALS_INTRO = (
  <p className={styles.note}>Globals apply to every request in this workspace, after the active environment and the collection.</p>
)

const GLOBALS_SCOPES = [GLOBALS_SCOPE]

export function GlobalsDialog() {
  const open = useEnvironmentsStore((state) => state.editor === 'globals')
  const closeEditor = useEnvironmentsStore((state) => state.closeEditor)
  return open ? <GlobalsContent onClose={closeEditor} /> : null
}

interface GlobalsContentProps {
  onClose: () => void
}

function GlobalsContent({ onClose }: GlobalsContentProps) {
  const { confirming, guard, discard, keepEditing } = useDiscardGuard(GLOBALS_SCOPES)
  const close = () => guard(onClose)
  return (
    <Dialog.Root open onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      <Dialog.Title>Globals</Dialog.Title>
      <Dialog.Body>
        {GLOBALS_INTRO}
        <VariableTable scope={GLOBALS_SCOPE} label="Global variables" />
      </Dialog.Body>
      <Dialog.Actions>
        {confirming ? (
          <DiscardChangesConfirm onKeepEditing={keepEditing} onDiscard={discard} />
        ) : (
          <Button variant="secondary" onClick={close}>
            Close
          </Button>
        )}
      </Dialog.Actions>
    </Dialog.Root>
  )
}
