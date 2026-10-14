import type { KeyboardEvent } from 'react'
import { HISTORY_SOURCE_FILTERS, type HistorySourceFilter as Source } from '@/platform/mappers/history'
import { cn } from '@/shared/lib/cn'
import { useSidebarStore } from '@/features/sidebar/store'
import styles from './HistorySourceFilter.module.css'

export const SOURCE_LABELS: Record<Source, string> = { all: 'All', user: 'User', ai: 'AI' }

/** All / User / AI: which entries the History view shows (and Clear History clears). */
export function HistorySourceFilter() {
  const source = useSidebarStore((state) => state.historySource)
  const setSource = useSidebarStore((state) => state.setHistorySource)

  /** Arrow keys move the choice, as in any radio group. */
  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const step = event.key === 'ArrowRight' || event.key === 'ArrowDown' ? 1 : event.key === 'ArrowLeft' || event.key === 'ArrowUp' ? -1 : 0
    if (step === 0) {
      return
    }
    event.preventDefault()
    const count = HISTORY_SOURCE_FILTERS.length
    const next = HISTORY_SOURCE_FILTERS[(HISTORY_SOURCE_FILTERS.indexOf(source) + step + count) % count]!
    setSource(next)
    event.currentTarget.querySelector<HTMLElement>(`[data-source="${next}"]`)?.focus()
  }

  return (
    <div role="radiogroup" aria-label="History source" className={styles.group} onKeyDown={onKeyDown}>
      {HISTORY_SOURCE_FILTERS.map((option) => (
        <button
          key={option}
          data-source={option}
          type="button"
          role="radio"
          aria-checked={source === option}
          tabIndex={source === option ? 0 : -1}
          className={cn(styles.option, source === option && styles.selected)}
          onClick={() => setSource(option)}
        >
          {SOURCE_LABELS[option]}
        </button>
      ))}
    </div>
  )
}
