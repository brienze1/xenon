import { useRequestBuilderStore } from '../store'
import styles from './Panel.module.css'

interface SettingsPanelProps {
  tabId: string
}

export function SettingsPanel({ tabId }: SettingsPanelProps) {
  const settings = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.settings)
  const updateSettings = useRequestBuilderStore((state) => state.updateSettings)
  if (!settings) {
    return null
  }

  return (
    <section className={styles.section}>
      <label className={styles.field}>
        Request timeout in ms (0 = no timeout)
        <input
          type="number"
          min={0}
          className={styles.input}
          value={settings.timeoutMs}
          onChange={(event) => updateSettings(tabId, { timeoutMs: Math.max(0, Math.floor(Number(event.target.value) || 0)) })}
        />
      </label>
      <label className={styles.toggle}>
        <input
          type="checkbox"
          checked={settings.followRedirects}
          onChange={(event) => updateSettings(tabId, { followRedirects: event.target.checked })}
        />
        Automatically follow redirects
      </label>
      <label className={styles.toggle}>
        <input
          type="checkbox"
          checked={settings.verifySsl}
          onChange={(event) => updateSettings(tabId, { verifySsl: event.target.checked })}
        />
        Enable SSL certificate verification
      </label>
      <label className={styles.toggle}>
        <input
          type="checkbox"
          checked={settings.encodeUrl}
          onChange={(event) => updateSettings(tabId, { encodeUrl: event.target.checked })}
        />
        Encode URL automatically
      </label>
    </section>
  )
}
