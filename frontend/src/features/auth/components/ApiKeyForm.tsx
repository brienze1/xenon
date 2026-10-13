import type { ApiKeyAuth, ApiKeyLocation } from '@/platform/mappers/request'
import styles from './AuthForm.module.css'

interface ApiKeyFormProps {
  apikey: ApiKeyAuth
  onChange: (apikey: ApiKeyAuth) => void
}

export function ApiKeyForm({ apikey, onChange }: ApiKeyFormProps) {
  return (
    <div className={styles.form}>
      <div className={styles.row}>
        <label className={styles.field}>
          Key
          <input
            name="apikey-key"
            autoComplete="off"
            spellCheck={false}
            className={styles.input}
            value={apikey.key}
            placeholder="X-Api-Key…"
            onChange={(event) => onChange({ ...apikey, key: event.target.value })}
          />
        </label>
        <label className={styles.field}>
          Value
          <input
            name="apikey-value"
            autoComplete="off"
            spellCheck={false}
            className={styles.input}
            value={apikey.value}
            placeholder="{{api_key}}…"
            onChange={(event) => onChange({ ...apikey, value: event.target.value })}
          />
        </label>
      </div>
      <label className={styles.field}>
        Add to
        <select name="apikey-in" className={styles.select} value={apikey.in} onChange={(event) => onChange({ ...apikey, in: event.target.value as ApiKeyLocation })}>
          <option value="header">Header</option>
          <option value="query">Query params</option>
        </select>
      </label>
      <p className={styles.note}>A header with the same name (in any case) or a query parameter with the same key is replaced.</p>
    </div>
  )
}
