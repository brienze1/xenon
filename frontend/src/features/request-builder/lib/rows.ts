import { newRowId } from '@/platform/mappers/request'

export function keepRowIds<T extends { id: string }>(current: T[], next: T[]): T[] {
  return next.map((row, index) => ({ ...row, id: current[index]?.id ?? row.id }))
}

export function patchRow<T extends { id: string }>(rows: T[], id: string, patch: NoInfer<Partial<T>>): T[] {
  return rows.map((row) => (row.id === id ? { ...row, ...patch } : row))
}

export function addRow<T extends { id: string }>(rows: T[], fields: NoInfer<Omit<T, 'id'>>): T[] {
  return [...rows, { ...fields, id: newRowId() } as T]
}

export function removeRow<T extends { id: string }>(rows: T[], id: string): T[] {
  return rows.filter((row) => row.id !== id)
}
