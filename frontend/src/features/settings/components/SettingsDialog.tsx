import { useId, useRef, useState, useTransition, type FormEvent } from 'react'
import { MAX_SCRIPT_TIMEOUT_MS, MIN_SCRIPT_TIMEOUT_MS } from '@/platform/mappers/settings'
import { SCRIPT_TIMEOUT_OUT_OF_RANGE } from '@/platform/services/settings'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import type { ShortcutEntry } from '@/features/settings/model'
import { useSettingsStore } from '@/features/settings/store'
import { GitAccounts } from './GitAccounts'
import { ShortcutsList } from './ShortcutsList'
import styles from './SettingsDialog.module.css'

const MS_PER_SECOND = 1000
const MIN_SECONDS = MIN_SCRIPT_TIMEOUT_MS / MS_PER_SECOND
const MAX_SECONDS = MAX_SCRIPT_TIMEOUT_MS / MS_PER_SECOND
const OUT_OF_RANGE_MESSAGE = `Enter a whole number of seconds from ${MIN_SECONDS} to ${MAX_SECONDS}.`

function secondsOf(text: string): number | null {
  const seconds = Number(text)
  return text.trim() !== '' && Number.isInteger(seconds) && seconds >= MIN_SECONDS && seconds <= MAX_SECONDS ? seconds : null
}

const NO_SHORTCUTS: readonly ShortcutEntry[] = []

interface SettingsDialogProps {
  /** The app's hotkey table, listed under Shortcuts. */
  shortcuts?: readonly ShortcutEntry[]
}

export function SettingsDialog({ shortcuts = NO_SHORTCUTS }: SettingsDialogProps) {
  const open = useSettingsStore((state) => state.open)
  const loading = useSettingsStore((state) => state.loading)
  const close = useSettingsStore((state) => state.close)
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : close())}>
      {open && !loading ? <SettingsForm onCancel={close} shortcuts={shortcuts} /> : null}
    </Dialog.Root>
  )
}

interface SettingsFormProps {
  onCancel: () => void
  shortcuts: readonly ShortcutEntry[]
}

function SettingsForm({ onCancel, shortcuts }: SettingsFormProps) {
  const settings = useSettingsStore((state) => state.settings)
  const loadError = useSettingsStore((state) => state.error)
  const save = useSettingsStore((state) => state.save)
  const [timeout, setTimeoutText] = useState(String(settings.scriptTimeoutMs / MS_PER_SECOND))
  const [fieldError, setFieldError] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saving, startSave] = useTransition()
  const timeoutRef = useRef<HTMLInputElement>(null)
  const hintId = useId()
  const errorId = useId()

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const seconds = secondsOf(timeout)
    if (seconds === null) {
      setFieldError(true)
      timeoutRef.current?.focus()
      return
    }
    startSave(async () => {
      const failure = await save({ scriptTimeoutMs: seconds * MS_PER_SECOND })
      if (failure?.code === SCRIPT_TIMEOUT_OUT_OF_RANGE) {
        setFieldError(true)
        timeoutRef.current?.focus()
      } else {
        setSaveError(failure ? withNextStep(failure.description || failure.code, 'Try again.') : null)
      }
    })
  }

  return (
    <form noValidate onSubmit={submit}>
      <Dialog.Title>Settings</Dialog.Title>
      <Dialog.Body>
        <fieldset className={styles.section}>
          <legend className={styles.legend}>Scripts</legend>
          <label className={styles.field}>
            Script Timeout
            <span className={styles.control}>
              <input
                ref={timeoutRef}
                name="script-timeout"
                type="number"
                inputMode="numeric"
                min={MIN_SECONDS}
                max={MAX_SECONDS}
                step={1}
                autoComplete="off"
                className={styles.input}
                aria-invalid={fieldError ? true : undefined}
                aria-describedby={fieldError ? `${errorId} ${hintId}` : hintId}
                value={timeout}
                onChange={(event) => {
                  setTimeoutText(event.target.value)
                  setFieldError(false)
                }}
              />
              <span className={styles.unit}>seconds</span>
            </span>
          </label>
          {fieldError ? (
            <p id={errorId} role="alert" className={styles.fieldError}>
              {OUT_OF_RANGE_MESSAGE}
            </p>
          ) : null}
          <p id={hintId} className={styles.hint}>
            How long one pre-request or post-response script may run, including waiting for its timers and requests.
          </p>
        </fieldset>
        <GitAccounts />
        {shortcuts.length > 0 ? <ShortcutsList shortcuts={shortcuts} /> : null}
        {loadError || saveError ? (
          <Banner.Root tone="error" className={styles.error}>
            <Banner.Message>{saveError ?? withNextStep(loadError?.description || loadError?.code || '', 'Close and open Settings again.')}</Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={saving}>
          Save
        </Button>
      </Dialog.Actions>
    </form>
  )
}
