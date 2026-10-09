import type { CSSProperties, ReactNode } from 'react'
import type { LayoutState } from '@/platform/mappers/ui-state'
import styles from './AppShell.module.css'
import { ShellLayoutContext } from './shell-layout'

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
      <div className={styles.shell} style={sizes} data-testid="app-shell">
        {children}
      </div>
    </ShellLayoutContext>
  )
}
