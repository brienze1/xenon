import type { TreeNode } from '@/platform/mappers/tree'
import type { Crumb } from '@/features/sidebar/model'
import { findExampleRequest, findPath } from './tree-rows'

const SEPARATOR = ' / '

function toCrumb(node: TreeNode): Crumb {
  return { id: node.id, kind: node.kind === 'error' ? 'request' : node.kind, name: node.name }
}

/** The collection and folders down to a request or an example (the example under its request); empty when the tree lacks the item. */
export function breadcrumbTrail(tree: TreeNode[], id: string): Crumb[] {
  const path = findPath(tree, id)
  if (path.length > 0) {
    return path.map(toCrumb)
  }
  const request = findExampleRequest(tree, id)
  const example = request?.examples?.find((candidate) => candidate.id === id)
  if (!request || !example) {
    return []
  }
  return [...findPath(tree, request.id).map(toCrumb), { id: example.id, kind: 'example', name: example.name }]
}

/** The whole trail as one line, for a tooltip. */
export function trailTitle(trail: Crumb[]): string {
  return trail.map((crumb) => crumb.name).join(SEPARATOR)
}
