import { useDndContext } from '@dnd-kit/core'
import { useSortable } from '@dnd-kit/sortable'
import type { KeyboardEvent, MouseEvent } from 'react'
import { isScopeDirty, useEnvironmentsStore } from '@/features/environments/store'
import { useExamplesStore } from '@/features/examples/store'
import { isItemDirty, useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { collectionScope } from '@/platform/mappers/variable'
import { cn } from '@/shared/lib/cn'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useTabTitle } from '@/features/tabs/hooks/use-tab-title'
import { environmentTabScope, useTabsStore } from '@/features/tabs/store'
import { TabContextMenu } from './TabContextMenu'
import styles from './TabItem.module.css'

const MIDDLE_BUTTON = 1

const GLOBALS_BADGE = <span className={styles.kindBadge}>GLB</span>

const COLLECTION_ICON = (
  <svg viewBox="0 0 16 16" className={styles.kindIcon} role="img" aria-label="Collection">
    <path d="M2.5 4.5h11M3.5 2.5h9M2.5 6.5h11v7h-11z" />
  </svg>
)

const FOLDER_ICON = (
  <svg viewBox="0 0 16 16" className={styles.kindIcon} role="img" aria-label="Folder">
    <path d="M2 4.5V12.5h12V5.5H7.5L6 3.5H2z" />
  </svg>
)

const DIRTY_DOT = <span className={styles.dirty} role="img" aria-label="Unsaved changes" />

const CLOSE_GLYPH = <span aria-hidden="true">×</span>

interface TabItemProps {
  tabId: string
}

export function TabItem({ tabId }: TabItemProps) {
  const active = useTabsStore((state) => state.activeTabId === tabId)
  const preview = useTabsStore((state) => state.previewTabId === tabId)
  const activate = useTabsStore((state) => state.activate)
  const pin = useTabsStore((state) => state.pin)
  const requestClose = useTabsStore((state) => state.requestClose)
  const name = useTabTitle(tabId)
  const kind = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.kind)
  const scope = useTabsStore((state) => environmentTabScope(state.tabs.find((tab) => tab.id === tabId)))
  const method = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.method ?? 'GET')
  const draftDirty = useRequestBuilderStore((state) => state.tabs[tabId]?.dirty === true)
  const responseDirty = useExamplesStore((state) => state.tabs[tabId]?.dirty === true)
  const variablesDirty = useEnvironmentsStore((state) => scope !== null && isScopeDirty(state, scope))
  const environmentId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.environmentId ?? null)
  const itemId = useTabsStore((state) => state.tabs.find((tab) => tab.id === tabId)?.itemId ?? null)
  const settingsDirty = useItemSettingsStore((state) => itemId !== null && isItemDirty(state, itemId))
  const collectionVariablesDirty = useEnvironmentsStore((state) => kind === 'collection' && itemId !== null && isScopeDirty(state, collectionScope(itemId)))
  const environmentColor = useEnvironmentsStore(
    (state) => state.environments.find((environment) => environment.id === environmentId)?.color ?? 'none',
  )
  const dirty = draftDirty || responseDirty || variablesDirty || settingsDirty || collectionVariablesDirty
  const { setNodeRef, attributes, listeners, transform, transition, isDragging } = useSortable({ id: tabId, attributes: { role: 'tab' } })
  // Only a pointer drag makes the moving tab inert: during a keyboard drag the tab must keep
  // focus so the arrow keys keep moving it and assistive tech keeps announcing it.
  const { activatorEvent } = useDndContext()
  const pointerDragging = isDragging && activatorEvent !== null && activatorEvent.type.startsWith('pointer')

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
    <TabContextMenu tabId={tabId} name={name}>
      <div
        ref={setNodeRef}
        {...attributes}
        {...listeners}
        role="tab"
        aria-selected={active}
        tabIndex={active ? 0 : -1}
        inert={pointerDragging}
        title={preview ? `${name} (preview)` : name}
        className={cn(styles.tab, active && styles.active, preview && styles.preview, isDragging && styles.dragging)}
        style={{ transform: transform ? `translate3d(${transform.x}px, 0, 0)` : undefined, transition }}
        onClick={() => activate(tabId)}
        onDoubleClick={() => pin(tabId)}
        onAuxClick={onAuxClick}
        onKeyDown={onKeyDown}
      >
        {kind === 'globals' ? (
          GLOBALS_BADGE
        ) : kind === 'collection' ? (
          COLLECTION_ICON
        ) : kind === 'folder' ? (
          FOLDER_ICON
        ) : kind === 'environment' ? (
          <span className={styles.kindBadge} data-color={environmentColor}>
            ENV
          </span>
        ) : (
          <MethodBadge method={method} className={styles.method} />
        )}
        <span className={styles.name}>{name}</span>
        {dirty ? DIRTY_DOT : null}
        <button type="button" tabIndex={-1} aria-label={`Close ${name}`} className={styles.close} onClick={onClose} onPointerDown={(event) => event.stopPropagation()}>
          {CLOSE_GLYPH}
        </button>
      </div>
    </TabContextMenu>
  )
}
