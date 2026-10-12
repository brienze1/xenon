import { EnvironmentTab } from '@/features/environments/components/EnvironmentTab'
import { GlobalsTab } from '@/features/environments/components/GlobalsTab'
import { VaultBrokenNotice } from '@/features/environments/components/VaultBrokenNotice'
import { SaveAsExampleDialog } from '@/features/examples/components/SaveAsExampleDialog'
import { CloseTabPrompt } from '@/features/tabs/components/CloseTabPrompt'
import { SaveRequestDialog } from '@/features/tabs/components/SaveRequestDialog'
import { TabNotice } from '@/features/tabs/components/TabNotice'
import { TabStrip } from '@/features/tabs/components/TabStrip'
import { useTabsStore } from '@/features/tabs/store'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { PanelPlusArt } from '@/shared/ui/empty-state/illustrations/PanelPlusArt'
import { Button } from '@/shared/ui/button/Button'
import { activeTabKind } from '@/app/active-tab'
import { ExampleWorkspace } from './ExampleWorkspace'
import { ItemWorkspace } from './ItemWorkspace'
import { RequestWorkspace } from './RequestWorkspace'
import styles from './TabsArea.module.css'

const TAB_STRIP = <TabStrip />

const VAULT_BROKEN_NOTICE = <VaultBrokenNotice />

function openCreatedExample(requestId: string, exampleId: string) {
  void useTabsStore.getState().openExamplePinned(requestId, exampleId)
}

const TAB_DIALOGS = (
  <>
    <CloseTabPrompt />
    <SaveRequestDialog />
    <SaveAsExampleDialog onCreated={openCreatedExample} />
  </>
)

const NO_TABS_INTRO = (
  <>
    <EmptyState.Art>
      <PanelPlusArt />
    </EmptyState.Art>
    <EmptyState.Title>No Open Tabs</EmptyState.Title>
    <EmptyState.Description>Open a request from the sidebar or start a new one.</EmptyState.Description>
  </>
)

export function TabsArea() {
  const activeTabId = useTabsStore((state) => state.activeTabId)
  const save = useTabsStore((state) => state.save)
  const newTab = useTabsStore((state) => state.newTab)
  const kind = useTabsStore(activeTabKind)
  const environmentId = useTabsStore((state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.environmentId ?? null)
  const itemId = useTabsStore((state) => state.tabs.find((tab) => tab.id === state.activeTabId)?.itemId ?? null)

  return (
    <div className={styles.area}>
      {VAULT_BROKEN_NOTICE}
      {TAB_STRIP}
      {activeTabId ? (
        <>
          <TabNotice tabId={activeTabId} />
          <div className={styles.workspace}>
            {kind === 'globals' ? (
              <GlobalsTab key={activeTabId} />
            ) : (kind === 'collection' || kind === 'folder') && itemId ? (
              <ItemWorkspace key={`${activeTabId}:${itemId}`} itemId={itemId} kind={kind} onSave={() => void save(activeTabId)} />
            ) : kind === 'environment' && environmentId ? (
              <EnvironmentTab key={activeTabId} environmentId={environmentId} />
            ) : kind === 'example' ? (
              <ExampleWorkspace key={activeTabId} tabId={activeTabId} onSave={() => void save(activeTabId)} />
            ) : (
              <RequestWorkspace key={activeTabId} tabId={activeTabId} onSave={() => void save(activeTabId)} />
            )}
          </div>
        </>
      ) : (
        <EmptyState.Root label="No open tabs">
          {NO_TABS_INTRO}
          <EmptyState.Actions>
            <Button variant="primary" onClick={() => newTab()}>
              New Tab
            </Button>
          </EmptyState.Actions>
        </EmptyState.Root>
      )}
      {TAB_DIALOGS}
    </div>
  )
}
