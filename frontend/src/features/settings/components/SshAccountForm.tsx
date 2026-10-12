import { useState, useTransition } from 'react'
import type { GitAccount, SshMode } from '@/platform/mappers/git-accounts'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { useGitAccountsStore } from '@/features/settings/store'
import { CommitIdentityFields } from './CommitIdentityFields'
import styles from './GitAccounts.module.css'

const MODES: Record<SshMode, string> = { generate: 'Generate a new key', key_file: 'Use an existing key file', agent: 'Use the SSH agent' }

interface SshAccountFormProps {
  onClose: () => void
}

export function SshAccountForm({ onClose }: SshAccountFormProps) {
  const addSsh = useGitAccountsStore((state) => state.addSsh)
  const [mode, setMode] = useState<SshMode>('generate')
  const [host, setHost] = useState('github.com')
  const [keyPath, setKeyPath] = useState('')
  const [passphrase, setPassphrase] = useState('')
  const [commitName, setCommitName] = useState('')
  const [commitEmail, setCommitEmail] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [added, setAdded] = useState<{ account: GitAccount; publicKey: string } | null>(null)
  const [saving, startSave] = useTransition()

  function save() {
    startSave(async () => {
      const result = await addSsh({ mode, host, keyPath, passphrase, commitName, commitEmail })
      if ('code' in result) {
        setError(withNextStep(result.description || result.code, 'Check the key and try again.'))
        return
      }
      if (result.publicKey === '') {
        onClose()
        return
      }
      setAdded(result)
    })
  }

  if (added) {
    return <PublicKeyPanel account={added.account} publicKey={added.publicKey} onClose={onClose} />
  }

  return (
    <section className={styles.panel} aria-label="Add SSH key">
      <fieldset className={styles.section}>
        <legend className={styles.legend}>Key</legend>
        {Object.entries(MODES).map(([value, label]) => (
          <label key={value} className={styles.meta}>
            <input type="radio" name="ssh-mode" value={value} checked={mode === value} onChange={() => setMode(value as SshMode)} /> {label}
          </label>
        ))}
      </fieldset>
      <label className={styles.field}>
        Host
        <input className={styles.input} name="ssh-host" autoComplete="off" spellCheck={false} value={host} onChange={(event) => setHost(event.target.value)} />
      </label>
      {mode === 'key_file' ? (
        <>
          <label className={styles.field}>
            Private key file
            <input className={styles.input} name="ssh-key-path" autoComplete="off" spellCheck={false} placeholder="~/.ssh/id_ed25519" value={keyPath} onChange={(event) => setKeyPath(event.target.value)} />
          </label>
          <label className={styles.field}>
            Key passphrase (if any)
            <input className={styles.input} name="ssh-passphrase" type="password" autoComplete="off" value={passphrase} onChange={(event) => setPassphrase(event.target.value)} />
          </label>
        </>
      ) : null}
      <CommitIdentityFields name={commitName} email={commitEmail} onName={setCommitName} onEmail={setCommitEmail} />
      {error ? (
        <p role="alert" className={styles.hint}>
          {error}
        </p>
      ) : null}
      <div className={styles.actions}>
        <Button size="sm" variant="secondary" onClick={onClose}>
          Cancel
        </Button>
        <Button size="sm" variant="primary" disabled={saving || host.trim() === ''} onClick={save}>
          Add SSH key
        </Button>
      </div>
    </section>
  )
}

interface PublicKeyPanelProps {
  account: GitAccount
  publicKey: string
  onClose: () => void
}

function PublicKeyPanel({ account, publicKey, onClose }: PublicKeyPanelProps) {
  const accounts = useGitAccountsStore((state) => state.accounts)
  const gitHubAccounts = accounts.filter((candidate) => candidate.kind === 'github' || candidate.provider === 'github')
  const uploadKey = useGitAccountsStore((state) => state.uploadKey)
  const copyKey = useGitAccountsStore((state) => state.copyKey)
  const [target, setTarget] = useState('')
  const [message, setMessage] = useState<string | null>(null)
  const chosen = target || gitHubAccounts[0]?.id || ''

  async function upload() {
    const failure = await uploadKey(account.id, chosen, 'Xenon')
    setMessage(failure ? withNextStep(failure.description || failure.code, 'Add the key on GitHub by hand.') : 'The key was added to GitHub.')
  }

  return (
    <section className={styles.panel} aria-label="Public key">
      <p className={styles.hint}>Add this public key to your Git host:</p>
      <p className={styles.publicKey}>{publicKey}</p>
      <div className={styles.actions}>
        <Button size="sm" onClick={() => void copyKey(publicKey)}>
          Copy public key
        </Button>
        {gitHubAccounts.length > 0 ? (
          <>
            <select className={styles.input} aria-label="GitHub account" value={chosen} onChange={(event) => setTarget(event.target.value)}>
              {gitHubAccounts.map((candidate) => (
                <option key={candidate.id} value={candidate.id}>
                  {candidate.displayName}
                </option>
              ))}
            </select>
            <Button size="sm" onClick={() => void upload()}>
              Add to GitHub
            </Button>
          </>
        ) : null}
        <Button size="sm" variant="primary" onClick={onClose}>
          Done
        </Button>
      </div>
      {message ? (
        <p role="status" className={styles.hint}>
          {message}
        </p>
      ) : null}
    </section>
  )
}
