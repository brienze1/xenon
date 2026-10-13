import { useId, useState } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import { ImportErrorBanner } from './ImportErrorBanner'
import styles from './ImportDialog.module.css'

const DROP_HINT = <p>Drop Postman or Bruno exports, folders or zips anywhere on this window</p>

export function ImportSourceStep() {
  const [text, setText] = useState('')
  const pasteId = useId()
  const detecting = useImportStore((state) => state.step === 'detecting')
  const pickSources = useImportStore((state) => state.pickSources)
  const detect = useImportStore((state) => state.detect)
  const close = useImportStore((state) => state.close)

  return (
    <>
      <Dialog.Body>
        <div className={styles.dropZone} data-file-drop-target>
          {DROP_HINT}
          <Button size="sm" disabled={detecting} onClick={() => void pickSources()}>
            Choose Files…
          </Button>
        </div>
        <div className={styles.field}>
          <label htmlFor={pasteId} className={styles.label}>
            Or paste Postman JSON or a cURL command
          </label>
          <textarea
            id={pasteId}
            name="import-text"
            className={styles.paste}
            spellCheck={false}
            autoComplete="off"
            value={text}
            onChange={(event) => setText(event.target.value)}
          />
        </div>
        <div className={styles.pasteActions}>
          <Button size="sm" disabled={detecting || text.trim() === ''} onClick={() => void detect({ text })}>
            Read Pasted Text
          </Button>
        </div>
        {detecting ? (
          <p className={styles.status} role="status">
            Reading the source…
          </p>
        ) : null}
        <ImportErrorBanner />
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={close}>
          Cancel
        </Button>
      </Dialog.Actions>
    </>
  )
}
