import type { ClashResolution, ImportClash, ImportPlan, ImportPreview, ImportWarning } from '@/platform/mappers/import'
import type { TreeNode } from '@/platform/mappers/tree'
import type { ImportLocation, ImportTarget } from '@/features/import-export/model'

const PATH_SEPARATOR = ' / '
const SKIPPED_PART = 'skipped_part'
const REQUEST_ITEM = 'request'
const WORKSPACE_ITEM = 'workspace'

/** A request can only go into an existing collection or folder, never become a collection itself. */
export function needsLocation(kind: string): boolean {
  return kind === REQUEST_ITEM
}

/** A workspace export becomes a new workspace in a folder of its own rather than going into an existing one. */
export function needsFolder(kind: string): boolean {
  return kind === WORKSPACE_ITEM
}

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
      targetLocalId: target.folder ? '' : target.localId,
      targetParentId: target.folder ? '' : target.parentId,
      ...(target.name?.trim() ? { name: target.name.trim() } : {}),
      ...(target.folder ? { targetFolder: target.folder } : {}),
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

/** True while an included request has no collection or folder to go into, or an included workspace no folder to become. */
export function missingLocations(preview: ImportPreview, targets: Readonly<Record<string, ImportTarget>>): boolean {
  return preview.items.some((item) => {
    const target = targets[item.key]
    if (target?.include !== true) {
      return false
    }
    return (needsLocation(item.kind) && target.parentId === '') || (needsFolder(item.kind) && !target.folder)
  })
}

/** True when an included item becomes a new workspace, so the workspace list must be reloaded after the import. */
export function createsWorkspaces(preview: ImportPreview, targets: Readonly<Record<string, ImportTarget>>): boolean {
  return preview.items.some((item) => needsFolder(item.kind) && targets[item.key]?.include === true)
}

/** Puts requests without a location yet into the first collection or folder of their workspace. */
export function withRequestLocations(
  preview: ImportPreview,
  targets: Readonly<Record<string, ImportTarget>>,
  localId: string,
  locations: readonly ImportLocation[],
): Record<string, ImportTarget> {
  const [first] = locations
  const kinds = new Map(preview.items.map((item) => [item.key, item.kind]))
  return Object.fromEntries(
    Object.entries(targets).map(([key, target]) => {
      const kind = kinds.get(key) ?? ''
      const placeable = first !== undefined && needsLocation(kind) && target.localId === localId && target.parentId === ''
      return [key, placeable ? { ...target, parentId: first.id } : target]
    }),
  )
}

export function assignIncluded(targets: Readonly<Record<string, ImportTarget>>, localId: string): Record<string, ImportTarget> {
  return Object.fromEntries(
    Object.entries(targets).map(([key, target]) => [key, target.include ? { ...target, localId, parentId: '' } : target]),
  )
}

export function skippedParts(preview: ImportPreview): ImportWarning[] {
  return preview.warnings.filter((warning) => warning.code === SKIPPED_PART)
}
