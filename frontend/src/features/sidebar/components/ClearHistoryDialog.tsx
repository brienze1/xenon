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
      <Dialog.Body>Every entry sent in this workspace is removed from this machine. This cannot be undone.</Dialog.Body>
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
