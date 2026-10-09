import { Dialog } from '@/shared/ui/dialog/Dialog'
import styles from './ItemSettingsDialog.module.css'
import { SettingsForm } from './SettingsForm'

interface ItemSettingsDialogProps {
  itemId: string | null
  onClose: () => void
}

export function ItemSettingsDialog({ itemId, onClose }: ItemSettingsDialogProps) {
  return (
    <Dialog.Root open={itemId !== null} onOpenChange={(open) => (open ? undefined : onClose())} className={styles.dialog}>
      {itemId ? <SettingsForm key={itemId} itemId={itemId} onClose={onClose} /> : null}
    </Dialog.Root>
  )
}
