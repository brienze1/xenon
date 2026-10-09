import { useId, useRef, useState, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { workspaceErrorMessage } from '@/features/workspaces/lib/error-message'
import { hasErrors, NO_ERRORS, validateWorkspaceForm } from '@/features/workspaces/lib/workspace-form'
import { useWorkspacesStore } from '@/features/workspaces/store'
import styles from './CreateWorkspaceDialog.module.css'

interface CreateWorkspaceDialogProps {
  open: boolean
  onClose: () => void
}

const PICKER_TITLE = 'Choose the folder for the new workspace'

export function CreateWorkspaceDialog({ open, onClose }: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState(NO_ERRORS)
  const nameRef = useRef<HTMLInputElement>(null)
  const folderRef = useRef<HTMLButtonElement>(null)
  const nameErrorId = useId()
  const folderErrorId = useId()
  const folderLabelId = useId()
  const error = useWorkspacesStore((state) => state.error)
  const create = useWorkspacesStore((state) => state.create)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)

  function close() {
    setName('')
    setPath(null)
    setErrors(NO_ERRORS)
    onClose()
  }

  async function onChooseFolder() {
    const picked = await chooseFolder(PICKER_TITLE)
    if (picked) {
      setPath(picked)
      setErrors((current) => ({ ...current, folder: null }))
    }
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const found = validateWorkspaceForm(name, path)
    setErrors(found)
    if (hasErrors(found) || path === null) {
      const firstInvalid = found.name ? nameRef.current : folderRef.current
      firstInvalid?.focus()
      return
    }
    setSubmitting(true)
    const created = await create(name, path)
    setSubmitting(false)
    if (created) {
      close()
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())}>
      <form aria-label="New workspace" className={styles.form} noValidate onSubmit={(event) => void onSubmit(event)}>
        <Dialog.Title>New Workspace</Dialog.Title>
        <Dialog.Body>
          <label className={styles.field}>
            <span className={styles.label}>Workspace name</span>
            <input
              ref={nameRef}
              className={styles.input}
              name="workspace-name"
              value={name}
              maxLength={200}
              autoComplete="off"
              aria-invalid={errors.name !== null}
              aria-describedby={errors.name ? nameErrorId : undefined}
              onChange={(event) => {
                setName(event.target.value)
                setErrors((current) => ({ ...current, name: null }))
              }}
            />
          </label>
          {errors.name ? (
            <p id={nameErrorId} className={styles.fieldError}>
              {errors.name}
            </p>
          ) : null}
          <div className={styles.field} role="group" aria-labelledby={folderLabelId}>
            <span className={styles.label} id={folderLabelId}>
              Folder
            </span>
            <div className={styles.folder}>
              <span className={styles.path} translate="no">
                {path ?? 'No folder chosen'}
              </span>
              <Button ref={folderRef} size="sm" aria-describedby={errors.folder ? folderErrorId : undefined} onClick={() => void onChooseFolder()}>
                Choose Folder…
              </Button>
            </div>
            {errors.folder ? (
              <p id={folderErrorId} className={styles.fieldError}>
                {errors.folder}
              </p>
            ) : null}
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
          <Button type="submit" variant="primary" disabled={submitting}>
            Create Workspace
          </Button>
        </Dialog.Actions>
      </form>
    </Dialog.Root>
  )
}
