import { useId } from 'react'
import { detectPlatform, formatHotkey, type HotkeyPlatform } from '@/shared/hooks/use-hotkey'
import type { ShortcutEntry } from '@/features/settings/model'
import styles from './ShortcutsList.module.css'

interface ShortcutsListProps {
  shortcuts: readonly ShortcutEntry[]
  platform?: HotkeyPlatform
}

/** Every shortcut with its key combination for the current OS (remapping comes after the MVP). */
export function ShortcutsList({ shortcuts, platform = detectPlatform() }: ShortcutsListProps) {
  const headingId = useId()
  return (
    <section className={styles.section} aria-labelledby={headingId}>
      <h3 id={headingId} className={styles.heading}>
        Shortcuts
      </h3>
      <table className={styles.table}>
        <thead className="visually-hidden">
          <tr>
            <th scope="col">Action</th>
            <th scope="col">Keys</th>
          </tr>
        </thead>
        <tbody>
          {shortcuts.map((shortcut) => (
            <tr key={shortcut.label}>
              <th scope="row" className={styles.action}>
                {shortcut.label}
              </th>
              <td className={styles.keys}>
                {shortcut.combos.map((combo, index) => (
                  <span key={combo}>
                    {index > 0 ? <span className={styles.or}> or </span> : null}
                    <kbd className={styles.kbd} translate="no">
                      {formatHotkey(combo, platform)}
                    </kbd>
                  </span>
                ))}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className={styles.note}>
        Shortcuts can’t be changed yet. Some desktops keep a few combinations for themselves (for example Ctrl+Alt shortcuts on Linux desktops or Ctrl+Tab
        in some window managers); use the menus or the other combination when one doesn’t reach Xenon.
      </p>
    </section>
  )
}
