import { describe, expect, it, vi } from 'vitest'
import { formatDateTime } from './date-format'

const SENT_AT = new Date(2026, 2, 4, 9, 5)

describe('formatDateTime', () => {
  it('formats a date for the locale', () => {
    expect(formatDateTime(SENT_AT, { day: 'numeric', month: 'short', year: 'numeric' }, 'en-US')).toBe('Mar 4, 2026')
    expect(formatDateTime(SENT_AT, { hour: '2-digit', minute: '2-digit', hour12: false }, 'en-GB')).toBe('09:05')
  })

  it('reuses one formatter per locale and options', () => {
    const spy = vi.spyOn(Intl, 'DateTimeFormat')
    formatDateTime(SENT_AT, { weekday: 'long' }, 'fr-FR')
    formatDateTime(new Date(2026, 2, 5), { weekday: 'long' }, 'fr-FR')
    expect(spy).toHaveBeenCalledTimes(1)
    spy.mockRestore()
  })
})
