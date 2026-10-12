import type { SyncConflict } from '@/platform/mappers/sync'
import { cn } from '@/shared/lib/cn'
import { isResolved, itemKindLabel, resolutionLabel } from '@/features/sync/lib/conflict-labels'
import styles from './ConflictView.module.css'

interface ConflictListProps {
  conflicts: readonly SyncConflict[]
  selectedId: string | null
  onSelect: (itemId: string) => void
}

export function ConflictList({ conflicts, selectedId, onSelect }: ConflictListProps) {
  return (
    <ul className={styles.list} aria-label="Conflicts">
      {conflicts.map((conflict) => (
        <li key={conflict.itemId}>
          <button
            type="button"
            className={cn(styles.listItem, conflict.itemId === selectedId && styles.selected)}
            aria-current={conflict.itemId === selectedId ? 'true' : undefined}
            onClick={() => onSelect(conflict.itemId)}
          >
            <span className={styles.listName} translate="no">
              {conflict.itemName}
            </span>
            <span className={styles.muted}>{itemKindLabel(conflict.itemKind)}</span>
            <span className={cn(styles.resolution, isResolved(conflict) && styles.resolved)}>{resolutionLabel(conflict.resolution)}</span>
          </button>
        </li>
      ))}
    </ul>
  )
}
