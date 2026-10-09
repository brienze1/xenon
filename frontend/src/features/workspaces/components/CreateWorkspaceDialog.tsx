import { useState, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { workspaceErrorMessage } from '../lib/error-message'
import { useWorkspacesStore } from '../store'
import styles from './WorkspaceDialog.module.css'

interface CreateWorkspaceDialogProps {
  open: boolean
  onClose: () => void
}

const PICKER_TITLE = 'Choose the folder for the new workspace'

export function CreateWorkspaceDialog({ open, onClose }: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const error = useWorkspacesStore((state) => state.error)
  const create = useWorkspacesStore((state) => state.create)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)

  function close() {
    setName('')
    setPath(null)
    onClose()
  }

  async function onChooseFolder() {
    const picked = await chooseFolder(PICKER_TITLE)
    if (picked) {
      setPath(picked)
    }
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (path && (await create(name, path))) {
      close()
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())}>
      <form aria-label="New workspace" className={styles.form} onSubmit={(event) => void onSubmit(event)}>
        <Dialog.Title>New workspace</Dialog.Title>
        <Dialog.Body>
          <label className={styles.field}>
            <span className={styles.label}>Workspace name</span>
            <input className={styles.input} value={name} maxLength={200} autoComplete="off" onChange={(event) => setName(event.target.value)} />
          </label>
          <div className={styles.field}>
            <span className={styles.label}>Folder</span>
            <div className={styles.folder}>
              <span className={styles.path}>{path ?? 'No folder chosen'}</span>
              <Button size="sm" onClick={() => void onChooseFolder()}>
                Choose folder…
              </Button>
            </div>
          </div>
          {error ? (
            <Banner.Root tone="error">
              <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
            </Banner.Root>
          ) : null}
        </Dialog.Body>
        <Dialog.Actions>
          <Button variant="ghost" onClick={close}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={name.trim() === '' || path === null}>
            Create workspace
          </Button>
        </Dialog.Actions>
      </form>
    </Dialog.Root>
  )
}
