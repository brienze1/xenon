import type { CSSProperties, KeyboardEvent, MouseEvent } from 'react'
import { cn } from '@/shared/lib/cn'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import type { TreeNode } from '@/platform/mappers/tree'
import { isContainer, isExpandable, type TreeRow } from '@/features/sidebar/lib/tree-rows'
import { useSidebarStore } from '@/features/sidebar/store'
import { RenameInput } from './RenameInput'
import styles from './TreeItem.module.css'

export interface TreeItemActions {
  onOpenRequest: (id: string) => void
  onPinRequest: (id: string) => void
  onOpenExample: (requestId: string, exampleId: string) => void
  onPinExample: (requestId: string, exampleId: string) => void
  onOpenSettings: (id: string) => void
  onDelete: (node: TreeNode) => void
  onRenamed: (id: string) => void
}

const ERROR_MARK = (
  <span aria-hidden="true" className={styles.errorMark}>
    !
  </span>
)

const DUPLICATE_ID_WARNING = <span className={styles.warning}>Duplicate ID</span>

interface TreeItemProps extends TreeItemActions {
  row: TreeRow
  focusable: boolean
}

export function TreeItem({ row, focusable, onOpenRequest, onPinRequest, onOpenExample, onPinExample, onOpenSettings, onDelete, onRenamed }: TreeItemProps) {
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
  const container = isContainer(node)
  const expandable = isExpandable(node)
  const isExample = node.kind === 'example'

  function activate() {
    select(node.id)
    if (container) {
      toggle(node.id)
    } else if (node.kind === 'request') {
      onOpenRequest(node.id)
    } else if (isExample && requestId) {
      onOpenExample(requestId, node.id)
    }
  }

  function pinItem() {
    if (node.kind === 'request') {
      onPinRequest(node.id)
    } else if (isExample && requestId) {
      onPinExample(requestId, node.id)
    }
  }

  function onChevronClick(event: MouseEvent<HTMLSpanElement>) {
    if (expandable && !container) {
      event.stopPropagation()
      toggle(node.id)
    }
  }

  async function commitRename(name: string) {
    const renamed = isExample ? await renameExample(node.id, name) : await rename(node.id, name)
    if (!renamed.ok) {
      return renamed.error
    }
    if (renamed.value) {
      onRenamed(node.id)
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

  return (
    <ContextMenu.Root>
      <ContextMenu.Trigger>
        <div
          role="treeitem"
          data-tree-id={node.id}
          aria-level={depth + 1}
          aria-expanded={expandable ? expanded : undefined}
          aria-selected={selected}
          tabIndex={focusable ? 0 : -1}
          title={node.error ?? node.warning ?? undefined}
          className={cn(styles.row, selected && styles.selected, node.kind === 'error' && styles.error)}
          style={{ '--depth': depth } as CSSProperties}
          onClick={activate}
          onDoubleClick={pinItem}
          onKeyDown={onKeyDown}
        >
          <span aria-hidden="true" className={styles.chevron} onClick={onChevronClick}>
            {expandable ? (expanded ? '▾' : '▸') : ''}
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
        </div>
      </ContextMenu.Trigger>
      {node.kind === 'error' ? null : (
        <ContextMenu.Menu label={`${node.name} actions`}>
          {container ? (
            <>
              <ContextMenu.Item onSelect={() => void createRequest(node.id)}>Add Request</ContextMenu.Item>
              <ContextMenu.Item onSelect={() => void createFolder(node.id)}>Add Folder</ContextMenu.Item>
              <ContextMenu.Item onSelect={() => onOpenSettings(node.id)}>Settings…</ContextMenu.Item>
            </>
          ) : null}
          <ContextMenu.Item onSelect={() => startRename(node.id)}>Rename</ContextMenu.Item>
          <ContextMenu.Item onSelect={() => void (isExample ? duplicateExample(node.id) : duplicate(node.id))}>Duplicate</ContextMenu.Item>
          <ContextMenu.Item tone="danger" onSelect={() => onDelete(node)}>
            Delete
          </ContextMenu.Item>
        </ContextMenu.Menu>
      )}
    </ContextMenu.Root>
  )
}
