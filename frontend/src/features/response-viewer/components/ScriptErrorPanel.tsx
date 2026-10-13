import type { ScriptFailure, SendOutcome } from '@/platform/mappers/request'
import styles from './ScriptErrorPanel.module.css'

const TITLES: Partial<Record<SendOutcome, string>> = {
  stopped_by_script_error: 'Request Not Sent: a Pre-request Script Failed',
  skipped_by_script: 'Request Skipped by a Script',
}

const NEXT_STEPS: Partial<Record<SendOutcome, string>> = {
  stopped_by_script_error: 'Fix the script, then send again. The console has its logs.',
  skipped_by_script: 'A pre-request script called pm.execution.skipRequest(), so nothing was sent.',
}

function location(failure: ScriptFailure): string {
  return failure.line === null ? '' : ` (line ${failure.line}, column ${failure.column ?? 0})`
}

interface ScriptErrorPanelProps {
  outcome: SendOutcome
  errors: readonly ScriptFailure[]
}

/** The script outcome of a send: why nothing was sent, or which post-response scripts failed. */
export function ScriptErrorPanel({ outcome, errors }: ScriptErrorPanelProps) {
  const title = TITLES[outcome] ?? (errors.length === 1 ? 'A Script Failed' : 'Scripts Failed')
  return (
    <section role="alert" aria-label="Script outcome" className={styles.panel}>
      <h3 className={styles.title}>{title}</h3>
      {errors.length > 0 ? (
        <ul className={styles.errors}>
          {errors.map((failure, index) => (
            <li key={index} className={styles.error}>
              <span className={styles.script} translate="no">
                {failure.scriptLabel}
                {location(failure)}
              </span>
              <span className={styles.message}>{failure.message}</span>
            </li>
          ))}
        </ul>
      ) : null}
      {NEXT_STEPS[outcome] ? <p className={styles.hint}>{NEXT_STEPS[outcome]}</p> : null}
    </section>
  )
}
