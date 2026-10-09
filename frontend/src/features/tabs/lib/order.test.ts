import { describe, expect, it } from 'vitest'
import { moveTab, neighbourOf } from './order'

const TABS = [{ id: 'a' }, { id: 'b' }, { id: 'c' }]

describe('moveTab', () => {
  it('moves a tab onto the position of another', () => {
    expect(moveTab(TABS, 'a', 'c').map((tab) => tab.id)).toEqual(['b', 'c', 'a'])
    expect(moveTab(TABS, 'c', 'a').map((tab) => tab.id)).toEqual(['c', 'a', 'b'])
  })

  it('keeps the list when a tab is unknown or dropped on itself', () => {
    expect(moveTab(TABS, 'a', 'a')).toBe(TABS)
    expect(moveTab(TABS, 'x', 'a')).toBe(TABS)
  })
})

describe('neighbourOf', () => {
  it('prefers the tab to the right, then the one to the left', () => {
    expect(neighbourOf(TABS, 'a')).toBe('b')
    expect(neighbourOf(TABS, 'c')).toBe('b')
  })

  it('has no neighbour for the last tab', () => {
    expect(neighbourOf([{ id: 'a' }], 'a')).toBeNull()
  })
})
