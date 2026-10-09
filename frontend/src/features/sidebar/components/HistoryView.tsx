import { useVirtualizer } from '@tanstack/react-virtual'
import { useMemo, useRef, useState } from 'react'
import type { HistoryEntry } from '@/platform/mappers/history'
import { cn } from '@/shared/lib/cn'
import { formatDuration } from '@/shared/lib/format-duration'
import { statusTone } from '@/shared/lib/status-tone'
import { Button } from '@/shared/ui/button/Button'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { MethodBadge } from '@/shared/ui/method-badge/MethodBadge'
import { historyRows, sentTime } from '@/features/sidebar/lib/history-rows'
import { useSidebarStore } from '@/features/sidebar/store'
import { ClearHistoryDialog } from './ClearHistoryDialog'
import styles from './HistoryView.module.css'

export const HISTORY_DAY_HEIGHT = 28
export const HISTORY_ENTRY_HEIGHT = 44

const EMPTY_HISTORY = (
  <EmptyState.Root label="No history" className={styles.empty}>
    <EmptyState.Art variant="history" />
    <EmptyState.Title>Nothing Sent Yet</EmptyState.Title>
    <EmptyState.Description>Every request you send in this workspace shows up here, newest first.</EmptyState.Description>
  </EmptyState.Root>
)

const NO_MATCHES = <p className={styles.noMatches}>No history entries match your search.</p>

interface HistoryViewProps {
  onOpenEntry: (id: string) => void
}

function outcome(entry: HistoryEntry): { label: string; tone: string } {
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
  return (
    <button
      type="button"
      className={styles.entry}
      aria-label={`${entry.method} ${entry.url || 'no URL'}, ${result.label}, ${time}`}
      onClick={() => onOpen(entry.id)}
    >
      <span className={styles.line}>
        <MethodBadge method={entry.method} className={styles.method} />
        <span className={styles.url} title={entry.url} translate="no">
          {entry.url || 'No URL'}
        </span>
      </span>
      <span className={styles.meta}>
        <span className={cn(styles.status, styles[result.tone])}>{result.label}</span>
        <span>{formatDuration(entry.totalMs)}</span>
        <time dateTime={entry.sentAt}>{time}</time>
      </span>
    </button>
  )
}

export function HistoryView({ onOpenEntry }: HistoryViewProps) {
  const history = useSidebarStore((state) => state.history)
  const status = useSidebarStore((state) => state.historyStatus)
  const query = useSidebarStore((state) => state.historyQuery)
  const clearHistory = useSidebarStore((state) => state.clearHistory)
  const [confirming, setConfirming] = useState(false)
  const rows = useMemo(() => historyRows(history, new Date()), [history])
  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: (index) => (rows[index]?.kind === 'day' ? HISTORY_DAY_HEIGHT : HISTORY_ENTRY_HEIGHT),
    overscan: 10,
  })

  async function confirmClear() {
    setConfirming(false)
    await clearHistory()
  }

  return (
    <div className={styles.view}>
      <div className={styles.header}>
        <span className={styles.count}>{history.length === 1 ? '1 entry' : `${history.length} entries`}</span>
        <Button variant="ghost" size="sm" disabled={history.length === 0} onClick={() => setConfirming(true)}>
          Clear History
        </Button>
      </div>
      {status === 'ready' && history.length === 0 ? (
        query.trim() === '' ? (
          EMPTY_HISTORY
        ) : (
          NO_MATCHES
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
                  ) : (
                    <HistoryEntryButton entry={row.entry} onOpen={onOpenEntry} />
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}
      <ClearHistoryDialog open={confirming} onCancel={() => setConfirming(false)} onConfirm={() => void confirmClear()} />
    </div>
  )
}
