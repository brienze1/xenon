import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import styles from './ImportDialog.module.css'

export function ImportProgress() {
  const progress = useImportStore((state) => state.progress)
  const cancel = useImportStore((state) => state.cancel)
  const total = progress?.total ?? 0

  return (
    <>
      <Dialog.Body>
        <progress className={styles.progress} aria-label="Import progress" max={Math.max(total, 1)} value={progress?.done ?? 0} />
        <p className={styles.status} role="status">
          {progress && progress.currentItem ? (
            <>
              Importing <span translate="no">{progress.currentItem}</span>… ({progress.done} of {total})
            </>
          ) : (
            'Preparing the import…'
          )}
        </p>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={cancel}>
          Cancel
        </Button>
      </Dialog.Actions>
    </>
  )
}
