import { useMemo, type ReactNode } from 'react'
import type { ContainerKind } from '@/platform/mappers/tree'
import { useSidebarStore } from '@/features/sidebar/store'
import { breadcrumbTrail, trailTitle } from '@/features/sidebar/lib/breadcrumb'
import type { Crumb, OpenMode } from '@/features/sidebar/model'
import styles from './ItemBreadcrumb.module.css'

const SEPARATOR = (
  <span className={styles.separator} aria-hidden="true">
    /
  </span>
)

interface ItemBreadcrumbProps {
  /** The saved request or example the tab shows; null for a request that was never saved. */
  itemId: string | null
  /** The breadcrumb's tooltip when the tree does not have the item (an unsaved request). */
  name: string
  /** The open item's own name, last in the trail: the tab's editable name, which the tab saves. */
  nameEditor: ReactNode
  onOpenItem: (id: string, kind: ContainerKind, mode: OpenMode) => void
  onOpenRequest: (id: string) => void
}

/** Where the open request or example lives: its collection and folders, then its own (editable) name. */
export function ItemBreadcrumb({ itemId, name, nameEditor, onOpenItem, onOpenRequest }: ItemBreadcrumbProps) {
  const tree = useSidebarStore((state) => state.tree)
  const trail = useMemo(() => (itemId === null ? [] : breadcrumbTrail(tree, itemId)), [tree, itemId])
  const ancestors = trail.slice(0, -1)

  function openAncestor(crumb: Crumb) {
    if (crumb.kind === 'collection' || crumb.kind === 'folder') {
      onOpenItem(crumb.id, crumb.kind, 'preview')
      return
    }
    onOpenRequest(crumb.id)
  }

  return (
    <nav aria-label="Breadcrumb" className={styles.breadcrumb} title={trail.length > 0 ? trailTitle(trail) : name}>
      <ol className={styles.trail}>
        {ancestors.map((crumb) => (
          <li key={crumb.id} className={styles.ancestor}>
            <button type="button" className={styles.link} onClick={() => openAncestor(crumb)}>
              {crumb.name}
            </button>
            {SEPARATOR}
          </li>
        ))}
        <li className={styles.current} aria-current="page">
          {nameEditor}
        </li>
      </ol>
    </nav>
  )
}
