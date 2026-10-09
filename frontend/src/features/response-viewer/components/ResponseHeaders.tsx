import styles from './ResponseHeaders.module.css'

interface ResponseHeadersProps {
  headers: Array<{ key: string; value: string }>
}

export function ResponseHeaders({ headers }: ResponseHeadersProps) {
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
            <td>{header.key}</td>
            <td>{header.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
