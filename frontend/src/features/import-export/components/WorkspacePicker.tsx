import { useShallow } from 'zustand/react/shallow'
import { useImportStore } from '@/features/import-export/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import styles from './ImportDialog.module.css'

const NEW_WORKSPACE = '__new__'

const NEW_WORKSPACE_OPTION = <option value={NEW_WORKSPACE}>New workspace…</option>

interface WorkspacePickerProps {
  /** The item key the picker targets, or ALL_INCLUDED for the bulk picker. */
  forKey: string
  label: string
  value: string
  disabled?: boolean
  onChange: (localId: string) => void
}

export function WorkspacePicker({ forKey, label, value, disabled = false, onChange }: WorkspacePickerProps) {
  const workspaces = useWorkspacesStore(useShallow((state) => state.workspaces.filter((workspace) => !workspace.missing)))
  const startNewWorkspace = useImportStore((state) => state.startNewWorkspace)

  return (
    <select
      name={`workspace-${forKey}`}
      aria-label={label}
      className={styles.select}
      value={value}
      disabled={disabled}
      onChange={(event) => (event.target.value === NEW_WORKSPACE ? startNewWorkspace(forKey) : onChange(event.target.value))}
    >
      {workspaces.map((workspace) => (
        <option key={workspace.localId} value={workspace.localId} translate="no">
          {workspace.name}
        </option>
      ))}
      {NEW_WORKSPACE_OPTION}
    </select>
  )
}
