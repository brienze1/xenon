import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { formatNumber } from '@/shared/lib/number-format'
import { showsProgress } from '@/features/export/lib/choices'
import { useExportStore } from '@/features/export/store'
import styles from './ExportDialog.module.css'

export function ExportProgress() {
  const progress = useExportStore((state) => state.progress)
  const cancel = useExportStore((state) => state.cancel)

  return (
    <>
      <Dialog.Body>
        {showsProgress(progress) ? (
          <>
            <progress className={styles.progress} aria-label="Export progress" max={progress.total} value={progress.done} />
            <p className={styles.status} role="status">
              Writing the export… ({formatNumber(progress.done, 0)} of {formatNumber(progress.total, 0)} files)
            </p>
          </>
        ) : (
          <p className={styles.status} role="status">
            Writing the export…
          </p>
        )}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={cancel}>
          Cancel
        </Button>
      </Dialog.Actions>
    </>
  )
}
