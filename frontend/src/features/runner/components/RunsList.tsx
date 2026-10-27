import { useState } from 'react'
import type { RunSummaryView } from '@/platform/mappers/runner'
import { formatDateTime } from '@/shared/lib/date-format'
import { formatDuration } from '@/shared/lib/format-duration'
import { Button } from '@/shared/ui/button/Button'
import { ContextMenu } from '@/shared/ui/context-menu/ContextMenu'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { STATUS_LABELS, TYPE_LABELS } from '@/features/runner/lib/labels'
import { isRunInProgress } from '@/features/runner/lib/results'
import { useRunnerStore } from '@/features/runner/store'
import { ClearRunsDialog } from './ClearRunsDialog'
import styles from './RunsList.module.css'

const DATE_TIME: Intl.DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'short' }

const LOADING = (
  <p role="status" className={styles.note}>
    Loading runs…
  </p>
)

const NO_RUNS = (
  <EmptyState.Root label="No stored runs" size="compact">
    <EmptyState.Title>No Runs Yet</EmptyState.Title>
    <EmptyState.Description>Runs of this collection or folder show up here.</EmptyState.Description>
  </EmptyState.Root>
)

const MORE_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.moreIcon}>
    <circle cx="5" cy="12" r="1" />
    <circle cx="12" cy="12" r="1" />
    <circle cx="19" cy="12" r="1" />
  </svg>
)

interface RunsListProps {
  itemId: string
}

/** The runs of the target, newest first, the one in progress included: open one in Results, delete a stored one, or clear them all. */
export function RunsList({ itemId }: RunsListProps) {
  const runs = useRunnerStore((state) => state.targets[itemId]?.runs ?? null)
  const openRun = useRunnerStore((state) => state.openRun)
  const deleteRun = useRunnerStore((state) => state.deleteRun)
  const [clearing, setClearing] = useState(false)

  if (runs === null) {
    return LOADING
  }
  return (
    <section className={styles.runs} aria-label="Stored runs">
      <header className={styles.header}>
        <span className={styles.note}>{runs.length === 1 ? '1 run' : `${runs.length} runs`}</span>
        <Button variant="ghost" size="sm" disabled={!runs.some((run) => !isRunInProgress(run))} onClick={() => setClearing(true)}>
          Clear runs
        </Button>
      </header>
      {runs.length === 0 ? (
        NO_RUNS
      ) : (
        <ul className={styles.list}>
          {runs.map((run) => (
            <RunItem key={run.id} run={run} onOpen={() => void openRun(itemId, run.id)} onDelete={() => void deleteRun(itemId, run.id)} />
          ))}
        </ul>
      )}
      <ClearRunsDialog itemId={itemId} open={clearing} onClose={() => setClearing(false)} />
    </section>
  )
}

interface RunItemProps {
  run: RunSummaryView
  onOpen: () => void
  onDelete: () => void
}

function RunItem({ run, onOpen, onDelete }: RunItemProps) {
  const started = run.startedAt ? formatDateTime(new Date(run.startedAt), DATE_TIME) : '—'
  const { summary } = run
  return (
    <li className={styles.item}>
      <button type="button" className={styles.open} onClick={onOpen}>
        <span className={styles.started}>{started}</span>
        <span className={styles.type}>{TYPE_LABELS[run.type]}</span>
        <span className={styles.badge} data-status={run.status}>
          {STATUS_LABELS[run.status]}
        </span>
        <span className={styles.environment} translate="no">
          {run.environmentName || 'No environment'}
        </span>
        <span className={styles.metric}>{summary.totalRequests} requests</span>
        <span className={styles.metric}>
          {summary.testsPassed} passed · {summary.testsFailed} failed · {summary.testsSkipped} skipped
        </span>
        <span className={styles.metric}>{formatDuration(summary.durationMs)}</span>
      </button>
      {isRunInProgress(run) ? null : (
        <ContextMenu.Root>
          <ContextMenu.Button label={`More actions for the run of ${started}`} title="More actions" className={styles.more}>
            {MORE_ICON}
          </ContextMenu.Button>
          <ContextMenu.Menu label={`Run of ${started}`}>
            <ContextMenu.Item tone="danger" onSelect={onDelete}>
              Delete
            </ContextMenu.Item>
          </ContextMenu.Menu>
        </ContextMenu.Root>
      )}
    </li>
  )
}
