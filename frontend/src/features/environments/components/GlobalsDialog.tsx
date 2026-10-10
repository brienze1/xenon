import { GLOBALS_SCOPE } from '@/platform/mappers/variable'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useEnvironmentsStore } from '@/features/environments/store'
import { VariableTable } from './VariableTable'
import styles from './GlobalsDialog.module.css'

const GLOBALS_INTRO = (
  <p className={styles.note}>Globals apply to every request in this workspace, after the active environment and the collection.</p>
)

export function GlobalsDialog() {
  const open = useEnvironmentsStore((state) => state.editor === 'globals')
  const closeEditor = useEnvironmentsStore((state) => state.closeEditor)
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : closeEditor())} className={styles.dialog}>
      <Dialog.Title>Globals</Dialog.Title>
      <Dialog.Body>
        {GLOBALS_INTRO}
        <VariableTable scope={GLOBALS_SCOPE} label="Global variables" />
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={closeEditor}>
          Close
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
