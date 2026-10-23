import { closestCenter, DndContext } from '@dnd-kit/core'
import { SortableContext, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useEffect, useRef } from 'react'
import type { RunPlanRow } from '@/platform/mappers/runner'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useRunOrderReorder } from '@/features/runner/hooks/use-run-order-reorder'
import { chooseAlias, moveRow, resetRows, selectedCount, selectionState, setAllSelected, toggleRow } from '@/features/runner/lib/run-order'
import { useRunnerStore } from '@/features/runner/store'
import { useSidebarStore } from '@/features/sidebar/store'
import styles from './RunOrderList.module.css'

const NO_ROWS: RunPlanRow[] = []

const EMPTY = (
  <EmptyState.Root label="No requests to run" size="compact">
    <EmptyState.Title>No Requests to Run</EmptyState.Title>
    <EmptyState.Description>Add requests to this collection or folder to run them in order.</EmptyState.Description>
  </EmptyState.Root>
)

const HANDLE_ICON = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.handleIcon}>
    <circle cx="6" cy="4" r="1" />
    <circle cx="10" cy="4" r="1" />
    <circle cx="6" cy="8" r="1" />
    <circle cx="10" cy="8" r="1" />
    <circle cx="6" cy="12" r="1" />
    <circle cx="10" cy="12" r="1" />
  </svg>
)

interface RunOrderListProps {
  itemId: string
}

/** The run order of the Setup sub-tab: checks, select all, Reset, alias choice and drag or keyboard reordering. */
export function RunOrderList({ itemId }: RunOrderListProps) {
  const rows = useRunnerStore((state) => state.targets[itemId]?.rows ?? NO_ROWS)
  const setRows = useRunnerStore((state) => state.setRows)
  const reorder = useRunOrderReorder(rows, (fromId, toId) => setRows(itemId, moveRow(rows, fromId, toId)))
  const selected = selectedCount(rows)
  const all = selectionState(rows)
  const selectAllRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate = all === 'mixed'
    }
  }, [all])

  if (rows.length === 0) {
    return <section className={styles.order}>{EMPTY}</section>
  }

  function reset() {
    setRows(itemId, resetRows(rows, useSidebarStore.getState().tree, itemId))
  }

  return (
    <section className={styles.order} aria-label="Run order">
      <header className={styles.header}>
        <label className={styles.selectAll}>
          <input
            ref={selectAllRef}
            type="checkbox"
            aria-label="Select all requests"
            aria-checked={all === 'mixed' ? 'mixed' : undefined}
            checked={all === true}
            onChange={() => setRows(itemId, setAllSelected(rows, all !== true))}
          />
        </label>
        <span className={styles.counter} aria-live="polite">
          {selected} of {rows.length} selected
        </span>
        <Button variant="ghost" size="sm" className={styles.reset} onClick={reset}>
          Reset
        </Button>
      </header>
      <DndContext
        sensors={reorder.sensors}
        collisionDetection={closestCenter}
        accessibility={reorder.accessibility}
        onDragStart={reorder.onDragStart}
        onDragEnd={reorder.onDragEnd}
        onDragCancel={reorder.onDragCancel}
      >
        <SortableContext items={rows.map((row) => row.requestId)} strategy={verticalListSortingStrategy}>
          <ol className={styles.list} aria-label="Requests in run order">
            {rows.map((row) => (
              <RunOrderRow
                key={row.requestId}
                row={row}
                onToggle={(checked) => setRows(itemId, toggleRow(rows, row.requestId, checked))}
                onAlias={(aliasId) => setRows(itemId, chooseAlias(rows, row.requestId, aliasId))}
              />
            ))}
          </ol>
        </SortableContext>
      </DndContext>
    </section>
  )
}

interface RunOrderRowProps {
  row: RunPlanRow
  onToggle: (checked: boolean) => void
  onAlias: (aliasId: string) => void
}

function RunOrderRow({ row, onToggle, onAlias }: RunOrderRowProps) {
  const { setNodeRef, attributes, listeners, transform, transition, isDragging } = useSortable({ id: row.requestId })
  const path = row.folderPath.join(' / ')
  return (
    <li
      ref={setNodeRef}
      className={cn(styles.row, !row.selected && styles.unselected, isDragging && styles.dragging)}
      style={{ transform: transform ? `translate3d(0, ${transform.y}px, 0)` : undefined, transition }}
    >
      <button type="button" className={styles.handle} {...attributes} {...listeners} aria-label={`Reorder ${row.name}`}>
        {HANDLE_ICON}
      </button>
      <input type="checkbox" aria-label={`Run ${row.name}`} checked={row.selected} onChange={(event) => onToggle(event.target.checked)} />
      <span className={styles.number}>{row.number}</span>
      <MethodBadge method={row.method} className={styles.method} />
      <span className={styles.label}>
        {path ? (
          <span className={styles.path} translate="no">
            {path} /{' '}
          </span>
        ) : null}
        <span className={styles.name} translate="no">
          {row.name}
        </span>
      </span>
      {row.aliases.length > 0 ? (
        <select aria-label={`Version of ${row.name}`} className={styles.alias} value={row.aliasId} onChange={(event) => onAlias(event.target.value)}>
          <option value="">Default</option>
          {row.aliases.map((alias) => (
            <option key={alias.id} value={alias.id}>
              {alias.name}
            </option>
          ))}
        </select>
      ) : null}
    </li>
  )
}
