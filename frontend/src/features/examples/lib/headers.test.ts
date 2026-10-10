import { describe, expect, it } from 'vitest'
import { headersToText, textToHeaders } from './headers'

describe('response header bulk text', () => {
  it('round trips rows through Key: value lines', () => {
    const rows = textToHeaders('Content-Type: application/json\n//X-Off: 1\n\nSet-Cookie: a=b: c')
    expect(rows.map(({ key, value, enabled }) => ({ key, value, enabled }))).toEqual([
      { key: 'Content-Type', value: 'application/json', enabled: true },
      { key: 'X-Off', value: '1', enabled: false },
      { key: 'Set-Cookie', value: 'a=b: c', enabled: true },
    ])
    expect(headersToText(rows)).toBe('Content-Type: application/json\n//X-Off: 1\nSet-Cookie: a=b: c')
    expect(textToHeaders('Bare')[0]).toMatchObject({ key: 'Bare', value: '' })
  })
})
