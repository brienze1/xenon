import { useId, useRef, useState, type FormEvent } from 'react'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { hasErrors, NO_ERRORS, validateWorkspaceForm } from '@/features/workspaces/lib/workspace-form'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'
import styles from './RenameWorkspaceDialog.module.css'

const RENAME_HINT = <p className={styles.hint}>Only the name changes. The folder stays where it is.</p>

interface RenameWorkspaceDialogProps {
  workspace: WorkspaceSummary
  open: boolean
  onClose: () => void
}

export function RenameWorkspaceDialog({ workspace, open, onClose }: RenameWorkspaceDialogProps) {
  const [name, setName] = useState(workspace.name)
  const [submitting, setSubmitting] = useState(false)
  const [errors, setErrors] = useState(NO_ERRORS)
  const nameRef = useRef<HTMLInputElement>(null)
  const nameErrorId = useId()
  const error = useWorkspacesStore((state) => state.error)
  const rename = useWorkspacesStore((state) => state.rename)

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const found = validateWorkspaceForm(name)
    setErrors(found)
    if (hasErrors(found)) {
      nameRef.current?.focus()
      return
    }
    setSubmitting(true)
    const renamed = await rename(workspace.localId, name)
    setSubmitting(false)
    if (renamed) {
      onClose()
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <form aria-label="Rename workspace" className={styles.form} noValidate onSubmit={(event) => void onSubmit(event)}>
        <Dialog.Title>Rename Workspace</Dialog.Title>
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
                setErrors(NO_ERRORS)
              }}
            />
          </label>
          {errors.name ? (
            <p id={nameErrorId} role="alert" className={styles.fieldError}>
              {errors.name}
            </p>
          ) : null}
          {RENAME_HINT}
          {error ? (
            <Banner.Root tone="error">
              <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
            </Banner.Root>
          ) : null}
        </Dialog.Body>
        <Dialog.Actions>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={submitting}>
            {submitting ? 'Renaming…' : 'Rename Workspace'}
          </Button>
        </Dialog.Actions>
      </form>
    </Dialog.Root>
  )
}
