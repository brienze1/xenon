import { useEffect } from 'react'
import { runResponseKey, type RunRowResult } from '@/platform/mappers/runner'
import { formatBytes } from '@/shared/lib/format-bytes'
import { formatDuration } from '@/shared/lib/format-duration'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import type { RunnerSlots } from '@/features/runner/model'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useRunnerStore } from '@/features/runner/store'
import styles from './RunRowDetail.module.css'

const CLOSE_GLYPH = <span aria-hidden="true">×</span>

const NOT_PERSISTED = <p className={styles.note}>Responses were not persisted for this run.</p>

const LOADING = (
  <p role="status" className={styles.note}>
    Loading response…
  </p>
)

interface RunRowDetailProps {
  runId: string
  row: RunRowResult
  response: RunnerSlots['response']
  onClose: () => void
}

/** An opened result row: the request as sent (secrets masked) and, when persisted, the response as the response viewer shows it. */
export function RunRowDetail({ runId, row, response, onClose }: RunRowDetailProps) {
  const key = runResponseKey(runId, row.index)
  const persisted = row.responseBodyUrl !== ''
  const status = useRunnerStore((state) => state.responses[key])
  const loadResponse = useRunnerStore((state) => state.loadResponse)
  const result = useResponseViewerStore((state) => state.entries[key]?.result ?? null)

  useEffect(() => {
    if (persisted) {
      void loadResponse(runId, row.index)
    }
  }, [persisted, runId, row.index, loadResponse])

  const title = row.aliasName ? `${row.name} · ${row.aliasName}` : row.name

  return (
    <aside className={styles.detail} aria-label={`${title} details`}>
      <header className={styles.header}>
        <h3 className={styles.title} translate="no">
          {title}
        </h3>
        <IconButton label="Close details" onClick={onClose}>
          {CLOSE_GLYPH}
        </IconButton>
      </header>
      <section className={styles.section} aria-label="Request">
        <h4 className={styles.heading}>Request</h4>
        <p className={styles.line}>
          <MethodBadge method={row.sentRequest.method || row.method} />
          <span className={styles.url} translate="no">
            {row.sentRequest.url || row.url}
          </span>
        </p>
        {row.sentRequest.headers.length > 0 ? (
          <table className={styles.headers}>
            <caption className="visually-hidden">Request headers</caption>
            <tbody>
              {row.sentRequest.headers.map((header, index) => (
                <tr key={`${header.key}-${index}`}>
                  <th scope="row" translate="no">
                    {header.key}
                  </th>
                  <td translate="no">{header.value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
        {row.sentRequest.body ? (
          <pre className={styles.body} aria-label="Request body" translate="no">
            {row.sentRequest.body}
          </pre>
        ) : null}
      </section>
      <section className={styles.section} aria-label="Response">
        <h4 className={styles.heading}>Response</h4>
        <p className={styles.summary}>
          <span>{row.status > 0 ? row.status : '—'}</span>
          <span>{formatDuration(row.durationMs)}</span>
          <span>{formatBytes(row.sizeBytes)}</span>
        </p>
        {row.error ? (
          <p role="alert" className={styles.error}>
            {row.error}
          </p>
        ) : null}
        {!persisted || status === 'missing' || status === 'error' ? NOT_PERSISTED : status === 'ready' && result !== null ? response(key, result) : LOADING}
      </section>
    </aside>
  )
}
