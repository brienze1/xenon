import { useState, type FormEvent } from 'react'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { workspaceErrorMessage } from '../lib/error-message'
import { useWorkspacesStore } from '../store'
import styles from './WelcomeScreen.module.css'

const PICKER_TITLE = 'Choose the folder for the new workspace'

export function WelcomeScreen() {
  const [name, setName] = useState('')
  const [path, setPath] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const error = useWorkspacesStore((state) => state.error)
  const create = useWorkspacesStore((state) => state.create)
  const chooseFolder = useWorkspacesStore((state) => state.chooseFolder)

  async function onChooseFolder() {
    const picked = await chooseFolder(PICKER_TITLE)
    if (picked) {
      setPath(picked)
    }
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!path) {
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
        Create a workspace in any folder — for example inside your company&apos;s git repository. Everything is saved
        there as plain files.
      </EmptyState.Description>
      <EmptyState.Actions>
        <form className={styles.form} aria-label="Create workspace" onSubmit={onSubmit}>
          <label className={styles.field}>
            <span className={styles.label}>Workspace name</span>
            <input
              className={styles.input}
              value={name}
              maxLength={200}
              autoComplete="off"
              onChange={(event) => setName(event.target.value)}
            />
          </label>
          <div className={styles.field}>
            <span className={styles.label} id="workspace-folder-label">
              Folder
            </span>
            <div className={styles.folder}>
              <span className={styles.path} aria-labelledby="workspace-folder-label">
                {path ?? <span className={styles.placeholder}>No folder chosen</span>}
              </span>
              <Button size="sm" onClick={onChooseFolder}>
                Choose folder…
              </Button>
            </div>
          </div>
          {error ? (
            <Banner.Root tone="error">
              <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
            </Banner.Root>
          ) : null}
          <Button type="submit" variant="primary" disabled={submitting || name.trim() === '' || path === null}>
            Create workspace
          </Button>
        </form>
      </EmptyState.Actions>
    </EmptyState.Root>
  )
}
