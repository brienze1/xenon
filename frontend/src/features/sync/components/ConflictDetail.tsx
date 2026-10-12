import { useId } from 'react'
import type { ChosenResolution, SyncConflict } from '@/platform/mappers/sync'
import { conflictChoices, conflictKindLabel, itemKindLabel, theirCopyName } from '@/features/sync/lib/conflict-labels'
import { ChoiceButtons } from './ChoiceButtons'
import { FieldConflicts } from './FieldConflicts'
import { ItemVersionView } from './ItemVersionView'
import styles from './ConflictView.module.css'

interface ConflictDetailProps {
  conflict: SyncConflict
  disabled: boolean
  onResolve: (resolution: ChosenResolution, fieldPath?: string) => void
}

export function ConflictDetail({ conflict, disabled, onResolve }: ConflictDetailProps) {
  const titleId = useId()
  return (
    <section className={styles.detail} aria-labelledby={titleId}>
      <header className={styles.detailHeader}>
        <h3 id={titleId} className={styles.detailTitle} translate="no">
          {conflict.itemName}
        </h3>
        <p className={styles.muted}>
          {itemKindLabel(conflict.itemKind)} · {conflictKindLabel(conflict)}
          {conflict.uncommitted ? ' · mine is your uncommitted edit' : ''}
        </p>
        <p className={styles.path} translate="no">
          {conflict.path}
        </p>
      </header>
      {conflict.kind === 'field' ? (
        <FieldConflicts fields={conflict.fields} disabled={disabled} onResolve={onResolve} />
      ) : (
        <>
          <ChoiceButtons label="Resolution" choices={conflictChoices(conflict)} selected={conflict.resolution} disabled={disabled} onChoose={(resolution) => onResolve(resolution)} />
          {conflict.resolution === 'both' ? <p className={styles.muted}>Theirs is saved next to yours as “{theirCopyName(conflict)}” with a new id.</p> : null}
        </>
      )}
      <div className={styles.sides}>
        <ItemVersionView title="Mine" version={conflict.mine} />
        <ItemVersionView title="Theirs" version={conflict.theirs} />
      </div>
      {conflict.base !== null ? (
        <details className={styles.base}>
          <summary>Common base</summary>
          <ItemVersionView title="Base" version={conflict.base} />
        </details>
      ) : null}
    </section>
  )
}
