import styles from './ResponseHeaders.module.css'

const NO_HEADERS = <p className={styles.empty}>No headers</p>

const HEADERS_HEAD = (
  <thead>
    <tr>
      <th>Key</th>
      <th>Value</th>
    </tr>
  </thead>
)

interface ResponseHeadersProps {
  headers: Array<{ key: string; value: string }>
}

export function ResponseHeaders({ headers }: ResponseHeadersProps) {
  if (headers.length === 0) {
    return NO_HEADERS
  }
  return (
    <table aria-label="Response headers" className={styles.table}>
      {HEADERS_HEAD}
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
