import type { ReactNode } from 'react'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import styles from './ItemSettingsDialog.module.css'
import { SettingsForm, type ItemAuthSlot, type ItemScriptsSlot } from './SettingsForm'

interface ItemSettingsDialogProps {
  itemId: string | null
  onClose: () => void
  collectionVariables?: (collectionId: string) => ReactNode
  scriptsEditor?: (slot: ItemScriptsSlot) => ReactNode
  authEditor?: (slot: ItemAuthSlot) => ReactNode
}

export function ItemSettingsDialog({ itemId, onClose, collectionVariables, scriptsEditor, authEditor }: ItemSettingsDialogProps) {
  return (
    <Dialog.Root open={itemId !== null} onOpenChange={(open) => (open ? undefined : onClose())} className={styles.dialog}>
      {itemId ? <SettingsForm key={itemId} itemId={itemId} onClose={onClose} collectionVariables={collectionVariables} scriptsEditor={scriptsEditor} authEditor={authEditor} /> : null}
    </Dialog.Root>
  )
}
