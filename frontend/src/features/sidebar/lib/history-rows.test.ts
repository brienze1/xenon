import { describe, expect, it } from 'vitest'
import type { HistoryEntry } from '@/platform/mappers/history'
import { dayLabel, historyRows, sentTime } from './history-rows'

function entry(id: string, sentAt: Date): HistoryEntry {
  return { id, sentAt: sentAt.toISOString(), method: 'GET', url: `https://x/${id}`, statusCode: 200, errorKind: null, totalMs: 5, sizeBytes: 2, sourceRequestId: null }
}

const now = new Date(2026, 9, 9, 15, 30)

describe('history rows', () => {
  it('labels today, yesterday and older days', () => {
    expect(dayLabel(new Date(2026, 9, 9, 0, 1), now)).toBe('Today')
    expect(dayLabel(new Date(2026, 9, 8, 23, 59), now)).toBe('Yesterday')
    expect(dayLabel(new Date(2026, 9, 6, 12), now)).toBe(new Date(2026, 9, 6, 12).toLocaleDateString(undefined, { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }))
  })

  it('groups newest-first entries under one header per day', () => {
    const rows = historyRows([entry('c', new Date(2026, 9, 9, 14)), entry('b', new Date(2026, 9, 9, 9)), entry('a', new Date(2026, 9, 8, 22))], now)
    expect(rows.map((row) => (row.kind === 'day' ? row.label : row.entry.id))).toEqual(['Today', 'c', 'b', 'Yesterday', 'a'])
    expect(new Set(rows.map((row) => row.key)).size).toBe(rows.length)
    expect(historyRows([], now)).toEqual([])
  })

  it('formats the time an entry was sent', () => {
    const sentAt = new Date(2026, 9, 9, 9, 5)
    expect(sentTime(sentAt.toISOString())).toBe(sentAt.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' }))
  })
})
