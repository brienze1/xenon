import type { SidebarView } from '@/platform/mappers/ui-state'
import { Dropdown } from '@/shared/ui/dropdown/Dropdown'
import { useSidebarStore } from '@/features/sidebar/store'
import styles from './SidebarViewSelect.module.css'

const VIEW_LABELS: Record<SidebarView, string> = {
  collections: 'Collections',
  environments: 'Environments',
  history: 'History',
}

const VIEWS: readonly SidebarView[] = ['collections', 'environments', 'history']

const CHEVRON = (
  <span className={styles.chevron} aria-hidden="true">
    ▾
  </span>
)

/** Picks what the sidebar shows; a menu instead of tabs so more views fit the narrow sidebar. */
export function SidebarViewSelect() {
  const view = useSidebarStore((state) => state.view)
  const setView = useSidebarStore((state) => state.setView)

  return (
    <div className={styles.bar}>
      <Dropdown.Root>
        <Dropdown.Trigger label={`Sidebar view: ${VIEW_LABELS[view]}`} className={styles.trigger}>
          <span className={styles.label}>{VIEW_LABELS[view]}</span>
          {CHEVRON}
        </Dropdown.Trigger>
        <Dropdown.Menu label="Sidebar view">
          {VIEWS.map((option) => (
            <Dropdown.RadioItem key={option} checked={option === view} onSelect={() => setView(option)}>
              {VIEW_LABELS[option]}
            </Dropdown.RadioItem>
          ))}
        </Dropdown.Menu>
      </Dropdown.Root>
    </div>
  )
}
