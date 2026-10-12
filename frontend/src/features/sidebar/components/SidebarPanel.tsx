import { Activity, useState, type ReactNode } from 'react'
import type { RequestDraft } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import type { OpenMode, VariablesItem } from '@/features/sidebar/model'
import { useSidebarStore } from '@/features/sidebar/store'
import { CollectionTree } from './CollectionTree'
import { DeleteItemDialog } from './DeleteItemDialog'
import { EnvironmentsView } from './EnvironmentsView'
import { HistoryView } from './HistoryView'
import { ItemSettingsDialog } from './ItemSettingsDialog'
import type { ItemAuthSlot, ItemScriptsSlot } from './SettingsForm'
import { SidebarErrorText } from './SidebarErrorText'
import { SidebarToolbar } from './SidebarToolbar'
import { SidebarViewSelect } from './SidebarViewSelect'
import styles from './SidebarPanel.module.css'

const SIDEBAR_NEXT_STEP = 'Try again, or check that the workspace folder exists and is writable.'

interface SidebarPanelProps {
  onOpenRequest: (id: string) => void
  onPinRequest: (id: string) => void
  onOpenExample: (requestId: string, exampleId: string) => void
  onPinExample: (requestId: string, exampleId: string) => void
  onExampleDeleted: (exampleId: string) => void
  onRequestsDeleted: (ids: string[]) => void
  onRenamed: (id: string) => void
  onOpenHistoryEntry: (request: RequestDraft) => void
  /** Opens the globals or an environment in a main-area tab. */
  onOpenVariables: (item: VariablesItem, mode: OpenMode) => void
  /** The environments-view entry whose tab is active. */
  selectedVariables: VariablesItem | null
  /** Renders a collection's Variables tab in its settings dialog. */
  collectionVariables?: (collectionId: string) => ReactNode
  /** Renders the scripts editor of a collection or folder's settings; composed by the app. */
  scriptsEditor?: (slot: ItemScriptsSlot) => ReactNode
  /** Renders the auth editor of a collection or folder; composed by the app. */
  authEditor?: (slot: ItemAuthSlot) => ReactNode
}

const SIDEBAR_VIEW_SELECT = <SidebarViewSelect />

const SIDEBAR_TOOLBAR = <SidebarToolbar />

interface SidebarSectionProps {
  visible: boolean
  label: string
  children: ReactNode
}

/** One sidebar view; Activity keeps a hidden view's state (scroll, expanded rows) for an instant switch back. */
function SidebarSection({ visible, label, children }: SidebarSectionProps) {
  return (
    <Activity mode={visible ? 'visible' : 'hidden'}>
      <section aria-label={label} hidden={!visible} className={styles.view}>
        {children}
      </section>
    </Activity>
  )
}

export function SidebarPanel({
  onOpenRequest,
  onPinRequest,
  onOpenExample,
  onPinExample,
  onExampleDeleted,
  onRequestsDeleted,
  onRenamed,
  onOpenHistoryEntry,
  onOpenVariables,
  selectedVariables,
  collectionVariables,
  scriptsEditor,
  authEditor,
}: SidebarPanelProps) {
  const view = useSidebarStore((state) => state.view)
  const openHistoryEntry = useSidebarStore((state) => state.openHistoryEntry)
  const error = useSidebarStore((state) => state.error)
  const clearError = useSidebarStore((state) => state.clearError)
  const remove = useSidebarStore((state) => state.remove)
  const removeExample = useSidebarStore((state) => state.removeExample)
  const [settingsId, setSettingsId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<TreeNode | null>(null)

  async function confirmDelete(item: TreeNode) {
    if (item.kind === 'example') {
      const failure = await removeExample(item.id)
      if (failure === null) {
        setDeleting(null)
        onExampleDeleted(item.id)
      }
      return failure
    }
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
    <div className={styles.panel}>
      {SIDEBAR_VIEW_SELECT}
      {view === 'environments' ? null : SIDEBAR_TOOLBAR}
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
      <SidebarSection visible={view === 'collections'} label="Collections">
        <CollectionTree
          onOpenRequest={onOpenRequest}
          onPinRequest={onPinRequest}
          onOpenExample={onOpenExample}
          onPinExample={onPinExample}
          onOpenSettings={setSettingsId}
          onDelete={setDeleting}
          onRenamed={onRenamed}
        />
      </SidebarSection>
      <SidebarSection visible={view === 'environments'} label="Environments">
        <EnvironmentsView selected={selectedVariables} onOpen={onOpenVariables} />
      </SidebarSection>
      <SidebarSection visible={view === 'history'} label="History">
        <HistoryView onOpenEntry={(id) => void openEntry(id)} />
      </SidebarSection>
      <ItemSettingsDialog itemId={settingsId} onClose={() => setSettingsId(null)} collectionVariables={collectionVariables} scriptsEditor={scriptsEditor} authEditor={authEditor} />
      <DeleteItemDialog item={deleting} onCancel={() => setDeleting(null)} onConfirm={confirmDelete} />
    </div>
  )
}
