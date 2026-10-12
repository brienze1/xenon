import { describe, expect, it } from 'vitest'
import { activeEnvironmentTarget, activeTabKind, isActiveTabExample, isActiveTabRequest } from './active-tab'

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

describe('active environment tab', () => {
  const tabs = [
    { id: 't1', requestId: 'r1', pinned: true },
    { id: 'env', requestId: null, pinned: false, kind: 'environment' as const, environmentId: 'e1' },
    { id: 'glb', requestId: null, pinned: true, kind: 'globals' as const },
  ]

  it('tells the kind of the active tab and what it edits', () => {
    expect(activeTabKind({ tabs, activeTabId: 'env' })).toBe('environment')
    expect(activeTabKind({ tabs, activeTabId: null })).toBe('request')
    expect(isActiveTabRequest({ tabs, activeTabId: 't1' })).toBe(true)
    expect(isActiveTabRequest({ tabs, activeTabId: 'glb' })).toBe(false)
    expect(activeEnvironmentTarget({ tabs, activeTabId: 'env' })).toEqual({ kind: 'environment', environmentId: 'e1' })
    expect(activeEnvironmentTarget({ tabs, activeTabId: 'glb' })).toEqual({ kind: 'globals' })
    expect(activeEnvironmentTarget({ tabs, activeTabId: 't1' })).toBeNull()
  })
})
