import { useEffect } from 'react'
import { Button } from '@/shared/ui/button/Button'
import { Toast } from '@/shared/ui/toast/Toast'
import { FINISHED_LABELS } from '@/features/runner/lib/labels'
import type { RunFinishedNotice } from '@/features/runner/model'
import { useRunnerStore } from '@/features/runner/store'

/** The finish toast closes itself after this long. */
export const RUN_TOAST_MS = 8000

interface RunFinishedToastProps {
  /** Focuses (or opens) the target's tab on Runner › Results for that run; composed by the app. */
  onOpenResults: (notice: RunFinishedNotice) => void
}

/** "<target> run completed · N passed · M failed" when a run ends, with Open results (FR-001b). */
export function RunFinishedToast({ onOpenResults }: RunFinishedToastProps) {
  const notice = useRunnerStore((state) => state.notice)
  const dismiss = useRunnerStore((state) => state.dismissNotice)

  useEffect(() => {
    if (notice === null) {
      return
    }
    const timer = setTimeout(dismiss, RUN_TOAST_MS)
    return () => clearTimeout(timer)
  }, [notice, dismiss])

  if (notice === null) {
    return null
  }

  function open() {
    if (notice !== null) {
      dismiss()
      onOpenResults(notice)
    }
  }

  return (
    <Toast.Root closeLabel="Close" onClose={dismiss}>
      <Toast.Message>
        <span translate="no">{notice.targetName}</span> run {FINISHED_LABELS[notice.status]} · {notice.passed} passed · {notice.failed} failed
      </Toast.Message>
      <Toast.Actions>
        <Button variant="primary" size="sm" onClick={open}>
          Open results
        </Button>
      </Toast.Actions>
    </Toast.Root>
  )
}
