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
import type { AppError } from '@/platform/errors'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { suspendTextSelection } from '@/shared/lib/text-selection'
import { useEnvironmentsStore } from '@/features/environments/store'

const DRAG_DISTANCE_PX = 5

// Space picks an environment up for a keyboard drag, the arrow keys move it, Space or Enter drops
// it and Escape cancels, as in the tab strip.
const KEYBOARD_CODES = { start: ['Space'], cancel: ['Escape'], end: ['Space', 'Enter'] }

const SCREEN_READER_INSTRUCTIONS: ScreenReaderInstructions = {
  draggable: 'To reorder an environment, press Space to pick it up, use the arrow keys to move it, then press Space again to drop it or Escape to cancel.',
}

/** The position last announced during a drag. */
interface AnnouncedPosition {
  overId: UniqueIdentifier | null
}

/**
 * Announcements that name the environment and its position among the readable ones. `announced`
 * keeps the position last announced, so the drag-over dnd-kit reports right after a pick-up (the
 * row over itself) does not drown out "Picked up".
 */
function announcementsFor(environments: EnvironmentSummary[], announced: AnnouncedPosition): Announcements {
  const total = environments.length
  const nameOf = (id: UniqueIdentifier) => environments.find((environment) => environment.id === id)?.name ?? ''
  const positionOf = (id: UniqueIdentifier) => environments.findIndex((environment) => environment.id === id) + 1
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

/** The DndContext setup of the Environments view: sensors, announcements and saving the order on drop. */
export function useEnvironmentReorder(environments: EnvironmentSummary[], onError: (error: AppError | null) => void) {
  const reorderEnvironment = useEnvironmentsStore((state) => state.reorderEnvironment)
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

  async function onDragEnd({ active, over }: DragEndEvent) {
    finish()
    if (!over || active.id === over.id) {
      return
    }
    const index = environments.findIndex((environment) => environment.id === over.id)
    onError(await reorderEnvironment(String(active.id), index))
  }

  return {
    sensors,
    accessibility: { screenReaderInstructions: SCREEN_READER_INSTRUCTIONS, announcements: announcementsFor(environments, announced) },
    onDragStart,
    onDragEnd,
    onDragCancel: finish,
  }
}
