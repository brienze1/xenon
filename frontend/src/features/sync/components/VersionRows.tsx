import styles from './ConflictView.module.css'

export interface VersionRow {
  key: string
  value: string
  enabled: boolean
  secret?: boolean
}

interface VersionRowsProps {
  label: string
  rows: readonly VersionRow[]
}

export function VersionRows({ label, rows }: VersionRowsProps) {
  if (rows.length === 0) {
    return null
  }
  return (
    <table className={styles.rows} aria-label={label}>
      <caption className={styles.caption}>{label}</caption>
      <thead>
        <tr>
          <th scope="col">Key</th>
          <th scope="col">Value</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row, index) => (
          <tr key={`${row.key}-${index}`} className={row.enabled ? undefined : styles.disabledRow}>
            <td translate="no">{row.key}</td>
            <td translate="no">{row.secret ? <span className={styles.secret}>secret - value stays on each machine</span> : row.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
