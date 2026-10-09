import { describe, expect, it } from 'vitest'
import { withoutNotice } from './notices'

describe('withoutNotice', () => {
  it('drops the notice of one tab without touching the others or the input', () => {
    const notices = { a: 'changed', b: 'deleted' } as const
    expect(withoutNotice(notices, 'a')).toEqual({ b: 'deleted' })
    expect(notices).toEqual({ a: 'changed', b: 'deleted' })
  })

  it('returns an equal copy when the tab has no notice', () => {
    expect(withoutNotice({ a: 'changed' }, 'x')).toEqual({ a: 'changed' })
  })
})
