import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { flushSync } from 'react-dom'
import type { AppError } from '@/platform/errors'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
import { useExamplesStore } from '@/features/examples/store'
import styles from './SaveAsExampleDialog.module.css'

const NAME_LABEL = <span className={styles.label}>Example name</span>

const HINT = (
  <p className={styles.hint}>
    Keeps the request as it was sent, with its {'{{variables}}'} unresolved, and the response this tab received.
  </p>
)

const NAME_REQUIRED = 'Enter a name for the example.'
const CREATE_NEXT_STEP = 'Send the request again, then try once more.'

interface SaveAsExampleDialogProps {
  /** Called with the new example so the caller can open it. */
  onCreated: (requestId: string, exampleId: string) => void
}

export function SaveAsExampleDialog({ onCreated }: SaveAsExampleDialogProps) {
  const dialog = useExamplesStore((state) => state.saveDialog)
  const close = useExamplesStore((state) => state.closeSaveDialog)
  return (
    <Dialog.Root open={dialog !== null} onOpenChange={(open) => (open ? undefined : close())}>
      {dialog ? <SaveAsExampleForm key={dialog.tabId} onCreated={onCreated} /> : null}
    </Dialog.Root>
  )
}

function SaveAsExampleForm({ onCreated }: SaveAsExampleDialogProps) {
  const createFromResponse = useExamplesStore((state) => state.createFromResponse)
  const close = useExamplesStore((state) => state.closeSaveDialog)
  const [name, setName] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)
  const [createError, setCreateError] = useState<AppError | null>(null)
  const [creating, startCreate] = useTransition()
  const nameRef = useRef<HTMLInputElement>(null)
  const errorRef = useRef<HTMLDivElement>(null)
  const nameErrorId = useId()

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setCreateError(null)
    const trimmed = name.trim()
    if (trimmed === '') {
      setNameError(NAME_REQUIRED)
      nameRef.current?.focus()
      return
    }
    startCreate(async () => {
      const created = await createFromResponse(trimmed)
      if (created.ok) {
        onCreated(created.value.requestId, created.value.exampleId)
        return
      }
      flushSync(() => setCreateError(created.error))
      errorRef.current?.focus()
    })
  }

  return (
    <form aria-label="Save response as example" className={styles.form} noValidate onSubmit={onSubmit}>
      <Dialog.Title>Save Response as Example</Dialog.Title>
      <Dialog.Body>
        <label className={styles.field}>
          {NAME_LABEL}
          <input
            ref={nameRef}
            className={styles.input}
            name="example-name"
            value={name}
            maxLength={200}
            autoComplete="off"
            placeholder="Created…"
            aria-invalid={nameError !== null}
            aria-describedby={nameError ? nameErrorId : undefined}
            onChange={(event) => {
              setName(event.target.value)
              setNameError(null)
            }}
          />
        </label>
        {nameError ? (
          <p id={nameErrorId} role="alert" className={styles.fieldError}>
            {nameError}
          </p>
        ) : null}
        {HINT}
        {createError ? (
          <div ref={errorRef} tabIndex={-1} className={styles.saveError}>
            <Banner.Root tone="error">
              <Banner.Message>
                <NoTranslateText text={withNextStep(createError.description || createError.code, CREATE_NEXT_STEP)} terms={[createError.code]} />
              </Banner.Message>
            </Banner.Root>
          </div>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={close}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={creating}>
          {creating ? 'Saving…' : 'Save Example'}
        </Button>
      </Dialog.Actions>
    </form>
  )
}
