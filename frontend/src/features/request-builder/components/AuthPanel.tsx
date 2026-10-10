import type { AuthType } from '@/platform/mappers/request'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import styles from './AuthPanel.module.css'

interface AuthPanelProps {
  tabId: string
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

const INHERIT_NOTE = <PanelSection.Note>The request uses the auth of its folder or collection. An unsaved request has none.</PanelSection.Note>

const NO_AUTH_NOTE = <PanelSection.Note>This request does not use any authorization.</PanelSection.Note>

export function AuthPanel({ tabId }: AuthPanelProps) {
  const auth = useRequestBuilderStore((state) => state.tabs[tabId]?.draft.auth)
  const updateAuth = useRequestBuilderStore((state) => state.updateAuth)
  if (!auth) {
    return null
  }

  return (
    <PanelSection.Root>
      <label className={styles.field}>
        Auth type
        <select
          name="auth-type"
          className={styles.select}
          value={auth.type}
          onChange={(event) => updateAuth(tabId, { type: event.target.value as AuthType })}
        >
          {AUTH_TYPE_OPTIONS}
          {auth.type === 'unsupported' ? UNSUPPORTED_OPTION : null}
        </select>
      </label>
      {auth.type === 'unsupported' ? (
        <PanelSection.Note>
          Unsupported auth type (kept): <span translate="no">{auth.unsupported?.type ?? 'unknown'}</span>. It is kept unchanged for export, and the
          request is sent without auth until <span translate="no">Xenon</span> supports it.
        </PanelSection.Note>
      ) : null}
      {auth.type === 'inherit' ? INHERIT_NOTE : null}
      {auth.type === 'none' ? NO_AUTH_NOTE : null}
      {auth.type === 'bearer' ? (
        <label className={styles.field}>
          Token
          <input
            name="bearer-token"
            autoComplete="off"
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
              name="basic-username"
              autoComplete="off"
              className={styles.input}
              spellCheck={false}
              value={auth.basic.username}
              onChange={(event) => updateAuth(tabId, { basic: { ...auth.basic, username: event.target.value } })}
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
              onChange={(event) => updateAuth(tabId, { basic: { ...auth.basic, password: event.target.value } })}
            />
          </label>
        </>
      ) : null}
    </PanelSection.Root>
  )
}
