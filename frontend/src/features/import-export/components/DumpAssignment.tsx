import { useId, useState } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { countsLabel, itemKindLabel } from '@/features/import-export/lib/labels'
import { skippedParts } from '@/features/import-export/lib/plan'
import { ALL_INCLUDED } from '@/features/import-export/model'
import { useImportStore } from '@/features/import-export/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { ImportPreview, ImportPreviewItem } from '@/platform/mappers/import'
import { NewWorkspaceForm } from './NewWorkspaceForm'
import { WorkspacePicker } from './WorkspacePicker'
import styles from './ImportDialog.module.css'

const TABLE_HEAD = (
  <thead>
    <tr>
      <th scope="col">Import</th>
      <th scope="col">Item</th>
      <th scope="col">Kind</th>
      <th scope="col">Workspace</th>
    </tr>
  </thead>
)

interface DumpRowProps {
  item: ImportPreviewItem
}

function DumpRow({ item }: DumpRowProps) {
  const target = useImportStore((state) => state.targets[item.key])
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
        <span className={styles.name} translate="no">
          {item.name}
        </span>
        <span className={styles.path}>{countsLabel(item.counts)}</span>
      </td>
      <td className={styles.muted}>{itemKindLabel(item.kind)}</td>
      <td>
        <WorkspacePicker
          forKey={item.key}
          label={`Workspace for ${item.name}`}
          value={target.localId}
          disabled={!target.include}
          onChange={(localId) => setTarget(item.key, { localId })}
        />
      </td>
    </tr>
  )
}

function BulkAssign() {
  const firstWorkspace = useWorkspacesStore((state) => state.workspaces.find((workspace) => !workspace.missing)?.localId ?? '')
  const activeLocalId = useWorkspacesStore((state) => state.activeLocalId)
  const assignIncluded = useImportStore((state) => state.assignIncluded)
  const [chosen, setChosen] = useState<string | null>(null)
  const localId = chosen ?? activeLocalId ?? firstWorkspace

  return (
    <div className={styles.bulk}>
      <span className={styles.label}>Assign the included items to</span>
      <WorkspacePicker forKey={ALL_INCLUDED} label="Workspace for the included items" value={localId} onChange={setChosen} />
      <Button size="sm" disabled={!localId} onClick={() => assignIncluded(localId)}>
        Assign
      </Button>
    </div>
  )
}

interface DumpAssignmentProps {
  preview: ImportPreview
}

export function DumpAssignment({ preview }: DumpAssignmentProps) {
  const newWorkspaceFor = useImportStore((state) => state.newWorkspaceFor)
  const skipped = skippedParts(preview)
  const skippedId = useId()

  return (
    <>
      <BulkAssign />
      {newWorkspaceFor !== null ? <NewWorkspaceForm /> : null}
      <table className={styles.table}>
        {TABLE_HEAD}
        <tbody>
          {preview.items.map((item) => (
            <DumpRow key={item.key} item={item} />
          ))}
        </tbody>
      </table>
      {skipped.length > 0 ? (
        <section className={styles.section} aria-labelledby={skippedId}>
          <h3 id={skippedId} className={styles.sectionTitle}>
            Not Imported
          </h3>
          <ul className={styles.list}>
            {skipped.map((warning) => (
              <li key={warning.path} className={styles.muted}>
                {warning.message}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </>
  )
}
