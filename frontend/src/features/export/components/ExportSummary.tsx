import { useId } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { formatNumber } from '@/shared/lib/number-format'
import { countsLabel, secretsLabel } from '@/features/export/lib/labels'
import { useExportStore } from '@/features/export/store'
import { ExportErrorBanner } from './ExportErrorBanner'
import styles from './ExportDialog.module.css'

export function ExportSummary() {
  const summary = useExportStore((state) => state.summary)
  const hasSecrets = useExportStore((state) => state.options?.hasSecrets === true)
  const close = useExportStore((state) => state.close)
  const showInFolder = useExportStore((state) => state.showInFolder)
  const droppedId = useId()
  const warningsId = useId()
  if (!summary) {
    return null
  }

  return (
    <>
      <Dialog.Body>
        <p className={styles.summaryLine} role="status">
          Exported {countsLabel(summary.counts)}.
          <span className={styles.path} translate="no">
            {summary.path}
          </span>
        </p>
        {hasSecrets ? <p className={styles.hint}>{secretsLabel(summary.secrets)}</p> : null}
        {summary.dropped.length > 0 ? (
          <section className={styles.section} aria-labelledby={droppedId}>
            <h3 id={droppedId} className={styles.sectionTitle}>
              Dropped or Changed ({formatNumber(summary.dropped.length, 0)})
            </h3>
            <ul className={styles.list}>
              {summary.dropped.map((drop, index) => (
                <li key={`${drop.itemPath}-${drop.field}-${index}`} className={styles.entryText}>
                  <span className={styles.code} translate="no">
                    {drop.field}
                  </span>
                  : {drop.reason}
                  <span className={styles.path} translate="no">
                    {drop.itemPath}
                  </span>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        {summary.warnings.length > 0 ? (
          <section className={styles.section} aria-labelledby={warningsId}>
            <h3 id={warningsId} className={styles.sectionTitle}>
              Needs Attention ({formatNumber(summary.warnings.length, 0)})
            </h3>
            <ul className={styles.list}>
              {summary.warnings.map((warning, index) => (
                <li key={`${warning.itemPath}-${warning.code}-${index}`} className={styles.entryText}>
                  {warning.message}
                  <span className={styles.path} translate="no">
                    {warning.itemPath}
                  </span>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        <ExportErrorBanner />
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={() => void showInFolder()}>
          Show in Folder
        </Button>
        <Button variant="primary" onClick={close}>
          Done
        </Button>
      </Dialog.Actions>
    </>
  )
}
