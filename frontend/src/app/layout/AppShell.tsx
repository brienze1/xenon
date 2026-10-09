import { useMemo, type CSSProperties, type ReactNode } from 'react'
import type { LayoutState } from '@/platform/mappers/ui-state'
import styles from './AppShell.module.css'
import { MAIN_CONTENT_ID } from './shell-layout'
import { ShellLayoutProvider } from './ShellLayoutProvider'

const SHELL_META = { mainContentId: MAIN_CONTENT_ID }

const SKIP_LINK = (
  <a href={`#${MAIN_CONTENT_ID}`} className={styles.skipLink}>
    Skip to Main Content
  </a>
)

const SHELL_HEADING = (
  <h1 className="visually-hidden" translate="no">
    Xenon
  </h1>
)

interface AppShellProps {
  layout: LayoutState
  onLayoutChange: (patch: Partial<LayoutState>) => void
  children: ReactNode
}

export function AppShell({ layout, onLayoutChange, children }: AppShellProps) {
  const state = useMemo(() => ({ layout }), [layout])
  const actions = useMemo(() => ({ update: onLayoutChange }), [onLayoutChange])
  const sizes = {
    '--sidebar-width': `${layout.sidebarWidth}px`,
    '--right-panel-width': `${layout.rightPanelWidth}px`,
    '--console-height': `${layout.consoleHeight}px`,
  } as CSSProperties
  return (
    <ShellLayoutProvider state={state} actions={actions} meta={SHELL_META}>
      <div className={styles.shell} style={sizes}>
        {SKIP_LINK}
        {SHELL_HEADING}
        {children}
      </div>
    </ShellLayoutProvider>
  )
}
