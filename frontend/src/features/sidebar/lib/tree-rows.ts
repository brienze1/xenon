import type { TreeNode, TreeSearchResult } from '@/platform/mappers/tree'

export interface TreeRow {
  node: TreeNode
  depth: number
  expanded: boolean
  /** The response status code of an example row. */
  code?: number
  /** The request an example row belongs to. */
  requestId?: string
}

export function isContainer(node: TreeNode): boolean {
  return node.kind === 'collection' || node.kind === 'folder'
}

/** Containers and requests with saved examples open and close. */
export function isExpandable(node: TreeNode): boolean {
  return isContainer(node) || (node.kind === 'request' && (node.examples?.length ?? 0) > 0)
}

function exampleRows(request: TreeNode, depth: number): TreeRow[] {
  return (request.examples ?? []).map((example) => ({
    node: { id: example.id, kind: 'example', name: example.name, method: null, url: null, error: null, warning: null, children: [] },
    depth,
    expanded: false,
    code: example.code,
    requestId: request.id,
  }))
}

/** The request that owns the example, or null. */
export function findExampleRequest(nodes: TreeNode[], exampleId: string): TreeNode | null {
  for (const node of nodes) {
    if (node.kind === 'request' && node.examples?.some((example) => example.id === exampleId)) {
      return node
    }
    const below = findExampleRequest(node.children, exampleId)
    if (below) {
      return below
    }
  }
  return null
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
      const open = isExpandable(node) && (ancestors ? ancestors.has(node.id) : expanded[node.id] === true)
      rows.push({ node, depth, expanded: open })
      if (open && node.kind === 'request') {
        rows.push(...exampleRows(node, depth + 1))
      } else if (open) {
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
