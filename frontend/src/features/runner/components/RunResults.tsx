import { useVirtualizer } from '@tanstack/react-virtual'
import { useMemo, useRef, useState } from 'react'
import type { Run, RunRowResult } from '@/platform/mappers/runner'
import { cn } from '@/shared/lib/cn'
import { formatDateTime } from '@/shared/lib/date-format'
import { formatBytes } from '@/shared/lib/format-bytes'
import { formatDuration } from '@/shared/lib/format-duration'
import { Banner } from '@/shared/ui/banner/Banner'
import { Button } from '@/shared/ui/button/Button'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { OUTCOME_LABELS, STATUS_LABELS } from '@/features/runner/lib/labels'
import { filterCounts, RESULT_FILTERS, resultItems, type ResultItem } from '@/features/runner/lib/results'
import type { ResultFilter, RunnerSlots } from '@/features/runner/model'
import { useRunnerStore } from '@/features/runner/store'
import { PerformanceResults } from './PerformanceResults'
import { RunRowDetail } from './RunRowDetail'
import styles from './RunResults.module.css'

export const RESULT_ROW_HEIGHT = 32
export const ITERATION_HEADING_HEIGHT = 28
/** First estimate of an expanded row before it is measured; the rendered row sets its real height. */
export const EXPANDED_ROW_HEIGHT = 240

const DATE_TIME: Intl.DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'medium' }

const NO_MATCHES = <p className={styles.none}>No results match this filter.</p>

interface RunResultsProps {
  itemId: string
  run: Run
  slots: RunnerSlots
}

/** A run's results: summary, Run again and Export, then the rows (functional) or the stats and chart (performance). */
export function RunResults({ itemId, run, slots }: RunResultsProps) {
  const running = useRunnerStore((state) => state.active[itemId] !== undefined)
  const starting = useRunnerStore((state) => state.targets[itemId]?.starting === true)
  const runAgain = useRunnerStore((state) => state.runAgain)
  const exportRun = useRunnerStore((state) => state.exportRun)
  const finished = run.status !== 'running' && run.status !== 'paused'
  const performance = run.config.type === 'performance'

  return (
    <div className={styles.results}>
      {finished ? <Summary run={run} /> : null}
      <div className={styles.actions}>
        <Button variant="secondary" size="sm" disabled={running || starting || !finished} onClick={() => void runAgain(itemId)}>
          Run again
        </Button>
        <Dropdown.Root>
          <Dropdown.Trigger className={styles.export}>Export</Dropdown.Trigger>
          <Dropdown.Menu label="Export run">
            <Dropdown.Item disabled={!finished} onSelect={() => void exportRun(itemId, 'json')}>
              JSON
            </Dropdown.Item>
            {performance ? null : (
              <Dropdown.Item disabled={!finished} onSelect={() => void exportRun(itemId, 'junit')}>
                JUnit XML
              </Dropdown.Item>
            )}
          </Dropdown.Menu>
        </Dropdown.Root>
      </div>
      {run.notices.map((notice) => (
        <Banner.Root key={notice} tone="info">
          <Banner.Message>{notice}</Banner.Message>
        </Banner.Root>
      ))}
      {performance ? run.performance ? <PerformanceResults report={run.performance} /> : null : <RowResults run={run} slots={slots} />}
    </div>
  )
}

interface SummaryProps {
  run: Run
}

function Summary({ run }: SummaryProps) {
  const { summary } = run
  const started = run.startedAt ? formatDateTime(new Date(run.startedAt), DATE_TIME) : '—'
  return (
    <dl className={styles.summary} aria-label="Run summary">
      <div>
        <dt>Status</dt>
        <dd className={styles.status} data-status={run.status}>
          {STATUS_LABELS[run.status]}
        </dd>
      </div>
      <div>
        <dt>Started</dt>
        <dd>{started}</dd>
      </div>
      <div>
        <dt>Duration</dt>
        <dd>{formatDuration(summary.durationMs)}</dd>
      </div>
      <div>
        <dt>Iterations</dt>
        <dd>{summary.iterations}</dd>
      </div>
      <div>
        <dt>Requests</dt>
        <dd>{summary.totalRequests}</dd>
      </div>
      <div>
        <dt>Tests</dt>
        <dd>
          <span className={styles.passed}>{summary.testsPassed} passed</span> · <span className={styles.failed}>{summary.testsFailed} failed</span> ·{' '}
          {summary.testsSkipped} skipped
        </dd>
      </div>
      <div>
        <dt>Avg. response</dt>
        <dd>{formatDuration(summary.avgResponseMs)}</dd>
      </div>
    </dl>
  )
}

interface RowResultsProps {
  run: Run
  slots: RunnerSlots
}

function RowResults({ run, slots }: RowResultsProps) {
  const [filter, setFilter] = useState<ResultFilter>('all')
  const [expanded, setExpanded] = useState<ReadonlySet<number>>(() => new Set())
  const [openIndex, setOpenIndex] = useState<number | null>(null)
  const counts = useMemo(() => filterCounts(run.rows), [run.rows])
  const items = useMemo(() => resultItems(run.rows, filter), [run.rows, filter])
  const opened = openIndex === null ? null : (run.rows.find((row) => row.index === openIndex) ?? null)
  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: (index) => itemHeight(items[index], expanded),
    getItemKey: (index) => items[index]?.key ?? index,
    overscan: 10,
  })

  function toggle(index: number) {
    setExpanded((current) => {
      const next = new Set(current)
      if (!next.delete(index)) {
        next.add(index)
      }
      return next
    })
  }

  return (
    <div className={cn(styles.rows, opened && styles.withDetail)}>
      <div className={styles.list}>
        <div role="group" aria-label="Filter results" className={styles.filters}>
          {RESULT_FILTERS.map((option) => (
            <button key={option.filter} type="button" aria-pressed={filter === option.filter} className={styles.chip} onClick={() => setFilter(option.filter)}>
              {option.label} <span className={styles.count}>{counts[option.filter]}</span>
            </button>
          ))}
        </div>
        {items.length === 0 ? (
          NO_MATCHES
        ) : (
          <div ref={scrollRef} className={styles.scroll}>
            <div role="list" aria-label="Run results" className={styles.virtual} style={{ height: virtualizer.getTotalSize() }}>
              {virtualizer.getVirtualItems().map((virtual) => {
                const item = items[virtual.index]
                return item ? (
                  <div
                    key={item.key}
                    ref={virtualizer.measureElement}
                    data-index={virtual.index}
                    role="listitem"
                    className={styles.slot}
                    style={{ transform: `translateY(${virtual.start}px)` }}
                  >
                    {item.kind === 'iteration' ? (
                      <p className={styles.iteration}>
                        Iteration {item.iteration + 1} <span className={styles.count}>{item.count}</span>
                      </p>
                    ) : (
                      <ResultRow
                        row={item.row}
                        expanded={expanded.has(item.row.index)}
                        selected={openIndex === item.row.index}
                        tests={slots.tests}
                        onToggle={() => toggle(item.row.index)}
                        onOpen={() => setOpenIndex(item.row.index)}
                      />
                    )}
                  </div>
                ) : null
              })}
            </div>
          </div>
        )}
      </div>
      {opened ? <RunRowDetail key={opened.index} runId={run.id} row={opened} response={slots.response} onClose={() => setOpenIndex(null)} /> : null}
    </div>
  )
}

function itemHeight(item: ResultItem | undefined, expanded: ReadonlySet<number>): number {
  if (item?.kind === 'iteration') {
    return ITERATION_HEADING_HEIGHT
  }
  return item && expanded.has(item.row.index) ? EXPANDED_ROW_HEIGHT : RESULT_ROW_HEIGHT
}

const CHEVRON = (
  <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.chevronIcon}>
    <path d="M6 4l4 4-4 4" />
  </svg>
)

interface ResultRowProps {
  row: RunRowResult
  expanded: boolean
  selected: boolean
  tests: RunnerSlots['tests']
  onToggle: () => void
  onOpen: () => void
}

function ResultRow({ row, expanded, selected, tests, onToggle, onOpen }: ResultRowProps) {
  const name = row.aliasName ? `${row.name} · ${row.aliasName}` : row.name
  return (
    <div className={cn(styles.row, selected && styles.selected)}>
      <div className={styles.line}>
        <button type="button" aria-expanded={expanded} aria-label={`Details of ${name}`} className={cn(styles.chevron, expanded && styles.open)} onClick={onToggle}>
          {CHEVRON}
        </button>
        <span role="img" aria-label={OUTCOME_LABELS[row.outcome]} className={styles.outcome} data-outcome={row.outcome} />
        <MethodBadge method={row.method} className={styles.method} />
        <button type="button" className={styles.name} onClick={onOpen} translate="no">
          {name}
        </button>
        <span className={styles.url} translate="no">
          {row.url}
        </span>
        <span className={styles.metric}>{row.status > 0 ? row.status : '—'}</span>
        <span className={styles.metric}>{formatDuration(row.durationMs)}</span>
        <span className={styles.metric}>{formatBytes(row.sizeBytes)}</span>
      </div>
      {expanded ? (
        <div role="group" aria-label={`Details of ${name}`} tabIndex={0} className={styles.expansion}>
          {row.error ? (
            <p role="alert" className={styles.rowError}>
              {row.error}
            </p>
          ) : null}
          {tests(row.tests, row.testSummary)}
          {row.scriptErrors.length > 0 ? (
            <ul className={styles.scriptErrors} aria-label="Script errors">
              {row.scriptErrors.map((failure, index) => (
                <li key={index}>
                  {failure.scriptLabel ? `${failure.scriptLabel}: ` : ''}
                  {failure.message}
                </li>
              ))}
            </ul>
          ) : null}
          {row.console.length > 0 ? (
            <ul className={styles.console} aria-label="Console">
              {row.console.map((entry) => (
                <li key={entry.id} data-level={entry.level} translate="no">
                  {entry.summary}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
