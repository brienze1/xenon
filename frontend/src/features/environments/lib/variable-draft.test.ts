import { describe, expect, it } from 'vitest'
import type { VariableRow } from '@/platform/mappers/variable'
import { displayedCurrent, inheritsCurrent, keyCounts, keyProblems, newDraftRow, toDraftRows, toVariableEdits, typeSwitches } from './variable-draft'

function saved(patch: Partial<VariableRow> = {}): VariableRow {
  return {
    key: 'base_url',
    initialValue: 'A',
    currentValue: 'A',
    hasCurrentOverride: false,
    secret: false,
    secretState: null,
    enabled: true,
    description: '',
    duplicate: false,
    surroundingSpaces: false,
    ...patch,
  }
}

describe('variable draft', () => {
  it('shows the inherited initial value until a current value is set or after it is cleared', () => {
    const [inherited] = toDraftRows([saved()])
    expect(inheritsCurrent(inherited!)).toBe(true)
    expect(displayedCurrent({ ...inherited!, initialValue: 'edited' })).toBe('edited')
    const [overridden] = toDraftRows([saved({ currentValue: 'B', hasCurrentOverride: true })])
    expect(inheritsCurrent(overridden!)).toBe(false)
    expect(displayedCurrent(overridden!)).toBe('B')
    expect(displayedCurrent({ ...overridden!, currentValue: 'C' })).toBe('C')
    expect(inheritsCurrent({ ...overridden!, currentValue: '' })).toBe(true)
    expect(displayedCurrent({ ...overridden!, currentValue: '' })).toBe('A')
  })

  it('flags empty, invalid, duplicate and space-surrounded names', () => {
    const rows = [
      { ...newDraftRow(), key: 'a' },
      { ...newDraftRow(), key: 'a' },
      { ...newDraftRow(), key: ' b ' },
      { ...newDraftRow(), key: '{{c}}' },
      { ...newDraftRow(), key: '  ' },
    ]
    const counts = keyCounts(rows)
    expect(rows.map((row) => keyProblems(row, counts))).toEqual([['duplicate'], ['duplicate'], ['spaces'], ['invalid'], ['empty']])
  })

  it('keeps the duplicate flag Go reported until the name is edited', () => {
    const [row] = toDraftRows([saved({ duplicate: true })])
    expect(keyProblems(row!, keyCounts([row!]))).toEqual(['duplicate'])
    const renamed = { ...row!, key: 'renamed' }
    expect(keyProblems(renamed, keyCounts([renamed]))).toEqual([])
  })

  it('counts how many rows use each name', () => {
    const counts = keyCounts([{ ...newDraftRow(), key: 'a' }, { ...newDraftRow(), key: 'b' }, { ...newDraftRow(), key: 'a' }])
    expect(Object.fromEntries(counts)).toEqual({ a: 2, b: 1 })
  })

  it('lists rows switching type, a switch to plain needing confirmation', () => {
    const [plain, secret] = toDraftRows([saved({ key: 'p' }), saved({ key: 's', secret: true, secretState: 'set' })])
    expect(typeSwitches([plain!, secret!, newDraftRow()])).toEqual({ toSecret: [], toPlain: [], reveal: false })
    expect(typeSwitches([{ ...plain!, secret: true }, { ...secret!, secret: false }])).toEqual({ toSecret: ['p'], toPlain: ['s'], reveal: true })
  })

  it('turns the draft into table edits carrying renames and only changed values', () => {
    const [plain, secret] = toDraftRows([saved({ key: 'old' }), saved({ key: 's', secret: true, secretState: 'set', initialValue: '' })])
    const added = { ...newDraftRow(), key: 'token', secret: true, secretValue: 'xyz' }
    expect(toVariableEdits([{ ...plain!, key: 'new', currentValue: 'B' }, secret!, added])).toEqual([
      { key: 'new', previousKey: 'old', initialValue: 'A', currentValue: 'B', secret: false, secretValue: null, enabled: true, description: '' },
      { key: 's', previousKey: 's', initialValue: '', currentValue: null, secret: true, secretValue: null, enabled: true, description: '' },
      { key: 'token', previousKey: null, initialValue: '', currentValue: null, secret: true, secretValue: 'xyz', enabled: true, description: '' },
    ])
  })
})
