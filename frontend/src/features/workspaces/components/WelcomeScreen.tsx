import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { FolderStarArt } from '@/shared/ui/empty-state/illustrations/FolderStarArt'
import { hasErrors, NO_ERRORS, validateWorkspaceForm } from '@/features/workspaces/lib/workspace-form'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'
import styles from './WelcomeScreen.module.css'

const PICKER_TITLE = 'Choose the folder for the new workspace'

const WELCOME_INTRO = (
  <>
    <EmptyState.Art>
      <FolderStarArt />
    </EmptyState.Art>
    <EmptyState.Title>
      <span translate="no">Xenon</span>
    </EmptyState.Title>
    <EmptyState.Description>
      Create a workspace in any folder — for example inside your company’s git repository. <span translate="no">Xenon</span> saves
      everything there as plain files.
    </EmptyState.Description>
  </>
)

const NAME_LABEL = <span className={styles.label}>Workspace name</span>

const FOLDER_LEGEND = <legend className={styles.label}>Folder</legend>

const NO_FOLDER = (
  <span className={styles.placeholder} translate="yes">
    No folder chosen
  </span>
)

export function WelcomeScreen() {
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
      await create(name, path)
    })
  }

  return (
    <EmptyState.Root label="Welcome">
      {WELCOME_INTRO}
      <EmptyState.Actions>
        <form className={styles.form} aria-label="Create workspace" noValidate onSubmit={onSubmit}>
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
              <span className={styles.path} translate="no">
                {path ?? NO_FOLDER}
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
          <Button type="submit" variant="primary" disabled={submitting}>
            {submitting ? 'Creating…' : 'Create Workspace'}
          </Button>
        </form>
      </EmptyState.Actions>
    </EmptyState.Root>
  )
}
