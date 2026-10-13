import { useConsoleStore } from '@/features/console/store'
import { CONSOLE_FILTERS } from '@/features/console/lib/filter'
import { formatNumber } from '@/shared/lib/number-format'
import { Button } from '@/shared/ui/button/Button'
import styles from './ConsoleToolbar.module.css'

interface ConsoleToolbarProps {
  shown: number
}

export function ConsoleToolbar({ shown }: ConsoleToolbarProps) {
  const filter = useConsoleStore((state) => state.filter)
  const search = useConsoleStore((state) => state.search)
  const total = useConsoleStore((state) => state.entries.length)
  const setFilter = useConsoleStore((state) => state.setFilter)
  const setSearch = useConsoleStore((state) => state.setSearch)
  const clear = useConsoleStore((state) => state.clear)
  return (
    <div className={styles.toolbar}>
      <div role="group" aria-label="Show" className={styles.filters}>
        {CONSOLE_FILTERS.map((option) => (
          <button
            key={option.value}
            type="button"
            aria-pressed={filter === option.value}
            className={styles.filter}
            onClick={() => setFilter(option.value)}
          >
            {option.label}
          </button>
        ))}
      </div>
      <input
        type="search"
        aria-label="Search console"
        placeholder="Search console…"
        autoComplete="off"
        spellCheck={false}
        value={search}
        className={styles.search}
        onChange={(event) => setSearch(event.target.value)}
      />
      <span className={styles.count} aria-live="polite">
        {shown === total ? `${formatNumber(total, 0)} entries` : `${formatNumber(shown, 0)} of ${formatNumber(total, 0)}`}
      </span>
      <Button variant="ghost" size="sm" disabled={total === 0} onClick={() => void clear()}>
        Clear
      </Button>
    </div>
  )
}
