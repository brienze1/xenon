import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useSidebarStore } from '../store'
import styles from './SidebarToolbar.module.css'

export function SidebarToolbar() {
  const view = useSidebarStore((state) => state.view)
  const query = useSidebarStore((state) => (state.view === 'history' ? state.historyQuery : state.query))
  const setQuery = useSidebarStore((state) => state.setQuery)
  const setHistoryQuery = useSidebarStore((state) => state.setHistoryQuery)
  const createCollection = useSidebarStore((state) => state.createCollection)
  const history = view === 'history'

  return (
    <div className={styles.toolbar}>
      {history ? null : (
        <IconButton label="New collection" className={styles.add} onClick={() => void createCollection()}>
          <span aria-hidden="true">+</span>
        </IconButton>
      )}
      <input
        type="search"
        name={history ? 'history-search' : 'request-search'}
        aria-label={history ? 'Search history' : 'Search requests'}
        placeholder={history ? 'Search URL or method…' : 'Search…'}
        autoComplete="off"
        className={styles.search}
        spellCheck={false}
        value={query}
        onChange={(event) => (history ? setHistoryQuery(event.target.value) : setQuery(event.target.value))}
      />
    </div>
  )
}
