import { useState } from 'react'
import type { RequestDraft } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'
import type { SidebarView } from '@/platform/mappers/ui-state'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Tabs } from '@/shared/ui/tabs/Tabs'
import { useSidebarStore } from '@/features/sidebar/store'
import { CollectionTree } from './CollectionTree'
import { DeleteItemDialog } from './DeleteItemDialog'
import { HistoryView } from './HistoryView'
import { ItemSettingsDialog } from './ItemSettingsDialog'
import { SidebarErrorText } from './SidebarErrorText'
import { SidebarToolbar } from './SidebarToolbar'
import styles from './SidebarPanel.module.css'

const SIDEBAR_NEXT_STEP = 'Try again, or check that the workspace folder exists and is writable.'

interface SidebarPanelProps {
  onOpenRequest: (id: string) => void
  onPinRequest: (id: string) => void
  onRequestsDeleted: (ids: string[]) => void
  onRenamed: (id: string) => void
  onOpenHistoryEntry: (request: RequestDraft) => void
}

const SIDEBAR_VIEWS = (
  <Tabs.List label="Sidebar view" className={styles.views}>
    <Tabs.Tab value="collections">Collections</Tabs.Tab>
    <Tabs.Tab value="history">History</Tabs.Tab>
  </Tabs.List>
)

const SIDEBAR_TOOLBAR = <SidebarToolbar />

export function SidebarPanel({ onOpenRequest, onPinRequest, onRequestsDeleted, onRenamed, onOpenHistoryEntry }: SidebarPanelProps) {
  const view = useSidebarStore((state) => state.view)
  const setView = useSidebarStore((state) => state.setView)
  const openHistoryEntry = useSidebarStore((state) => state.openHistoryEntry)
  const error = useSidebarStore((state) => state.error)
  const clearError = useSidebarStore((state) => state.clearError)
  const remove = useSidebarStore((state) => state.remove)
  const [settingsId, setSettingsId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<TreeNode | null>(null)

  async function confirmDelete(item: TreeNode) {
    const deleted = await remove(item.id)
    if (!deleted.ok) {
      return deleted.error
    }
    setDeleting(null)
    if (deleted.value.length > 0) {
      onRequestsDeleted(deleted.value)
    }
    return null
  }

  async function openEntry(id: string) {
    const request = await openHistoryEntry(id)
    if (request) {
      onOpenHistoryEntry(request)
    }
  }

  return (
    <Tabs.Root value={view} onValueChange={(next) => setView(next as SidebarView)} className={styles.panel}>
      {SIDEBAR_VIEWS}
      {SIDEBAR_TOOLBAR}
      {error ? (
        <Banner.Root tone="error" className={styles.banner}>
          <Banner.Message>
            <SidebarErrorText error={error} nextStep={SIDEBAR_NEXT_STEP} />
          </Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss Error
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <Tabs.Panel value="collections" className={styles.view}>
        <CollectionTree onOpenRequest={onOpenRequest} onPinRequest={onPinRequest} onOpenSettings={setSettingsId} onDelete={setDeleting} onRenamed={onRenamed} />
      </Tabs.Panel>
      <Tabs.Panel value="history" className={styles.view}>
        <HistoryView onOpenEntry={(id) => void openEntry(id)} />
      </Tabs.Panel>
      <ItemSettingsDialog itemId={settingsId} onClose={() => setSettingsId(null)} />
      <DeleteItemDialog item={deleting} onCancel={() => setDeleting(null)} onConfirm={confirmDelete} />
    </Tabs.Root>
  )
}
