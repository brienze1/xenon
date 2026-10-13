import { useState, useTransition } from 'react'
import type { GitProvider } from '@/platform/mappers/git-accounts'
import { withNextStep } from '@/shared/lib/next-step'
import { Button } from '@/shared/ui/button/Button'
import { useGitAccountsStore } from '@/features/settings/store'
import { CommitIdentityFields } from './CommitIdentityFields'
import styles from './GitAccounts.module.css'

interface ProviderPreset {
  label: string
  host: string
  needsUsername: boolean
}

const PRESETS: Record<GitProvider, ProviderPreset> = {
  github: { label: 'GitHub', host: 'github.com', needsUsername: false },
  gitlab: { label: 'GitLab', host: 'gitlab.com', needsUsername: false },
  bitbucket: { label: 'Bitbucket (API token)', host: 'bitbucket.org', needsUsername: false },
  azure_devops: { label: 'Azure DevOps', host: 'dev.azure.com', needsUsername: false },
  other: { label: 'Other host', host: '', needsUsername: true },
}

interface TokenAccountFormProps {
  onClose: () => void
}

export function TokenAccountForm({ onClose }: TokenAccountFormProps) {
  const addToken = useGitAccountsStore((state) => state.addToken)
  const [provider, setProvider] = useState<GitProvider>('github')
  const [host, setHost] = useState(PRESETS.github.host)
  const [username, setUsername] = useState('')
  const [token, setToken] = useState('')
  const [repository, setRepository] = useState('')
  const [commitName, setCommitName] = useState('')
  const [commitEmail, setCommitEmail] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, startSave] = useTransition()

  function choose(next: GitProvider) {
    setProvider(next)
    setHost(PRESETS[next].host)
  }

  function save() {
    startSave(async () => {
      const failure = await addToken({ provider, host, username, token, repository, commitName, commitEmail })
      if (failure) {
        setError(withNextStep(failure.description || failure.code, 'Check the token and try again.'))
        return
      }
      onClose()
    })
  }

  return (
    <section className={styles.panel} aria-label="Add token account">
      <label className={styles.field}>
        Host
        <select className={styles.input} value={provider} onChange={(event) => choose(event.target.value as GitProvider)}>
          {Object.entries(PRESETS).map(([value, preset]) => (
            <option key={value} value={value}>
              {preset.label}
            </option>
          ))}
        </select>
      </label>
      <label className={styles.field}>
        Host name
        <input className={styles.input} name="git-host" autoComplete="off" spellCheck={false} value={host} onChange={(event) => setHost(event.target.value)} />
      </label>
      {PRESETS[provider].needsUsername ? (
        <label className={styles.field}>
          Username
          <input className={styles.input} name="git-username" autoComplete="off" spellCheck={false} value={username} onChange={(event) => setUsername(event.target.value)} />
        </label>
      ) : null}
      <label className={styles.field}>
        Token
        <input className={styles.input} name="git-token" type="password" autoComplete="off" value={token} onChange={(event) => setToken(event.target.value)} />
      </label>
      <label className={styles.field}>
        Repository to verify the token with
        <input
          className={styles.input}
          name="git-repository"
          autoComplete="off"
          spellCheck={false}
          placeholder="https://host/owner/repo.git"
          value={repository}
          onChange={(event) => setRepository(event.target.value)}
        />
      </label>
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
        <Button size="sm" variant="primary" disabled={saving || token.trim() === '' || host.trim() === ''} onClick={save}>
          Verify and add
        </Button>
      </div>
    </section>
  )
}
