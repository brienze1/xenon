import { describe, expect, it } from 'vitest'
import type { ImportPreview } from '@/platform/mappers/import'
import type { TreeNode } from '@/platform/mappers/tree'
import {
  assignIncluded,
  buildPlan,
  containerLocations,
  defaultTargets,
  hasIncludedItems,
  missingLocations,
  needsLocation,
  pendingClashes,
  skippedParts,
  withRequestLocations,
} from './plan'

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
  encryptedSecrets: false,
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

  it('assigns every included item to one workspace as a new collection', () => {
    const targets = { '0': { include: true, localId: 'L', parentId: 'F' }, '1': { include: false, localId: 'M', parentId: 'G' } }
    expect(assignIncluded(targets, 'N')).toEqual({ '0': { include: true, localId: 'N', parentId: '' }, '1': { include: false, localId: 'M', parentId: 'G' } })
  })

  it('picks the skipped parts out of the source warnings', () => {
    const skipped = { code: 'skipped_part', severity: 'warning' as const, message: 'mock skipped', itemKey: '', itemId: '', localId: '', path: 'mock', location: '' }
    const other = { ...skipped, code: 'missing_file', path: 'x' }
    expect(skippedParts({ ...preview, warnings: [other, skipped] })).toEqual([skipped])
  })

  it('sends a trimmed new name only when one was typed', () => {
    const targets = { '0': { include: true, localId: 'L', parentId: 'F', name: '  Invite  ' }, '1': { include: true, localId: 'L', parentId: 'F', name: ' ' } }
    expect(buildPlan('S', targets, {}, null).items).toEqual([
      { key: '0', include: true, targetLocalId: 'L', targetParentId: 'F', name: 'Invite', resolution: null },
      { key: '1', include: true, targetLocalId: 'L', targetParentId: 'F', resolution: null },
    ])
  })

  it('needs a collection or folder for every included request', () => {
    const curl: ImportPreview = { ...preview, kind: 'curl', items: [{ key: '0', kind: 'request', name: 'GET /x', counts, warningsCount: 0 }] }
    expect(needsLocation('request')).toBe(true)
    expect(needsLocation('collection')).toBe(false)
    expect(missingLocations(curl, { '0': { include: true, localId: 'L', parentId: '' } })).toBe(true)
    expect(missingLocations(curl, { '0': { include: false, localId: 'L', parentId: '' } })).toBe(false)
    expect(missingLocations(curl, { '0': { include: true, localId: 'L', parentId: 'F' } })).toBe(false)
    expect(missingLocations(preview, defaultTargets(preview, 'L'))).toBe(false)
  })

  it('puts requests without a location into the first collection of their workspace', () => {
    const mixed: ImportPreview = { ...preview, items: [...preview.items, { key: '2', kind: 'request', name: 'R', counts, warningsCount: 0 }, { key: '3', kind: 'request', name: 'S', counts, warningsCount: 0 }] }
    const targets = {
      '0': { include: true, localId: 'L', parentId: '' },
      '2': { include: true, localId: 'L', parentId: '' },
      '3': { include: true, localId: 'M', parentId: '' },
      '4': { include: true, localId: 'L', parentId: '' },
    }
    const locations = [{ id: 'C', label: 'Main' }, { id: 'F', label: 'Main / F' }]
    expect(withRequestLocations(mixed, targets, 'L', locations)).toEqual({ ...targets, '2': { include: true, localId: 'L', parentId: 'C' } })
    expect(withRequestLocations(mixed, targets, 'L', [])).toEqual(targets)
    expect(withRequestLocations(mixed, { '2': { include: true, localId: 'L', parentId: 'F' } }, 'L', locations)).toEqual({ '2': { include: true, localId: 'L', parentId: 'F' } })
  })
})
