import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useExportStore } from '@/features/export/store'
import { ExportOptionsStep } from './ExportOptionsStep'
import { ExportProgress } from './ExportProgress'
import { ExportSummary } from './ExportSummary'
import styles from './ExportDialog.module.css'

const LOADING = (
  <Dialog.Body>
    <p className={styles.status} role="status">
      Loading the export options…
    </p>
  </Dialog.Body>
)

export function ExportDialog() {
  const step = useExportStore((state) => state.step)
  const name = useExportStore((state) => state.target?.name ?? '')
  const close = useExportStore((state) => state.close)

  return (
    <Dialog.Root open={step !== 'closed'} onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      <Dialog.Title>
        Export “<span translate="no">{name}</span>”
      </Dialog.Title>
      {step === 'loading' ? LOADING : null}
      {step === 'configure' ? <ExportOptionsStep /> : null}
      {step === 'exporting' ? <ExportProgress /> : null}
      {step === 'summary' ? <ExportSummary /> : null}
    </Dialog.Root>
  )
}
