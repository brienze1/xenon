import { useVirtualizer } from '@tanstack/react-virtual'
import { useMemo, useRef } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { visibleRows } from '../lib/tree-rows'
import { useSidebarStore } from '../store'
import styles from './CollectionTree.module.css'
import { TreeItem, type TreeItemActions } from './TreeItem'

export const TREE_ROW_HEIGHT = 28

export function CollectionTree(actions: TreeItemActions) {
  const tree = useSidebarStore((state) => state.tree)
  const expanded = useSidebarStore((state) => state.expanded)
  const search = useSidebarStore((state) => state.search)
  const status = useSidebarStore((state) => state.status)
  const selectedId = useSidebarStore((state) => state.selectedId)
  const createCollection = useSidebarStore((state) => state.createCollection)
  const rows = useMemo(() => visibleRows(tree, expanded, search), [tree, expanded, search])
  const focusId = rows.some((row) => row.node.id === selectedId) ? selectedId : (rows[0]?.node.id ?? null)
  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => TREE_ROW_HEIGHT,
    overscan: 12,
  })

  if (status === 'ready' && tree.length === 0) {
    return (
      <EmptyState.Root label="No collections" className={styles.empty}>
        <EmptyState.Art variant="collections" />
        <EmptyState.Title>No collections yet</EmptyState.Title>
        <EmptyState.Description>Collections group your requests. Each one is a folder in your workspace.</EmptyState.Description>
        <EmptyState.Actions>
          <Button variant="primary" onClick={() => void createCollection()}>
            Create collection
          </Button>
        </EmptyState.Actions>
      </EmptyState.Root>
    )
  }

  if (search && rows.length === 0) {
    return <p className={styles.noMatches}>No requests match your search.</p>
  }

  return (
    <div ref={scrollRef} className={styles.scroll}>
      <div role="tree" aria-label="Collections" className={styles.tree} style={{ height: virtualizer.getTotalSize() }}>
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
