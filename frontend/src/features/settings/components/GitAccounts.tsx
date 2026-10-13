import { useEffect, useState } from 'react'
import { withNextStep } from '@/shared/lib/next-step'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { useGitAccountsStore } from '@/features/settings/store'
import { GitHubSignIn } from './GitHubSignIn'
import { SshAccountForm } from './SshAccountForm'
import { TokenAccountForm } from './TokenAccountForm'
import styles from './GitAccounts.module.css'

type AddPanel = 'none' | 'github' | 'token' | 'ssh'

const KIND_LABELS: Record<string, string> = { github: 'GitHub sign-in', token: 'Token', ssh: 'SSH key' }

export function GitAccounts() {
  const accounts = useGitAccountsStore((state) => state.accounts)
  const loadError = useGitAccountsStore((state) => state.error)
  const load = useGitAccountsStore((state) => state.load)
  const remove = useGitAccountsStore((state) => state.remove)
  const [panel, setPanel] = useState<AddPanel>('none')
  const [removeError, setRemoveError] = useState<string | null>(null)

  useEffect(() => {
    void load()
  }, [load])

  async function removeAccount(id: string) {
    const failure = await remove(id)
    setRemoveError(failure ? withNextStep(failure.description || failure.code, 'Try again.') : null)
  }

  const error = removeError ?? (loadError ? withNextStep(loadError.description || loadError.code, 'Close and open Settings again.') : null)
  const close = () => setPanel('none')

  return (
    <fieldset className={styles.section}>
      <legend className={styles.legend}>Git accounts</legend>
      {accounts.length === 0 ? (
        <p className={styles.empty}>No Git accounts yet. Any workspace can sync with any account you add here.</p>
      ) : (
        <ul className={styles.list} aria-label="Git accounts">
          {accounts.map((account) => (
            <li key={account.id} className={styles.account}>
              {account.avatarUrl ? <img className={styles.avatar} src={account.avatarUrl} alt="" width={24} height={24} /> : null}
              <span className={styles.identity}>
                <span className={styles.name}>{account.displayName}</span>
                <span className={styles.meta}>
                  {KIND_LABELS[account.kind] ?? account.kind} · commits as {account.commitName || '—'} {account.commitEmail ? `<${account.commitEmail}>` : ''}
                </span>
              </span>
              <Button size="sm" variant="ghost" aria-label={`Remove ${account.displayName}`} onClick={() => void removeAccount(account.id)}>
                Remove
              </Button>
            </li>
          ))}
        </ul>
      )}
      {error ? (
        <Banner.Root tone="error">
          <Banner.Message>{error}</Banner.Message>
        </Banner.Root>
      ) : null}
      {panel === 'none' ? (
        <div className={styles.actions}>
          <Button size="sm" onClick={() => setPanel('github')}>
            Add GitHub account
          </Button>
          <Button size="sm" onClick={() => setPanel('token')}>
            Add token account
          </Button>
          <Button size="sm" onClick={() => setPanel('ssh')}>
            Add SSH key
          </Button>
        </div>
      ) : null}
      {panel === 'github' ? <GitHubSignIn onClose={close} /> : null}
      {panel === 'token' ? <TokenAccountForm onClose={close} /> : null}
      {panel === 'ssh' ? <SshAccountForm onClose={close} /> : null}
    </fieldset>
  )
}
