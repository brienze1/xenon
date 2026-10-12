import type { ApiOperation, ApiTagGroup } from '@/platform/mappers/documentation'

export const UNTAGGED = ''

export interface OperationGroup {
  /** The folder ("Products / Admin"), or empty for requests directly in the documented collection or folder. */
  tag: string
  operations: ApiOperation[]
}

/** Operations grouped by their folder tag, in the order the groups first appear. */
export function groupOperations(operations: ApiOperation[]): OperationGroup[] {
  const groups: OperationGroup[] = []
  for (const operation of operations) {
    const tag = operation.tag || UNTAGGED
    const group = groups.find((candidate) => candidate.tag === tag)
    if (group) {
      group.operations.push(operation)
    } else {
      groups.push({ tag, operations: [operation] })
    }
  }
  return groups
}

export interface OperationSection {
  /** The x-tagGroups group, or empty for the operations no group holds. */
  group: string
  tags: OperationGroup[]
}

/**
 * Operations sectioned by the document's x-tagGroups: first the folders no group lists (and the requests without a
 * folder), then each group with its folders in the group's order. Without groups, one unnamed section of all folders.
 */
export function sectionOperations(operations: ApiOperation[], tagGroups: ApiTagGroup[]): OperationSection[] {
  const groups = groupOperations(operations)
  if (tagGroups.length === 0) {
    return groups.length === 0 ? [] : [{ group: UNTAGGED, tags: groups }]
  }
  const grouped = new Set(tagGroups.flatMap((group) => group.tags))
  const sections: OperationSection[] = []
  const ungrouped = groups.filter((group) => group.tag === UNTAGGED || !grouped.has(group.tag))
  if (ungrouped.length > 0) {
    sections.push({ group: UNTAGGED, tags: ungrouped })
  }
  for (const tagGroup of tagGroups) {
    const tags = tagGroup.tags.flatMap((tag) => groups.filter((group) => group.tag === tag && tag !== UNTAGGED))
    if (tags.length > 0) {
      sections.push({ group: tagGroup.name, tags })
    }
  }
  return sections
}
