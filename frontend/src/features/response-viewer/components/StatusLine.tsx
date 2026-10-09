import type { SendResult } from '@/platform/mappers/request'
import { cn } from '@/shared/lib/cn'
import { formatBytes } from '@/shared/lib/format-bytes'
import { formatDuration } from '@/shared/lib/format-duration'
import { Tooltip } from '@/shared/ui/tooltip/Tooltip'
import { statusTone } from '../lib/body-view'
import styles from './StatusLine.module.css'

interface StatusLineProps {
  result: SendResult
}

export function StatusLine({ result }: StatusLineProps) {
  const timings = result.timings
  const phases = [
    ['DNS Lookup', timings.dnsMs],
    ['TCP Connection', timings.connectMs],
    ['TLS Handshake', timings.tlsMs],
    ['Time to First Byte', timings.firstByteMs],
    ['Download', timings.downloadMs],
  ] as const

  return (
    <div className={styles.line} aria-label="Response status">
      <span className={cn(styles.status, styles[statusTone(result.statusCode)])}>
        {result.statusCode} {result.statusText}
      </span>
      <Tooltip
        content={
          <dl className={styles.timings}>
            {phases.map(([label, value]) => (
              <div key={label} className={styles.timing}>
                <dt>{label}</dt>
                <dd>{formatDuration(value)}</dd>
              </div>
            ))}
            <div className={cn(styles.timing, styles.total)}>
              <dt>Total</dt>
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
