import { useEffect } from 'react'
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
import { RightPanel } from './layout/regions/RightPanel'
import { Sidebar } from './layout/regions/Sidebar'
import { TopBar } from './layout/regions/TopBar'

const NO_HOTKEY_HANDLERS = {}

export function App() {
  const status = useWorkspacesStore((state) => state.status)
  const active = useWorkspacesStore((state) =>
    state.workspaces.find((workspace) => workspace.localId === state.activeLocalId),
  )

  useEffect(() => bootstrap(), [])
  useAppHotkeys(NO_HOTKEY_HANDLERS)

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
            <EmptyState.Root label="No request open">
              <EmptyState.Art />
              <EmptyState.Title>{active.name}</EmptyState.Title>
              <EmptyState.Description>Open a new tab to build and send a request.</EmptyState.Description>
            </EmptyState.Root>
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
