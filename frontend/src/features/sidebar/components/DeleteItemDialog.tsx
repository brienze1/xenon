import type { TreeNode } from '@/platform/mappers/tree'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { isContainer } from '@/features/sidebar/lib/tree-rows'

interface DeleteItemDialogProps {
  item: TreeNode | null
  onCancel: () => void
  onConfirm: (item: TreeNode) => void
}

export function DeleteItemDialog({ item, onCancel, onConfirm }: DeleteItemDialogProps) {
  return (
    <Dialog.Root open={item !== null} onOpenChange={(open) => (open ? undefined : onCancel())}>
      {item ? (
        <>
          <Dialog.Title>Delete “{item.name}”?</Dialog.Title>
          <Dialog.Body>
            {isContainer(item)
              ? 'Its folder and everything inside it are deleted from disk.'
              : 'Its file is deleted from disk.'}
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={onCancel}>
              Cancel
            </Button>
            <Button variant="danger" onClick={() => onConfirm(item)}>
              Delete
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
