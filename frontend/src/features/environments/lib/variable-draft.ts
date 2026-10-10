import type { VariableEdit, VariableRow } from '@/platform/mappers/variable'
import { newRowId } from '@/shared/lib/rows'

/** One editable row of a variable table; `saved` is the row as Go last returned it (null when new). */
export interface DraftRow {
  id: string
  saved: VariableRow | null
  key: string
  initialValue: string
  /** `null` keeps the stored current value; `""` clears the local override. */
  currentValue: string | null
  secret: boolean
  /** `null` keeps the stored secret value. */
  secretValue: string | null
  enabled: boolean
  description: string
}

export type KeyProblem = 'empty' | 'invalid' | 'duplicate' | 'spaces'

export interface TypeSwitches {
  toSecret: string[]
  toPlain: string[]
  /** Whether saving writes a secret value into the workspace files (needs `confirm_reveal`). */
  reveal: boolean
}

const INVALID_KEY = /[{}\r\n]/

export function toDraftRows(rows: VariableRow[]): DraftRow[] {
  return rows.map((row) => ({
    id: newRowId(),
    saved: row,
    key: row.key,
    initialValue: row.initialValue,
    currentValue: null,
    secret: row.secret,
    secretValue: null,
    enabled: row.enabled,
    description: row.description,
  }))
}

export function newDraftRow(): DraftRow {
  return {
    id: newRowId(),
    saved: null,
    key: '',
    initialValue: '',
    currentValue: null,
    secret: false,
    secretValue: null,
    enabled: true,
    description: '',
  }
}

/** Whether the current value cell shows the inherited initial value (no local override). */
export function inheritsCurrent(row: DraftRow): boolean {
  if (row.currentValue !== null) {
    return row.currentValue === ''
  }
  return !(row.saved?.hasCurrentOverride ?? false) || (row.saved?.secret ?? false)
}

/** The value a plain row's current value cell shows. */
export function displayedCurrent(row: DraftRow): string {
  if (inheritsCurrent(row)) {
    return row.initialValue
  }
  return row.currentValue ?? row.saved?.currentValue ?? ''
}

/** How many rows use each name; build it once per render and pass it to `keyProblems`. */
export function keyCounts(rows: DraftRow[]): Map<string, number> {
  const counts = new Map<string, number>()
  for (const row of rows) {
    counts.set(row.key, (counts.get(row.key) ?? 0) + 1)
  }
  return counts
}

export function keyProblems(row: DraftRow, counts: Map<string, number>): KeyProblem[] {
  const problems: KeyProblem[] = []
  if (row.key.trim() === '') {
    return ['empty']
  }
  if (INVALID_KEY.test(row.key)) {
    problems.push('invalid')
  }
  const edited = row.key !== row.saved?.key
  if ((!edited && row.saved?.duplicate) || (counts.get(row.key) ?? 0) > 1) {
    problems.push('duplicate')
  }
  if (row.key !== row.key.trim()) {
    problems.push('spaces')
  }
  return problems
}

export function typeSwitches(rows: DraftRow[]): TypeSwitches {
  const toSecret: string[] = []
  const toPlain: string[] = []
  for (const row of rows) {
    if (row.saved === null || row.saved.secret === row.secret) {
      continue
    }
    if (row.secret) {
      toSecret.push(row.key)
    } else {
      toPlain.push(row.key)
    }
  }
  return { toSecret, toPlain, reveal: toPlain.length > 0 }
}

export function toVariableEdits(rows: DraftRow[]): VariableEdit[] {
  return rows.map((row) => ({
    key: row.key,
    previousKey: row.saved?.key ?? null,
    initialValue: row.secret ? '' : row.initialValue,
    currentValue: row.secret ? null : row.currentValue,
    secret: row.secret,
    secretValue: row.secret ? row.secretValue : null,
    enabled: row.enabled,
    description: row.description,
  }))
}
