import type { KeyValueRow } from '@/platform/mappers/request'
import { newRowId } from '@/shared/lib/rows'

const SEPARATOR = ':'

/** One `Key: value` line per row; disabled rows are prefixed with `//`. */
export function headersToText(rows: readonly KeyValueRow[]): string {
  return rows.map((row) => `${row.enabled ? '' : '//'}${row.key}${SEPARATOR} ${row.value}`).join('\n')
}

export function textToHeaders(text: string): KeyValueRow[] {
  return text
    .split('\n')
    .filter((line) => line.trim() !== '')
    .map((line) => {
      const disabled = line.trimStart().startsWith('//')
      const content = disabled ? line.trimStart().slice(2) : line
      const index = content.indexOf(SEPARATOR)
      const key = (index < 0 ? content : content.slice(0, index)).trim()
      const value = index < 0 ? '' : content.slice(index + 1).trim()
      return { id: newRowId(), key, value, description: '', enabled: !disabled }
    })
}
