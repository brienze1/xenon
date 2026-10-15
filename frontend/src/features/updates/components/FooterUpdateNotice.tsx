import { useTabsStore } from '@/features/tabs/store'
import { useUpdatesStore } from '@/features/updates/store'
import styles from './FooterUpdateNotice.module.css'

function openSettings() {
  useTabsStore.getState().openSettingsTab()
}

/** The footer's "Update ready — Restart" / "Update available" notice; nothing otherwise. */
export function FooterUpdateNotice() {
  const state = useUpdatesStore((store) => store.status.state)
  const restart = useUpdatesStore((store) => store.restart)
  if (state === 'ready') {
    return (
      <button type="button" className={styles.notice} onClick={() => void restart()}>
        <span className={styles.dot} aria-hidden="true" />
        Update ready — Restart
      </button>
    )
  }
  if (state === 'available') {
    return (
      <button type="button" className={styles.notice} title="Open Settings" onClick={openSettings}>
        <span className={styles.dot} aria-hidden="true" />
        Update available
      </button>
    )
  }
  return null
}
