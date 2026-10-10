import { cn } from '@/shared/lib/cn'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { displayedCurrent, inheritsCurrent, type DraftRow, type KeyProblem } from '@/features/environments/lib/variable-draft'
import { SecretValueCell } from './SecretValueCell'
import styles from './VariableTableRow.module.css'

const PROBLEM_TEXT: Record<KeyProblem, string> = {
  empty: 'No name',
  invalid: 'Invalid name',
  duplicate: 'Duplicate',
  spaces: 'Spaces',
}

const PROBLEM_HINT: Record<KeyProblem, string> = {
  empty: 'Give the variable a name, or remove the row.',
  invalid: 'Names cannot contain {, } or line breaks.',
  duplicate: 'Another row has the same name; only the last enabled one is used.',
  spaces: 'The name starts or ends with a space, so {{name}} without the space will not match it.',
}

const NO_INITIAL = (
  <span className={styles.none} title="Secret values are kept only in the vault on this machine.">
    Not shared
  </span>
)

const RESET_GLYPH = <span aria-hidden="true">↺</span>
const PERSIST_GLYPH = <span aria-hidden="true">⤒</span>
const REMOVE_GLYPH = <span aria-hidden="true">×</span>

export interface RowActions {
  onChange: (id: string, patch: Partial<DraftRow>) => void
  onRemove: (id: string) => void
  onReset: (key: string) => void
  onPersist: (key: string) => void
}

interface VariableTableRowProps {
  scope: string
  row: DraftRow
  index: number
  problems: KeyProblem[]
  /** Reset and Persist act on what is saved, so they wait while the table has unsaved edits. */
  savedActionsDisabled: boolean
  actions: RowActions
}

export function VariableTableRow({ scope, row, index, problems, savedActionsDisabled, actions }: VariableTableRowProps) {
  const name = row.key.trim() || `row ${index + 1}`
  const inherited = inheritsCurrent(row)
  const current = displayedCurrent(row)
  const overridden = row.saved !== null && !row.saved.secret && row.saved.hasCurrentOverride
  const change = (patch: Partial<DraftRow>) => actions.onChange(row.id, patch)

  return (
    <tr className={cn(!row.enabled && styles.disabled)}>
      <td>
        <div className={styles.keyCell}>
          <input
            name="key"
            autoComplete="off"
            aria-label={`Name of ${name}`}
            aria-invalid={problems.length > 0 ? true : undefined}
            className={styles.cell}
            spellCheck={false}
            translate="no"
            placeholder="name"
            title={row.key}
            value={row.key}
            onChange={(event) => change({ key: event.target.value })}
          />
          {problems.map((problem) => (
            <span key={problem} className={styles.problem} title={PROBLEM_HINT[problem]}>
              {PROBLEM_TEXT[problem]}
            </span>
          ))}
        </div>
      </td>
      <td>
        {row.secret ? (
          NO_INITIAL
        ) : (
          <input
            name="initial-value"
            autoComplete="off"
            aria-label={`Initial value of ${name}`}
            className={styles.cell}
            spellCheck={false}
            title={row.initialValue}
            value={row.initialValue}
            onChange={(event) => change({ initialValue: event.target.value })}
          />
        )}
      </td>
      <td>
        {row.secret ? (
          <SecretValueCell scope={scope} row={row} name={name} onChange={(secretValue) => change({ secretValue })} />
        ) : (
          <input
            name="current-value"
            autoComplete="off"
            aria-label={`Current value of ${name}`}
            aria-description={inherited ? 'Inherited from the initial value' : undefined}
            data-inherited={inherited ? 'true' : undefined}
            className={cn(styles.cell, inherited && styles.inherited)}
            spellCheck={false}
            title={current}
            value={current}
            onChange={(event) => change({ currentValue: event.target.value })}
          />
        )}
      </td>
      <td>
        <select
          name="type"
          aria-label={`Type of ${name}`}
          className={styles.select}
          value={row.secret ? 'secret' : 'plain'}
          onChange={(event) => change({ secret: event.target.value === 'secret' })}
        >
          <option value="plain">Plain</option>
          <option value="secret">Secret</option>
        </select>
      </td>
      <td>
        <div className={styles.check}>
          <input
            name="enabled"
            type="checkbox"
            aria-label={`Enable ${name}`}
            checked={row.enabled}
            onChange={(event) => change({ enabled: event.target.checked })}
          />
        </div>
      </td>
      <td>
        <input
          name="description"
          autoComplete="off"
          aria-label={`Description of ${name}`}
          className={styles.cell}
          title={row.description}
          value={row.description}
          onChange={(event) => change({ description: event.target.value })}
        />
      </td>
      <td>
        <div className={styles.actions}>
          {row.secret || row.saved === null ? null : (
            <>
              <IconButton
                label={`Reset ${name}`}
                title={`Reset ${name} to its initial value`}
                disabled={savedActionsDisabled || !overridden}
                onClick={() => actions.onReset(row.saved?.key ?? row.key)}
              >
                {RESET_GLYPH}
              </IconButton>
              <IconButton
                label={`Persist ${name}`}
                title={`Persist the current value of ${name} as its initial value`}
                disabled={savedActionsDisabled || !overridden}
                onClick={() => actions.onPersist(row.saved?.key ?? row.key)}
              >
                {PERSIST_GLYPH}
              </IconButton>
            </>
          )}
          <IconButton label={`Remove ${name}`} onClick={() => actions.onRemove(row.id)}>
            {REMOVE_GLYPH}
          </IconButton>
        </div>
      </td>
    </tr>
  )
}
