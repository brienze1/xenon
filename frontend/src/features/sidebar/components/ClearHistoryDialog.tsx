import { useState, useTransition, type ReactNode } from 'react'
import type { AppError } from '@/platform/errors'
import type { HistorySourceFilter } from '@/platform/mappers/history'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { CLEAR_SCOPES } from '@/features/sidebar/lib/history-labels'
import { SidebarErrorText } from './SidebarErrorText'
import styles from './ClearHistoryDialog.module.css'

const CLEAR_INTROS: Record<HistorySourceFilter, ReactNode> = {
  all: (
    <>
      <span translate="no">Xenon</span> removes every entry in this workspace’s history, yours and the AI’s, from this machine. You can’t undo this.
    </>
  ),
  user: (
    <>
      <span translate="no">Xenon</span> removes every request you sent in this workspace from this machine. AI entries stay. You can’t undo this.
    </>
  ),
  ai: (
    <>
      <span translate="no">Xenon</span> removes every AI send and action in this workspace from this machine. Your own sends stay. You can’t undo this.
    </>
  ),
}

const CLEAR_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'

interface ClearHistoryDialogProps {
  open: boolean
  /** The source filter in effect: only its entries are cleared. */
  source: HistorySourceFilter
  onCancel: () => void
  onConfirm: () => Promise<AppError | null>
}

export function ClearHistoryDialog({ open, source, onCancel, onConfirm }: ClearHistoryDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(next) => (next ? undefined : onCancel())}>
      {open ? <ClearHistoryConfirm source={source} onCancel={onCancel} onConfirm={onConfirm} /> : null}
    </Dialog.Root>
  )
}

interface ClearHistoryConfirmProps {
  source: HistorySourceFilter
  onCancel: () => void
  onConfirm: () => Promise<AppError | null>
}

function ClearHistoryConfirm({ source, onCancel, onConfirm }: ClearHistoryConfirmProps) {
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
      <Dialog.Title>{CLEAR_SCOPES[source].title}</Dialog.Title>
      <Dialog.Body>
        {CLEAR_INTROS[source]}
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
          {clearing ? 'Clearing…' : CLEAR_SCOPES[source].action}
        </Button>
      </Dialog.Actions>
    </>
  )
}
