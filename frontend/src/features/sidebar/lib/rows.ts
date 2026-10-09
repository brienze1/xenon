import { newRowId, type KeyValueRow } from '@/platform/mappers/request'

export function patchRow(rows: KeyValueRow[], id: string, patch: Partial<KeyValueRow>): KeyValueRow[] {
  return rows.map((row) => (row.id === id ? { ...row, ...patch } : row))
}

export function addRow(rows: KeyValueRow[], fields: Omit<KeyValueRow, 'id'>): KeyValueRow[] {
  return [...rows, { ...fields, id: newRowId() }]
}

export function removeRow(rows: KeyValueRow[], id: string): KeyValueRow[] {
  return rows.filter((row) => row.id !== id)
}
