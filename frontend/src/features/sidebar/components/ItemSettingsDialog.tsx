import type { ReactNode } from 'react'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import styles from './ItemSettingsDialog.module.css'
import { SettingsForm } from './SettingsForm'

interface ItemSettingsDialogProps {
  itemId: string | null
  onClose: () => void
  collectionVariables?: (collectionId: string) => ReactNode
}

export function ItemSettingsDialog({ itemId, onClose, collectionVariables }: ItemSettingsDialogProps) {
  return (
    <Dialog.Root open={itemId !== null} onOpenChange={(open) => (open ? undefined : onClose())} className={styles.dialog}>
      {itemId ? <SettingsForm key={itemId} itemId={itemId} onClose={onClose} collectionVariables={collectionVariables} /> : null}
    </Dialog.Root>
  )
}
