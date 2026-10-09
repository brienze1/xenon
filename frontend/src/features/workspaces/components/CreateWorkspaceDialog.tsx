import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { hasErrors, NO_ERRORS, validateWorkspaceForm } from '@/features/workspaces/lib/workspace-form'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'
import styles from './CreateWorkspaceDialog.module.css'

interface CreateWorkspaceDialogProps {
  open: boolean
  onClose: () => void
}

const PICKER_TITLE = 'Choose the folder for the new workspace'

const CREATE_TITLE = <Dialog.Title>New Workspace</Dialog.Title>

const NAME_LABEL = <span className={styles.label}>Workspace name</span>

const FOLDER_LEGEND = <legend className={styles.label}>Folder</legend>

export function CreateWorkspaceDialog({ open, onClose }: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const [submitting, startSubmit] = useTransition()
  const [errors, setErrors] = useState(NO_ERRORS)
  const nameRef = useRef<HTMLInputElement>(null)
  const folderRef = useRef<HTMLButtonElement>(null)
  const nameErrorId = useId()
  const folderErrorId = useId()
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

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const found = validateWorkspaceForm(name, path)
    setErrors(found)
    if (hasErrors(found) || path === null) {
      const firstInvalid = found.name ? nameRef.current : folderRef.current
      firstInvalid?.focus()
      return
    }
    startSubmit(async () => {
      if (await create(name, path)) {
        close()
      }
    })
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())}>
      <form aria-label="New workspace" className={styles.form} noValidate onSubmit={onSubmit}>
        {CREATE_TITLE}
        <Dialog.Body>
          <label className={styles.field}>
            {NAME_LABEL}
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
            <p id={nameErrorId} role="alert" className={styles.fieldError}>
              {errors.name}
            </p>
          ) : null}
          <fieldset className={styles.fieldset}>
            {FOLDER_LEGEND}
            <div className={styles.folder}>
              <span className={styles.path} translate={path ? 'no' : undefined}>
                {path ?? 'No folder chosen'}
              </span>
              <Button ref={folderRef} size="sm" aria-describedby={errors.folder ? folderErrorId : undefined} onClick={() => void onChooseFolder()}>
                Choose Folder…
              </Button>
            </div>
            {errors.folder ? (
              <p id={folderErrorId} role="alert" className={styles.fieldError}>
                {errors.folder}
              </p>
            ) : null}
          </fieldset>
          {error ? (
            <Banner.Root tone="error">
              <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
            </Banner.Root>
          ) : null}
        </Dialog.Body>
        <Dialog.Actions>
          <Button variant="ghost" onClick={close}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={submitting}>
            {submitting ? 'Creating…' : 'Create Workspace'}
          </Button>
        </Dialog.Actions>
      </form>
    </Dialog.Root>
  )
}
