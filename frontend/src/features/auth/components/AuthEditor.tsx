import type { AuthType, RequestAuth } from '@/platform/mappers/request'
import { ApiKeyForm } from './ApiKeyForm'
import styles from './AuthForm.module.css'
import { AwsV4Form } from './AwsV4Form'
import { DigestForm } from './DigestForm'
import { OAuth2Form } from './OAuth2Form'

export type AuthOwnerKind = 'request' | 'collection' | 'folder'

interface AuthEditorProps {
  auth: RequestAuth
  /** The item whose auth this is (or the tab of an unsaved request): OAuth 2.0 tokens are kept per owner. */
  ownerId: string
  kind: AuthOwnerKind
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
    <option value="apikey" translate="no">
      API key
    </option>
    <option value="oauth2" translate="no">
      OAuth 2.0
    </option>
    <option value="awsv4" translate="no">
      AWS signature
    </option>
    <option value="digest" translate="no">
      Digest auth
    </option>
  </>
)

const UNSUPPORTED_OPTION = (
  <option value="unsupported" disabled>
    Unsupported (kept)
  </option>
)

const INHERIT_NOTES: Record<AuthOwnerKind, string> = {
  request: 'The request uses the auth of its folder or collection. An unsaved request has none.',
  collection: 'A collection has no parent, so its requests use no auth.',
  folder: 'Uses the auth of the folder or collection above.',
}

export function AuthEditor({ auth, ownerId, kind, onChange }: AuthEditorProps) {
  return (
    <div className={styles.form}>
      <label className={styles.field}>
        Auth type
        <select name="auth-type" className={styles.select} value={auth.type} onChange={(event) => onChange({ ...auth, type: event.target.value as AuthType })}>
          {AUTH_TYPE_OPTIONS}
          {auth.type === 'unsupported' ? UNSUPPORTED_OPTION : null}
        </select>
      </label>
      {auth.type === 'unsupported' ? (
        <p className={styles.note}>
          Unsupported auth type (kept): <span translate="no">{auth.unsupported?.type ?? 'unknown'}</span>. It is kept unchanged for export, and requests are
          sent without auth until <span translate="no">Xenon</span> supports it.
        </p>
      ) : null}
      {auth.type === 'inherit' ? <p className={styles.note}>{INHERIT_NOTES[kind]}</p> : null}
      {auth.type === 'none' ? <p className={styles.note}>No authorization is sent.</p> : null}
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
        <div className={styles.row}>
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
        </div>
      ) : null}
      {auth.type === 'apikey' ? <ApiKeyForm apikey={auth.apikey} onChange={(apikey) => onChange({ ...auth, apikey })} /> : null}
      {auth.type === 'oauth2' ? <OAuth2Form oauth2={auth.oauth2} ownerId={ownerId} onChange={(oauth2) => onChange({ ...auth, oauth2 })} /> : null}
      {auth.type === 'awsv4' ? <AwsV4Form awsv4={auth.awsv4} onChange={(awsv4) => onChange({ ...auth, awsv4 })} /> : null}
      {auth.type === 'digest' ? <DigestForm digest={auth.digest} onChange={(digest) => onChange({ ...auth, digest })} /> : null}
    </div>
  )
}
