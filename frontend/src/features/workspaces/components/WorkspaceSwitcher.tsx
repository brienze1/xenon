import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { workspaceErrorMessage } from '../lib/error-message'
import { useWorkspacesStore } from '../store'
import { CreateWorkspaceDialog } from './CreateWorkspaceDialog'
import { RemoveWorkspaceDialog } from './RemoveWorkspaceDialog'
import { RenameWorkspaceDialog } from './RenameWorkspaceDialog'
import styles from './WorkspaceSwitcher.module.css'

type SwitcherDialog = 'create' | 'rename' | 'remove' | null

const OPEN_TITLE = 'Choose an existing workspace folder'

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
        <Dropdown.Trigger label={`Workspace: ${current?.name ?? 'none'}`}>
          <span className={styles.name}>{current?.name ?? 'No workspace'}</span>
          <span className={styles.chevron} aria-hidden="true">
            ▾
          </span>
        </Dropdown.Trigger>
        <Dropdown.Menu label="Workspaces">
          {workspaces.map((workspace) => (
            <Dropdown.Item
              key={workspace.localId}
              checked={workspace.localId === activeLocalId}
              onSelect={() => void activate(workspace.localId)}
            >
              {workspace.name}
              {workspace.missing ? <span className={styles.missing}> (missing)</span> : null}
            </Dropdown.Item>
          ))}
          <div role="separator" className={styles.separator} />
          <Dropdown.Item onSelect={() => show('create')}>New workspace…</Dropdown.Item>
          <Dropdown.Item onSelect={() => void onOpen()}>Open workspace…</Dropdown.Item>
          {current ? <Dropdown.Item onSelect={() => show('rename')}>Rename workspace…</Dropdown.Item> : null}
          {current ? <Dropdown.Item onSelect={() => show('remove')}>Remove workspace…</Dropdown.Item> : null}
        </Dropdown.Menu>
      </Dropdown.Root>
      {error && dialog === null ? (
        <Banner.Root tone="error" className={styles.error}>
          <Banner.Message>{workspaceErrorMessage(error)}</Banner.Message>
          <Banner.Actions>
            <Button variant="ghost" size="sm" onClick={clearError}>
              Dismiss
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
