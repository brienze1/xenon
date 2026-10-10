import { describe, expect, it } from 'vitest'
import type { ImportPreview } from '@/platform/mappers/import'
import type { TreeNode } from '@/platform/mappers/tree'
import { buildPlan, containerLocations, defaultTargets, hasIncludedItems, pendingClashes } from './plan'

const counts = { collections: 1, folders: 0, requests: 1, examples: 0, variables: 0, scripts: 0, environments: 0 }
const preview: ImportPreview = {
  sourceId: 'S',
  kind: 'postman_collection',
  version: '2.1',
  items: [
    { key: '0', kind: 'collection', name: 'A', counts, warningsCount: 0 },
    { key: '1', kind: 'collection', name: 'B', counts, warningsCount: 0 },
  ],
  warnings: [],
  pmUsages: [],
}

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children }
}

describe('import plan helpers', () => {
  it('targets every item at the given workspace as a new collection', () => {
    expect(defaultTargets(preview, 'L')).toEqual({ '0': { include: true, localId: 'L', parentId: '' }, '1': { include: true, localId: 'L', parentId: '' } })
  })

  it('builds the plan with per-item resolutions and apply to all', () => {
    const targets = { '0': { include: true, localId: 'L', parentId: 'F' }, '1': { include: false, localId: 'M', parentId: '' } }
    expect(buildPlan('S', targets, { '0': 'replace' }, 'skip')).toEqual({
      sourceId: 'S',
      applyToAll: 'skip',
      items: [
        { key: '0', include: true, targetLocalId: 'L', targetParentId: 'F', resolution: 'replace' },
        { key: '1', include: false, targetLocalId: 'M', targetParentId: '', resolution: null },
      ],
    })
    expect(hasIncludedItems(targets)).toBe(true)
    expect(hasIncludedItems({ '1': targets['1'] })).toBe(false)
  })

  it('lists the clashes of items without a resolution yet', () => {
    const clashes = [
      { itemKey: '0', kind: 'collection', existingId: 'C', existingName: 'A', name: 'A' },
      { itemKey: '1', kind: 'request', existingId: 'R', existingName: 'B', name: 'B' },
    ]
    expect(pendingClashes(clashes, { '0': 'skip' })).toEqual([clashes[1]])
  })

  it('lists collections and folders with their path as import locations', () => {
    const tree = [node('C', 'collection', 'Main', [node('F', 'folder', 'Target', [node('R', 'request', 'Ping')]), node('E', 'error', 'bad')])]
    expect(containerLocations(tree)).toEqual([
      { id: 'C', label: 'Main' },
      { id: 'F', label: 'Main / Target' },
    ])
  })
})
