import { DndContext, KeyboardSensor, PointerSensor, closestCenter, useSensor, useSensors, type Announcements, type DragEndEvent, type ScreenReaderInstructions } from '@dnd-kit/core'
import { SortableContext, horizontalListSortingStrategy, sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import { useRef, type KeyboardEvent } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { moveFocus } from '@/shared/lib/roving-focus'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useTabsStore } from '@/features/tabs/store'
import { TabItem } from './TabItem'
import styles from './TabStrip.module.css'

const DRAG_DISTANCE_PX = 5

// Enter activates a tab; Space picks it up for a keyboard drag, the arrow keys move it while
// dragging (roving focus is paused), Space or Enter drops it and Escape cancels.
const KEYBOARD_CODES = { start: ['Space'], cancel: ['Escape'], end: ['Space', 'Enter'] }

const SCREEN_READER_INSTRUCTIONS: ScreenReaderInstructions = {
  draggable: 'To reorder a tab, press Space to pick it up, use the arrow keys to move it, then press Space again to drop it or Escape to cancel.',
}

const ANNOUNCEMENTS: Announcements = {
  onDragStart: () => 'Picked up tab.',
  onDragOver: ({ over }) => (over ? 'Tab moved.' : 'Tab is no longer over a position.'),
  onDragEnd: ({ over }) => (over ? 'Tab dropped.' : 'Tab dropped back in place.'),
  onDragCancel: () => 'Reordering cancelled.',
}

const ACCESSIBILITY = { screenReaderInstructions: SCREEN_READER_INSTRUCTIONS, announcements: ANNOUNCEMENTS }

export function TabStrip() {
  const tabIds = useTabsStore(useShallow((state) => state.tabs.map((tab) => tab.id)))
  const reorder = useTabsStore((state) => state.reorder)
  const newTab = useTabsStore((state) => state.newTab)
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: DRAG_DISTANCE_PX } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates, keyboardCodes: KEYBOARD_CODES }),
  )
  const draggingRef = useRef(false)

  function onDragEnd({ active, over }: DragEndEvent) {
    draggingRef.current = false
    if (over && active.id !== over.id) {
      reorder(String(active.id), String(over.id))
    }
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (draggingRef.current || (event.target as HTMLElement).getAttribute('role') !== 'tab') {
      return
    }
    if (moveFocus(event.currentTarget, event.key, '[role="tab"]', 'horizontal')) {
      event.preventDefault()
    }
  }

  return (
    <div className={styles.strip}>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        accessibility={ACCESSIBILITY}
        onDragStart={() => {
          draggingRef.current = true
        }}
        onDragEnd={onDragEnd}
        onDragCancel={() => {
          draggingRef.current = false
        }}
      >
        <SortableContext items={tabIds} strategy={horizontalListSortingStrategy}>
          <div role="tablist" aria-label="Open requests" className={styles.list} onKeyDown={onKeyDown}>
            {tabIds.map((tabId) => (
              <TabItem key={tabId} tabId={tabId} />
            ))}
          </div>
        </SortableContext>
      </DndContext>
      <IconButton label="New Tab" className={styles.add} onClick={() => newTab()}>
        <span aria-hidden="true">+</span>
      </IconButton>
    </div>
  )
}
