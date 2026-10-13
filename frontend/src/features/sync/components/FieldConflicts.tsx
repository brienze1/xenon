import type { ChosenResolution, SyncFieldConflict } from '@/platform/mappers/sync'
import { FIELD_CHOICES } from '@/features/sync/lib/conflict-labels'
import { ChoiceButtons } from './ChoiceButtons'
import styles from './ConflictView.module.css'

interface FieldConflictsProps {
  fields: readonly SyncFieldConflict[]
  disabled: boolean
  onResolve: (resolution: ChosenResolution, fieldPath: string) => void
}

function FieldValue({ value }: { value: string | null }) {
  return value === null ? (
    <span className={styles.muted}>(removed)</span>
  ) : (
    <pre className={styles.fieldValue} translate="no">
      {value}
    </pre>
  )
}

export function FieldConflicts({ fields, disabled, onResolve }: FieldConflictsProps) {
  return (
    <ul className={styles.fields} aria-label="Conflicting fields">
      {fields.map((field) => (
        <li key={field.path} className={styles.field}>
          <p className={styles.fieldPath} translate="no">
            {field.path}
          </p>
          <div className={styles.fieldSides}>
            <div>
              <p className={styles.sideTitle}>Mine</p>
              <FieldValue value={field.mine} />
            </div>
            <div>
              <p className={styles.sideTitle}>Theirs</p>
              <FieldValue value={field.theirs} />
            </div>
          </div>
          <ChoiceButtons
            label={`Resolution for ${field.path}`}
            choices={FIELD_CHOICES}
            selected={field.resolution}
            disabled={disabled}
            onChoose={(resolution) => onResolve(resolution, field.path)}
          />
        </li>
      ))}
    </ul>
  )
}
