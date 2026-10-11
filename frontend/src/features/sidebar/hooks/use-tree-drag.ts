import type { DragEndEvent, DragMoveEvent, DragStartEvent } from '@dnd-kit/core'
import { useEffect, useRef } from 'react'
import { suspendTextSelection } from '@/shared/lib/text-selection'
import { dropPositionAt, resolveDrop, type DropTarget } from '@/features/sidebar/lib/tree-drop'
import { findPath, isContainer } from '@/features/sidebar/lib/tree-rows'
import { useSidebarStore } from '@/features/sidebar/store'

/** How long a dragged item hovers inside a closed folder or collection before it opens. */
export const AUTO_EXPAND_DELAY_MS = 600

interface PendingExpand {
  id: string
  timer: ReturnType<typeof setTimeout>
}

function pointerY(event: DragMoveEvent): number | null {
  const start = event.activatorEvent
  return start instanceof MouseEvent ? start.clientY + event.delta.y : null
}

/** The drop the pointer means right now, or null when it is refused or changes nothing. */
function dropTargetOf(event: DragMoveEvent): DropTarget | null {
  const { tree } = useSidebarStore.getState()
  const y = pointerY(event)
  const dragged = findPath(tree, String(event.active.id)).at(-1)
  const over = event.over ? findPath(tree, String(event.over.id)).at(-1) : undefined
  if (!event.over || y === null || !dragged || !over) {
    return null
  }
  const position = dropPositionAt(y - event.over.rect.top, event.over.rect.height, dragged, over)
  const target = position ? { id: over.id, position } : null
  return target && resolveDrop(tree, dragged.id, target) ? target : null
}

/** The DndContext handlers of the sidebar tree: drop indicator, auto-expand after a hover, move on drop. */
export function useTreeDrag() {
  const pending = useRef<PendingExpand | null>(null)
  const restoreSelection = useRef<(() => void) | null>(null)

  function stopExpanding() {
    if (pending.current) {
      clearTimeout(pending.current.timer)
      pending.current = null
    }
  }

  function expandLater(id: string) {
    if (pending.current?.id === id) {
      return
    }
    stopExpanding()
    const timer = setTimeout(() => {
      pending.current = null
      const sidebar = useSidebarStore.getState()
      if (sidebar.draggingId !== null && !sidebar.expanded[id]) {
        sidebar.toggle(id)
      }
    }, AUTO_EXPAND_DELAY_MS)
    pending.current = { id, timer }
  }

  function finish() {
    stopExpanding()
    restoreSelection.current?.()
    restoreSelection.current = null
    useSidebarStore.getState().endDrag()
  }

  useEffect(() => {
    const drag = { pending, restoreSelection }
    return () => {
      clearTimeout(drag.pending.current?.timer)
      drag.restoreSelection.current?.()
    }
  }, [])

  function onDragStart({ active }: DragStartEvent) {
    restoreSelection.current = suspendTextSelection()
    useSidebarStore.getState().startDrag(String(active.id))
  }

  function onDragMove(event: DragMoveEvent) {
    const target = dropTargetOf(event)
    const sidebar = useSidebarStore.getState()
    sidebar.setDropTarget(target)
    const node = target?.position === 'inside' ? findPath(sidebar.tree, target.id).at(-1) : undefined
    if (target && node && isContainer(node) && !sidebar.expanded[target.id]) {
      expandLater(target.id)
    } else {
      stopExpanding()
    }
  }

  function onDragEnd(event: DragEndEvent) {
    const target = dropTargetOf(event)
    const draggedId = String(event.active.id)
    finish()
    const move = target ? resolveDrop(useSidebarStore.getState().tree, draggedId, target) : null
    if (move) {
      void useSidebarStore.getState().move(draggedId, move)
    }
  }

  return { onDragStart, onDragMove, onDragEnd, onDragCancel: finish }
}
