import { useEffect, useId, useRef, type FormEvent } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { FIELD_ORDER } from '@/features/settings/lib/settings-form'
import type { SettingsField, ShortcutEntry } from '@/features/settings/model'
import { useSettingsStore } from '@/features/settings/store'
import { AI_PORT_FIELD, AiAccessSection } from './AiAccessSection'
import { GitAccounts } from './GitAccounts'
import { HISTORY_AI_CAP_FIELD, HISTORY_USER_CAP_FIELD, HistorySection } from './HistorySection'
import { SCRIPT_TIMEOUT_FIELD, ScriptsSection } from './ScriptsSection'
import { ShortcutsList } from './ShortcutsList'
import styles from './SettingsTab.module.css'

const FIELD_NAMES: Record<SettingsField, string> = {
  scriptTimeout: SCRIPT_TIMEOUT_FIELD,
  historyUserCap: HISTORY_USER_CAP_FIELD,
  historyAiCap: HISTORY_AI_CAP_FIELD,
  aiPort: AI_PORT_FIELD,
}

const NO_SHORTCUTS: readonly ShortcutEntry[] = []

const SECTIONS = (
  <>
    <ScriptsSection />
    <HistorySection />
    <AiAccessSection />
  </>
)

const CLOSE_GLYPH = <span aria-hidden="true">×</span>

interface SettingsTabProps {
  /** The app's hotkey table, listed under Shortcuts. */
  shortcuts?: readonly ShortcutEntry[]
  /** Set when Settings fills the main area with no workspace open: shows a close button. */
  onClose?: () => void
}

export function SettingsTab({ shortcuts = NO_SHORTCUTS, onClose }: SettingsTabProps) {
  const status = useSettingsStore((state) => state.status)
  const loadError = useSettingsStore((state) => state.loadError)
  const saveError = useSettingsStore((state) => state.saveError)
  const saving = useSettingsStore((state) => state.saving)
  const load = useSettingsStore((state) => state.load)
  const save = useSettingsStore((state) => state.save)
  const formRef = useRef<HTMLFormElement>(null)
  const titleId = useId()
  const formId = useId()

  useEffect(() => {
    void load()
  }, [load])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const saved = await save()
    if (saved.ok && !saved.value) {
      const errors = useSettingsStore.getState().fieldErrors
      const first = FIELD_ORDER.find((field) => errors[field] !== undefined)
      const control = first ? formRef.current?.elements.namedItem(FIELD_NAMES[first]) : null
      if (control instanceof HTMLElement) {
        control.focus()
      }
    }
  }

  const error = saveError
    ? withNextStep(saveError.description || saveError.code, 'Try again.')
    : loadError
      ? withNextStep(loadError.description || loadError.code, 'Close Settings and open it again.')
      : null

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h2 id={titleId} className={styles.title}>
          Settings
        </h2>
        <div className={styles.actions}>
          <Button type="submit" form={formId} variant="primary" disabled={saving || status !== 'ready'}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
          {onClose ? (
            <IconButton label="Close Settings" onClick={onClose}>
              {CLOSE_GLYPH}
            </IconButton>
          ) : null}
        </div>
      </header>
      {error ? (
        <Banner.Root tone="error" className={styles.error}>
          <Banner.Message>{error}</Banner.Message>
        </Banner.Root>
      ) : null}
      <div className={styles.body}>
        <form ref={formRef} id={formId} noValidate aria-labelledby={titleId} className={styles.form} onSubmit={(event) => void submit(event)}>
          {SECTIONS}
        </form>
        <GitAccounts />
        {shortcuts.length > 0 ? <ShortcutsList shortcuts={shortcuts} /> : null}
      </div>
    </div>
  )
}
