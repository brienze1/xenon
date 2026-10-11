import { describe, expect, it } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'
import { dropPositionAt, isDraggable, resolveDrop } from './tree-drop'

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: null, url: null, error: null, warning: null, children }
}

const tree = [
  node('c1', 'collection', [node('f1', 'folder', [node('f2', 'folder', [node('r3', 'request')]), node('r1', 'request')]), node('r2', 'request'), node('e1', 'error')]),
  node('c2', 'collection', [node('r4', 'request')]),
  node('c3', 'collection'),
]

describe('isDraggable', () => {
  it('drags collections, folders and requests only', () => {
    expect(['collection', 'folder', 'request'].map((kind) => isDraggable(node('x', kind as TreeNode['kind'])))).toEqual([true, true, true])
    expect(isDraggable(node('x', 'example'))).toBe(false)
    expect(isDraggable(node('x', 'error'))).toBe(false)
  })
})

describe('dropPositionAt', () => {
  const request = node('r', 'request')
  const folder = node('f', 'folder')
  const collection = node('c', 'collection')

  it.each([
    [0, 'before'],
    [6, 'before'],
    [7, 'inside'],
    [21, 'inside'],
    [22, 'after'],
    [28, 'after'],
  ] as const)('splits a folder row into quarters: %i px of 28 is %s', (offset, position) => {
    expect(dropPositionAt(offset, 28, request, folder)).toBe(position)
  })

  it.each([
    [request, request, 13, 'before'],
    [request, request, 14, 'after'],
    [request, node('e', 'error'), 20, 'after'],
    [request, collection, 1, 'inside'],
    [folder, collection, 27, 'inside'],
    [collection, collection, 13, 'before'],
    [collection, collection, 14, 'after'],
  ] as const)('drops %s over %s at %i px as %s', (dragged, target, offset, position) => {
    expect(dropPositionAt(offset, 28, dragged, target)).toBe(position)
  })

  it('refuses examples, collections over other rows and treats a row without height as its top', () => {
    expect(dropPositionAt(5, 28, request, node('x', 'example'))).toBeNull()
    expect(dropPositionAt(5, 28, collection, folder)).toBeNull()
    expect(dropPositionAt(5, 28, collection, request)).toBeNull()
    expect(dropPositionAt(5, 0, request, folder)).toBe('before')
  })
})

describe('resolveDrop', () => {
  it.each([
    ['r2', { id: 'r4', position: 'before' }, { parentId: 'c2', index: 0 }],
    ['r2', { id: 'r4', position: 'after' }, { parentId: 'c2', index: 1 }],
    ['r2', { id: 'c3', position: 'inside' }, { parentId: 'c3', index: 0 }],
    ['r2', { id: 'f1', position: 'inside' }, { parentId: 'f1', index: 2 }],
    ['r2', { id: 'f1', position: 'before' }, { parentId: 'c1', index: 0 }],
    ['r2', { id: 'e1', position: 'after' }, { parentId: 'c1', index: 2 }],
    ['r1', { id: 'f2', position: 'before' }, { parentId: 'f1', index: 0 }],
    ['f2', { id: 'c1', position: 'inside' }, { parentId: 'c1', index: 3 }],
    ['f1', { id: 'r2', position: 'after' }, { parentId: 'c1', index: 1 }],
    ['c3', { id: 'c1', position: 'before' }, { parentId: '', index: 0 }],
    ['c1', { id: 'c3', position: 'after' }, { parentId: '', index: 2 }],
  ] as const)('moves %s for %o to %o', (dragged, target, move) => {
    expect(resolveDrop(tree, dragged, target)).toEqual(move)
  })

  it.each([
    ['f1', { id: 'f1', position: 'inside' }],
    ['f1', { id: 'f2', position: 'inside' }],
    ['f1', { id: 'r3', position: 'before' }],
    ['r2', { id: 'c1', position: 'before' }],
    ['c2', { id: 'r1', position: 'before' }],
    ['c2', { id: 'c1', position: 'inside' }],
    ['r2', { id: 'r2', position: 'after' }],
    ['r2', { id: 'f1', position: 'after' }],
    ['r2', { id: 'e1', position: 'before' }],
    ['c2', { id: 'c3', position: 'before' }],
    ['missing', { id: 'r1', position: 'before' }],
    ['r1', { id: 'missing', position: 'before' }],
  ] as const)('refuses or ignores %s for %o', (dragged, target) => {
    expect(resolveDrop(tree, dragged, target)).toBeNull()
  })
})
