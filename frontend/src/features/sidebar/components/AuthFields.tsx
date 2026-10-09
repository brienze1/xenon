import type { AuthType, RequestAuth } from '@/platform/mappers/request'
import styles from './ItemSettingsDialog.module.css'

interface AuthFieldsProps {
  kind: 'collection' | 'folder'
  auth: RequestAuth
  onChange: (auth: RequestAuth) => void
}

const AUTH_OPTIONS: Array<{ type: AuthType; label: string }> = [
  { type: 'inherit', label: 'Inherit auth from parent' },
  { type: 'none', label: 'No Auth' },
  { type: 'bearer', label: 'Bearer Token' },
  { type: 'basic', label: 'Basic Auth' },
]

export function AuthFields({ kind, auth, onChange }: AuthFieldsProps) {
  return (
    <>
      <label className={styles.field}>
        Auth Type
        <select
          className={styles.input}
          value={auth.type}
          onChange={(event) => onChange({ ...auth, type: event.target.value as AuthType })}
        >
          {AUTH_OPTIONS.map((option) => (
            <option key={option.type} value={option.type}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
      {auth.type === 'inherit' ? (
        <p className={styles.note}>
          {kind === 'collection' ? 'A collection has no parent, so its requests use no auth.' : 'Uses the auth of the folder or collection above.'}
        </p>
      ) : null}
      {auth.type === 'bearer' ? (
        <label className={styles.field}>
          Token
          <input
            className={styles.input}
            spellCheck={false}
            value={auth.bearer.token}
            onChange={(event) => onChange({ ...auth, bearer: { token: event.target.value } })}
          />
        </label>
      ) : null}
      {auth.type === 'basic' ? (
        <>
          <label className={styles.field}>
            Username
            <input
              className={styles.input}
              spellCheck={false}
              value={auth.basic.username}
              onChange={(event) => onChange({ ...auth, basic: { ...auth.basic, username: event.target.value } })}
            />
          </label>
          <label className={styles.field}>
            Password
            <input
              type="password"
              className={styles.input}
              value={auth.basic.password}
              onChange={(event) => onChange({ ...auth, basic: { ...auth.basic, password: event.target.value } })}
            />
          </label>
        </>
      ) : null}
    </>
  )
}
