import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { useEnvironmentsStore } from '@/features/environments/store'
import { useRunnerStore } from '@/features/runner/store'
import styles from './RunMissingSecretsDialog.module.css'

const TITLE = <Dialog.Title>Run Without Some Secrets?</Dialog.Title>

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

interface RunMissingSecretsDialogProps {
  itemId: string
}

/** The run uses secrets with no value (SND-06409): Run anyway restarts with them acknowledged, as Send anyway does. */
export function RunMissingSecretsDialog({ itemId }: RunMissingSecretsDialogProps) {
  const missingSecrets = useRunnerStore((state) => state.targets[itemId]?.missingSecrets ?? null)
  const runAnyway = useRunnerStore((state) => state.runAnyway)
  const dismiss = useRunnerStore((state) => state.dismissMissingSecrets)
  const openPassphrase = useEnvironmentsStore((state) => state.openPassphrase)
  const close = () => dismiss(itemId)

  function unlock() {
    close()
    openPassphrase('unlock')
  }

  return (
    <Dialog.Root open={missingSecrets !== null} onOpenChange={(open) => (open ? undefined : close())}>
      {missingSecrets ? (
        <>
          {TITLE}
          <Dialog.Body>
            <p>Requests of this run use secrets that have no value right now. If you run anyway, their placeholders are sent as written.</p>
            {missingSecrets.missing.length > 0 ? <SecretList title="No Value on This Machine" names={missingSecrets.missing} /> : null}
            {missingSecrets.locked.length > 0 ? <SecretList title="In a Locked Vault" names={missingSecrets.locked} /> : null}
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
            <Button variant="primary" onClick={() => void runAnyway(itemId)}>
              Run Anyway
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
