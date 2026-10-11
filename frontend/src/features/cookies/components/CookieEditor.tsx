import { useId, useState, useTransition, type FormEvent } from 'react'
import type { AppError } from '@/platform/errors'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { CookieErrorText } from './CookieErrorText'
import styles from './CookiesDialog.module.css'

const SAVE_NEXT_STEP = 'Write the cookie as name=value followed by attributes such as Path=/; Expires=…; HttpOnly; Secure.'

interface CookieEditorProps {
  initial: string
  onSave: (raw: string) => Promise<AppError | null>
  onCancel: () => void
}

export function CookieEditor({ initial, onSave, onCancel }: CookieEditorProps) {
  const fieldId = useId()
  const [raw, setRaw] = useState(initial)
  const [error, setError] = useState<AppError | null>(null)
  const [saving, startSave] = useTransition()

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    startSave(async () => {
      const failure = await onSave(raw)
      setError(failure)
    })
  }

  return (
    <form className={styles.editor} onSubmit={submit}>
      <label htmlFor={fieldId} className={styles.label}>
        Cookie
      </label>
      <textarea
        id={fieldId}
        className={styles.raw}
        value={raw}
        spellCheck={false}
        autoComplete="off"
        translate="no"
        rows={3}
        onChange={(event) => setRaw(event.target.value)}
      />
      {error ? (
        <Banner.Root tone="error">
          <Banner.Message>
            <CookieErrorText error={error} nextStep={SAVE_NEXT_STEP} />
          </Banner.Message>
        </Banner.Root>
      ) : null}
      <div className={styles.editorActions}>
        <Button variant="secondary" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" size="sm" disabled={saving || raw.trim() === ''}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </form>
  )
}
