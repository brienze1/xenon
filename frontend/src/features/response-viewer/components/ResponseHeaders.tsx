import styles from './ResponseHeaders.module.css'

interface ResponseHeadersProps {
  headers: Array<{ key: string; value: string }>
}

export function ResponseHeaders({ headers }: ResponseHeadersProps) {
  if (headers.length === 0) {
    return <p className={styles.empty}>No headers</p>
  }
  return (
    <table aria-label="Response headers" className={styles.table}>
      <thead>
        <tr>
          <th>Key</th>
          <th>Value</th>
        </tr>
      </thead>
      <tbody>
        {headers.map((header, index) => (
          <tr key={`${header.key}-${index}`}>
            <td translate="no">{header.key}</td>
            <td translate="no">{header.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
