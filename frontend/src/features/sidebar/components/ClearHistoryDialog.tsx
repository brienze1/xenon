import { useState, useTransition } from 'react'
import type { AppError } from '@/platform/errors'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './ClearHistoryDialog.module.css'

const CLEAR_HISTORY_TITLE = <Dialog.Title>Clear History?</Dialog.Title>

const CLEAR_HISTORY_INTRO = (
  <>
    <span translate="no">Xenon</span> removes every entry you sent in this workspace from this machine. You can’t undo this.
  </>
)

const CLEAR_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

interface ClearHistoryDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => Promise<AppError | null>
}

export function ClearHistoryDialog({ open, onCancel, onConfirm }: ClearHistoryDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onCancel())}>
      {open ? <ClearHistoryConfirm onCancel={onCancel} onConfirm={onConfirm} /> : null}
    </Dialog.Root>
  )
}

interface ClearHistoryConfirmProps {
  onCancel: () => void
  onConfirm: () => Promise<AppError | null>
}

function ClearHistoryConfirm({ onCancel, onConfirm }: ClearHistoryConfirmProps) {
  const [clearing, startClear] = useTransition()
  const [error, setError] = useState<AppError | null>(null)

  function confirm() {
    startClear(async () => {
      const failure = await onConfirm()
      setError(failure)
    })
  }

  return (
    <>
      {CLEAR_HISTORY_TITLE}
      <Dialog.Body>
        {CLEAR_HISTORY_INTRO}
        {error ? (
          <Banner.Root tone="error" className={styles.error}>
            <Banner.Message>
              <SidebarErrorText error={error} nextStep={CLEAR_NEXT_STEP} />
            </Banner.Message>
          </Banner.Root>
        ) : null}
      </Dialog.Body>
      <Dialog.Actions>
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="danger" disabled={clearing} onClick={confirm}>
          {clearing ? 'Clearing…' : 'Clear History'}
        </Button>
      </Dialog.Actions>
    </>
  )
}
