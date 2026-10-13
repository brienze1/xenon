import { describe, expect, it } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'
import { breadcrumbTrail, trailTitle } from './breadcrumb'

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = [], examples: TreeNode['examples'] = []): TreeNode {
  return { id, kind, name: `Name ${id}`, method: kind === 'request' ? 'GET' : null, url: null, error: null, warning: null, children, examples }
}

const tree = [
  node('c1', 'collection', [node('f1', 'folder', [node('f2', 'folder', [node('r1', 'request', [], [{ id: 'x1', name: 'Created', code: 201 }])])]), node('r2', 'request')]),
  node('c2', 'collection'),
]

describe('breadcrumb trail', () => {
  it('lists the collection and folders down to a request', () => {
    expect(breadcrumbTrail(tree, 'r1')).toEqual([
      { id: 'c1', kind: 'collection', name: 'Name c1' },
      { id: 'f1', kind: 'folder', name: 'Name f1' },
      { id: 'f2', kind: 'folder', name: 'Name f2' },
      { id: 'r1', kind: 'request', name: 'Name r1' },
    ])
    expect(breadcrumbTrail(tree, 'r2').map((crumb) => crumb.id)).toEqual(['c1', 'r2'])
  })

  it('ends an example trail with the example under its request', () => {
    expect(breadcrumbTrail(tree, 'x1')).toEqual([
      { id: 'c1', kind: 'collection', name: 'Name c1' },
      { id: 'f1', kind: 'folder', name: 'Name f1' },
      { id: 'f2', kind: 'folder', name: 'Name f2' },
      { id: 'r1', kind: 'request', name: 'Name r1' },
      { id: 'x1', kind: 'example', name: 'Created' },
    ])
  })

  it('is empty for an item the tree does not have', () => {
    expect(breadcrumbTrail(tree, 'missing')).toEqual([])
    expect(breadcrumbTrail([], 'r1')).toEqual([])
  })

  it('joins the names into one readable path', () => {
    expect(trailTitle(breadcrumbTrail(tree, 'r2'))).toBe('Name c1 / Name r2')
  })
})
