import type { TreeNode } from '@/platform/mappers/tree'

export interface ContainerOption {
  id: string
  name: string
  depth: number
}

export function containerOptions(tree: TreeNode[], depth = 0): ContainerOption[] {
  return tree.flatMap((node) =>
    node.kind === 'collection' || node.kind === 'folder'
      ? [{ id: node.id, name: node.name, depth }, ...containerOptions(node.children, depth + 1)]
      : [],
  )
}
