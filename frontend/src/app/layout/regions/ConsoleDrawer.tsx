import type { ReactNode } from 'react'
import { consoleMaxHeight, LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { ResizeHandle } from '@/shared/ui/resize-handle/ResizeHandle'
import { useShellLayout } from '../shell-layout'
import styles from './ConsoleDrawer.module.css'

export function ConsoleDrawer({ children }: { children?: ReactNode }) {
  const { layout, onLayoutChange } = useShellLayout()
  if (!layout.consoleVisible) {
    return null
  }
  return (
    <section aria-label="Console" className={styles.drawer}>
      <ResizeHandle
        label="Resize console"
        orientation="horizontal"
        edge="start"
        value={layout.consoleHeight}
        min={LAYOUT_LIMITS.console.min}
        max={consoleMaxHeight(window.innerHeight)}
        onChange={(consoleHeight) => onLayoutChange({ consoleHeight })}
      />
      <div className={styles.content}>{children}</div>
    </section>
  )
}
