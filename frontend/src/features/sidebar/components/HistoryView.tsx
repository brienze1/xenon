import { useVirtualizer } from '@tanstack/react-virtual'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { HistoryEntry } from '@/platform/mappers/history'
import { cn } from '@/shared/lib/cn'
import { formatDuration } from '@/shared/lib/format-duration'
import { formatNumber } from '@/shared/lib/number-format'
import { statusTone } from '@/shared/lib/status-tone'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { ClockArt } from '@/shared/ui/empty-state/illustrations/ClockArt'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { Swatch } from '@/shared/ui/swatch/Swatch'
import { aiTag } from '@/features/sidebar/lib/history-labels'
import { historyRows, sentTime } from '@/features/sidebar/lib/history-rows'
import { selectLocation, useSidebarStore } from '@/features/sidebar/store'
import { ClearHistoryDialog } from './ClearHistoryDialog'
import { HistoryActionRow } from './HistoryActionRow'
import { HistorySourceFilter } from './HistorySourceFilter'
import styles from './HistoryView.module.css'

export const HISTORY_DAY_HEIGHT = 28
export const HISTORY_ENTRY_HEIGHT = 44
/** An expanded AI action row: the row plus its (scrollable) masked arguments. */
export const HISTORY_EXPANDED_HEIGHT = HISTORY_ENTRY_HEIGHT + 168

const EMPTY_HISTORY = (
  <EmptyState.Root label="No history" className={styles.empty}>
    <EmptyState.Art>
      <ClockArt />
    </EmptyState.Art>
    <EmptyState.Title>Nothing Sent Yet</EmptyState.Title>
    <EmptyState.Description>Every request you send in this workspace shows up here, newest first.</EmptyState.Description>
  </EmptyState.Root>
)

const NOTHING_FROM_SOURCE = <p role="status" className={styles.noMatches}>No entries from this source yet.</p>

const NO_MATCHES = <p role="status" className={styles.noMatches}>No history entries match your search.</p>

interface HistoryViewProps {
  onOpenEntry: (id: string) => void
}

function outcome(entry: HistoryEntry): { label: string; tone: string } {
  if (entry.outcome === 'skipped_by_script') {
    return { label: 'skipped', tone: 'info' }
  }
  if (entry.outcome === 'stopped_by_script_error') {
    return { label: 'script error', tone: 'error' }
  }
  if (entry.statusCode === null) {
    return { label: entry.errorKind ?? 'error', tone: 'error' }
  }
  return { label: String(entry.statusCode), tone: statusTone(entry.statusCode) }
}

interface HistoryEntryButtonProps {
  entry: HistoryEntry
  onOpen: (id: string) => void
}

function HistoryEntryButton({ entry, onOpen }: HistoryEntryButtonProps) {
  const result = outcome(entry)
  const time = sentTime(entry.sentAt)
  const requestName = useSidebarStore((state) => (entry.sourceRequestId ? (selectLocation(state.tree, entry.sourceRequestId).at(-1) ?? null) : null))
  const version = entry.aliasName === null ? null : `${requestName ?? 'Request'} · ${entry.aliasName}`
  return (
    <button
      type="button"
      className={styles.entry}
      aria-label={`${entry.source === 'ai' ? `${aiTag(entry.client)} — ` : ''}${version ? `${version}, ` : ''}${entry.method} ${entry.url || 'no URL'}${entry.environment ? `, in ${entry.environment.name}` : ''}, ${result.label}, ${time}`}
      onClick={() => onOpen(entry.id)}
    >
      {version || entry.environment ? (
        <span className={styles.labels}>
          {version ? (
            <span className={styles.version} translate="no">
              {version}
            </span>
          ) : null}
          {entry.environment ? (
            <span className={styles.environment}>
              <Swatch color={entry.environment.color} />
              <span translate="no">{entry.environment.name}</span>
            </span>
          ) : null}
        </span>
      ) : null}
      <span className={styles.line}>
        {entry.source === 'ai' ? (
          <span className={styles.aiTag} translate="no">
            {aiTag(entry.client)}
          </span>
        ) : null}
        <MethodBadge method={entry.method} className={styles.method} />
        <span className={styles.url} title={entry.url} translate="no">
          {entry.url || 'No URL'}
        </span>
      </span>
      <span className={styles.meta}>
        <span className={cn(styles.status, styles[result.tone])} translate="no">
          {result.label}
        </span>
        <span>{formatDuration(entry.totalMs)}</span>
        <time dateTime={entry.sentAt}>{time}</time>
      </span>
    </button>
  )
}

export function HistoryView({ onOpenEntry }: HistoryViewProps) {
  const history = useSidebarStore((state) => state.history)
  const status = useSidebarStore((state) => state.historyStatus)
  const hasQuery = useSidebarStore((state) => state.historyQuery.trim() !== '')
  const source = useSidebarStore((state) => state.historySource)
  const clearHistory = useSidebarStore((state) => state.clearHistory)
  const [confirming, setConfirming] = useState(false)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const rows = useMemo(() => historyRows(history, new Date()), [history])
  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: (index) => {
      const row = rows[index]
      if (row?.kind === 'day') {
        return HISTORY_DAY_HEIGHT
      }
      return row?.key === expandedId ? HISTORY_EXPANDED_HEIGHT : HISTORY_ENTRY_HEIGHT
    },
    overscan: 10,
  })

  useEffect(() => {
    virtualizer.measure()
  }, [expandedId, virtualizer])

  function toggle(id: string) {
    setExpandedId((current) => (current === id ? null : id))
  }

  async function confirmClear() {
    const failure = await clearHistory()
    if (failure === null) {
      setConfirming(false)
    }
    return failure
  }

  return (
    <div className={styles.view}>
      <div className={styles.header}>
        <HistorySourceFilter />
        <Button variant="ghost" size="sm" disabled={history.length === 0} onClick={() => setConfirming(true)}>
          Clear History
        </Button>
      </div>
      <span className={styles.count}>{history.length === 1 ? '1 entry' : `${formatNumber(history.length, 0)} entries`}</span>
      {status === 'ready' && history.length === 0 ? (
        hasQuery ? (
          NO_MATCHES
        ) : source === 'all' ? (
          EMPTY_HISTORY
        ) : (
          NOTHING_FROM_SOURCE
        )
      ) : (
        <div ref={scrollRef} className={styles.scroll}>
          <div role="list" aria-label="History" className={styles.list} style={{ height: virtualizer.getTotalSize() }}>
            {virtualizer.getVirtualItems().map((item) => {
              const row = rows[item.index]
              if (!row) {
                return null
              }
              return (
                <div
                  key={row.key}
                  role={row.kind === 'day' ? 'presentation' : 'listitem'}
                  className={styles.slot}
                  style={{ height: item.size, transform: `translateY(${item.start}px)` }}
                >
                  {row.kind === 'day' ? (
                    <h3 className={styles.day}>{row.label}</h3>
                  ) : row.entry.kind === 'action' ? (
                    <HistoryActionRow entry={row.entry} expanded={row.entry.id === expandedId} onToggle={toggle} />
                  ) : (
                    <HistoryEntryButton entry={row.entry} onOpen={onOpenEntry} />
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}
      <ClearHistoryDialog open={confirming} source={source} onCancel={() => setConfirming(false)} onConfirm={confirmClear} />
    </div>
  )
}
