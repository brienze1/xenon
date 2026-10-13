import { useState, useTransition } from 'react'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { TRASH_FAILED, useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'
import styles from './RemoveWorkspaceDialog.module.css'

const XENON = <span translate="no">Xenon</span>

const MODE_LEGEND = <legend className={styles.label}>What happens to the folder</legend>

const KEEP_FILES_TEXT = <>Remove from {XENON} (the files stay on disk)</>

const KEEP_FILES_ACTION = <>Remove from {XENON}</>

interface RemoveWorkspaceDialogProps {
  workspace: WorkspaceSummary
  open: boolean
  onClose: () => void
}

export function RemoveWorkspaceDialog({ workspace, open, onClose }: RemoveWorkspaceDialogProps) {
  const [moveToTrash, setMoveToTrash] = useState(false)
  const [removing, startRemove] = useTransition()
  const error = useWorkspacesStore((state) => state.error)
  const remove = useWorkspacesStore((state) => state.remove)

  function confirm(trash: boolean) {
    startRemove(async () => {
      if (await remove(workspace.localId, trash)) {
        onClose()
      }
    })
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <Dialog.Title>Remove “{workspace.name}”?</Dialog.Title>
      <Dialog.Body>
        <fieldset className={styles.choices}>
          {MODE_LEGEND}
          <label className={styles.choice}>
            <input type="radio" name="remove-mode" checked={!moveToTrash} onChange={() => setMoveToTrash(false)} />
            {KEEP_FILES_TEXT}
          </label>
          <label className={styles.choice}>
            <input type="radio" name="remove-mode" checked={moveToTrash} onChange={() => setMoveToTrash(true)} />
            Remove and move the folder to the trash
          </label>
        </fieldset>
        <p className={styles.path} translate="no">
          {workspace.path}
        </p>
        {error ? (
          <Banner.Root tone="error">
            <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
            {error.code === TRASH_FAILED ? (
              <Banner.Actions>
                <Button variant="secondary" size="sm" disabled={removing} onClick={() => confirm(false)}>
                  {KEEP_FILES_ACTION}
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
        <Button variant="danger" disabled={removing} onClick={() => confirm(moveToTrash)}>
          {removing ? 'Removing…' : 'Remove Workspace'}
        </Button>
      </Dialog.Actions>
    </Dialog.Root>
  )
}
