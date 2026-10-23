import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dialog } from '@/shared/ui/dialog/Dialog'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { useAIAccessStore } from '@/features/ai-access/store'
import styles from './ApprovalPrompt.module.css'

const TITLE = <Dialog.Title>Approve AI Request?</Dialog.Title>

const ANSWER_NEXT_STEP = 'Try again. If no answer arrives within 2 minutes, the request is denied.'

/** Approve / Deny for an AI send to a "Confirm before send" environment; one at a time, oldest first. */
export function ApprovalPrompt() {
  const head = useAIAccessStore((state) => state.pending[0] ?? null)
  const waiting = useAIAccessStore((state) => state.pending.length - 1)
  const answering = useAIAccessStore((state) => state.answering)
  const error = useAIAccessStore((state) => state.error)
  const answer = useAIAccessStore((state) => state.answer)
  const busy = head !== null && answering === head.id

  return (
    <Dialog.Root open={head !== null} onOpenChange={(open) => (open || head === null ? undefined : void answer(head.id, false))}>
      {head ? (
        <>
          {TITLE}
          <Dialog.Body>
            <p className={styles.intro}>
              <strong translate="no">{head.client || 'An AI client'}</strong> wants to {head.summary ? 'run this' : 'send this request'} with the environment{' '}
              <span className={styles.environment}>
                <span className={styles.chip} data-color={head.environment.color} aria-hidden="true" />
                <strong translate="no">{head.environment.name}</strong>
              </span>
              , which asks for confirmation before every send.
            </p>
            {head.summary ? (
              <p className={styles.request} translate="no">
                {head.summary}
              </p>
            ) : (
              <p className={styles.request}>
                <MethodBadge method={head.method} className={styles.method} />
                <span className={styles.url} translate="no">
                  {head.url}
                </span>
              </p>
            )}
            {waiting > 0 ? <p className={styles.more}>+{waiting} more waiting</p> : null}
            {error ? (
              <Banner.Root tone="error" className={styles.error}>
                <Banner.Message>{withNextStep(error.description || error.code, ANSWER_NEXT_STEP)}</Banner.Message>
              </Banner.Root>
            ) : null}
          </Dialog.Body>
          <Dialog.Actions>
            <Button variant="secondary" disabled={busy} onClick={() => void answer(head.id, false)}>
              Deny
            </Button>
            <Button variant="primary" disabled={busy} onClick={() => void answer(head.id, true)}>
              Approve
            </Button>
          </Dialog.Actions>
        </>
      ) : null}
    </Dialog.Root>
  )
}
