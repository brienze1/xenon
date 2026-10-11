import { useId, type FormEvent } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useImportStore } from '@/features/import-export/store'
import styles from './ImportDialog.module.css'

const ASK = 'This export carries secret values encrypted with a passphrase. Enter it to put them in the target workspace’s vault, or import without them; they then arrive as missing.'
const WRONG = 'That passphrase does not open the export’s secrets. Try again, or import without them.'

export function PassphrasePrompt() {
  const passphrase = useImportStore((state) => state.passphrase)
  const wrong = useImportStore((state) => state.passphraseRequest === 'wrong')
  const setPassphrase = useImportStore((state) => state.setPassphrase)
  const submitPassphrase = useImportStore((state) => state.submitPassphrase)
  const importWithoutSecrets = useImportStore((state) => state.importWithoutSecrets)
  const back = useImportStore((state) => state.back)
  const inputId = useId()
  const messageId = useId()

  function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void submitPassphrase()
  }

  return (
    <>
      <Dialog.Body>
        <form id={`${inputId}-form`} className={styles.field} noValidate onSubmit={onSubmit}>
          <p id={messageId} className={wrong ? styles.fieldError : styles.source} role={wrong ? 'alert' : undefined}>
            {wrong ? WRONG : ASK}
          </p>
          <label htmlFor={inputId} className={styles.label}>
            Passphrase
          </label>
          <input
            id={inputId}
            name="import-passphrase"
            type="password"
            autoComplete="off"
            spellCheck={false}
            className={styles.input}
            aria-describedby={messageId}
            aria-invalid={wrong ? true : undefined}
            value={passphrase}
            onChange={(event) => setPassphrase(event.target.value)}
          />
        </form>
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="ghost" onClick={back}>
          Cancel
        </Button>
        <Button onClick={() => void importWithoutSecrets()}>Import Without Secrets</Button>
        <Button type="submit" form={`${inputId}-form`} variant="primary" disabled={passphrase === ''}>
          Import
        </Button>
      </Dialog.Actions>
    </>
  )
}
