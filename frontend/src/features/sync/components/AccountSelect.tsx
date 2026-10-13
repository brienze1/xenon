import type { GitAccount } from '@/platform/mappers/git-accounts'
import { NO_ACCOUNT } from '@/features/sync/lib/status-label'
import styles from './SyncPanel.module.css'

interface AccountSelectProps {
  accounts: GitAccount[]
  value: string
  disabled?: boolean
  onChange: (accountId: string) => void
}

export function AccountSelect({ accounts, value, disabled = false, onChange }: AccountSelectProps) {
  return (
    <label className={styles.field}>
      Account
      <select className={styles.input} name="sync-account" value={value} disabled={disabled} onChange={(event) => onChange(event.target.value)}>
        <option value="">{NO_ACCOUNT}</option>
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.displayName}
          </option>
        ))}
      </select>
    </label>
  )
}
