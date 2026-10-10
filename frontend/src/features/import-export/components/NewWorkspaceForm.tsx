import { useEffect, useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useImportStore } from '@/features/import-export/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import styles from './ImportDialog.module.css'

const PICKER_TITLE = 'Choose the folder for the new workspace'

const CREATE_NEXT_STEP = 'Choose another folder or name, then try again.'

const NAME_LABEL = <span className={styles.label}>Workspace name</span>

export function NewWorkspaceForm() {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const [submitting, startSubmit] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const errorId = useId()
  const error = useWorkspacesStore((state) => state.error)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)
  const createWorkspace = useImportStore((state) => state.createWorkspace)
  const cancel = useImportStore((state) => state.cancelNewWorkspace)
  const canCreate = name.trim() !== '' && path !== null && !submitting

  useEffect(() => nameRef.current?.focus(), [])

  async function onChooseFolder() {
    const picked = await chooseFolder(PICKER_TITLE)
    if (picked) {
      setPath(picked)
    }
  }

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (path === null || name.trim() === '') {
      nameRef.current?.focus()
      return
    }
    startSubmit(async () => {
      await createWorkspace(name.trim(), path)
    })
  }

  return (
    <form aria-label="New workspace" className={styles.newWorkspace} noValidate onSubmit={onSubmit}>
      <label className={styles.field}>
        {NAME_LABEL}
        <input
          ref={nameRef}
          className={styles.input}
          name="new-workspace-name"
          value={name}
          maxLength={200}
          autoComplete="off"
          aria-invalid={error ? true : undefined}
          aria-describedby={error ? errorId : undefined}
          onChange={(event) => setName(event.target.value)}
        />
      </label>
      <div className={styles.folderRow}>
        <span className={styles.path} translate={path ? 'no' : undefined}>
          {path ?? 'No folder chosen'}
        </span>
        <Button size="sm" onClick={() => void onChooseFolder()}>
          Choose Folder…
        </Button>
      </div>
      {error ? (
        <p id={errorId} role="alert" className={styles.fieldError}>
          <NoTranslateText text={withNextStep(error.description || error.code, CREATE_NEXT_STEP)} terms={[error.code]} />
        </p>
      ) : null}
      <div className={styles.formActions}>
        <Button variant="ghost" size="sm" onClick={cancel}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" size="sm" disabled={!canCreate}>
          {submitting ? 'Creating…' : 'Create Workspace'}
        </Button>
      </div>
    </form>
  )
}
