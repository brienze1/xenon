import { CloseTabPrompt } from '@/features/tabs/components/CloseTabPrompt'
import { SaveRequestDialog } from '@/features/tabs/components/SaveRequestDialog'
import { TabNotice } from '@/features/tabs/components/TabNotice'
import { TabStrip } from '@/features/tabs/components/TabStrip'
import { useTabsStore } from '@/features/tabs/store'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { PanelPlusArt } from '@/shared/ui/empty-state/illustrations/PanelPlusArt'
import { Button } from '@/shared/ui/button/Button'
import { RequestWorkspace } from './RequestWorkspace'
import styles from './TabsArea.module.css'

const TAB_STRIP = <TabStrip />

const TAB_DIALOGS = (
  <>
    <CloseTabPrompt />
    <SaveRequestDialog />
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

  return (
    <div className={styles.area}>
      {TAB_STRIP}
      {activeTabId ? (
        <>
          <TabNotice tabId={activeTabId} />
          <div className={styles.workspace}>
            <RequestWorkspace key={activeTabId} tabId={activeTabId} onSave={() => void save(activeTabId)} />
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
