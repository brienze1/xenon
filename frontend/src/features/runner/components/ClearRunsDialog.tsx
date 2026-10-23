import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useRunnerStore } from '@/features/runner/store'

const TITLE = <Dialog.Title>Clear Stored Runs?</Dialog.Title>

const BODY = (
  <Dialog.Body>
    <p>Every stored run of this collection or folder is deleted from this machine, with any responses it kept for this session. This cannot be undone.</p>
  </Dialog.Body>
)

interface ClearRunsDialogProps {
  itemId: string
  open: boolean
  onClose: () => void
}

/** Confirms Clear runs (FR-041a). */
export function ClearRunsDialog({ itemId, open, onClose }: ClearRunsDialogProps) {
  const clearRuns = useRunnerStore((state) => state.clearRuns)

  function confirm() {
    onClose()
    void clearRuns(itemId)
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      {open ? (
        <>
          {TITLE}
          {BODY}
          <Dialog.Actions>
            <Button variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="danger" onClick={confirm}>
              Clear runs
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
