import type { SyncChange } from '@/platform/mappers/sync'
import { cn } from '@/shared/lib/cn'
import { changeLabel } from '@/features/sync/lib/status-label'
import styles from './SyncPanel.module.css'

interface ChangeListProps {
  changes: SyncChange[]
  excluded: string[]
  disabled: boolean
  onToggle: (itemId: string) => void
}

export function ChangeList({ changes, excluded, disabled, onToggle }: ChangeListProps) {
  if (changes.length === 0) {
    return <p className={styles.hint}>No local changes.</p>
  }
  return (
    <ul className={styles.changes} aria-label="Changed items">
      {changes.map((change) => (
        <li key={change.itemId}>
          <label className={styles.change}>
            <input type="checkbox" checked={!excluded.includes(change.itemId)} disabled={disabled} onChange={() => onToggle(change.itemId)} />
            <span className={styles.changeName} translate="no">
              {change.itemName}
            </span>
            <span className={styles.kind}>{change.itemKind}</span>
            <span className={cn(styles.badge, styles[change.change])}>{changeLabel(change.change)}</span>
          </label>
        </li>
      ))}
    </ul>
  )
}
