import { useShallow } from 'zustand/react/shallow'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import { countsLabel, sourceLabel } from '@/features/import-export/lib/labels'
import { hasIncludedItems } from '@/features/import-export/lib/plan'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { ImportPreviewItem } from '@/platform/mappers/import'
import { ImportErrorBanner } from './ImportErrorBanner'
import styles from './ImportDialog.module.css'

const NEW_COLLECTION = ''

const TABLE_HEAD = (
  <thead>
    <tr>
      <th scope="col">Import</th>
      <th scope="col">Item</th>
      <th scope="col">Workspace</th>
      <th scope="col">Location</th>
    </tr>
  </thead>
)

interface ItemRowProps {
  item: ImportPreviewItem
}

function ItemRow({ item }: ItemRowProps) {
  const target = useImportStore((state) => state.targets[item.key])
  const locations = useImportStore((state) => (target ? state.locations[target.localId] : undefined))
  const setTarget = useImportStore((state) => state.setTarget)
  const workspaces = useWorkspacesStore(useShallow((state) => state.workspaces.filter((workspace) => !workspace.missing)))
  if (!target) {
    return null
  }

  return (
    <tr>
      <td>
        <input
          type="checkbox"
          name={`include-${item.key}`}
          aria-label={`Import ${item.name}`}
          checked={target.include}
          onChange={(event) => setTarget(item.key, { include: event.target.checked })}
        />
      </td>
      <td>
        <span className={styles.name} translate="no">
          {item.name}
        </span>
        <span className={styles.path}>
          {countsLabel(item.counts)}
          {item.warningsCount > 0 ? ` · ${item.warningsCount} to review` : ''}
        </span>
      </td>
      <td>
        <select
          name={`workspace-${item.key}`}
          aria-label={`Workspace for ${item.name}`}
          className={styles.select}
          value={target.localId}
          disabled={!target.include}
          onChange={(event) => setTarget(item.key, { localId: event.target.value })}
        >
          {workspaces.map((workspace) => (
            <option key={workspace.localId} value={workspace.localId} translate="no">
              {workspace.name}
            </option>
          ))}
        </select>
      </td>
      <td>
        <select
          name={`location-${item.key}`}
          aria-label={`Location for ${item.name}`}
          className={styles.select}
          value={target.parentId}
          disabled={!target.include}
          onChange={(event) => setTarget(item.key, { parentId: event.target.value })}
        >
          <option value={NEW_COLLECTION}>New collection</option>
          {(locations ?? []).map((location) => (
            <option key={location.id} value={location.id} translate="no">
              Into {location.label}
            </option>
          ))}
        </select>
      </td>
    </tr>
  )
}

export function ImportPreviewStep() {
  const preview = useImportStore((state) => state.preview)
  const canImport = useImportStore((state) => hasIncludedItems(state.targets))
  const start = useImportStore((state) => state.start)
  const back = useImportStore((state) => state.back)
  if (!preview) {
    return null
  }

  return (
    <>
      <Dialog.Body>
        <p className={styles.source}>{sourceLabel(preview.kind, preview.version)}</p>
        {preview.items.length > 0 ? (
          <table className={styles.table}>
            {TABLE_HEAD}
            <tbody>
              {preview.items.map((item) => (
                <ItemRow key={item.key} item={item} />
              ))}
            </tbody>
          </table>
        ) : (
          <p className={styles.status}>Importing this kind of source arrives in a later update.</p>
        )}
        <ImportErrorBanner />
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={back}>
          Back
        </Button>
        <Button variant="primary" disabled={!canImport} onClick={() => void start()}>
          Import
        </Button>
      </Dialog.Actions>
    </>
  )
}
