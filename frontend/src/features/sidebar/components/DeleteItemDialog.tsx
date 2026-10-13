import { useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import type { TreeNode, TreeNodeKind } from '@/platform/mappers/tree'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { isContainer } from '@/features/sidebar/lib/tree-rows'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './DeleteItemDialog.module.css'

const DELETE_LABELS: Record<TreeNodeKind, string> = {
  collection: 'Delete Collection',
  folder: 'Delete Folder',
  request: 'Delete Request',
  example: 'Delete Example',
  error: 'Delete File',
}

const XENON = <span translate="no">Xenon</span>

const DELETE_NEXT_STEP = 'Check that the file is not open in another program, then try again.'

function deleteScope(item: TreeNode): string {
  if (item.kind === 'example') {
    return 'removes it from its request’s file.'
  }
  return isContainer(item) ? 'deletes its folder and everything inside it from disk.' : 'deletes its file from disk.'
}

interface DeleteItemDialogProps {
  item: TreeNode | null
  onCancel: () => void
  onConfirm: (item: TreeNode) => Promise<AppError | null>
}

export function DeleteItemDialog({ item, onCancel, onConfirm }: DeleteItemDialogProps) {
  return (
    <Dialog.Root open={item !== null} onOpenChange={(open) => (open ? undefined : onCancel())}>
      {item ? <DeleteItemConfirm key={item.id} item={item} onCancel={onCancel} onConfirm={onConfirm} /> : null}
    </Dialog.Root>
  )
}

interface DeleteItemConfirmProps {
  item: TreeNode
  onCancel: () => void
  onConfirm: (item: TreeNode) => Promise<AppError | null>
}

function DeleteItemConfirm({ item, onCancel, onConfirm }: DeleteItemConfirmProps) {
  const [deleting, startDelete] = useTransition()
  const [error, setError] = useState<AppError | null>(null)

  function confirm() {
    startDelete(async () => {
      const failure = await onConfirm(item)
      setError(failure)
    })
  }

  return (
    <>
      <Dialog.Title>Delete “{item.name}”?</Dialog.Title>
      <Dialog.Body>
        {XENON} {deleteScope(item)}
        {error ? (
          <Banner.Root tone="error" className={styles.error}>
            <Banner.Message>
              <SidebarErrorText error={error} nextStep={DELETE_NEXT_STEP} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="danger" disabled={deleting} onClick={confirm}>
          {deleting ? 'Deleting…' : DELETE_LABELS[item.kind]}
        </Button>
      </Dialog.Actions>
    </>
  )
}
