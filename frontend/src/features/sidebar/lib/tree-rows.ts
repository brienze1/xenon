import type { TreeNode, TreeSearchResult } from '@/platform/mappers/tree'

export interface TreeRow {
  node: TreeNode
  depth: number
  expanded: boolean
}

export function isContainer(node: TreeNode): boolean {
  return node.kind === 'collection' || node.kind === 'folder'
}

export function visibleRows(nodes: TreeNode[], expanded: Readonly<Record<string, boolean>>, search: TreeSearchResult | null): TreeRow[] {
  const rows: TreeRow[] = []
  const matches = search ? new Set(search.matchIds) : null
  const ancestors = search ? new Set(search.ancestorIds) : null
  function visit(level: TreeNode[], depth: number) {
    for (const node of level) {
      if (matches && ancestors && !matches.has(node.id) && !ancestors.has(node.id)) {
        continue
      }
      const open = isContainer(node) && (ancestors ? ancestors.has(node.id) : expanded[node.id] === true)
      rows.push({ node, depth, expanded: open })
      if (open) {
        visit(node.children, depth + 1)
      }
    }
  }
  visit(nodes, 0)
  return rows
}

export function findPath(nodes: TreeNode[], id: string): TreeNode[] {
  for (const node of nodes) {
    if (node.id === id) {
      return [node]
    }
    const below = findPath(node.children, id)
    if (below.length > 0) {
      return [node, ...below]
    }
  }
  return []
}
