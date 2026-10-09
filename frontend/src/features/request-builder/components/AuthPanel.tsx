import type { AuthType } from '@/platform/mappers/request'
import { useRequestBuilderStore } from '../store'
import styles from './Panel.module.css'

interface AuthPanelProps {
  tabId: string
}

const AUTH_OPTIONS: Array<{ type: AuthType; label: string; available: boolean }> = [
  { type: 'inherit', label: 'Inherit auth from parent', available: true },
  { type: 'none', label: 'No Auth', available: true },
  { type: 'bearer', label: 'Bearer Token', available: true },
  { type: 'basic', label: 'Basic Auth', available: true },
  { type: 'apikey', label: 'API Key (later stage)', available: false },
  { type: 'oauth2', label: 'OAuth 2.0 (later stage)', available: false },
  { type: 'awsv4', label: 'AWS Signature (later stage)', available: false },
  { type: 'digest', label: 'Digest Auth (later stage)', available: false },
]

export function AuthPanel({ tabId }: AuthPanelProps) {
  const auth = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.auth)
  const updateAuth = useRequestBuilderStore((state) => state.updateAuth)
  if (!auth) {
    return null
  }

  return (
    <section className={styles.section}>
      <label className={styles.field}>
        Auth Type
        <select
          className={styles.select}
          value={auth.type}
          onChange={(event) => updateAuth(tabId, { type: event.target.value as AuthType })}
        >
          {AUTH_OPTIONS.map((option) => (
            <option key={option.type} value={option.type} disabled={!option.available}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
      {auth.type === 'inherit' ? (
        <p className={styles.note}>The request uses the auth of its folder or collection. An unsaved request has none.</p>
      ) : null}
      {auth.type === 'none' ? <p className={styles.note}>This request does not use any authorization.</p> : null}
      {auth.type === 'bearer' ? (
        <label className={styles.field}>
          Token
          <input
            className={styles.input}
            spellCheck={false}
            value={auth.bearer.token}
            onChange={(event) => updateAuth(tabId, { bearer: { token: event.target.value } })}
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
              onChange={(event) => updateAuth(tabId, { basic: { ...auth.basic, username: event.target.value } })}
            />
          </label>
          <label className={styles.field}>
            Password
            <input
              type="password"
              className={styles.input}
              value={auth.basic.password}
              onChange={(event) => updateAuth(tabId, { basic: { ...auth.basic, password: event.target.value } })}
            />
          </label>
        </>
      ) : null}
    </section>
  )
}
