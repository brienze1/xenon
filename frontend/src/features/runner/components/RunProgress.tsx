import type { RunSummaryView } from '@/platform/mappers/runner'
import { formatDuration } from '@/shared/lib/format-duration'
import { formatNumber } from '@/shared/lib/number-format'
import { Button } from '@/shared/ui/button/Button'
import { useRunnerStore } from '@/features/runner/store'
import styles from './RunProgress.module.css'

const STATUS_LABELS: Record<string, string> = { running: 'Running', paused: 'Paused' }

interface RunProgressProps {
  itemId: string
  run: RunSummaryView
}

/** The live progress of the active run: status, current request, the completed bar, counts, elapsed time and the controls. */
export function RunProgress({ itemId, run }: RunProgressProps) {
  const pause = useRunnerStore((state) => state.pause)
  const resume = useRunnerStore((state) => state.resume)
  const stop = useRunnerStore((state) => state.stop)
  const { progress } = run
  const functional = run.type === 'functional'
  const planned = Math.max(progress.planned, progress.completed)
  const metrics = progress.metrics

  return (
    <section className={styles.progress} aria-label="Run progress">
      <div className={styles.head}>
        <span className={styles.status} data-status={run.status} role="status">
          {STATUS_LABELS[run.status] ?? run.status}
        </span>
        {functional && progress.currentName ? (
          <span className={styles.current}>
            <span translate="no">{progress.currentName}</span> · iteration {progress.iteration + 1}
          </span>
        ) : null}
        <span className={styles.elapsed}>{formatDuration(progress.elapsedMs)}</span>
        <div className={styles.controls}>
          {functional ? (
            run.status === 'paused' ? (
              <Button variant="secondary" size="sm" onClick={() => void resume(itemId)}>
                Resume
              </Button>
            ) : (
              <Button variant="secondary" size="sm" onClick={() => void pause(itemId)}>
                Pause
              </Button>
            )
          ) : null}
          <Button variant="danger" size="sm" onClick={() => void stop(itemId)}>
            Stop
          </Button>
        </div>
      </div>
      {functional ? (
        <>
          <div
            role="progressbar"
            aria-label="Completed requests"
            aria-valuemin={0}
            aria-valuemax={planned}
            aria-valuenow={progress.completed}
            className={styles.bar}
          >
            <span className={styles.fill} style={{ width: planned > 0 ? `${(100 * progress.completed) / planned}%` : '0%' }} />
          </div>
          <p className={styles.counts}>
            <span>
              {progress.completed} of {planned} requests
            </span>
            <span className={styles.passed}>{progress.testsPassed} passed</span>
            <span className={styles.failed}>{progress.testsFailed} failed</span>
          </p>
        </>
      ) : metrics ? (
        <dl className={styles.metrics}>
          <div>
            <dt>Active users</dt>
            <dd>{metrics.activeUsers}</dd>
          </div>
          <div>
            <dt>Requests</dt>
            <dd>{formatNumber(metrics.requests, 0)}</dd>
          </div>
          <div>
            <dt>Avg. response</dt>
            <dd>{formatDuration(metrics.avgMs)}</dd>
          </div>
          <div>
            <dt>Throughput</dt>
            <dd>{formatNumber(metrics.throughput, 1)} req/s</dd>
          </div>
          <div>
            <dt>Error rate</dt>
            <dd>{formatNumber(metrics.errorRate * 100, 1)}%</dd>
          </div>
        </dl>
      ) : null}
    </section>
  )
}
