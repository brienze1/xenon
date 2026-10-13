import type { TreeNode } from '@/platform/mappers/tree'

export interface ContainerOption {
  id: string
  kind: 'collection' | 'folder'
  name: string
  depth: number
}

export function containerOptions(tree: TreeNode[], depth = 0): ContainerOption[] {
  return tree.flatMap((node) =>
    node.kind === 'collection' || node.kind === 'folder'
      ? [{ id: node.id, kind: node.kind, name: node.name, depth }, ...containerOptions(node.children, depth + 1)]
      : [],
  )
}
