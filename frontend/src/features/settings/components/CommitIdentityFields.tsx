import styles from './GitAccounts.module.css'

interface CommitIdentityFieldsProps {
  name: string
  email: string
  onName: (name: string) => void
  onEmail: (email: string) => void
}

export function CommitIdentityFields({ name, email, onName, onEmail }: CommitIdentityFieldsProps) {
  return (
    <>
      <label className={styles.field}>
        Commit name
        <input className={styles.input} name="commit-name" autoComplete="name" value={name} onChange={(event) => onName(event.target.value)} />
      </label>
      <label className={styles.field}>
        Commit email
        <input className={styles.input} name="commit-email" type="email" autoComplete="email" spellCheck={false} value={email} onChange={(event) => onEmail(event.target.value)} />
      </label>
    </>
  )
}
