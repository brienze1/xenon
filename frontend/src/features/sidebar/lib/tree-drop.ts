import type { TreeNode } from '@/platform/mappers/tree'
import { findPath, isContainer } from './tree-rows'

export type DropPosition = 'before' | 'inside' | 'after'

export interface DropTarget {
  id: string
  position: DropPosition
}

/** Where a drop lands: the new parent ('' for the top level) and the index among its other children. */
export interface MoveTarget {
  parentId: string
  index: number
}

/** A row of a container takes drops inside its middle half, before its top quarter and after its bottom quarter. */
const EDGE_SHARE = 0.25

export function isDraggable(node: TreeNode): boolean {
  return isContainer(node) || node.kind === 'request'
}

/** The drop position the pointer at `offset` px from the top of a `height` px row of `target` means for `dragged`. */
export function dropPositionAt(offset: number, height: number, dragged: TreeNode, target: TreeNode): DropPosition | null {
  const share = height > 0 ? offset / height : 0
  if (target.kind === 'example') {
    return null
  }
  if (dragged.kind === 'collection') {
    return target.kind === 'collection' ? (share < 0.5 ? 'before' : 'after') : null
  }
  if (target.kind === 'collection') {
    return 'inside'
  }
  if (target.kind === 'folder') {
    return share < EDGE_SHARE ? 'before' : share > 1 - EDGE_SHARE ? 'after' : 'inside'
  }
  return share < 0.5 ? 'before' : 'after'
}

function childrenOf(tree: TreeNode[], parentId: string): TreeNode[] {
  return parentId === '' ? tree : (findPath(tree, parentId).at(-1)?.children ?? [])
}

/** The parent and index the drop resolves to, or null when it is refused or changes nothing. */
export function resolveDrop(tree: TreeNode[], draggedId: string, target: DropTarget): MoveTarget | null {
  const draggedPath = findPath(tree, draggedId)
  const targetPath = findPath(tree, target.id)
  const dragged = draggedPath.at(-1)
  const over = targetPath.at(-1)
  if (!dragged || !over || targetPath.some((node) => node.id === draggedId)) {
    return null
  }
  const parentId = target.position === 'inside' ? over.id : (targetPath.at(-2)?.id ?? '')
  if ((parentId === '') !== (dragged.kind === 'collection')) {
    return null
  }
  const siblings = childrenOf(tree, parentId).filter((node) => node.id !== draggedId)
  const index = target.position === 'inside' ? siblings.length : siblings.findIndex((node) => node.id === over.id) + (target.position === 'after' ? 1 : 0)
  const currentParentId = draggedPath.at(-2)?.id ?? ''
  const currentIndex = childrenOf(tree, currentParentId).findIndex((node) => node.id === draggedId)
  return parentId === currentParentId && index === currentIndex ? null : { parentId, index }
}
