import type { ConsoleEntry } from '@/platform/mappers/console'
import { cn } from '@/shared/lib/cn'
import { formatBytes } from '@/shared/lib/format-bytes'
import { formatDuration } from '@/shared/lib/format-duration'
import { statusTone } from '@/shared/lib/status-tone'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { consoleTime } from '@/features/console/lib/format'
import styles from './ConsoleEntryRow.module.css'

const KIND_LABELS: Record<ConsoleEntry['kind'], string> = { request: 'request', log: 'log', info: 'info', error: 'error', test_failure: 'test' }

interface RequestSummaryProps {
  entry: ConsoleEntry
}

function RequestSummary({ entry }: RequestSummaryProps) {
  const request = entry.request
  const response = entry.response
  if (!request) {
    return null
  }
  return (
    <span className={styles.request}>
      <MethodBadge method={request.method} />
      <span className={styles.url} translate="no">
        {request.url}
      </span>
      {response ? (
        <span className={styles.meta} translate="no">
          {response.error ? (
            <span className={styles.error}>{response.error}</span>
          ) : (
            <span className={styles[statusTone(response.statusCode)]}>
              {response.statusCode} {response.statusText}
            </span>
          )}
          <span>{formatDuration(response.timeMs)}</span>
          <span>{formatBytes(response.sizeBytes)}</span>
        </span>
      ) : null}
    </span>
  )
}

interface ConsoleEntryRowProps {
  entry: ConsoleEntry
  selected: boolean
  onSelect: (id: string) => void
}

export function ConsoleEntryRow({ entry, selected, onSelect }: ConsoleEntryRowProps) {
  const label = entry.kind === 'request' ? entry.summary : `${entry.level}: ${entry.summary}`
  return (
    <button
      type="button"
      aria-pressed={selected}
      aria-label={label}
      className={cn(styles.row, styles[`level-${entry.level}`], entry.kind === 'test_failure' ? styles['level-error'] : undefined)}
      onClick={() => onSelect(entry.id)}
    >
      <time className={styles.time} dateTime={entry.createdAt}>
        {consoleTime(entry.createdAt)}
      </time>
      <span className={styles.kind}>{entry.request?.nested ? 'script request' : KIND_LABELS[entry.kind]}</span>
      {entry.kind === 'request' ? <RequestSummary entry={entry} /> : <span className={styles.summary}>{entry.summary}</span>}
      {entry.scriptLabel ? (
        <span className={styles.script} translate="no">
          {entry.scriptLabel}
        </span>
      ) : null}
    </button>
  )
}
