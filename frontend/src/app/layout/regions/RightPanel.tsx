import type { ReactNode } from 'react'
import { LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { ResizeHandle } from '@/shared/ui/resize-handle/ResizeHandle'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './RightPanel.module.css'

interface RightPanelProps {
  children?: ReactNode
}

export function RightPanel({ children }: RightPanelProps) {
  const { layout, onLayoutChange } = useShellLayout()
  if (!layout.rightPanelVisible) {
    return null
  }
  return (
    <aside aria-label="Side panel" className={styles.panel}>
      <ResizeHandle
        label="Resize side panel"
        orientation="vertical"
        edge="start"
        value={layout.rightPanelWidth}
        min={LAYOUT_LIMITS.rightPanel.min}
        max={LAYOUT_LIMITS.rightPanel.max}
        onChange={(rightPanelWidth) => onLayoutChange({ rightPanelWidth })}
      />
      <div className={styles.content}>{children}</div>
    </aside>
  )
}
