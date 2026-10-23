import {
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type Announcements,
  type DragEndEvent,
  type ScreenReaderInstructions,
  type UniqueIdentifier,
} from '@dnd-kit/core'
import { sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import { useEffect, useRef, useState } from 'react'
import type { RunPlanRow } from '@/platform/mappers/runner'
import { suspendTextSelection } from '@/shared/lib/text-selection'

const DRAG_DISTANCE_PX = 5

// Space picks a row up for a keyboard drag, the arrow keys move it, Space or Enter drops it and
// Escape cancels, as in the Environments view.
const KEYBOARD_CODES = { start: ['Space'], cancel: ['Escape'], end: ['Space', 'Enter'] }

const SCREEN_READER_INSTRUCTIONS: ScreenReaderInstructions = {
  draggable: 'To reorder a request, press Space to pick it up, use the arrow keys to move it, then press Space again to drop it or Escape to cancel.',
}

/** The position last announced during a drag. */
interface AnnouncedPosition {
  overId: UniqueIdentifier | null
}

/** Announcements naming the request and its position in the run order. */
export function announcementsFor(rows: readonly RunPlanRow[], announced: AnnouncedPosition): Announcements {
  const total = rows.length
  const nameOf = (id: UniqueIdentifier) => rows.find((row) => row.requestId === id)?.name ?? ''
  const positionOf = (id: UniqueIdentifier) => rows.findIndex((row) => row.requestId === id) + 1
  return {
    onDragStart: ({ active }) => {
      announced.overId = active.id
      return `Picked up ${nameOf(active.id)}.`
    },
    onDragOver: ({ active, over }) => {
      const overId = over?.id ?? null
      if (overId === announced.overId) {
        return undefined
      }
      announced.overId = overId
      return overId === null ? `${nameOf(active.id)} is no longer over a position.` : `${nameOf(active.id)} moved to position ${positionOf(overId)} of ${total}.`
    },
    onDragEnd: ({ active, over }) => `${nameOf(active.id)} dropped at position ${positionOf((over ?? active).id)} of ${total}.`,
    onDragCancel: ({ active }) => `Reordering cancelled. ${nameOf(active.id)} is back at position ${positionOf(active.id)} of ${total}.`,
  }
}

/** The DndContext setup of the run order: sensors, announcements and moving the row on drop. */
export function useRunOrderReorder(rows: readonly RunPlanRow[], onMove: (fromId: string, toId: string) => void) {
  const restoreSelection = useRef<(() => void) | null>(null)
  const [announced] = useState<AnnouncedPosition>(() => ({ overId: null }))
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: DRAG_DISTANCE_PX } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates, keyboardCodes: KEYBOARD_CODES }),
  )

  useEffect(() => {
    const drag = { restoreSelection }
    return () => drag.restoreSelection.current?.()
  }, [])

  function finish() {
    restoreSelection.current?.()
    restoreSelection.current = null
  }

  function onDragStart() {
    restoreSelection.current = suspendTextSelection()
  }

  function onDragEnd({ active, over }: DragEndEvent) {
    finish()
    if (over && active.id !== over.id) {
      onMove(String(active.id), String(over.id))
    }
  }

  return {
    sensors,
    accessibility: { screenReaderInstructions: SCREEN_READER_INSTRUCTIONS, announcements: announcementsFor(rows, announced) },
    onDragStart,
    onDragEnd,
    onDragCancel: finish,
  }
}
