import { describe, expect, it } from 'vitest'
import { statusTone } from './status-tone'

describe('statusTone', () => {
  it('classifies status codes by class', () => {
    expect([101, 200, 299, 300, 399, 400, 500].map(statusTone)).toEqual(['info', 'success', 'success', 'redirect', 'redirect', 'error', 'error'])
  })
})
