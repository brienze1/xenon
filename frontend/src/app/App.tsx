import { useEffect, useMemo } from 'react'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { WelcomeScreen } from '@/features/workspaces/components/WelcomeScreen'
import { WorkspaceSwitcher } from '@/features/workspaces/components/WorkspaceSwitcher'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import styles from './App.module.css'
import { bootstrap } from './bootstrap'
import { useAppHotkeys } from './hotkeys'
import { AppShell } from './layout/AppShell'
import { ConsoleDrawer } from './layout/regions/ConsoleDrawer'
import { Footer } from './layout/regions/Footer'
import { MainArea } from './layout/regions/MainArea'
import { RequestWorkspace } from './layout/RequestWorkspace'
import { RightPanel } from './layout/regions/RightPanel'
import { Sidebar } from './layout/regions/Sidebar'
import { TopBar } from './layout/regions/TopBar'

export function App() {
  const status = useWorkspacesStore((state) => state.status)
  const active = useWorkspacesStore((state) =>
    state.workspaces.find((workspace) => workspace.localId === state.activeLocalId),
  )

  const activeTabId = useRequestBuilderStore((state) => state.activeTabId)
  const send = useRequestBuilderStore((state) => state.send)
  const hotkeys = useMemo(() => (active ? { send: () => void send(activeTabId) } : {}), [active, send, activeTabId])

  useEffect(() => bootstrap(), [])
  useAppHotkeys(hotkeys)

  return (
    <AppShell>
      <TopBar>{active ? <WorkspaceSwitcher /> : null}</TopBar>
      {active ? (
        <>
          <Sidebar>
            <EmptyState.Root label="No collections">
              <EmptyState.Description>No collections yet.</EmptyState.Description>
            </EmptyState.Root>
          </Sidebar>
          <MainArea>
            <RequestWorkspace key={active.localId} tabId={activeTabId} />
          </MainArea>
          <RightPanel>
            <h2 className={styles.infoTitle}>Info</h2>
          </RightPanel>
        </>
      ) : (
        <MainArea>{status === 'ready' ? <WelcomeScreen /> : null}</MainArea>
      )}
      <ConsoleDrawer open={false} />
      <Footer>{active ? <span className={styles.path}>{active.path}</span> : null}</Footer>
    </AppShell>
  )
}
