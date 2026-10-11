import { useEffect } from 'react'
import type { OAuth2Auth, TokenUse } from '@/platform/mappers/request'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { emptyTokenState, useAuthStore } from '@/features/auth/store'
import styles from './TokenPanel.module.css'

interface TokenPanelProps {
  ownerId: string
  oauth2: OAuth2Auth
}

const MASK = '•••••'
const EXPIRY_FORMAT = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' })

function expiryText(expiresAt: string | null): string {
  if (!expiresAt) {
    return 'No expiry'
  }
  return EXPIRY_FORMAT.format(new Date(expiresAt))
}

export function TokenPanel({ ownerId, oauth2 }: TokenPanelProps) {
  const token = useAuthStore((state) => state.tokens[ownerId]) ?? emptyTokenState()
  const load = useAuthStore((state) => state.load)
  const getNew = useAuthStore((state) => state.getNew)
  const cancel = useAuthStore((state) => state.cancel)
  const refresh = useAuthStore((state) => state.refresh)
  const use = useAuthStore((state) => state.use)
  const clear = useAuthStore((state) => state.clear)
  const toggleReveal = useAuthStore((state) => state.toggleReveal)
  const copy = useAuthStore((state) => state.copy)

  useEffect(() => {
    void load(ownerId)
  }, [load, ownerId])

  const info = token.info
  return (
    <section className={styles.panel} aria-label="Access token">
      <header className={styles.header}>
        <h3 className={styles.title}>Current Token</h3>
        {token.waiting ? (
          <Button variant="secondary" size="sm" onClick={() => cancel(ownerId)}>
            Cancel
          </Button>
        ) : (
          <Button variant="primary" size="sm" disabled={token.busy} onClick={() => void getNew(ownerId, oauth2)}>
            Get New Access Token
          </Button>
        )}
      </header>
      {token.waiting ? (
        <p className={styles.waiting} role="status" aria-live="polite">
          Waiting for the sign-in in your browser…
        </p>
      ) : null}
      {token.error ? (
        <Banner.Root tone="error">
          <Banner.Message>
            {token.error.description}
            {token.error.details.map((detail) => (
              <span key={detail.attribute} className={styles.detail}>
                {detail.attribute}: {detail.messages.join(', ')}
              </span>
            ))}
          </Banner.Message>
        </Banner.Root>
      ) : null}
      {info ? (
        <>
          <dl className={styles.facts}>
            <dt>Token</dt>
            <dd className={styles.value} translate="no">
              {token.revealed ?? MASK}
            </dd>
            <dt>Type</dt>
            <dd translate="no">{info.tokenType || 'unknown'}</dd>
            <dt>Expires</dt>
            <dd>
              {expiryText(info.expiresAt)}
              {info.expired ? <span className={styles.expired}>Expired</span> : null}
            </dd>
            <dt>Scopes</dt>
            <dd translate="no">{info.scope || 'none'}</dd>
          </dl>
          {info.tokenType.toLowerCase() === 'mac' ? <p className={styles.note}>A mac token is not sent with requests, as in Postman.</p> : null}
          <div className={styles.actions}>
            <Button size="sm" variant="ghost" onClick={() => void toggleReveal(ownerId)}>
              {token.revealed === null ? 'Reveal' : 'Hide'}
            </Button>
            <Button size="sm" variant="ghost" onClick={() => void copy(ownerId)}>
              {token.copied ? 'Copied' : 'Copy'}
            </Button>
            <Button size="sm" variant="secondary" disabled={!info.hasRefreshToken || token.busy} onClick={() => void refresh(ownerId, oauth2)}>
              Refresh
            </Button>
            {info.hasIdToken ? (
              <label className={styles.use}>
                Use
                <select
                  name="use-token"
                  className={styles.select}
                  value={info.useToken}
                  onChange={(event) => void use(ownerId, event.target.value as TokenUse)}
                >
                  <option value="access_token">Access token</option>
                  <option value="id_token">ID token</option>
                </select>
              </label>
            ) : null}
            <Button size="sm" variant="danger" onClick={() => void clear(ownerId)}>
              Clear
            </Button>
          </div>
        </>
      ) : token.waiting ? null : (
        <p className={styles.note}>No token yet. It is kept in the workspace vault, never in the workspace files.</p>
      )}
    </section>
  )
}
