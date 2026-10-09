import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { useSidebarStore } from '../store'
import styles from './SidebarToolbar.module.css'

export function SidebarToolbar() {
  const query = useSidebarStore((state) => state.query)
  const setQuery = useSidebarStore((state) => state.setQuery)
  const createCollection = useSidebarStore((state) => state.createCollection)

  return (
    <div className={styles.toolbar}>
      <IconButton label="New collection" className={styles.add} onClick={() => void createCollection()}>
        <span aria-hidden="true">+</span>
      </IconButton>
      <input
        type="search"
        aria-label="Search requests"
        placeholder="Search"
        className={styles.search}
        spellCheck={false}
        value={query}
        onChange={(event) => setQuery(event.target.value)}
      />
    </div>
  )
}
