import { describe, expect, it } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'
import { findExampleRequest, findPath, isContainer, isExpandable, visibleRows } from './tree-rows'

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: kind === 'request' ? 'GET' : null, url: null, error: null, warning: null, children }
}

const tree = [
  node('c1', 'collection', [node('f1', 'folder', [node('r1', 'request')]), node('r2', 'request'), node('e1', 'error')]),
  node('c2', 'collection', [node('r3', 'request')]),
]

describe('tree rows', () => {
  it('shows only expanded containers', () => {
    expect(visibleRows(tree, {}, null).map((row) => row.node.id)).toEqual(['c1', 'c2'])
    const rows = visibleRows(tree, { c1: true, f1: true }, null)
    expect(rows.map((row) => [row.node.id, row.depth, row.expanded])).toEqual([
      ['c1', 0, true],
      ['f1', 1, true],
      ['r1', 2, false],
      ['r2', 1, false],
      ['e1', 1, false],
      ['c2', 0, false],
    ])
  })

  it('shows matches with their ancestors opened while searching', () => {
    const rows = visibleRows(tree, {}, { matchIds: ['r1'], ancestorIds: ['c1', 'f1'] })
    expect(rows.map((row) => row.node.id)).toEqual(['c1', 'f1', 'r1'])
    expect(visibleRows(tree, { c1: true }, { matchIds: [], ancestorIds: [] })).toEqual([])
  })

  it('finds the path to an item', () => {
    expect(findPath(tree, 'r1').map((item) => item.id)).toEqual(['c1', 'f1', 'r1'])
    expect(findPath(tree, 'missing')).toEqual([])
    expect(isContainer(node('x', 'folder'))).toBe(true)
    expect(isContainer(node('x', 'request'))).toBe(false)
  })

  it('nests the examples of an expanded request', () => {
    const request = { ...node('r9', 'request'), examples: [{ id: 'x1', name: 'Created', code: 201 }] }
    const withExamples = [node('c9', 'collection', [request])]
    expect(isExpandable(request)).toBe(true)
    expect(isExpandable(node('r0', 'request'))).toBe(false)
    expect(visibleRows(withExamples, { c9: true }, null).map((row) => row.node.id)).toEqual(['c9', 'r9'])
    const rows = visibleRows(withExamples, { c9: true, r9: true }, null)
    expect(rows.map((row) => [row.node.id, row.node.kind, row.depth, row.code])).toEqual([
      ['c9', 'collection', 0, undefined],
      ['r9', 'request', 1, undefined],
      ['x1', 'example', 2, 201],
    ])
    expect(findExampleRequest(withExamples, 'x1')?.id).toBe('r9')
    expect(findExampleRequest(withExamples, 'nope')).toBeNull()
  })
})
