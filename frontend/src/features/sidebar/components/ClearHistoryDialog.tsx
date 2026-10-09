import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'

interface ClearHistoryDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}

export function ClearHistoryDialog({ open, onCancel, onConfirm }: ClearHistoryDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onCancel())}>
      <Dialog.Title>Clear History?</Dialog.Title>
      <Dialog.Body>
        <span translate="no">Xenon</span> removes every entry you sent in this workspace from this machine. You can’t undo this.
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="danger" onClick={onConfirm}>
          Clear History
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
