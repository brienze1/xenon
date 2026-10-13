import { DIGEST_ALGORITHMS, type DigestAlgorithm, type DigestAuth, type DigestQop } from '@/platform/mappers/request'
import styles from './AuthForm.module.css'

interface DigestFormProps {
  digest: DigestAuth
  onChange: (digest: DigestAuth) => void
}

type DigestText = 'username' | 'password' | 'realm' | 'nonce' | 'opaque' | 'nonceCount' | 'clientNonce'

export function DigestForm({ digest, onChange }: DigestFormProps) {
  const text = (key: DigestText) => (event: { target: { value: string } }) => onChange({ ...digest, [key]: event.target.value })
  return (
    <div className={styles.form}>
      <div className={styles.row}>
        <label className={styles.field}>
          Username
          <input name="digest-username" autoComplete="off" spellCheck={false} className={styles.input} value={digest.username} onChange={text('username')} />
        </label>
        <label className={styles.field}>
          Password
          <input name="digest-password" type="password" autoComplete="new-password" className={styles.input} value={digest.password} onChange={text('password')} />
        </label>
      </div>
      <label className={styles.check}>
        <input type="checkbox" name="digest-disable-retry" checked={digest.disableRetryRequest} onChange={(event) => onChange({ ...digest, disableRetryRequest: event.target.checked })} />
        Disable retrying the request (send without answering the challenge)
      </label>
      <details className={styles.advanced}>
        <summary>Advanced</summary>
        <p className={styles.note}>These are filled from the server&apos;s challenge on the first send; type them only to answer on the first request.</p>
        <div className={styles.row}>
          <label className={styles.field}>
            Realm
            <input name="digest-realm" autoComplete="off" spellCheck={false} className={styles.input} value={digest.realm} onChange={text('realm')} />
          </label>
          <label className={styles.field}>
            Nonce
            <input name="digest-nonce" autoComplete="off" spellCheck={false} className={styles.input} value={digest.nonce} onChange={text('nonce')} />
          </label>
        </div>
        <div className={styles.row}>
          <label className={styles.field}>
            Algorithm
            <select name="digest-algorithm" className={styles.select} value={digest.algorithm} onChange={(event) => onChange({ ...digest, algorithm: event.target.value as DigestAlgorithm })}>
              {DIGEST_ALGORITHMS.map((algorithm) => (
                <option key={algorithm} value={algorithm} translate="no">
                  {algorithm}
                </option>
              ))}
            </select>
          </label>
          <label className={styles.field}>
            qop
            <select name="digest-qop" className={styles.select} value={digest.qop} onChange={(event) => onChange({ ...digest, qop: event.target.value as DigestQop })}>
              <option value="">None</option>
              <option value="auth">auth</option>
              <option value="auth-int">auth-int</option>
            </select>
          </label>
        </div>
        <div className={styles.row}>
          <label className={styles.field}>
            Nonce count
            <input name="digest-nonce-count" autoComplete="off" spellCheck={false} className={styles.input} value={digest.nonceCount} placeholder="00000001…" onChange={text('nonceCount')} />
          </label>
          <label className={styles.field}>
            Client nonce
            <input name="digest-client-nonce" autoComplete="off" spellCheck={false} className={styles.input} value={digest.clientNonce} placeholder="Random when empty…" onChange={text('clientNonce')} />
          </label>
        </div>
        <label className={styles.field}>
          Opaque
          <input name="digest-opaque" autoComplete="off" spellCheck={false} className={styles.input} value={digest.opaque} onChange={text('opaque')} />
        </label>
      </details>
    </div>
  )
}
