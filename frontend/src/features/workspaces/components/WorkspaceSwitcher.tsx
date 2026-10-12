import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { useExportStore } from '@/features/export/store'
import { useSyncStore } from '@/features/sync/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { WorkspaceErrorText } from './WorkspaceErrorText'
import { CreateWorkspaceDialog } from './CreateWorkspaceDialog'
import { RemoveWorkspaceDialog } from './RemoveWorkspaceDialog'
import { RenameWorkspaceDialog } from './RenameWorkspaceDialog'
import styles from './WorkspaceSwitcher.module.css'

type SwitcherDialog = 'create' | 'rename' | 'remove' | null

const OPEN_TITLE = 'Choose an existing workspace folder'

const MISSING_MARK = <span className={styles.missing}> (missing)</span>

const CHEVRON = (
  <span className={styles.chevron} aria-hidden="true">
    ▾
  </span>
)

const MENU_SEPARATOR = <hr className={styles.separator} />

export function WorkspaceSwitcher() {
  const { workspaces, activeLocalId, missingLocalId, activate, open, chooseFolder } = useWorkspacesStore(
    useShallow((state) => ({
      workspaces: state.workspaces,
      activeLocalId: state.activeLocalId,
      missingLocalId: state.missingLocalId,
      activate: state.activate,
      open: state.open,
      chooseFolder: state.chooseFolder,
    })),
  )
  const openExport = useExportStore((state) => state.openWorkspace)
  const openSync = useSyncStore((state) => state.openPanel)
  const error = useWorkspacesStore((state) => state.error)
  const clearError = useWorkspacesStore((state) => state.clearError)
  const [dialog, setDialog] = useState<SwitcherDialog>(null)
  const current = workspaces.find((workspace) => workspace.localId === (activeLocalId ?? missingLocalId))

  async function onOpen() {
    const path = await chooseFolder(OPEN_TITLE)
    if (path) {
      await open(path)
    }
  }

  function show(next: SwitcherDialog) {
    clearError()
    setDialog(next)
  }

  return (
    <div className={styles.container}>
      <Dropdown.Root className={styles.switcher}>
        <Dropdown.Trigger label={`Workspace: ${current?.name ?? 'None'}`}>
          <span className={styles.name}>{current?.name ?? 'No workspace'}</span>
          {CHEVRON}
        </Dropdown.Trigger>
        <Dropdown.Menu label="Workspaces">
          {workspaces.map((workspace) => (
            <Dropdown.RadioItem
              key={workspace.localId}
              checked={workspace.localId === activeLocalId}
              onSelect={() => void activate(workspace.localId)}
            >
              {workspace.name}
              {workspace.missing ? MISSING_MARK : null}
            </Dropdown.RadioItem>
          ))}
          {MENU_SEPARATOR}
          <Dropdown.Item onSelect={() => show('create')}>New Workspace…</Dropdown.Item>
          <Dropdown.Item onSelect={() => void onOpen()}>Open Workspace…</Dropdown.Item>
          {current && current.localId === activeLocalId ? <Dropdown.Item onSelect={() => void openExport()}>Export Workspace…</Dropdown.Item> : null}
          {current && current.localId === activeLocalId ? <Dropdown.Item onSelect={openSync}>Sync Settings…</Dropdown.Item> : null}
          {current ? <Dropdown.Item onSelect={() => show('rename')}>Rename Workspace…</Dropdown.Item> : null}
          {current ? <Dropdown.Item onSelect={() => show('remove')}>Remove Workspace…</Dropdown.Item> : null}
        </Dropdown.Menu>
      </Dropdown.Root>
      {error && dialog === null ? (
        <Banner.Root tone="error" className={styles.error}>
          <Banner.Message><WorkspaceErrorText error={error} /></Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss Error
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <CreateWorkspaceDialog open={dialog === 'create'} onClose={() => setDialog(null)} />
      {current ? <RenameWorkspaceDialog key={`rename-${current.localId}`} workspace={current} open={dialog === 'rename'} onClose={() => setDialog(null)} /> : null}
      {current ? <RemoveWorkspaceDialog key={`remove-${current.localId}`} workspace={current} open={dialog === 'remove'} onClose={() => setDialog(null)} /> : null}
    </div>
  )
}
