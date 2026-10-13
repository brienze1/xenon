import { formatDateTime } from '@/shared/lib/date-format'

const TIME_OPTIONS: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3, hour12: false }

/** The local wall-clock time of an entry, to the millisecond. */
export function consoleTime(createdAt: string): string {
  const date = new Date(createdAt)
  return Number.isNaN(date.getTime()) ? '' : formatDateTime(date, TIME_OPTIONS)
}
