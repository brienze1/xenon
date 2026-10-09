import { useId, useRef, useState, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { workspaceErrorMessage } from '@/features/workspaces/lib/error-message'
import { hasErrors, NO_ERRORS, validateWorkspaceForm } from '@/features/workspaces/lib/workspace-form'
import { useWorkspacesStore } from '@/features/workspaces/store'
import styles from './WelcomeScreen.module.css'

const PICKER_TITLE = 'Choose the folder for the new workspace'

export function WelcomeScreen() {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState(NO_ERRORS)
  const nameRef = useRef<HTMLInputElement>(null)
  const folderRef = useRef<HTMLButtonElement>(null)
  const nameErrorId = useId()
  const folderErrorId = useId()
  const error = useWorkspacesStore((state) => state.error)
  const create = useWorkspacesStore((state) => state.create)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)

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
    await create(name, path)
    setSubmitting(false)
  }

  return (
    <EmptyState.Root label="Welcome">
      <EmptyState.Art variant="workspace" />
      <EmptyState.Title>Xenon</EmptyState.Title>
      <EmptyState.Description>
        Create a workspace in any folder — for example inside your company’s git repository. Everything is saved there as
        plain files.
      </EmptyState.Description>
      <EmptyState.Actions>
        <form className={styles.form} aria-label="Create workspace" noValidate onSubmit={(event) => void onSubmit(event)}>
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
          <div className={styles.field}>
            <span className={styles.label} id="workspace-folder-label">
              Folder
            </span>
            <div className={styles.folder}>
              <span className={styles.path} aria-labelledby="workspace-folder-label">
                {path ?? <span className={styles.placeholder}>No folder chosen</span>}
              </span>
              <Button ref={folderRef} size="sm" aria-describedby={errors.folder ? folderErrorId : undefined} onClick={onChooseFolder}>
                Choose folder…
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
          <Button type="submit" variant="primary" disabled={submitting}>
            Create workspace
          </Button>
        </form>
      </EmptyState.Actions>
    </EmptyState.Root>
  )
}
