import type { ConsoleEntry, ConsoleFilter } from '@/platform/mappers/console'

export const CONSOLE_FILTERS: ReadonlyArray<{ value: ConsoleFilter; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'requests', label: 'Requests' },
  { value: 'logs', label: 'Logs' },
  { value: 'errors', label: 'Errors' },
]

/** The same rules Go's ConsoleService.List applies, so live entries filter the way a reload would. */
export function matchesFilter(entry: ConsoleEntry, filter: ConsoleFilter): boolean {
  switch (filter) {
    case 'requests':
      return entry.kind === 'request'
    case 'logs':
      return entry.kind === 'log' || entry.kind === 'info'
    case 'errors':
      return entry.kind === 'error' || entry.kind === 'test_failure' || entry.level === 'error'
    default:
      return true
  }
}

export function matchesSearch(entry: ConsoleEntry, search: string): boolean {
  const needle = search.trim().toLowerCase()
  if (needle === '') {
    return true
  }
  const texts = [entry.summary, ...entry.arguments, entry.request?.url ?? '', entry.scriptLabel ?? '']
  return texts.some((text) => text.toLowerCase().includes(needle))
}

export function visibleEntries(entries: readonly ConsoleEntry[], filter: ConsoleFilter, search: string): ConsoleEntry[] {
  return entries.filter((entry) => matchesFilter(entry, filter) && matchesSearch(entry, search))
}

export const CONSOLE_CAPACITY = 5000

/** Appends entries not seen yet and keeps the newest CONSOLE_CAPACITY. */
export function appendEntries(entries: readonly ConsoleEntry[], added: readonly ConsoleEntry[]): ConsoleEntry[] {
  const known = new Set(entries.map((entry) => entry.id))
  const fresh = added.filter((entry) => !known.has(entry.id))
  if (fresh.length === 0) {
    return entries as ConsoleEntry[]
  }
  const all = [...entries, ...fresh]
  return all.length > CONSOLE_CAPACITY ? all.slice(all.length - CONSOLE_CAPACITY) : all
}
