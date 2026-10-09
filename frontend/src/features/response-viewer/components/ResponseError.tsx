import type { SendFailure } from '@/platform/mappers/request'
import { NoTranslateText } from '@/shared/ui/no-translate/NoTranslateText'
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

const NEXT_STEPS: Record<SendFailure['kind'], string> = {
  dns: 'Check the host name in the URL and your network connection, then send again.',
  connection_refused: 'Check that the server is running and listening on that port, then send again.',
  timeout: 'Increase the timeout in the request Settings, or check that the server responds, then send again.',
  tls: 'Check the server certificate, or turn off SSL certificate verification in Settings, then send again.',
  cancelled: 'Send the request again when you are ready.',
  missing_file: 'Select the file again in the Body tab, then send again.',
  invalid_url: 'Fix the URL in the request bar, then send again.',
  other: 'Check the request and your connection, then send again.',
}

const TECHNICAL_TERMS: readonly string[] = ['SSL/TLS', 'SSL']

interface ResponseErrorProps {
  failure: SendFailure
}

export function ResponseError({ failure }: ResponseErrorProps) {
  return (
    <section role="alert" aria-label="Request error" className={styles.error}>
      <h3 className={styles.title}>
        <NoTranslateText text={TITLES[failure.kind]} terms={TECHNICAL_TERMS} />
      </h3>
      <p className={styles.message}>{failure.message}</p>
      <p className={styles.hint}>
        <NoTranslateText text={failure.hint || NEXT_STEPS[failure.kind]} terms={TECHNICAL_TERMS} />
      </p>
    </section>
  )
}
