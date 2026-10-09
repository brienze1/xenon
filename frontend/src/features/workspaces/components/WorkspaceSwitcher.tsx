import { useShallow } from 'zustand/react/shallow'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { useWorkspacesStore } from '../store'
import styles from './WorkspaceSwitcher.module.css'

export function WorkspaceSwitcher() {
  const { workspaces, activeLocalId, activate } = useWorkspacesStore(
    useShallow((state) => ({ workspaces: state.workspaces, activeLocalId: state.activeLocalId, activate: state.activate })),
  )
  const active = workspaces.find((workspace) => workspace.localId === activeLocalId)

  return (
    <Dropdown.Root className={styles.switcher}>
      <Dropdown.Trigger label={`Workspace: ${active?.name ?? 'none'}`}>
        <span className={styles.name}>{active?.name ?? 'No workspace'}</span>
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
      </Dropdown.Menu>
    </Dropdown.Root>
  )
}
