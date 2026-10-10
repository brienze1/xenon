import type { ClashResolution, ImportClash, ImportPlan, ImportPreview } from '@/platform/mappers/import'
import type { TreeNode } from '@/platform/mappers/tree'
import type { ImportLocation, ImportTarget } from '@/features/import-export/model'

const PATH_SEPARATOR = ' / '

export function defaultTargets(preview: ImportPreview, localId: string): Record<string, ImportTarget> {
  return Object.fromEntries(preview.items.map((item) => [item.key, { include: true, localId, parentId: '' }]))
}

export function buildPlan(
  sourceId: string,
  targets: Readonly<Record<string, ImportTarget>>,
  resolutions: Readonly<Record<string, ClashResolution>>,
  applyToAll: ClashResolution | null,
): ImportPlan {
  return {
    sourceId,
    applyToAll,
    items: Object.entries(targets).map(([key, target]) => ({
      key,
      include: target.include,
      targetLocalId: target.localId,
      targetParentId: target.parentId,
      resolution: resolutions[key] ?? null,
    })),
  }
}

export function pendingClashes(clashes: readonly ImportClash[], resolutions: Readonly<Record<string, ClashResolution>>): ImportClash[] {
  return clashes.filter((clash) => resolutions[clash.itemKey] === undefined)
}

export function containerLocations(nodes: readonly TreeNode[], parentLabel = ''): ImportLocation[] {
  const locations: ImportLocation[] = []
  for (const node of nodes) {
    if (node.kind !== 'collection' && node.kind !== 'folder') {
      continue
    }
    const label = parentLabel ? parentLabel + PATH_SEPARATOR + node.name : node.name
    locations.push({ id: node.id, label }, ...containerLocations(node.children, label))
  }
  return locations
}

export function hasIncludedItems(targets: Readonly<Record<string, ImportTarget>>): boolean {
  return Object.values(targets).some((target) => target.include)
}
