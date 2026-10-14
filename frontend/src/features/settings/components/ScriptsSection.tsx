import { useSettingsStore } from '@/features/settings/store'
import { MAX_TIMEOUT_SECONDS, MIN_TIMEOUT_SECONDS } from '@/features/settings/lib/settings-form'
import { NumberField } from './NumberField'
import styles from './SettingsSection.module.css'

export const SCRIPT_TIMEOUT_FIELD = 'script-timeout'

const HINT = 'How long one pre-request or post-response script may run, including waiting for its timers and requests.'

export function ScriptsSection() {
  const value = useSettingsStore((state) => state.draft.scriptTimeoutSeconds)
  const error = useSettingsStore((state) => state.fieldErrors.scriptTimeout)
  const edit = useSettingsStore((state) => state.edit)
  return (
    <fieldset className={styles.section}>
      <legend className={styles.legend}>Scripts</legend>
      <NumberField
        label="Script Timeout"
        name={SCRIPT_TIMEOUT_FIELD}
        value={value}
        min={MIN_TIMEOUT_SECONDS}
        max={MAX_TIMEOUT_SECONDS}
        unit="seconds"
        hint={HINT}
        error={error}
        onChange={(scriptTimeoutSeconds) => edit({ scriptTimeoutSeconds })}
      />
    </fieldset>
  )
}
