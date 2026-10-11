import { describe, expect, it } from 'vitest'
import type { ConsoleEntry } from '@/platform/mappers/console'
import { appendEntries, CONSOLE_CAPACITY, matchesFilter, matchesSearch, visibleEntries } from './filter'

function entry(id: string, overrides: Partial<ConsoleEntry> = {}): ConsoleEntry {
  return { id, sendId: 's', kind: 'log', level: 'log', summary: id, createdAt: '', arguments: [], masked: false, scriptLabel: null, line: null, column: null, request: null, response: null, values: null, ...overrides }
}

describe('console filters', () => {
  it('follows Go: requests, logs (log and info) and errors (errors, failed tests, error levels)', () => {
    const failedRequest = entry('r', { kind: 'request', level: 'error' })
    expect(matchesFilter(failedRequest, 'requests')).toBe(true)
    expect(matchesFilter(failedRequest, 'errors')).toBe(true)
    expect(matchesFilter(failedRequest, 'logs')).toBe(false)
    expect(matchesFilter(entry('i', { kind: 'info' }), 'logs')).toBe(true)
    expect(matchesFilter(entry('t', { kind: 'test_failure' }), 'errors')).toBe(true)
    expect(matchesFilter(entry('e', { kind: 'error' }), 'logs')).toBe(false)
    expect(matchesFilter(entry('l'), 'errors')).toBe(false)
    expect(matchesFilter(entry('l'), 'all')).toBe(true)
  })

  it('searches summaries, arguments, URLs and script names ignoring case', () => {
    const request = entry('r', { kind: 'request', summary: 'GET', request: { method: 'GET', url: 'https://X.io/Invoices', nested: false, headers: null, body: null } })
    expect(matchesSearch(request, ' invoices ')).toBe(true)
    expect(matchesSearch(entry('a', { arguments: ['Token'] }), 'token')).toBe(true)
    expect(matchesSearch(entry('s', { scriptLabel: 'Billing · Pre-request' }), 'billing')).toBe(true)
    expect(matchesSearch(entry('n'), 'zzz')).toBe(false)
    expect(matchesSearch(entry('n'), '  ')).toBe(true)
    expect(visibleEntries([request, entry('x')], 'requests', 'x.io').map((item) => item.id)).toEqual(['r'])
  })

  it('appends new entries only and keeps the newest 5,000', () => {
    const first = [entry('a')]
    expect(appendEntries(first, [entry('a')])).toBe(first)
    expect(appendEntries(first, [entry('a'), entry('b')]).map((item) => item.id)).toEqual(['a', 'b'])
    const full = Array.from({ length: CONSOLE_CAPACITY }, (_, index) => entry(String(index)))
    const appended = appendEntries(full, [entry('new')])
    expect(appended).toHaveLength(CONSOLE_CAPACITY)
    expect(appended[0]?.id).toBe('1')
    expect(appended.at(-1)?.id).toBe('new')
  })
})
