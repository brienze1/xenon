import type { HistoryEntry } from '@/platform/mappers/history'
import { formatDateTime } from '@/shared/lib/date-format'

export type HistoryRow =
  | { kind: 'day'; key: string; label: string }
  | { kind: 'entry'; key: string; entry: HistoryEntry }

const DAY_MS = 24 * 60 * 60 * 1000
const DAY_FORMAT: Intl.DateTimeFormatOptions = { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }
const TIME_FORMAT: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' }

function startOfDay(date: Date): number {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()
}

export function dayLabel(sentAt: Date, now: Date): string {
  const days = Math.round((startOfDay(now) - startOfDay(sentAt)) / DAY_MS)
  if (days === 0) {
    return 'Today'
  }
  if (days === 1) {
    return 'Yesterday'
  }
  return formatDateTime(sentAt, DAY_FORMAT)
}

export function sentTime(sentAt: string): string {
  return formatDateTime(new Date(sentAt), TIME_FORMAT)
}

export function historyRows(entries: HistoryEntry[], now: Date): HistoryRow[] {
  const rows: HistoryRow[] = []
  let currentDay: number | null = null
  for (const entry of entries) {
    const sentAt = new Date(entry.sentAt)
    const day = startOfDay(sentAt)
    if (day !== currentDay) {
      currentDay = day
      rows.push({ kind: 'day', key: `day-${day}`, label: dayLabel(sentAt, now) })
    }
    rows.push({ kind: 'entry', key: entry.id, entry })
  }
  return rows
}
