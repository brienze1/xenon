import { describe, expect, it } from 'vitest'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { aliasEnvironmentLabels, isAliasAvailable } from './alias-availability'

function environment(id: string, name: string, color: EnvironmentSummary['color'] = 'none', error: string | null = null): EnvironmentSummary {
  return { id, name, color, confirmBeforeSend: false, error }
}

describe('isAliasAvailable', () => {
  it('makes an alias for all environments available with or without an active environment', () => {
    expect(isAliasAvailable({ environmentIds: [] }, null)).toBe(true)
    expect(isAliasAvailable({ environmentIds: [] }, 'prod')).toBe(true)
  })

  it('makes a linked alias available only while one of its environments is active', () => {
    const alias = { environmentIds: ['hml', 'dev'] }
    expect(isAliasAvailable(alias, 'hml')).toBe(true)
    expect(isAliasAvailable(alias, 'dev')).toBe(true)
    expect(isAliasAvailable(alias, 'prod')).toBe(false)
    expect(isAliasAvailable(alias, null)).toBe(false)
  })
})

describe('aliasEnvironmentLabels', () => {
  it('names each environment by its current name and colour and flags deleted or broken ones', () => {
    const environments = [environment('hml', 'Staging', 'green'), environment('broken', 'broken.yaml', 'none', 'unreadable')]
    expect(aliasEnvironmentLabels({ environmentIds: ['hml', 'gone', 'broken'] }, environments)).toEqual([
      { id: 'hml', name: 'Staging', color: 'green', deleted: false },
      { id: 'gone', name: 'Deleted environment', color: 'none', deleted: true },
      { id: 'broken', name: 'Deleted environment', color: 'none', deleted: true },
    ])
  })
})
