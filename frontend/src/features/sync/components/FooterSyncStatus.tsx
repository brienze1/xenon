import { useSyncStore } from '@/features/sync/store'
import { cn } from '@/shared/lib/cn'
import { statusLabel, statusTone } from '@/features/sync/lib/status-label'
import styles from './FooterSyncStatus.module.css'

export function FooterSyncStatus() {
  const status = useSyncStore((state) => state.status)
  const busy = useSyncStore((state) => state.busy)
  const openPanel = useSyncStore((state) => state.openPanel)
  const openConflicts = useSyncStore((state) => state.openConflicts)
  if (status === null) {
    return null
  }
  const label = busy !== null && busy !== 'status' ? 'Syncing…' : (statusLabel(status) ?? 'Not synced')
  const tone = status.state === 'not_linked' ? 'idle' : statusTone(status)
  const inConflict = status.state === 'conflict'

  return (
    <button
      type="button"
      className={styles.status}
      title={inConflict ? 'Resolve the conflicts' : 'Open the Sync panel'}
      onClick={inConflict ? () => void openConflicts() : openPanel}
    >
      <span className={cn(styles.dot, styles[tone])} aria-hidden="true" />
      <span aria-live="polite">Git: {label}</span>
    </button>
  )
}
