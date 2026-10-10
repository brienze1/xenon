import { describe, expect, it } from 'vitest'
import { isActiveTabExample } from './active-tab'

describe('isActiveTabExample', () => {
  const tabs = [
    { id: 't1', requestId: 'r1', pinned: true },
    { id: 't2', requestId: 'r1', pinned: true, kind: 'example' as const, exampleId: 'e1' },
  ]

  it('is true only when the active tab holds an example', () => {
    expect(isActiveTabExample({ tabs, activeTabId: 't2' })).toBe(true)
    expect(isActiveTabExample({ tabs, activeTabId: 't1' })).toBe(false)
    expect(isActiveTabExample({ tabs, activeTabId: null })).toBe(false)
    expect(isActiveTabExample({ tabs, activeTabId: 'gone' })).toBe(false)
  })
})
