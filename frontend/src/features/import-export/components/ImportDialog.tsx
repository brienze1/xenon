import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import { ClashPrompt } from './ClashPrompt'
import { ImportPreviewStep } from './ImportPreviewStep'
import { ImportProgress } from './ImportProgress'
import { ImportSourceStep } from './ImportSourceStep'
import { ImportSummary } from './ImportSummary'
import { PassphrasePrompt } from './PassphrasePrompt'
import styles from './ImportDialog.module.css'

const IMPORT_TITLE = <Dialog.Title>Import</Dialog.Title>

export function ImportDialog() {
  const step = useImportStore((state) => state.step)
  const close = useImportStore((state) => state.close)

  return (
    <Dialog.Root open={step !== 'closed'} onOpenChange={(next) => (next ? undefined : close())} className={styles.dialog}>
      {IMPORT_TITLE}
      {step === 'choose' || step === 'detecting' ? <ImportSourceStep /> : null}
      {step === 'preview' ? <ImportPreviewStep /> : null}
      {step === 'clashes' ? <ClashPrompt /> : null}
      {step === 'passphrase' ? <PassphrasePrompt /> : null}
      {step === 'importing' ? <ImportProgress /> : null}
      {step === 'summary' ? <ImportSummary /> : null}
    </Dialog.Root>
  )
}
