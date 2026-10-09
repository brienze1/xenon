import type { CSSProperties, ReactNode } from 'react'
import type { LayoutState } from '@/platform/mappers/ui-state'
import styles from './AppShell.module.css'
import { MAIN_CONTENT_ID, ShellLayoutContext } from './shell-layout'

interface AppShellProps {
  layout: LayoutState
  onLayoutChange: (patch: Partial<LayoutState>) => void
  children: ReactNode
}

export function AppShell({ layout, onLayoutChange, children }: AppShellProps) {
  const sizes = {
    '--sidebar-width': `${layout.sidebarWidth}px`,
    '--right-panel-width': `${layout.rightPanelWidth}px`,
    '--console-height': `${layout.consoleHeight}px`,
  } as CSSProperties
  return (
    <ShellLayoutContext value={{ layout, onLayoutChange }}>
      <div className={styles.shell} style={sizes}>
        <a href={`#${MAIN_CONTENT_ID}`} className={styles.skipLink}>
          Skip to Main Content
        </a>
        <h1 className="visually-hidden" translate="no">
          Xenon
        </h1>
        {children}
      </div>
    </ShellLayoutContext>
  )
}
