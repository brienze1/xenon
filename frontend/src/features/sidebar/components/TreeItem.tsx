import type { CSSProperties, KeyboardEvent } from 'react'
import { cn } from '@/shared/lib/cn'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import type { TreeNode } from '@/platform/mappers/tree'
import { isContainer, type TreeRow } from '@/features/sidebar/lib/tree-rows'
import { useSidebarStore } from '@/features/sidebar/store'
import { RenameInput } from './RenameInput'
import styles from './TreeItem.module.css'

export interface TreeItemActions {
  onOpenRequest: (id: string) => void
  onPinRequest: (id: string) => void
  onOpenSettings: (id: string) => void
  onDelete: (node: TreeNode) => void
  onRenamed: (id: string) => void
}

interface TreeItemProps extends TreeItemActions {
  row: TreeRow
  focusable: boolean
}

export function TreeItem({ row, focusable, onOpenRequest, onPinRequest, onOpenSettings, onDelete, onRenamed }: TreeItemProps) {
  const { node, depth, expanded } = row
  const selected = useSidebarStore((state) => state.selectedId === node.id)
  const renaming = useSidebarStore((state) => state.renamingId === node.id)
  const select = useSidebarStore((state) => state.select)
  const toggle = useSidebarStore((state) => state.toggle)
  const startRename = useSidebarStore((state) => state.startRename)
  const cancelRename = useSidebarStore((state) => state.cancelRename)
  const rename = useSidebarStore((state) => state.rename)
  const duplicate = useSidebarStore((state) => state.duplicate)
  const createFolder = useSidebarStore((state) => state.createFolder)
  const createRequest = useSidebarStore((state) => state.createRequest)
  const container = isContainer(node)

  function activate() {
    select(node.id)
    if (container) {
      toggle(node.id)
    } else if (node.kind === 'request') {
      onOpenRequest(node.id)
    }
  }

  async function commitRename(name: string) {
    if (await rename(node.id, name)) {
      onRenamed(node.id)
    }
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
          aria-expanded={container ? expanded : undefined}
          aria-selected={selected}
          tabIndex={focusable ? 0 : -1}
          title={node.error ?? node.warning ?? undefined}
          className={cn(styles.row, selected && styles.selected, node.kind === 'error' && styles.error)}
          style={{ '--depth': depth } as CSSProperties}
          onClick={activate}
          onDoubleClick={() => (node.kind === 'request' ? onPinRequest(node.id) : undefined)}
          onKeyDown={onKeyDown}
        >
          <span aria-hidden="true" className={styles.chevron}>
            {container ? (expanded ? '▾' : '▸') : ''}
          </span>
          {node.kind === 'request' ? <MethodBadge method={node.method ?? 'GET'} className={styles.method} /> : null}
          {node.kind === 'error' ? (
            <span aria-hidden="true" className={styles.errorMark}>
              !
            </span>
          ) : null}
          {renaming ? (
            <RenameInput name={node.name} onCommit={(name) => void commitRename(name)} onCancel={cancelRename} />
          ) : (
            <span className={styles.name}>{node.name}</span>
          )}
          {node.warning ? <span className={styles.warning}>duplicate id</span> : null}
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
          <ContextMenu.Item onSelect={() => void duplicate(node.id)}>Duplicate</ContextMenu.Item>
          <ContextMenu.Item tone="danger" onSelect={() => onDelete(node)}>
            Delete
          </ContextMenu.Item>
        </ContextMenu.Menu>
      )}
    </ContextMenu.Root>
  )
}
