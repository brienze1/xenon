import { useId } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { countsLabel } from '@/features/import-export/lib/labels'
import { useImportStore } from '@/features/import-export/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { ImportWarning, ImportWorkspaceResult, PmUsage } from '@/platform/mappers/import'
import { formatNumber } from '@/shared/lib/number-format'
import styles from './ImportDialog.module.css'

interface OpenButtonProps {
  localId: string
  itemId: string
  label: string
}

function OpenButton({ localId, itemId, label }: OpenButtonProps) {
  const openItem = useImportStore((state) => state.openItem)
  if (!itemId || !localId) {
    return null
  }
  return (
    <Button size="sm" variant="ghost" aria-label={`Open ${label}`} onClick={() => void openItem(localId, itemId)}>
      Open
    </Button>
  )
}

interface WarningEntryProps {
  warning: ImportWarning
}

function WarningEntry({ warning }: WarningEntryProps) {
  return (
    <li className={styles.entry}>
      <span className={styles.entryText}>
        {warning.message}
        <span className={styles.path} translate="no">
          {warning.path}
        </span>
      </span>
      <OpenButton localId={warning.localId} itemId={warning.itemId} label={warning.path} />
    </li>
  )
}

interface UsageEntryProps {
  usage: PmUsage
}

function UsageEntry({ usage }: UsageEntryProps) {
  return (
    <li>
      <span className={styles.code} translate="no">
        {usage.api}
      </span>
      <ul className={styles.list}>
        {usage.occurrences.map((occurrence) => (
          <li key={`${occurrence.itemId}-${occurrence.phase}-${occurrence.line}`} className={styles.entry}>
            <span className={styles.path} translate="no">
              {occurrence.path} · {occurrence.phase} line {occurrence.line}
            </span>
            <OpenButton localId={occurrence.localId} itemId={occurrence.itemId} label={occurrence.path} />
          </li>
        ))}
      </ul>
    </li>
  )
}

interface WorkspaceEntryProps {
  workspace: ImportWorkspaceResult
}

function WorkspaceEntry({ workspace }: WorkspaceEntryProps) {
  const name = useWorkspacesStore((state) => state.workspaces.find((candidate) => candidate.localId === workspace.localId)?.name ?? workspace.localId)
  return (
    <li>
      <span className={styles.name} translate="no">
        {name}
      </span>
      : {countsLabel(workspace.created)}
      {workspace.replaced > 0 ? `, ${formatNumber(workspace.replaced, 0)} replaced` : ''}
      {workspace.skipped > 0 ? `, ${formatNumber(workspace.skipped, 0)} skipped` : ''}
    </li>
  )
}

export function ImportSummary() {
  const result = useImportStore((state) => state.result)
  const close = useImportStore((state) => state.close)
  const failedId = useId()
  const warningsId = useId()
  const usagesId = useId()
  if (!result) {
    return null
  }

  return (
    <>
      <Dialog.Body>
        <ul className={styles.list}>
          {result.workspaces.map((workspace) => (
            <WorkspaceEntry key={workspace.localId} workspace={workspace} />
          ))}
        </ul>
        {result.failed.length > 0 ? (
          <section className={styles.section} aria-labelledby={failedId}>
            <h3 id={failedId} className={styles.sectionTitle}>
              Not Imported
            </h3>
            <ul className={styles.list}>
              {result.failed.map((failure) => (
                <li key={failure.itemKey}>
                  <span translate="no">{failure.name}</span>: {failure.message}
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        {result.warnings.length > 0 ? (
          <section className={styles.section} aria-labelledby={warningsId}>
            <h3 id={warningsId} className={styles.sectionTitle}>
              Needs Attention ({formatNumber(result.warnings.length, 0)})
            </h3>
            <ul className={styles.list}>
              {result.warnings.map((warning, index) => (
                <WarningEntry key={`${warning.code}-${warning.itemId}-${index}`} warning={warning} />
              ))}
            </ul>
          </section>
        ) : null}
        {result.pmUsages.length > 0 ? (
          <section className={styles.section} aria-labelledby={usagesId}>
            <h3 id={usagesId} className={styles.sectionTitle}>
              <span translate="no">pm</span> APIs Outside <span translate="no">Xenon</span>’s Script Subset
            </h3>
            <ul className={styles.list}>
              {result.pmUsages.map((usage) => (
                <UsageEntry key={usage.api} usage={usage} />
              ))}
            </ul>
          </section>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="primary" onClick={close}>
          Done
        </Button>
      </Dialog.Actions>
    </>
  )
}
