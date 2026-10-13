import { MAX_HISTORY_CAP, MIN_HISTORY_CAP } from '@/platform/mappers/settings'
import { useSettingsStore } from '@/features/settings/store'
import { NumberField } from './NumberField'
import styles from './SettingsSection.module.css'

export const HISTORY_USER_CAP_FIELD = 'history-user-cap'
export const HISTORY_AI_CAP_FIELD = 'history-ai-cap'

export function HistorySection() {
  const userCap = useSettingsStore((state) => state.draft.historyUserCap)
  const aiCap = useSettingsStore((state) => state.draft.historyAiCap)
  const userError = useSettingsStore((state) => state.fieldErrors.historyUserCap)
  const aiError = useSettingsStore((state) => state.fieldErrors.historyAiCap)
  const edit = useSettingsStore((state) => state.edit)
  return (
    <fieldset className={styles.section}>
      <legend className={styles.legend}>History</legend>
      <NumberField
        label="User Cap"
        name={HISTORY_USER_CAP_FIELD}
        value={userCap}
        min={MIN_HISTORY_CAP}
        max={MAX_HISTORY_CAP}
        unit="entries"
        hint="The most of your own sends each workspace keeps. The oldest go first."
        error={userError}
        onChange={(historyUserCap) => edit({ historyUserCap })}
      />
      <NumberField
        label="AI Cap"
        name={HISTORY_AI_CAP_FIELD}
        value={aiCap}
        min={MIN_HISTORY_CAP}
        max={MAX_HISTORY_CAP}
        unit="entries"
        hint="The most AI sends and actions each workspace keeps. The oldest go first."
        error={aiError}
        onChange={(historyAiCap) => edit({ historyAiCap })}
      />
    </fieldset>
  )
}
