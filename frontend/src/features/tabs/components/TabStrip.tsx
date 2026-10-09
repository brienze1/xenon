import { DndContext, PointerSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core'
import { SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable'
import { useShallow } from 'zustand/react/shallow'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useTabsStore } from '../store'
import { TabItem } from './TabItem'
import styles from './TabStrip.module.css'

const DRAG_DISTANCE_PX = 5

export function TabStrip() {
  const tabIds = useTabsStore(useShallow((state) => state.tabs.map((tab) => tab.id)))
  const reorder = useTabsStore((state) => state.reorder)
  const newTab = useTabsStore((state) => state.newTab)
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: DRAG_DISTANCE_PX } }))

  function onDragEnd({ active, over }: DragEndEvent) {
    if (over && active.id !== over.id) {
      reorder(String(active.id), String(over.id))
    }
  }

  return (
    <div className={styles.strip}>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
        <SortableContext items={tabIds} strategy={horizontalListSortingStrategy}>
          <div role="tablist" aria-label="Open requests" className={styles.list}>
            {tabIds.map((tabId) => (
              <TabItem key={tabId} tabId={tabId} />
            ))}
          </div>
        </SortableContext>
      </DndContext>
      <IconButton label="New tab" className={styles.add} onClick={() => newTab()}>
        +
      </IconButton>
    </div>
  )
}
