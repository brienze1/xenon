import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import type { TypeSwitches } from '@/features/environments/lib/variable-draft'
import styles from './TypeSwitchConfirm.module.css'

interface TypeSwitchConfirmProps {
  switches: TypeSwitches
  busy: boolean
  onCancel: () => void
  onConfirm: () => void
}

function names(keys: string[]) {
  return keys.map((key, index) => (
    <span key={key}>
      {index > 0 ? ', ' : null}
      <code translate="no">{key}</code>
    </span>
  ))
}

/**
 * Inline confirmation before a save that changes a row's type. It stays inline (not a nested
 * dialog) because the table itself lives inside dialogs.
 */
export function TypeSwitchConfirm({ switches, busy, onCancel, onConfirm }: TypeSwitchConfirmProps) {
  const reveals = switches.reveal
  return (
    <Banner.Root tone="warning" className={styles.confirm}>
      <Banner.Message>
        {switches.toSecret.length > 0 ? (
          <p>
            {names(switches.toSecret)} will become secret: the value moves into the vault on this machine and is removed from the
            workspace files. The old value may still be in git history.
          </p>
        ) : null}
        {reveals ? (
          <p>
            {switches.toPlain.length > 0 ? names(switches.toPlain) : 'A secret'} will become plain: the value will be written to the
            workspace files and shared with everyone who has the folder.
          </p>
        ) : null}
      </Banner.Message>
      <Banner.Actions>
        <Button variant="secondary" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant={reveals ? 'danger' : 'primary'} size="sm" disabled={busy} onClick={onConfirm}>
          {reveals ? 'Write Values to Files' : 'Make Secret'}
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
