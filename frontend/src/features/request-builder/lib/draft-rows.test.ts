import { describe, expect, it } from 'vitest'
import { emptyDraft, type KeyValueRow } from '@/platform/mappers/request'
import { rowsOf, withRows, withSavedRows } from './draft-rows'

function row(id: string, key: string): KeyValueRow {
  return { id, key, value: '', description: '', enabled: true }
}

describe('withSavedRows', () => {
  it('takes the saved draft but keeps the row ids the editor already has', () => {
    const base = emptyDraft()
    const current = {
      ...base,
      params: [row('p1', 'a')],
      pathVariables: [{ id: 'v1', key: 'id', value: '', description: '' }],
      headers: [row('h1', 'Accept')],
      body: { ...base.body, formdata: [{ ...row('f1', 'file'), type: 'file' as const, src: '', srcState: 'ok' as const }], urlencoded: [row('u1', 'q')] },
    }
    const saved = {
      ...base,
      id: 'r1',
      params: [row('new-1', 'a'), row('new-2', 'b')],
      pathVariables: [{ id: 'new-3', key: 'id', value: '7', description: '' }],
      headers: [row('new-4', 'Accept')],
      body: { ...base.body, formdata: [{ ...row('new-5', 'file'), type: 'file' as const, src: '/x', srcState: 'ok' as const }], urlencoded: [row('new-6', 'q')] },
    }
    const merged = withSavedRows(current, saved)
    expect(merged.id).toBe('r1')
    expect(merged.params.map((item) => item.id)).toEqual(['p1', 'new-2'])
    expect(merged.pathVariables[0]).toEqual({ id: 'v1', key: 'id', value: '7', description: '' })
    expect(merged.headers[0]!.id).toBe('h1')
    expect(merged.body.formdata[0]).toMatchObject({ id: 'f1', src: '/x' })
    expect(merged.body.urlencoded[0]!.id).toBe('u1')
  })
})

describe('rowsOf and withRows', () => {
  it('read and replace the params, headers and urlencoded tables', () => {
    const rows = [row('r', 'k')]
    for (const table of ['params', 'headers', 'urlencoded'] as const) {
      const draft = withRows(emptyDraft(), table, rows)
      expect(rowsOf(draft, table)).toBe(rows)
    }
    expect(withRows(emptyDraft(), 'urlencoded', rows).params).toEqual([])
  })
})
