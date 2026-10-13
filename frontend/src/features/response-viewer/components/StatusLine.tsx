import type { ReactNode } from 'react'
import type { SendResult, SendTimings } from '@/platform/mappers/request'
import { cn } from '@/shared/lib/cn'
import { formatBytes } from '@/shared/lib/format-bytes'
import { formatDuration } from '@/shared/lib/format-duration'
import { statusTone } from '@/shared/lib/status-tone'
import { Tooltip } from '@/shared/ui/tooltip/Tooltip'
import styles from './StatusLine.module.css'

interface StatusLineProps {
  result: SendResult
}

const PHASES: Array<{ key: keyof SendTimings; label: ReactNode }> = [
  {
    key: 'dnsMs',
    label: (
      <>
        <span translate="no">DNS</span> lookup
      </>
    ),
  },
  {
    key: 'connectMs',
    label: (
      <>
        <span translate="no">TCP</span> connection
      </>
    ),
  },
  {
    key: 'tlsMs',
    label: (
      <>
        <span translate="no">TLS</span> handshake
      </>
    ),
  },
  { key: 'firstByteMs', label: 'Time to first byte' },
  { key: 'downloadMs', label: 'Download' },
]

const TOTAL_TERM = <dt>Total</dt>

export function StatusLine({ result }: StatusLineProps) {
  const timings = result.timings

  return (
    <div role="group" className={styles.line} aria-label="Response status">
      <span className={cn(styles.status, styles[statusTone(result.statusCode)])} translate="no" title={result.statusText}>
        {result.statusCode} {result.statusText}
      </span>
      <Tooltip
        content={
          <dl className={styles.timings}>
            {PHASES.map((phase) => (
              <div key={phase.key} className={styles.timing}>
                <dt>{phase.label}</dt>
                <dd>{formatDuration(timings[phase.key])}</dd>
              </div>
            ))}
            <div className={cn(styles.timing, styles.total)}>
              {TOTAL_TERM}
              <dd>{formatDuration(timings.totalMs)}</dd>
            </div>
          </dl>
        }
      >
        <button type="button" className={styles.metric} aria-label={`Time ${formatDuration(timings.totalMs)}`}>
          {formatDuration(timings.totalMs)}
        </button>
      </Tooltip>
      <span className={styles.metric}>{formatBytes(result.sizeBytes)}</span>
    </div>
  )
}
