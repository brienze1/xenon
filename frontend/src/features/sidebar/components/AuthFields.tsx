import type { AuthType, RequestAuth } from '@/platform/mappers/request'
import styles from './AuthFields.module.css'

interface AuthFieldsProps {
  kind: 'collection' | 'folder'
  auth: RequestAuth
  onChange: (auth: RequestAuth) => void
}

const AUTH_TYPE_OPTIONS = (
  <>
    <option value="inherit">Inherit auth from parent</option>
    <option value="none">No auth</option>
    <option value="bearer" translate="no">
      Bearer token
    </option>
    <option value="basic" translate="no">
      Basic auth
    </option>
    <optgroup label="Later stage">
      <option value="apikey" translate="no" disabled>
        API key
      </option>
      <option value="oauth2" translate="no" disabled>
        OAuth 2.0
      </option>
      <option value="awsv4" translate="no" disabled>
        AWS signature
      </option>
      <option value="digest" translate="no" disabled>
        Digest auth
      </option>
    </optgroup>
  </>
)

const UNSUPPORTED_OPTION = (
  <option value="unsupported" disabled>
    Unsupported (kept)
  </option>
)

export function AuthFields({ kind, auth, onChange }: AuthFieldsProps) {
  return (
    <>
      <label className={styles.field}>
        Auth type
        <select
          name="auth-type"
          className={styles.input}
          value={auth.type}
          onChange={(event) => onChange({ ...auth, type: event.target.value as AuthType })}
        >
          {AUTH_TYPE_OPTIONS}
          {auth.type === 'unsupported' ? UNSUPPORTED_OPTION : null}
        </select>
      </label>
      {auth.type === 'unsupported' ? (
        <p className={styles.note}>
          Unsupported auth type (kept): <span translate="no">{auth.unsupported?.type ?? 'unknown'}</span>. It is kept unchanged for export, and requests
          are sent without auth until <span translate="no">Xenon</span> supports it.
        </p>
      ) : null}
      {auth.type === 'inherit' ? (
        <p className={styles.note}>
          {kind === 'collection' ? 'A collection has no parent, so its requests use no auth.' : 'Uses the auth of the folder or collection above.'}
        </p>
      ) : null}
      {auth.type === 'bearer' ? (
        <label className={styles.field}>
          Token
          <input
            name="bearer-token"
            autoComplete="off"
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
              name="basic-username"
              autoComplete="off"
              className={styles.input}
              spellCheck={false}
              value={auth.basic.username}
              onChange={(event) => onChange({ ...auth, basic: { ...auth.basic, username: event.target.value } })}
            />
          </label>
          <label className={styles.field}>
            Password
            <input
              name="basic-password"
              autoComplete="new-password"
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
