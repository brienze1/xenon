import type { AwsV4Auth } from '@/platform/mappers/request'
import styles from './AuthForm.module.css'

interface AwsV4FormProps {
  awsv4: AwsV4Auth
  onChange: (awsv4: AwsV4Auth) => void
}

export function AwsV4Form({ awsv4, onChange }: AwsV4FormProps) {
  const text = (key: 'accessKey' | 'secretKey' | 'sessionToken' | 'region' | 'service') => (event: { target: { value: string } }) => onChange({ ...awsv4, [key]: event.target.value })
  return (
    <div className={styles.form}>
      <div className={styles.row}>
        <label className={styles.field}>
          Access key
          <input name="awsv4-access-key" autoComplete="off" spellCheck={false} className={styles.input} value={awsv4.accessKey} placeholder="{{aws_access_key}}…" onChange={text('accessKey')} />
        </label>
        <label className={styles.field}>
          Secret key
          <input
            name="awsv4-secret-key"
            type="password"
            autoComplete="new-password"
            className={styles.input}
            value={awsv4.secretKey}
            placeholder="{{aws_secret_key}}…"
            onChange={text('secretKey')}
          />
        </label>
      </div>
      <label className={styles.check}>
        <input type="checkbox" name="awsv4-query" checked={awsv4.addAuthDataToQuery} onChange={(event) => onChange({ ...awsv4, addAuthDataToQuery: event.target.checked })} />
        Add auth data to the query (presigned URL)
      </label>
      <details className={styles.advanced}>
        <summary>Advanced</summary>
        <div className={styles.row}>
          <label className={styles.field}>
            AWS region
            <input name="awsv4-region" autoComplete="off" spellCheck={false} className={styles.input} value={awsv4.region} placeholder="us-east-1…" onChange={text('region')} />
          </label>
          <label className={styles.field}>
            Service name
            <input name="awsv4-service" autoComplete="off" spellCheck={false} className={styles.input} value={awsv4.service} placeholder="execute-api…" onChange={text('service')} />
          </label>
        </div>
        <label className={styles.field}>
          Session token
          <input name="awsv4-session-token" type="password" autoComplete="new-password" className={styles.input} value={awsv4.sessionToken} onChange={text('sessionToken')} />
        </label>
      </details>
      <p className={styles.note}>The request is signed last, after every header is set. Form-data bodies are signed as an empty payload, as in Postman.</p>
    </div>
  )
}
