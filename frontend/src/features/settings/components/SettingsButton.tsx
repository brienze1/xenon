import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useSettingsStore } from '@/features/settings/store'
import styles from './SettingsButton.module.css'

const GEAR = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.icon}>
    <circle cx="8" cy="8" r="2.25" />
    <path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4" />
  </svg>
)

export function SettingsButton() {
  const openSettings = useSettingsStore((state) => state.openSettings)
  return (
    <IconButton label="Settings" onClick={() => void openSettings()}>
      {GEAR}
    </IconButton>
  )
}
