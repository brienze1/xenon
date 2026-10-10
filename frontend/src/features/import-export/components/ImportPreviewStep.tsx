import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import { countsLabel, sourceLabel } from '@/features/import-export/lib/labels'
import { hasIncludedItems, missingLocations, needsLocation } from '@/features/import-export/lib/plan'
import type { ImportPreview, ImportPreviewItem } from '@/platform/mappers/import'
import { formatNumber } from '@/shared/lib/number-format'
import { DumpAssignment } from './DumpAssignment'
import { ImportErrorBanner } from './ImportErrorBanner'
import { NewWorkspaceForm } from './NewWorkspaceForm'
import { WorkspacePicker } from './WorkspacePicker'
import styles from './ImportDialog.module.css'

const NEW_COLLECTION = ''
const DUMP_KIND = 'postman_dump'

const VARIABLES_PLACES: Readonly<Record<string, string>> = {
  environment: 'Environments',
  globals: 'Workspace globals',
}

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
        {needsLocation(item.kind) ? (
          <input
            name={`name-${item.key}`}
            autoComplete="off"
            aria-label={`Name for ${item.name}`}
            className={styles.nameInput}
            value={target.name ?? item.name}
            placeholder={item.name}
            spellCheck={false}
            disabled={!target.include}
            onChange={(event) => setTarget(item.key, { name: event.target.value })}
          />
        ) : (
          <span className={styles.name} translate="no">
            {item.name}
          </span>
        )}
        <span className={styles.path}>
          {countsLabel(item.counts)}
          {item.warningsCount > 0 ? ` · ${formatNumber(item.warningsCount, 0)} to review` : ''}
        </span>
      </td>
      <td>
        <WorkspacePicker
          forKey={item.key}
          label={`Workspace for ${item.name}`}
          value={target.localId}
          disabled={!target.include}
          onChange={(localId) => setTarget(item.key, { localId })}
        />
      </td>
      <td>
        {VARIABLES_PLACES[item.kind] ? (
          <span className={styles.muted}>{VARIABLES_PLACES[item.kind]}</span>
        ) : (
          <select
            name={`location-${item.key}`}
            aria-label={`Location for ${item.name}`}
            className={styles.select}
            value={target.parentId}
            disabled={!target.include}
            onChange={(event) => setTarget(item.key, { parentId: event.target.value })}
          >
            {needsLocation(item.kind) ? (
              <option value={NEW_COLLECTION} disabled>
                Choose a collection or folder
              </option>
            ) : (
              <option value={NEW_COLLECTION}>New collection</option>
            )}
            {(locations ?? []).map((location) => (
              <option key={location.id} value={location.id} translate="no">
                Into {location.label}
              </option>
            ))}
          </select>
        )}
      </td>
    </tr>
  )
}

interface ItemTableProps {
  preview: ImportPreview
}

function ItemTable({ preview }: ItemTableProps) {
  const newWorkspaceFor = useImportStore((state) => state.newWorkspaceFor)
  return (
    <>
      {newWorkspaceFor !== null ? <NewWorkspaceForm /> : null}
      <table className={styles.table}>
        {TABLE_HEAD}
        <tbody>
          {preview.items.map((item) => (
            <ItemRow key={item.key} item={item} />
          ))}
        </tbody>
      </table>
    </>
  )
}

export function ImportPreviewStep() {
  const preview = useImportStore((state) => state.preview)
  const canImport = useImportStore((state) => hasIncludedItems(state.targets) && (state.preview === null || !missingLocations(state.preview, state.targets)))
  const start = useImportStore((state) => state.start)
  const back = useImportStore((state) => state.back)
  if (!preview) {
    return null
  }

  return (
    <>
      <Dialog.Body>
        <p className={styles.source}>{sourceLabel(preview.kind, preview.version)}</p>
        {preview.kind === DUMP_KIND ? <DumpAssignment preview={preview} /> : null}
        {preview.kind !== DUMP_KIND && preview.items.length > 0 ? <ItemTable preview={preview} /> : null}
        {preview.items.length === 0 ? <p className={styles.status}>There is nothing to import in this source.</p> : null}
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
