import { useVirtualizer } from '@tanstack/react-virtual'
import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react'
import { rovingIndex } from '@/shared/lib/roving-focus'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { StackArt } from '@/shared/ui/empty-state/illustrations/StackArt'
import { isExpandable, visibleRows, type TreeRow } from '@/features/sidebar/lib/tree-rows'
import { useSidebarStore } from '@/features/sidebar/store'
import styles from './CollectionTree.module.css'
import { TreeItem, type TreeItemActions } from './TreeItem'

export const TREE_ROW_HEIGHT = 28

// eslint-disable-next-line @typescript-eslint/no-empty-object-type -- the tree takes exactly the row actions it forwards
interface CollectionTreeProps extends TreeItemActions {}

function parentIndex(rows: TreeRow[], index: number): number {
  const depth = rows[index]?.depth ?? 0
  for (let candidate = index - 1; candidate >= 0; candidate--) {
    if ((rows[candidate]?.depth ?? 0) < depth) {
      return candidate
    }
  }
  return -1
}

const NO_COLLECTIONS_INTRO = (
  <>
    <EmptyState.Art>
      <StackArt />
    </EmptyState.Art>
    <EmptyState.Title>No Collections Yet</EmptyState.Title>
    <EmptyState.Description>Collections group your requests. Each one is a folder in your workspace.</EmptyState.Description>
  </>
)

const NO_MATCHES = <p role="status" className={styles.noMatches}>No requests match your search.</p>

export function CollectionTree(actions: CollectionTreeProps) {
  const tree = useSidebarStore((state) => state.tree)
  const expanded = useSidebarStore((state) => state.expanded)
  const search = useSidebarStore((state) => state.search)
  const status = useSidebarStore((state) => state.status)
  const selectedId = useSidebarStore((state) => state.selectedId)
  const createCollection = useSidebarStore((state) => state.createCollection)
  const toggle = useSidebarStore((state) => state.toggle)
  const rows = useMemo(() => visibleRows(tree, expanded, search), [tree, expanded, search])
  const [activeId, setActiveId] = useState<string | null>(null)
  const focusId = [activeId, selectedId].find((id) => rows.some((row) => row.node.id === id)) ?? rows[0]?.node.id ?? null
  const scrollRef = useRef<HTMLDivElement>(null)
  const pendingFocus = useRef<string | null>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => TREE_ROW_HEIGHT,
    overscan: 12,
  })

  useEffect(() => {
    const id = pendingFocus.current
    if (id === null || id !== focusId) {
      return
    }
    pendingFocus.current = null
    scrollRef.current?.querySelector<HTMLElement>(`[data-tree-id="${CSS.escape(id)}"]`)?.focus()
  })

  function focusRow(index: number) {
    const row = rows[index]
    if (!row) {
      return
    }
    pendingFocus.current = row.node.id
    setActiveId(row.node.id)
    virtualizer.scrollToIndex(index)
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const target = event.target as HTMLElement
    const id = target.getAttribute('role') === 'treeitem' ? target.dataset.treeId : undefined
    const index = id === undefined ? -1 : rows.findIndex((row) => row.node.id === id)
    const row = rows[index]
    if (!row) {
      return
    }
    const next = rovingIndex(rows.length, index, event.key, { wrap: false })
    if (next !== null) {
      event.preventDefault()
      focusRow(next)
      return
    }
    if (event.key === 'ArrowRight' && isExpandable(row.node)) {
      event.preventDefault()
      if (row.expanded) {
        focusRow(rows[index + 1]?.depth === row.depth + 1 ? index + 1 : index)
      } else {
        toggle(row.node.id)
      }
    }
    if (event.key === 'ArrowLeft') {
      event.preventDefault()
      if (isExpandable(row.node) && row.expanded) {
        toggle(row.node.id)
      } else {
        focusRow(parentIndex(rows, index))
      }
    }
  }

  if (status === 'ready' && tree.length === 0) {
    return (
      <EmptyState.Root label="No collections" className={styles.empty}>
        {NO_COLLECTIONS_INTRO}
        <EmptyState.Actions>
          <Button variant="primary" onClick={() => void createCollection()}>
            Create Collection
          </Button>
        </EmptyState.Actions>
      </EmptyState.Root>
    )
  }

  if (search && rows.length === 0) {
    return NO_MATCHES
  }

  return (
    <div ref={scrollRef} className={styles.scroll}>
      <div role="tree" aria-label="Collections" className={styles.tree} style={{ height: virtualizer.getTotalSize() }} onKeyDown={onKeyDown}>
        {virtualizer.getVirtualItems().map((item) => {
          const row = rows[item.index]
          return row ? (
            <div
              key={row.node.id}
              className={styles.slot}
              style={{ height: item.size, transform: `translateY(${item.start}px)` }}
            >
              <TreeItem row={row} focusable={row.node.id === focusId} {...actions} />
            </div>
          ) : null
        })}
      </div>
    </div>
  )
}
