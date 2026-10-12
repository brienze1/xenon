import type { ChosenResolution, ConflictResolution, SyncConflict } from '@/platform/mappers/sync'

export interface ConflictChoice {
  resolution: ChosenResolution
  label: string
}

const KIND_LABELS: Record<SyncConflict['kind'], string> = {
  both_modified: 'Changed on both sides',
  delete_modify: 'Deleted on one side, changed on the other',
  field: 'Fields changed on both sides',
}

const RESOLUTION_LABELS: Record<ConflictResolution, string> = {
  unresolved: 'Unresolved',
  mine: 'Keep mine',
  theirs: 'Keep theirs',
  both: 'Keep both',
  keep: 'Keep',
  delete: 'Delete',
  merged: 'Merged',
}

const ITEM_KIND_LABELS: Record<string, string> = {
  request: 'Request',
  environment: 'Environment',
  collection: 'Collection settings',
  folder: 'Folder settings',
  workspace: 'Workspace settings',
}

export const FIELD_CHOICES: readonly ConflictChoice[] = [
  { resolution: 'mine', label: 'Keep mine' },
  { resolution: 'theirs', label: 'Keep theirs' },
]

export function conflictKindLabel(conflict: SyncConflict): string {
  return KIND_LABELS[conflict.kind]
}

export function resolutionLabel(resolution: ConflictResolution): string {
  return RESOLUTION_LABELS[resolution]
}

export function itemKindLabel(itemKind: string): string {
  return ITEM_KIND_LABELS[itemKind] ?? 'Item'
}

/** The choices of an item conflict (spec FR-011); field conflicts are chosen per field. */
export function conflictChoices(conflict: SyncConflict): readonly ConflictChoice[] {
  if (conflict.kind === 'delete_modify') {
    return [
      { resolution: 'keep', label: 'Keep' },
      { resolution: 'delete', label: 'Delete' },
    ]
  }
  if (conflict.kind === 'field') {
    return []
  }
  if (conflict.itemKind === 'request' || conflict.itemKind === 'environment') {
    return [...FIELD_CHOICES, { resolution: 'both', label: 'Keep both' }]
  }
  return FIELD_CHOICES
}

export function isResolved(conflict: SyncConflict): boolean {
  return conflict.resolution !== 'unresolved'
}

export function resolvedCount(conflicts: readonly SyncConflict[]): number {
  return conflicts.filter(isResolved).length
}

/** What Keep both adds, e.g. "Users (theirs)". */
export function theirCopyName(conflict: SyncConflict): string {
  return `${conflict.itemName} (theirs)`
}
