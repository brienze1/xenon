import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './MissingSecretsDialog.module.css'

interface MissingSecretsDialogProps {
  tabId: string
}

interface SecretListProps {
  title: string
  names: string[]
}

function SecretList({ title, names }: SecretListProps) {
  return (
    <section className={styles.group}>
      <h3 className={styles.heading}>{title}</h3>
      <ul className={styles.list}>
        {names.map((name) => (
          <li key={name}>
            <code translate="no">{`{{${name}}}`}</code>
          </li>
        ))}
      </ul>
    </section>
  )
}

export function MissingSecretsDialog({ tabId }: MissingSecretsDialogProps) {
  const missingSecrets = useRequestBuilderStore((state) => state.tabs[tabId]?.missingSecrets ?? null)
  const send = useRequestBuilderStore((state) => state.send)
  const dismiss = useRequestBuilderStore((state) => state.dismissMissingSecrets)
  const openPassphrase = useEnvironmentsStore((state) => state.openPassphrase)
  const close = () => dismiss(tabId)

  function sendAnyway() {
    dismiss(tabId)
    void send(tabId, true)
  }

  function unlock() {
    dismiss(tabId)
    openPassphrase('unlock')
  }

  return (
    <Dialog.Root open={missingSecrets !== null} onOpenChange={(open) => (open ? undefined : close())}>
      {missingSecrets ? (
        <>
          <Dialog.Title>Send without some secrets?</Dialog.Title>
          <Dialog.Body>
            <p>This request uses secrets that have no value right now. If you send anyway, their placeholders are sent as written.</p>
            {missingSecrets.missing.length > 0 ? <SecretList title="No value on this machine" names={missingSecrets.missing} /> : null}
            {missingSecrets.locked.length > 0 ? <SecretList title="In a locked vault" names={missingSecrets.locked} /> : null}
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" onClick={close}>
              Cancel
            </Button>
            {missingSecrets.locked.length > 0 ? (
              <Button variant="secondary" onClick={unlock}>
                Unlock Vault
              </Button>
            ) : null}
            <Button variant="primary" onClick={sendAnyway}>
              Send Anyway
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
