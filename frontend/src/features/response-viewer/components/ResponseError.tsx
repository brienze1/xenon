import type { SendFailure } from '@/platform/mappers/request'
import styles from './ResponseError.module.css'

const TITLES: Record<SendFailure['kind'], string> = {
  dns: 'Could not resolve host',
  connection_refused: 'Connection refused',
  timeout: 'Request timed out',
  tls: 'SSL/TLS error',
  cancelled: 'Request cancelled',
  missing_file: 'Body file missing',
  invalid_url: 'Invalid URL',
  other: 'Could not send request',
}

interface ResponseErrorProps {
  failure: SendFailure
}

export function ResponseError({ failure }: ResponseErrorProps) {
  return (
    <section role="alert" aria-label="Request error" className={styles.error}>
      <h2 className={styles.title}>{TITLES[failure.kind]}</h2>
      <p className={styles.message}>{failure.message}</p>
      {failure.hint ? <p className={styles.hint}>{failure.hint}</p> : null}
    </section>
  )
}
