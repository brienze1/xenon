import { isLegacyGrant, usesBrowser, type OAuth2Auth, type OAuth2Grant } from '@/platform/mappers/request'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import styles from './AuthForm.module.css'
import { OAuth2Params } from './OAuth2Params'
import { TokenPanel } from './TokenPanel'

interface OAuth2FormProps {
  oauth2: OAuth2Auth
  ownerId: string
  onChange: (oauth2: OAuth2Auth) => void
}

interface TextFieldProps {
  label: string
  name: string
  value: string
  placeholder?: string
  secret?: boolean
  onChange: (value: string) => void
}

const GRANT_LABELS: Record<OAuth2Grant, string> = {
  authorization_code: 'Authorization Code',
  authorization_code_with_pkce: 'Authorization Code (With PKCE)',
  implicit: 'Implicit',
  password_credentials: 'Password Credentials',
  client_credentials: 'Client Credentials',
}

const GRANTS = Object.keys(GRANT_LABELS) as OAuth2Grant[]

function TextField({ label, name, value, placeholder, secret = false, onChange }: TextFieldProps) {
  return (
    <label className={styles.field}>
      {label}
      <input
        name={name}
        type={secret ? 'password' : 'text'}
        autoComplete={secret ? 'new-password' : 'off'}
        spellCheck={false}
        className={styles.input}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  )
}

export function OAuth2Form({ oauth2, ownerId, onChange }: OAuth2FormProps) {
  const set = <K extends keyof OAuth2Auth>(key: K) => (value: OAuth2Auth[K]) => onChange({ ...oauth2, [key]: value })
  const grant = oauth2.grantType
  const browser = usesBrowser(grant)
  const pkce = grant === 'authorization_code_with_pkce'
  return (
    <div className={styles.form}>
      <TokenPanel ownerId={ownerId} oauth2={oauth2} />
      <h3 className={styles.note}>Configure New Token</h3>
      {oauth2.usesPostmanCallback ? (
        <Banner.Root tone="warning">
          <Banner.Message>
            Postman&apos;s web callback cannot reach <span translate="no">Xenon</span>. Register this callback URL with your provider:{' '}
            <code translate="no">{oauth2.suggestedCallback}</code>
          </Banner.Message>
          <Banner.Actions>
            <Button size="sm" variant="secondary" onClick={() => onChange({ ...oauth2, redirectUri: oauth2.suggestedCallback, usesPostmanCallback: false, suggestedCallback: '' })}>
              Use This Callback
            </Button>
          </Banner.Actions>
        </Banner.Root>
      ) : null}
      <label className={styles.field}>
        <span>
          Grant type
          {isLegacyGrant(grant) ? <span className={styles.badge}>Legacy</span> : null}
        </span>
        <select name="oauth2-grant" className={styles.select} value={grant} onChange={(event) => set('grantType')(event.target.value as OAuth2Grant)}>
          {GRANTS.map((value) => (
            <option key={value} value={value}>
              {GRANT_LABELS[value]}
              {isLegacyGrant(value) ? ' (legacy)' : ''}
            </option>
          ))}
        </select>
      </label>
      <TextField label="Token name" name="oauth2-token-name" value={oauth2.tokenName} onChange={set('tokenName')} />
      {browser ? (
        <>
          <TextField label="Callback URL" name="oauth2-callback" value={oauth2.redirectUri} placeholder="http://127.0.0.1:53682/callback…" onChange={set('redirectUri')} />
          <p className={styles.note}>
            The provider must allow this loopback address; <span translate="no">Xenon</span> catches the redirect there after you sign in in your browser.
          </p>
          <TextField label="Auth URL" name="oauth2-auth-url" value={oauth2.authUrl} placeholder="https://…/authorize" onChange={set('authUrl')} />
        </>
      ) : null}
      {grant === 'implicit' ? null : (
        <TextField label="Access token URL" name="oauth2-token-url" value={oauth2.accessTokenUrl} placeholder="https://…/token" onChange={set('accessTokenUrl')} />
      )}
      <div className={styles.row}>
        <TextField label="Client ID" name="oauth2-client-id" value={oauth2.clientId} onChange={set('clientId')} />
        <TextField label="Client secret" name="oauth2-client-secret" value={oauth2.clientSecret} placeholder="{{client_secret}}…" secret onChange={set('clientSecret')} />
      </div>
      {grant === 'password_credentials' ? (
        <div className={styles.row}>
          <TextField label="Username" name="oauth2-username" value={oauth2.username} onChange={set('username')} />
          <TextField label="Password" name="oauth2-password" value={oauth2.password} secret onChange={set('password')} />
        </div>
      ) : null}
      {pkce ? (
        <div className={styles.row}>
          <label className={styles.field}>
            Code challenge method
            <select name="oauth2-challenge" className={styles.select} value={oauth2.challengeAlgorithm} onChange={(event) => set('challengeAlgorithm')(event.target.value === 'plain' ? 'plain' : 'S256')}>
              <option value="S256">SHA-256</option>
              <option value="plain">Plain</option>
            </select>
          </label>
          <TextField label="Code verifier" name="oauth2-verifier" value={oauth2.codeVerifier} placeholder="Generated when empty…" onChange={set('codeVerifier')} />
        </div>
      ) : null}
      <div className={styles.row}>
        <TextField label="Scope" name="oauth2-scope" value={oauth2.scope} placeholder="openid profile…" onChange={set('scope')} />
        {browser ? <TextField label="State" name="oauth2-state" value={oauth2.state} placeholder="Random when empty…" onChange={set('state')} /> : null}
      </div>
      <label className={styles.field}>
        Client authentication
        <select
          name="oauth2-client-authentication"
          className={styles.select}
          value={oauth2.clientAuthentication}
          onChange={(event) => set('clientAuthentication')(event.target.value === 'body' ? 'body' : 'header')}
        >
          <option value="header">Send as Basic auth header</option>
          <option value="body">Send client credentials in body</option>
        </select>
      </label>
      <div className={styles.row}>
        <label className={styles.field}>
          Add auth data to
          <select name="oauth2-add-token-to" className={styles.select} value={oauth2.addTokenTo} onChange={(event) => set('addTokenTo')(event.target.value === 'queryParams' ? 'queryParams' : 'header')}>
            <option value="header">Request headers</option>
            <option value="queryParams">Request URL</option>
          </select>
        </label>
        {oauth2.addTokenTo === 'header' ? <TextField label="Header prefix" name="oauth2-header-prefix" value={oauth2.headerPrefix} onChange={set('headerPrefix')} /> : null}
      </div>
      <label className={styles.check}>
        <input type="checkbox" name="oauth2-auto-refresh" checked={oauth2.autoRefresh} onChange={(event) => set('autoRefresh')(event.target.checked)} />
        Refresh the token automatically when it expires
      </label>
      <details className={styles.advanced}>
        <summary>Advanced</summary>
        <TextField label="Refresh token URL" name="oauth2-refresh-url" value={oauth2.refreshTokenUrl} placeholder="Access token URL when empty…" onChange={set('refreshTokenUrl')} />
        <div className={styles.row}>
          <TextField label="Audience" name="oauth2-audience" value={oauth2.audience} onChange={set('audience')} />
          <TextField label="Resource" name="oauth2-resource" value={oauth2.resource} onChange={set('resource')} />
        </div>
        {browser ? (
          <OAuth2Params label="Auth request" name="oauth2-auth-params" params={oauth2.authRequestParams} targets={['request_url']} onChange={set('authRequestParams')} />
        ) : null}
        <OAuth2Params label="Token request" name="oauth2-token-params" params={oauth2.tokenRequestParams} targets={['request_body', 'request_url', 'request_header']} onChange={set('tokenRequestParams')} />
        <OAuth2Params label="Refresh request" name="oauth2-refresh-params" params={oauth2.refreshRequestParams} targets={['request_body', 'request_url', 'request_header']} onChange={set('refreshRequestParams')} />
      </details>
    </div>
  )
}
