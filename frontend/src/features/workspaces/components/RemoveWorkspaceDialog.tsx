import { useState } from 'react'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { workspaceErrorMessage } from '../lib/error-message'
import { TRASH_FAILED, useWorkspacesStore } from '../store'
import styles from './WorkspaceDialog.module.css'

interface RemoveWorkspaceDialogProps {
  workspace: WorkspaceSummary
  open: boolean
  onClose: () => void
}

export function RemoveWorkspaceDialog({ workspace, open, onClose }: RemoveWorkspaceDialogProps) {
  const [moveToTrash, setMoveToTrash] = useState(false)
  const [removing, setRemoving] = useState(false)
  const error = useWorkspacesStore((state) => state.error)
  const remove = useWorkspacesStore((state) => state.remove)

  async function confirm(trash: boolean) {
    setRemoving(true)
    const removed = await remove(workspace.localId, trash)
    setRemoving(false)
    if (removed) {
      onClose()
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <Dialog.Title>Remove “{workspace.name}”?</Dialog.Title>
      <Dialog.Body>
        <fieldset className={styles.choices}>
          <legend className={styles.label}>What happens to the folder</legend>
          <label className={styles.choice}>
            <input type="radio" name="remove-mode" checked={!moveToTrash} onChange={() => setMoveToTrash(false)} />
            Remove from Xenon (the files stay on disk)
          </label>
          <label className={styles.choice}>
            <input type="radio" name="remove-mode" checked={moveToTrash} onChange={() => setMoveToTrash(true)} />
            Remove and move the folder to the trash
          </label>
        </fieldset>
        <p className={styles.path}>{workspace.path}</p>
        {error ? (
          <Banner.Root tone="error">
            <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
            {error.code === TRASH_FAILED ? (
              <Banner.Actions>
                <Button variant="secondary" size="sm" onClick={() => void confirm(false)}>
                  Remove from Xenon
                </Button>
              </Banner.Actions>
            ) : null}
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button variant="danger" disabled={removing} onClick={() => void confirm(moveToTrash)}>
          Remove
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
