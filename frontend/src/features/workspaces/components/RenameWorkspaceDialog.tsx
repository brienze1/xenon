import { useState, type FormEvent } from 'react'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { workspaceErrorMessage } from '../lib/error-message'
import { useWorkspacesStore } from '../store'
import styles from './WorkspaceDialog.module.css'

interface RenameWorkspaceDialogProps {
  workspace: WorkspaceSummary
  open: boolean
  onClose: () => void
}

export function RenameWorkspaceDialog({ workspace, open, onClose }: RenameWorkspaceDialogProps) {
  const [name, setName] = useState(workspace.name)
  const error = useWorkspacesStore((state) => state.error)
  const rename = useWorkspacesStore((state) => state.rename)

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (await rename(workspace.localId, name)) {
      onClose()
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <form aria-label="Rename workspace" className={styles.form} onSubmit={(event) => void onSubmit(event)}>
        <Dialog.Title>Rename workspace</Dialog.Title>
        <Dialog.Body>
          <label className={styles.field}>
            <span className={styles.label}>Workspace name</span>
            <input className={styles.input} value={name} maxLength={200} autoComplete="off" onChange={(event) => setName(event.target.value)} />
          </label>
          <p className={styles.hint}>Only the name changes. The folder stays where it is.</p>
          {error ? (
            <Banner.Root tone="error">
              <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
            </Banner.Root>
          ) : null}
        </Dialog.Body>
        <Dialog.Actions>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={name.trim() === ''}>
            Rename
          </Button>
        </Dialog.Actions>
      </form>
    </Dialog.Root>
  )
}
