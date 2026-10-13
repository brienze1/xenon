import { useVirtualizer } from '@tanstack/react-virtual'
import { useEffect, useMemo, useRef } from 'react'
import { visibleEntries } from '@/features/console/lib/filter'
import { useConsoleStore } from '@/features/console/store'
import { EmptyState } from '@/shared/ui/empty-state/EmptyState'
import { TerminalArt } from '@/shared/ui/empty-state/illustrations/TerminalArt'
import { ConsoleEntryDetail } from './ConsoleEntryDetail'
import { ConsoleEntryRow } from './ConsoleEntryRow'
import { ConsoleToolbar } from './ConsoleToolbar'
import styles from './ConsoleDrawer.module.css'

export const CONSOLE_ROW_HEIGHT = 26

const EMPTY_CONSOLE = (
  <EmptyState.Root label="Console is empty" size="compact">
    <EmptyState.Art>
      <TerminalArt />
    </EmptyState.Art>
    <EmptyState.Title>Nothing Logged Yet</EmptyState.Title>
    <EmptyState.Description>Requests you send and script logs show up here.</EmptyState.Description>
  </EmptyState.Root>
)

const NO_MATCHES = <p className={styles.none}>No entries match.</p>

export function ConsoleDrawer() {
  const entries = useConsoleStore((state) => state.entries)
  const filter = useConsoleStore((state) => state.filter)
  const search = useConsoleStore((state) => state.search)
  const selectedId = useConsoleStore((state) => state.selectedId)
  const load = useConsoleStore((state) => state.load)
  const select = useConsoleStore((state) => state.select)
  const shown = useMemo(() => visibleEntries(entries, filter, search), [entries, filter, search])
  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- the virtualizer is read during render only, never memoized
  const virtualizer = useVirtualizer({
    count: shown.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => CONSOLE_ROW_HEIGHT,
    overscan: 20,
  })

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    if (shown.length > 0) {
      virtualizer.scrollToIndex(shown.length - 1, { align: 'end' })
    }
  }, [shown.length, virtualizer])

  return (
    <div className={styles.console}>
      <ConsoleToolbar shown={shown.length} />
      <div className={styles.body}>
        {entries.length === 0 ? (
          EMPTY_CONSOLE
        ) : shown.length === 0 ? (
          NO_MATCHES
        ) : (
          <div ref={scrollRef} className={styles.scroll}>
            <div role="list" aria-label="Console entries" className={styles.list} style={{ height: virtualizer.getTotalSize() }}>
              {virtualizer.getVirtualItems().map((item) => {
                const entry = shown[item.index]
                return entry ? (
                  <div key={entry.id} role="listitem" className={styles.slot} style={{ height: item.size, transform: `translateY(${item.start}px)` }}>
                    <ConsoleEntryRow entry={entry} selected={entry.id === selectedId} onSelect={(id) => void select(id)} />
                  </div>
                ) : null
              })}
            </div>
          </div>
        )}
        <ConsoleEntryDetail />
      </div>
    </div>
  )
}
