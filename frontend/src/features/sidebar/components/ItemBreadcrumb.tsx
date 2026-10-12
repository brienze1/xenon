import { useMemo, useState } from 'react'
import type { ContainerKind } from '@/platform/mappers/tree'
import { useSidebarStore } from '@/features/sidebar/store'
import { breadcrumbTrail, trailTitle } from '@/features/sidebar/lib/breadcrumb'
import type { Crumb, OpenMode } from '@/features/sidebar/model'
import { RenameInput } from './RenameInput'
import styles from './ItemBreadcrumb.module.css'

const SEPARATOR = (
  <span className={styles.separator} aria-hidden="true">
    /
  </span>
)

interface ItemBreadcrumbProps {
  /** The saved request or example the tab shows; null for a request that was never saved. */
  itemId: string | null
  /** The name shown alone when the tree does not have the item (an unsaved request). */
  name: string
  onOpenItem: (id: string, kind: ContainerKind, mode: OpenMode) => void
  onOpenRequest: (id: string) => void
  onRenamed: (id: string) => void
}

/** Where the open request or example lives: its collection and folders, then its own name, renamed in place. */
export function ItemBreadcrumb({ itemId, name, onOpenItem, onOpenRequest, onRenamed }: ItemBreadcrumbProps) {
  const tree = useSidebarStore((state) => state.tree)
  const rename = useSidebarStore((state) => state.rename)
  const renameExample = useSidebarStore((state) => state.renameExample)
  const [editing, setEditing] = useState(false)
  const trail = useMemo(() => (itemId === null ? [] : breadcrumbTrail(tree, itemId)), [tree, itemId])
  const current = trail.at(-1) ?? null
  const ancestors = trail.slice(0, -1)

  function openAncestor(crumb: Crumb) {
    if (crumb.kind === 'collection' || crumb.kind === 'folder') {
      onOpenItem(crumb.id, crumb.kind, 'preview')
      return
    }
    onOpenRequest(crumb.id)
  }

  async function commitRename(next: string) {
    if (!current) {
      return null
    }
    const renamed = current.kind === 'example' ? await renameExample(current.id, next) : await rename(current.id, next)
    if (!renamed.ok) {
      return renamed.error
    }
    setEditing(false)
    if (renamed.value) {
      onRenamed(current.id)
    }
    return null
  }

  return (
    <nav aria-label="Breadcrumb" className={styles.breadcrumb} title={current ? trailTitle(trail) : name}>
      <ol className={styles.trail}>
        {ancestors.map((crumb) => (
          <li key={crumb.id} className={styles.ancestor}>
            <button type="button" className={styles.link} onClick={() => openAncestor(crumb)}>
              {crumb.name}
            </button>
            {SEPARATOR}
          </li>
        ))}
        <li className={editing ? styles.editing : styles.current}>
          {current === null ? (
            <span className={styles.name} aria-current="page">
              {name}
            </span>
          ) : editing ? (
            <RenameInput name={current.name} onCommit={commitRename} onCancel={() => setEditing(false)} />
          ) : (
            <button type="button" className={styles.rename} aria-current="page" title="Click to rename" onClick={() => setEditing(true)}>
              {current.name}
            </button>
          )}
        </li>
      </ol>
    </nav>
  )
}
