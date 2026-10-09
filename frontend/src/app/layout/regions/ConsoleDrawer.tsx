import { useId, type ReactNode } from 'react'
import { consoleMaxHeight, LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { ResizeHandle } from '@/shared/ui/resize-handle/ResizeHandle'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './ConsoleDrawer.module.css'

interface ConsoleDrawerProps {
  children?: ReactNode
}

export function ConsoleDrawer({ children }: ConsoleDrawerProps) {
  const { layout, onLayoutChange } = useShellLayout()
  const headingId = useId()
  if (!layout.consoleVisible) {
    return null
  }
  return (
    <section aria-labelledby={headingId} className={styles.drawer}>
      <h2 id={headingId} className="visually-hidden">
        Console
      </h2>
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
