import { useDraggable, useDroppable } from '@dnd-kit/core'
import type { CSSProperties, KeyboardEvent, MouseEvent } from 'react'
import { cn } from '@/shared/lib/cn'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import type { ContainerKind, TreeNode } from '@/platform/mappers/tree'
import { useExportStore } from '@/features/export/store'
import { isDraggable } from '@/features/sidebar/lib/tree-drop'
import { isExpandable, type TreeRow } from '@/features/sidebar/lib/tree-rows'
import type { OpenMode } from '@/features/sidebar/model'
import { useSidebarStore } from '@/features/sidebar/store'
import { RenameInput } from './RenameInput'
import styles from './TreeItem.module.css'

export interface TreeItemActions {
  onOpenRequest: (id: string) => void
  onPinRequest: (id: string) => void
  onOpenExample: (requestId: string, exampleId: string) => void
  onPinExample: (requestId: string, exampleId: string) => void
  /** Opens a collection or folder in a main-area tab (a single click previews, a double click pins). */
  onOpenItem: (id: string, kind: ContainerKind, mode: OpenMode) => void
  onDelete: (node: TreeNode) => void
  /** A row was renamed; `previousName` lets open tabs tell a pending name of their own from the old one. */
  onRenamed: (id: string, previousName: string) => void
}

const ERROR_MARK = (
  <span aria-hidden="true" className={styles.errorMark}>
    !
  </span>
)

const PLUS_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.actionIcon}>
    <path d="M5 12h14" />
    <path d="M12 5v14" />
  </svg>
)

const MORE_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.actionIcon}>
    <circle cx="5" cy="12" r="1" />
    <circle cx="12" cy="12" r="1" />
    <circle cx="19" cy="12" r="1" />
  </svg>
)

const DUPLICATE_ID_WARNING = <span className={styles.warning}>Duplicate ID</span>

function exportNode(node: TreeNode) {
  if (node.kind === 'collection' || node.kind === 'folder' || node.kind === 'request') {
    void useExportStore.getState().openItem(node.kind, node.id, node.name)
  }
}

interface TreeItemProps extends TreeItemActions {
  row: TreeRow
  focusable: boolean
}

// Points right when closed and down when open, like Postman; it rotates instead of swapping glyphs.
const CHEVRON = (
  <svg viewBox="0 0 16 16" className={styles.chevronIcon}>
    <path d="M6 4l4 4-4 4" />
  </svg>
)

export function TreeItem({ row, focusable, onOpenRequest, onPinRequest, onOpenExample, onPinExample, onOpenItem, onDelete, onRenamed }: TreeItemProps) {
  const { node, depth, expanded, code, requestId } = row
  const selected = useSidebarStore((state) => state.selectedId === node.id)
  const renaming = useSidebarStore((state) => state.renamingId === node.id)
  const select = useSidebarStore((state) => state.select)
  const toggle = useSidebarStore((state) => state.toggle)
  const startRename = useSidebarStore((state) => state.startRename)
  const cancelRename = useSidebarStore((state) => state.cancelRename)
  const rename = useSidebarStore((state) => state.rename)
  const duplicate = useSidebarStore((state) => state.duplicate)
  const renameExample = useSidebarStore((state) => state.renameExample)
  const duplicateExample = useSidebarStore((state) => state.duplicateExample)
  const createFolder = useSidebarStore((state) => state.createFolder)
  const createRequest = useSidebarStore((state) => state.createRequest)
  const dropPosition = useSidebarStore((state) => (state.dropTarget?.id === node.id ? state.dropTarget.position : null))
  const dragging = useSidebarStore((state) => state.draggingId === node.id)
  const draggable = useDraggable({ id: node.id, disabled: renaming || !isDraggable(node) })
  const droppable = useDroppable({ id: node.id, disabled: node.kind === 'example' })
  const expandable = isExpandable(node)
  const isExample = node.kind === 'example'
  const containerKind: ContainerKind | null = node.kind === 'collection' || node.kind === 'folder' ? node.kind : null

  function activate() {
    select(node.id)
    if (containerKind) {
      if (!expanded) {
        toggle(node.id)
      }
      onOpenItem(node.id, containerKind, 'preview')
    } else if (node.kind === 'request') {
      onOpenRequest(node.id)
    } else if (isExample && requestId) {
      onOpenExample(requestId, node.id)
    }
  }

  function pinItem() {
    if (containerKind) {
      onOpenItem(node.id, containerKind, 'pinned')
    } else if (node.kind === 'request') {
      onPinRequest(node.id)
    } else if (isExample && requestId) {
      onPinExample(requestId, node.id)
    }
  }

  // The chevron only expands and collapses; the rest of a collection or folder row opens its tab.
  function onChevronClick(event: MouseEvent<HTMLSpanElement>) {
    if (expandable) {
      event.stopPropagation()
      select(node.id)
      toggle(node.id)
    }
  }

  function onChevronDoubleClick(event: MouseEvent<HTMLSpanElement>) {
    if (expandable) {
      event.stopPropagation()
    }
  }

  async function commitRename(name: string) {
    const previousName = node.name
    const renamed = isExample ? await renameExample(node.id, name) : await rename(node.id, name)
    if (!renamed.ok) {
      return renamed.error
    }
    if (renamed.value) {
      onRenamed(node.id, previousName)
    }
    return null
  }

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.target !== event.currentTarget) {
      return
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      activate()
    }
    if (event.key === 'F2' && node.kind !== 'error') {
      event.preventDefault()
      startRename(node.id)
    }
  }

  // The right-click menu has everything; the row's + button only the adds and its … button the rest.
  const openItem = containerKind ? <ContextMenu.Item onSelect={() => onOpenItem(node.id, containerKind, 'pinned')}>Open</ContextMenu.Item> : null
  const addItems = containerKind ? (
    <>
      <ContextMenu.Item onSelect={() => void createRequest(node.id)}>Add Request</ContextMenu.Item>
      <ContextMenu.Item onSelect={() => void createFolder(node.id)}>Add Folder</ContextMenu.Item>
    </>
  ) : null
  const restItems = (
    <>
      <ContextMenu.Item onSelect={() => startRename(node.id)}>Rename</ContextMenu.Item>
      <ContextMenu.Item onSelect={() => void (isExample ? duplicateExample(node.id) : duplicate(node.id))}>Duplicate</ContextMenu.Item>
      {isExample ? null : <ContextMenu.Item onSelect={() => exportNode(node)}>Export…</ContextMenu.Item>}
      <ContextMenu.Item tone="danger" onSelect={() => onDelete(node)}>
        Delete
      </ContextMenu.Item>
    </>
  )

  return (
    <ContextMenu.Root>
      <ContextMenu.Trigger>
        <div
          ref={(element) => {
            draggable.setNodeRef(element)
            droppable.setNodeRef(element)
          }}
          {...draggable.listeners}
          aria-describedby={draggable.attributes['aria-describedby']}
          role="treeitem"
          data-tree-id={node.id}
          data-drop={dropPosition ?? undefined}
          aria-level={depth + 1}
          aria-expanded={expandable ? expanded : undefined}
          aria-selected={selected}
          tabIndex={focusable ? 0 : -1}
          title={node.error ?? node.warning ?? undefined}
          className={cn(
            styles.row,
            selected && styles.selected,
            node.kind === 'error' && styles.error,
            dragging && styles.dragging,
            dropPosition && styles[dropPosition],
          )}
          style={{ '--depth': depth } as CSSProperties}
          onClick={activate}
          onDoubleClick={pinItem}
          onKeyDown={onKeyDown}
        >
          <span aria-hidden="true" className={cn(styles.chevron, expanded && styles.open)} onClick={onChevronClick} onDoubleClick={onChevronDoubleClick}>
            {expandable ? CHEVRON : null}
          </span>
          {node.kind === 'request' ? <MethodBadge method={node.method ?? 'GET'} className={styles.method} /> : null}
          {isExample ? (
            <span className={styles.code}>
              <span className="visually-hidden">Status </span>
              {code}
            </span>
          ) : null}
          {node.kind === 'error' ? ERROR_MARK : null}
          {renaming ? (
            <RenameInput name={node.name} onCommit={commitRename} onCancel={cancelRename} />
          ) : (
            <span className={styles.name}>{node.name}</span>
          )}
          {node.warning ? DUPLICATE_ID_WARNING : null}
          {node.kind === 'error' || renaming ? null : (
            <span className={styles.rowActions}>
              {addItems ? (
                <ContextMenu.Root>
                  <ContextMenu.Button label={`Add to ${node.name}`} title="Add request or folder" className={styles.rowAction} tabIndex={-1}>
                    {PLUS_ICON}
                  </ContextMenu.Button>
                  <ContextMenu.Menu label={`Add to ${node.name}`}>{addItems}</ContextMenu.Menu>
                </ContextMenu.Root>
              ) : null}
              <ContextMenu.Root>
                <ContextMenu.Button label={`More actions for ${node.name}`} title="More actions" className={styles.rowAction} tabIndex={-1}>
                  {MORE_ICON}
                </ContextMenu.Button>
                <ContextMenu.Menu label={`More actions for ${node.name}`}>
                  {openItem}
                  {restItems}
                </ContextMenu.Menu>
              </ContextMenu.Root>
            </span>
          )}
        </div>
      </ContextMenu.Trigger>
      {node.kind === 'error' ? null : (
        <ContextMenu.Menu label={`${node.name} actions`}>
          {openItem}
          {addItems}
          {restItems}
        </ContextMenu.Menu>
      )}
    </ContextMenu.Root>
  )
}
