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

export function keyProblems(rows: DraftRow[], row: DraftRow): KeyProblem[] {
  const problems: KeyProblem[] = []
  if (row.key.trim() === '') {
    return ['empty']
  }
  if (INVALID_KEY.test(row.key)) {
    problems.push('invalid')
  }
  const edited = row.key !== row.saved?.key
  if ((!edited && row.saved?.duplicate) || rows.some((other) => other.id !== row.id && other.key === row.key)) {
    problems.push('duplicate')
  }
  if (row.key !== row.key.trim()) {
    problems.push('spaces')
  }
  return problems
}

export function typeSwitches(rows: DraftRow[]): TypeSwitches {
  const switched = rows.filter((row) => row.saved !== null && row.saved.secret !== row.secret)
  const toPlain = switched.filter((row) => !row.secret).map((row) => row.key)
  return { toSecret: switched.filter((row) => row.secret).map((row) => row.key), toPlain, reveal: toPlain.length > 0 }
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
