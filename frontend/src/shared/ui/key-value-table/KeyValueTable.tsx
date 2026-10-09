import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import styles from './KeyValueTable.module.css'

export interface KeyValueRow {
  id: string
  key: string
  value: string
  description: string
  enabled: boolean
}

export type KeyValueFields = Omit<KeyValueRow, 'id'>
export type KeyValueMode = 'table' | 'bulk'

interface KeyValueTableProps {
  label: string
  rows: KeyValueRow[]
  mode: KeyValueMode
  bulkText: string
  onModeChange: (mode: KeyValueMode) => void
  onBulkTextChange: (text: string) => void
  onRowChange: (id: string, patch: Partial<KeyValueFields>) => void
  onRowAdd: (row: KeyValueFields) => void
  onRowRemove: (id: string) => void
  examples: KeyValueExamples
}

export type KeyValueExamples = Record<TextField, string>

const EMPTY_ROW: KeyValueFields = { key: '', value: '', description: '', enabled: true }

export type TextField = 'key' | 'value' | 'description'

const COLUMNS: Array<{ field: TextField; title: string; code: boolean }> = [
  { field: 'key', title: 'Key', code: true },
  { field: 'value', title: 'Value', code: true },
  { field: 'description', title: 'Description', code: false },
]

const REMOVE_GLYPH = <span aria-hidden="true">×</span>

const EMPTY_CELL = <td />

const TABLE_HEAD = (
  <thead>
    <tr>
      <th className={styles.checkColumn}>
        <span className="visually-hidden">Enabled</span>
      </th>
      {COLUMNS.map((column) => (
        <th key={column.field}>{column.title}</th>
      ))}
      <th className={styles.actionColumn}>
        <span className="visually-hidden">Actions</span>
      </th>
    </tr>
  </thead>
)

export function KeyValueTable({
  label,
  rows,
  mode,
  bulkText,
  onModeChange,
  onBulkTextChange,
  onRowChange,
  onRowAdd,
  onRowRemove,
  examples,
}: KeyValueTableProps) {
  const bulk = mode === 'bulk'
  return (
    <div className={styles.root}>
      <div className={styles.toolbar}>
        <Button variant="ghost" size="sm" aria-pressed={bulk} onClick={() => onModeChange(bulk ? 'table' : 'bulk')}>
          {bulk ? 'Key-Value Edit' : 'Bulk Edit'}
        </Button>
      </div>
      {bulk ? (
        <textarea
          name="bulk-text"
          autoComplete="off"
          aria-label={`${label} bulk edit`}
          className={styles.bulk}
          spellCheck={false}
          value={bulkText}
          onChange={(event) => onBulkTextChange(event.target.value)}
        />
      ) : (
        <table aria-label={label} className={styles.table}>
          {TABLE_HEAD}
          <tbody>
            {rows.map((row) => (
              <tr key={row.id} className={cn(!row.enabled && styles.disabled)}>
                <td>
                  <div className={styles.check}>
                    <input
                      name="enabled"
                      type="checkbox"
                      aria-label={`Enable ${row.key || 'row'}`}
                      checked={row.enabled}
                      onChange={(event) => onRowChange(row.id, { enabled: event.target.checked })}
                    />
                  </div>
                </td>
                {COLUMNS.map((column) => (
                  <td key={column.field}>
                    <input
                      name={column.field}
                      autoComplete="off"
                      aria-label={column.title}
                      className={styles.cell}
                      spellCheck={false}
                      value={row[column.field]}
                      onChange={(event) => onRowChange(row.id, { [column.field]: event.target.value })}
                    />
                  </td>
                ))}
                <td>
                  <IconButton label={`Remove ${row.key || 'Row'}`} onClick={() => onRowRemove(row.id)}>
                    {REMOVE_GLYPH}
                  </IconButton>
                </td>
              </tr>
            ))}
            <tr>
              {EMPTY_CELL}
              {COLUMNS.map((column) => (
                <td key={column.field}>
                  <input
                    name={`new-${column.field}`}
                    autoComplete="off"
                    aria-label={`New ${column.title.toLowerCase()}`}
                    placeholder={`${examples[column.field]}…`}
                    translate={column.code ? 'no' : undefined}
                    className={styles.cell}
                    spellCheck={false}
                    value=""
                    onChange={(event) => onRowAdd({ ...EMPTY_ROW, [column.field]: event.target.value })}
                  />
                </td>
              ))}
              {EMPTY_CELL}
            </tr>
          </tbody>
        </table>
      )}
    </div>
  )
}
