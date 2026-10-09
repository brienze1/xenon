import type { TreeNode, TreeNodeKind } from '@/platform/mappers/tree'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { isContainer } from '@/features/sidebar/lib/tree-rows'

const DELETE_LABELS: Record<TreeNodeKind, string> = {
  collection: 'Delete Collection',
  folder: 'Delete Folder',
  request: 'Delete Request',
  error: 'Delete File',
}

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
            <span translate="no">Xenon</span> deletes {isContainer(item) ? 'its folder and everything inside it' : 'its file'} from disk.
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={onCancel}>
              Cancel
            </Button>
            <Button variant="danger" onClick={() => onConfirm(item)}>
              {DELETE_LABELS[item.kind]}
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
