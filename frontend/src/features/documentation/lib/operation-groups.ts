import type { ApiOperation } from '@/platform/mappers/documentation'

export const UNTAGGED = ''

export interface OperationGroup {
  /** The folder ("Products / Admin"), or empty for requests directly in the collection. */
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
