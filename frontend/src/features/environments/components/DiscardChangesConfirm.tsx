import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import styles from './DiscardChangesConfirm.module.css'

interface DiscardChangesConfirmProps {
  onKeepEditing: () => void
  onDiscard: () => void
}

/** Inline confirmation before unsaved variable table edits are dropped; inline because the tables live inside dialogs. */
export function DiscardChangesConfirm({ onKeepEditing, onDiscard }: DiscardChangesConfirmProps) {
  return (
    <Banner.Root tone="warning" className={styles.confirm}>
      <Banner.Message>Discard unsaved variable changes?</Banner.Message>
      <Banner.Actions>
        <Button variant="secondary" size="sm" autoFocus onClick={onKeepEditing}>
          Keep Editing
        </Button>
        <Button variant="danger" size="sm" onClick={onDiscard}>
          Discard
        </Button>
      </Banner.Actions>
    </Banner.Root>
  )
}
