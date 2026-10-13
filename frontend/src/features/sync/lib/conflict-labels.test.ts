import { describe, expect, it } from 'vitest'
import { environmentConflict, fieldConflict, requestConflict } from '@/features/sync/test-fixtures'
import { conflictChoices, conflictKindLabel, isResolved, itemKindLabel, resolutionLabel, resolvedCount, theirCopyName } from './conflict-labels'

describe('conflict labels', () => {
  it('offers the choices of each conflict kind', () => {
    expect(conflictChoices(requestConflict()).map((choice) => choice.resolution)).toEqual(['mine', 'theirs', 'both'])
    expect(conflictChoices(requestConflict({ itemKind: 'environment' })).map((choice) => choice.resolution)).toEqual(['mine', 'theirs', 'both'])
    expect(conflictChoices(requestConflict({ itemKind: 'folder' })).map((choice) => choice.resolution)).toEqual(['mine', 'theirs'])
    expect(conflictChoices(environmentConflict).map((choice) => choice.label)).toEqual(['Keep', 'Delete'])
    expect(conflictChoices(fieldConflict)).toEqual([])
  })

  it('names kinds, resolutions and copies', () => {
    expect(conflictKindLabel(requestConflict())).toBe('Changed on both sides')
    expect(conflictKindLabel(environmentConflict)).toBe('Deleted on one side, changed on the other')
    expect(conflictKindLabel(fieldConflict)).toBe('Fields changed on both sides')
    expect(resolutionLabel('both')).toBe('Keep both')
    expect(resolutionLabel('merged')).toBe('Merged')
    expect(itemKindLabel('workspace')).toBe('Workspace settings')
    expect(itemKindLabel('other')).toBe('Item')
    expect(theirCopyName(requestConflict())).toBe('Users (theirs)')
  })

  it('counts resolved conflicts', () => {
    const conflicts = [requestConflict({ resolution: 'mine' }), fieldConflict, { ...fieldConflict, resolution: 'merged' as const }]
    expect(conflicts.map(isResolved)).toEqual([true, false, true])
    expect(resolvedCount(conflicts)).toBe(2)
  })
})
