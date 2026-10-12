import { useEffect, useMemo } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { ConsoleDrawer as ConsoleContent } from '@/features/console/components/ConsoleDrawer'
import { CookiesButton } from '@/features/cookies/components/CookiesButton'
import { CookiesDialog } from '@/features/cookies/components/CookiesDialog'
import { EnvironmentSwitcher } from '@/features/environments/components/EnvironmentSwitcher'
import { ExportDialog } from '@/features/export/components/ExportDialog'
import { ImportDialog } from '@/features/import-export/components/ImportDialog'
import { useImportStore } from '@/features/import-export/store'
import { PassphraseDialog } from '@/features/environments/components/PassphraseDialog'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { QuickSearch } from '@/features/search/components/QuickSearch'
import { useSearchStore } from '@/features/search/store'
import { SidePanel } from '@/features/side-panel/components/SidePanel'
import { useExamplesStore } from '@/features/examples/store'
import { SettingsButton } from '@/features/settings/components/SettingsButton'
import { HostKeyPrompt } from '@/features/settings/components/HostKeyPrompt'
import { SettingsDialog } from '@/features/settings/components/SettingsDialog'
import { useSettingsStore } from '@/features/settings/store'
import { SidebarPanel } from '@/features/sidebar/components/SidebarPanel'
import { useSidebarStore } from '@/features/sidebar/store'
import { FooterSyncStatus } from '@/features/sync/components/FooterSyncStatus'
import { CloneWorkspaceDialog } from '@/features/sync/components/CloneWorkspaceDialog'
import { ConflictView } from '@/features/sync/components/ConflictView'
import { LinkRepoDialog } from '@/features/sync/components/LinkRepoDialog'
import { SyncPanel } from '@/features/sync/components/SyncPanel'
import type { EnvironmentTabTarget, OpenMode } from '@/features/tabs/model'
import type { ContainerKind } from '@/platform/mappers/tree'
import { useTabsStore } from '@/features/tabs/store'
import { MissingFolderNotice } from '@/features/workspaces/components/MissingFolderNotice'
import { WelcomeScreen } from '@/features/workspaces/components/WelcomeScreen'
import { WorkspaceSwitcher } from '@/features/workspaces/components/WorkspaceSwitcher'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { activeEnvironmentTarget, activeTabKind, isActiveTabRequest } from './active-tab'
import styles from './App.module.css'
import { bootstrap } from './bootstrap'
import { GO_TO_TAB_POSITIONS, HOTKEYS, type HotkeyHandlers } from './hotkeys'
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

function openEnvironmentTab(target: EnvironmentTabTarget, mode: OpenMode) {
  useTabsStore.getState().openEnvironmentTab(target, mode)
}

function openGlobals() {
  openEnvironmentTab({ kind: 'globals' }, 'pinned')
}

/** "Manage Environments…": the environments live in the sidebar's Environments view. */
function manageEnvironments() {
  useSidebarStore.getState().setView('environments')
  const workspaces = useWorkspacesStore.getState()
  if (!workspaces.layout.sidebarVisible) {
    workspaces.setLayout({ sidebarVisible: true })
  }
}

const ENVIRONMENT_SWITCHER = <EnvironmentSwitcher onOpenGlobals={openGlobals} onManage={manageEnvironments} />

const PASSPHRASE_DIALOG = <PassphraseDialog />

const IMPORT_DIALOG = <ImportDialog />

const EXPORT_DIALOG = <ExportDialog />

const SETTINGS_DIALOG = <SettingsDialog shortcuts={HOTKEYS} />

const HOST_KEY_PROMPT = <HostKeyPrompt />

const QUICK_SEARCH = <QuickSearch />

const SETTINGS_BUTTON = <SettingsButton />

const COOKIES_DIALOG = <CookiesDialog />

const COOKIES_BUTTON = <CookiesButton />

const FOOTER_SYNC_STATUS = <FooterSyncStatus />

const SYNC_PANEL = <SyncPanel />

const LINK_REPO_DIALOG = <LinkRepoDialog />

const CLONE_WORKSPACE_DIALOG = <CloneWorkspaceDialog />

const CONFLICT_VIEW = <ConflictView />

const TOP_BAR_END_WITH_ENVIRONMENTS = (
  <>
    {ENVIRONMENT_SWITCHER}
    {SETTINGS_BUTTON}
  </>
)

function openItemTab(itemId: string, kind: ContainerKind, mode: OpenMode) {
  useTabsStore.getState().openItemTab({ kind, itemId }, mode)
}

const CONSOLE = (
  <ConsoleDrawer>
    <ConsoleContent />
  </ConsoleDrawer>
)

function openSettings() {
  void useSettingsStore.getState().openSettings()
}

const NO_WORKSPACE_HOTKEYS: HotkeyHandlers = { settings: openSettings }

function goToTabHandlers(): HotkeyHandlers {
  return Object.fromEntries([...GO_TO_TAB_POSITIONS].map(([action, position]) => [action, () => useTabsStore.getState().activateIndex(position)]))
}

function toggleLayout(key: 'sidebarVisible' | 'consoleVisible') {
  const workspaces = useWorkspacesStore.getState()
  workspaces.setLayout({ [key]: !workspaces.layout[key] })
}

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
  const selectedVariables = useTabsStore(useShallow(activeEnvironmentTarget))
  const activeKind = useTabsStore(activeTabKind)
  // Only requests and examples have side panel content; other tabs (environments, collections, folders) hide it.
  const sidePanelTabId = activeTabId !== null && (activeKind === 'request' || activeKind === 'example') ? activeTabId : null
  const tabs = useTabsStore.getState
  const send = useRequestBuilderStore((state) => state.send)
  const adoptRename = useRequestBuilderStore((state) => state.adoptRename)
  const hotkeys = useMemo<HotkeyHandlers>(
    () =>
      hasActive
        ? {
            send: () => {
              const openTabs = useTabsStore.getState()
              if (openTabs.activeTabId && isActiveTabRequest(openTabs)) {
                void send(openTabs.activeTabId)
              }
            },
            save: () => void useTabsStore.getState().save(),
            saveAs: () => useTabsStore.getState().saveAs(),
            newTab: () => void useTabsStore.getState().newTab(),
            closeTab: () => useTabsStore.getState().requestClose(),
            forceCloseTab: () => useTabsStore.getState().forceClose(),
            reopenClosedTab: () => void useTabsStore.getState().reopenClosed(),
            nextTab: () => useTabsStore.getState().cycle(1),
            previousTab: () => useTabsStore.getState().cycle(-1),
            ...goToTabHandlers(),
            lastTab: () => useTabsStore.getState().activateIndex(9),
            focusUrl: () => useRequestBuilderStore.getState().focusUrl(),
            searchRequests: () => useSearchStore.getState().openSearch(),
            findInResponse: () => useResponseViewerStore.getState().requestFind(),
            toggleSidebar: () => toggleLayout('sidebarVisible'),
            console: () => toggleLayout('consoleVisible'),
            renameSelected: () => useSidebarStore.getState().renameSelected(),
            duplicateSelected: () => void useSidebarStore.getState().duplicateSelected(),
            environmentSwitcher: () => useEnvironmentsStore.getState().openSwitcher(),
            import: () => useImportStore.getState().open(),
            settings: openSettings,
          }
        : NO_WORKSPACE_HOTKEYS,
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
      <TopBar end={hasActive ? TOP_BAR_END_WITH_ENVIRONMENTS : SETTINGS_BUTTON}>{hasWorkspaces ? WORKSPACE_SWITCHER : null}</TopBar>
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
              onOpenVariables={openEnvironmentTab}
              selectedVariables={selectedVariables}
              onOpenItem={openItemTab}
            />
          </Sidebar>
          <MainArea>
            <TabsArea key={active.localId} />
          </MainArea>
          <RightPanel applies={sidePanelTabId !== null}>
            {sidePanelTabId !== null ? (
              <SidePanel key={active.localId} tabId={sidePanelTabId} view={layout.rightPanelView} onViewChange={(rightPanelView) => setLayout({ rightPanelView })} />
            ) : null}
          </RightPanel>
        </>
      ) : (
        <MainArea>
          {missingLocalId ? <MissingFolderNotice localId={missingLocalId} /> : status === 'ready' ? WELCOME_SCREEN : null}
        </MainArea>
      )}
      {active ? CONSOLE : null}
      {PASSPHRASE_DIALOG}
      {SETTINGS_DIALOG}
      {HOST_KEY_PROMPT}
      {active ? IMPORT_DIALOG : null}
      {active ? EXPORT_DIALOG : null}
      {active ? COOKIES_DIALOG : null}
      {active ? SYNC_PANEL : null}
      {active ? LINK_REPO_DIALOG : null}
      {active ? CONFLICT_VIEW : null}
      {CLONE_WORKSPACE_DIALOG}
      {active ? QUICK_SEARCH : null}
      <Footer>
        {active ? (
          <>
            <span className={styles.path} translate="no">
              {active.path}
            </span>
            {FOOTER_SYNC_STATUS}
            {COOKIES_BUTTON}
          </>
        ) : null}
      </Footer>
    </AppShell>
  )
}
