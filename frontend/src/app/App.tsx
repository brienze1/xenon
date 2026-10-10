import { useEffect, useMemo } from 'react'
import { ConsoleDrawer as ConsoleContent } from '@/features/console/components/ConsoleDrawer'
import { EnvironmentSwitcher } from '@/features/environments/components/EnvironmentSwitcher'
import { ImportDialog } from '@/features/import-export/components/ImportDialog'
import { useImportStore } from '@/features/import-export/store'
import { PassphraseDialog } from '@/features/environments/components/PassphraseDialog'
import { VariableTable } from '@/features/environments/components/VariableTable'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { SidePanel } from '@/features/side-panel/components/SidePanel'
import { useExamplesStore } from '@/features/examples/store'
import { SidebarPanel } from '@/features/sidebar/components/SidebarPanel'
import { useTabsStore } from '@/features/tabs/store'
import { MissingFolderNotice } from '@/features/workspaces/components/MissingFolderNotice'
import { WelcomeScreen } from '@/features/workspaces/components/WelcomeScreen'
import { WorkspaceSwitcher } from '@/features/workspaces/components/WorkspaceSwitcher'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { collectionScope } from '@/platform/mappers/variable'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import styles from './App.module.css'
import { bootstrap } from './bootstrap'
import { type HotkeyHandlers } from './hotkeys'
import { AppShell } from './layout/AppShell'
import { ConsoleDrawer } from './layout/regions/ConsoleDrawer'
import { Footer } from './layout/regions/Footer'
import { MainArea } from './layout/regions/MainArea'
import { RightPanel } from './layout/regions/RightPanel'
import { Sidebar } from './layout/regions/Sidebar'
import { TopBar } from './layout/regions/TopBar'
import { TabsArea } from './layout/TabsArea'
import { useAppHotkeys } from './use-app-hotkeys'

let didInit = false

const WORKSPACE_SWITCHER = <WorkspaceSwitcher />

const WELCOME_SCREEN = <WelcomeScreen />

const ENVIRONMENT_SWITCHER = <EnvironmentSwitcher />

const PASSPHRASE_DIALOG = <PassphraseDialog />

const IMPORT_DIALOG = <ImportDialog />

function collectionVariables(collectionId: string) {
  return <VariableTable scope={collectionScope(collectionId)} label="Collection variables" />
}

const CONSOLE = (
  <ConsoleDrawer>
    <ConsoleContent />
  </ConsoleDrawer>
)

const NO_REQUEST_SELECTED = (
  <EmptyState.Root label="No request selected" size="compact">
    <EmptyState.Description>Open a request to see its details here.</EmptyState.Description>
  </EmptyState.Root>
)

export function App() {
  const status = useWorkspacesStore((state) => state.status)
  const hasWorkspaces = useWorkspacesStore((state) => state.workspaces.length > 0)
  const missingLocalId = useWorkspacesStore((state) => state.missingLocalId)
  const layout = useWorkspacesStore((state) => state.layout)
  const setLayout = useWorkspacesStore((state) => state.setLayout)
  const active = useWorkspacesStore((state) =>
    state.workspaces.find((workspace) => workspace.localId === state.activeLocalId),
  )

  const hasActive = active !== undefined

  const activeTabId = useTabsStore((state) => state.activeTabId)
  const tabs = useTabsStore.getState
  const send = useRequestBuilderStore((state) => state.send)
  const adoptRename = useRequestBuilderStore((state) => state.adoptRename)
  const hotkeys = useMemo<HotkeyHandlers>(
    () =>
      hasActive
        ? {
            send: () => {
              const { activeTabId: tabId, tabs: openTabs } = useTabsStore.getState()
              const isExample = openTabs.some((tab) => tab.id === tabId && tab.kind === 'example')
              if (tabId && !isExample) {
                void send(tabId)
              }
            },
            save: () => void useTabsStore.getState().save(),
            saveAs: () => useTabsStore.getState().saveAs(),
            newTab: () => void useTabsStore.getState().newTab(),
            closeTab: () => useTabsStore.getState().requestClose(),
            environmentSwitcher: () => useEnvironmentsStore.getState().openSwitcher(),
            import: () => useImportStore.getState().open(),
          }
        : {},
    [hasActive, send],
  )

  useEffect(() => {
    if (didInit) {
      return
    }
    didInit = true
    bootstrap()
  }, [])
  useAppHotkeys(hotkeys)

  return (
    <AppShell layout={layout} onLayoutChange={setLayout}>
      <TopBar end={hasActive ? ENVIRONMENT_SWITCHER : null}>{hasWorkspaces ? WORKSPACE_SWITCHER : null}</TopBar>
      {active ? (
        <>
          <Sidebar>
            <SidebarPanel
              key={active.localId}
              onOpenRequest={(id) => void tabs().openPreview(id)}
              onPinRequest={(id) => void tabs().openPinned(id)}
              onOpenExample={(requestId, exampleId) => void tabs().openExamplePreview(requestId, exampleId)}
              onPinExample={(requestId, exampleId) => void tabs().openExamplePinned(requestId, exampleId)}
              onExampleDeleted={(exampleId) => tabs().handleExampleDeleted(exampleId)}
              onRequestsDeleted={(ids) => tabs().handleDeleted(ids)}
              onRenamed={(id) => {
                void adoptRename(id)
                void useExamplesStore.getState().adoptRename(id)
              }}
              onOpenHistoryEntry={(request) => void tabs().openDraft(request)}
              collectionVariables={collectionVariables}
            />
          </Sidebar>
          <MainArea>
            <TabsArea key={active.localId} />
          </MainArea>
          <RightPanel>
            {activeTabId ? (
              <SidePanel
                key={active.localId}
                tabId={activeTabId}
                view={layout.rightPanelView}
                onViewChange={(rightPanelView) => setLayout({ rightPanelView })}
              />
            ) : (
              NO_REQUEST_SELECTED
            )}
          </RightPanel>
        </>
      ) : (
        <MainArea>
          {missingLocalId ? <MissingFolderNotice localId={missingLocalId} /> : status === 'ready' ? WELCOME_SCREEN : null}
        </MainArea>
      )}
      {active ? CONSOLE : null}
      {PASSPHRASE_DIALOG}
      {active ? IMPORT_DIALOG : null}
      <Footer>
        {active ? (
          <span className={styles.path} translate="no">
            {active.path}
          </span>
        ) : null}
      </Footer>
    </AppShell>
  )
}
