import { useState } from 'react'
import type { TreeNode } from '@/platform/mappers/tree'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useSidebarStore } from '../store'
import { CollectionTree } from './CollectionTree'
import { DeleteItemDialog } from './DeleteItemDialog'
import { ItemSettingsDialog } from './ItemSettingsDialog'
import { SidebarToolbar } from './SidebarToolbar'
import styles from './SidebarPanel.module.css'

interface SidebarPanelProps {
  onOpenRequest: (id: string) => void
  onRequestsDeleted: (ids: string[]) => void
  onRenamed: (id: string) => void
}

export function SidebarPanel({ onOpenRequest, onRequestsDeleted, onRenamed }: SidebarPanelProps) {
  const error = useSidebarStore((state) => state.error)
  const clearError = useSidebarStore((state) => state.clearError)
  const remove = useSidebarStore((state) => state.remove)
  const [settingsId, setSettingsId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<TreeNode | null>(null)

  async function confirmDelete(item: TreeNode) {
    setDeleting(null)
    const deleted = await remove(item.id)
    if (deleted.length > 0) {
      onRequestsDeleted(deleted)
    }
  }

  return (
    <div className={styles.panel}>
      <SidebarToolbar />
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>{error.description || error.code}</Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <CollectionTree onOpenRequest={onOpenRequest} onOpenSettings={setSettingsId} onDelete={setDeleting} onRenamed={onRenamed} />
      <ItemSettingsDialog itemId={settingsId} onClose={() => setSettingsId(null)} />
      <DeleteItemDialog item={deleting} onCancel={() => setDeleting(null)} onConfirm={(item) => void confirmDelete(item)} />
    </div>
  )
}
