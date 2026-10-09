import { describe, expect, it } from 'vitest'
import { addRow, keepRowIds, newRowId, patchRow, removeRow } from './rows'

interface Row {
  id: string
  key: string
}

describe('rows', () => {
  it('hands out unique row ids', () => {
    expect(newRowId()).not.toBe(newRowId())
    expect(newRowId()).toMatch(/^row-\d+$/)
  })

  it('patches, adds and removes rows by id without mutating the input', () => {
    const rows: Row[] = [
      { id: 'a', key: 'one' },
      { id: 'b', key: 'two' },
    ]
    expect(patchRow(rows, 'b', { key: 'TWO' })).toEqual([
      { id: 'a', key: 'one' },
      { id: 'b', key: 'TWO' },
    ])
    const added = addRow(rows, { key: 'three' })
    expect(added).toHaveLength(3)
    expect(added[2]).toMatchObject({ key: 'three', id: expect.stringMatching(/^row-/) })
    expect(removeRow(rows, 'a')).toEqual([{ id: 'b', key: 'two' }])
    expect(rows).toHaveLength(2)
  })

  it('keeps the current ids of rows that line up by position', () => {
    const current: Row[] = [{ id: 'a', key: 'one' }]
    const next: Row[] = [
      { id: 'x', key: 'uno' },
      { id: 'y', key: 'dos' },
    ]
    expect(keepRowIds(current, next)).toEqual([
      { id: 'a', key: 'uno' },
      { id: 'y', key: 'dos' },
    ])
  })
})
