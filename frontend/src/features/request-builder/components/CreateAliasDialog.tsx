import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { flushSync } from 'react-dom'
import type { AppError } from '@/platform/errors'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useTabsStore } from '@/features/tabs/store'
import styles from './AliasDialog.module.css'

const NAME_REQUIRED = 'Enter a name for the alias.'

interface CreateAliasDialogProps {
  tabId: string
  open: boolean
  onClose: () => void
}

/** Asks for the name of a new alias, a copy of the saved Default, and shows it in the tab once created. */
export function CreateAliasDialog({ tabId, open, onClose }: CreateAliasDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      {open ? <CreateAliasForm tabId={tabId} onClose={onClose} /> : null}
    </Dialog.Root>
  )
}

interface CreateAliasFormProps {
  tabId: string
  onClose: () => void
}

function CreateAliasForm({ tabId, onClose }: CreateAliasFormProps) {
  const createAlias = useRequestBuilderStore((state) => state.createAlias)
  const pin = useTabsStore((state) => state.pin)
  const [name, setName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [creating, startCreate] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const errorId = useId()

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (name.trim() === '') {
      setError(NAME_REQUIRED)
      nameRef.current?.focus()
      return
    }
    startCreate(async () => {
      const failure: AppError | null = await createAlias(tabId, name)
      if (failure === null) {
        pin(tabId)
        onClose()
        return
      }
      flushSync(() => setError(failure.description || failure.code))
      nameRef.current?.focus()
    })
  }

  return (
    <form aria-label="Create new alias" className={styles.form} noValidate onSubmit={onSubmit}>
      <Dialog.Title>Create New Alias</Dialog.Title>
      <Dialog.Body>
        <label className={styles.field}>
          <span className={styles.label}>Alias name</span>
          <input
            ref={nameRef}
            className={styles.input}
            name="alias-name"
            value={name}
            maxLength={200}
            autoComplete="off"
            placeholder="mx-test-user…"
            aria-invalid={error ? true : undefined}
            aria-describedby={error ? errorId : undefined}
            onChange={(event) => {
              setName(event.target.value)
              setError(null)
            }}
          />
        </label>
        {error ? (
          <p id={errorId} role="alert" className={styles.fieldError}>
            <NoTranslateText text={error} terms={[]} />
          </p>
        ) : null}
        <p className={styles.hint}>The alias starts as a copy of the saved Default. Editing it never changes the Default.</p>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={creating}>
          {creating ? 'Creating…' : 'Create Alias'}
        </Button>
      </Dialog.Actions>
    </form>
  )
}
