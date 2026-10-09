import type { SendFailure } from '@/platform/mappers/request'
import styles from './ResponseError.module.css'

const TITLES: Record<SendFailure['kind'], string> = {
  dns: 'Could Not Resolve Host',
  connection_refused: 'Connection Refused',
  timeout: 'Request Timed Out',
  tls: 'SSL/TLS Error',
  cancelled: 'Request Cancelled',
  missing_file: 'Body File Missing',
  invalid_url: 'Invalid URL',
  other: 'Could Not Send Request',
}

interface ResponseErrorProps {
  failure: SendFailure
}

export function ResponseError({ failure }: ResponseErrorProps) {
  return (
    <section role="alert" aria-label="Request error" className={styles.error}>
      <h3 className={styles.title}>{TITLES[failure.kind]}</h3>
      <p className={styles.message}>{failure.message}</p>
      {failure.hint ? <p className={styles.hint}>{failure.hint}</p> : null}
    </section>
  )
}
