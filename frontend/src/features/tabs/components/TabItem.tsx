import { useSortable } from '@dnd-kit/sortable'
import type { KeyboardEvent, MouseEvent } from 'react'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { cn } from '@/shared/lib/cn'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useTabsStore } from '@/features/tabs/store'
import styles from './TabItem.module.css'

const MIDDLE_BUTTON = 1
const UNTITLED = 'Untitled Request'

interface TabItemProps {
  tabId: string
}

export function TabItem({ tabId }: TabItemProps) {
  const active = useTabsStore((state) => state.activeTabId === tabId)
  const preview = useTabsStore((state) => state.previewTabId === tabId)
  const activate = useTabsStore((state) => state.activate)
  const pin = useTabsStore((state) => state.pin)
  const requestClose = useTabsStore((state) => state.requestClose)
  const name = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.name || UNTITLED)
  const method = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.method ?? 'GET')
  const dirty = useRequestBuilderStore((state) => state.tabs[tabId]?.dirty === true)
  const { setNodeRef, attributes, listeners, transform, transition, isDragging } = useSortable({ id: tabId, attributes: { role: 'tab' } })

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (isDragging) {
      return
    }
    if (event.key === 'Delete') {
      event.preventDefault()
      requestClose(tabId)
      return
    }
    if (event.key === 'Enter') {
      event.preventDefault()
      activate(tabId)
      return
    }
    if (event.key === ' ') {
      activate(tabId)
      // Space also picks the tab up for a keyboard drag (dnd-kit KeyboardSensor); the sensor
      // ignores events that were already default-prevented, so it runs first.
      listeners?.onKeyDown?.(event)
      event.preventDefault()
    }
  }

  function onAuxClick(event: MouseEvent<HTMLDivElement>) {
    if (event.button === MIDDLE_BUTTON) {
      event.preventDefault()
      requestClose(tabId)
    }
  }

  function onClose(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation()
    requestClose(tabId)
  }

  return (
    <div
      ref={setNodeRef}
      {...attributes}
      {...listeners}
      role="tab"
      aria-selected={active}
      tabIndex={active ? 0 : -1}
      title={preview ? `${name} (preview)` : name}
      className={cn(styles.tab, active && styles.active, preview && styles.preview, isDragging && styles.dragging)}
      style={{ transform: transform ? `translate3d(${transform.x}px, 0, 0)` : undefined, transition }}
      onClick={() => activate(tabId)}
      onDoubleClick={() => pin(tabId)}
      onAuxClick={onAuxClick}
      onKeyDown={onKeyDown}
    >
      <MethodBadge method={method} className={styles.method} />
      <span className={styles.name}>{name}</span>
      {dirty ? <span className={styles.dirty} role="img" aria-label="Unsaved changes" /> : null}
      <button type="button" tabIndex={-1} aria-label={`Close ${name}`} className={styles.close} onClick={onClose} onPointerDown={(event) => event.stopPropagation()}>
        <span aria-hidden="true">×</span>
      </button>
    </div>
  )
}
