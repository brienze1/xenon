import type { ReactNode } from 'react'
import { LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { IconButton } from '@/shared/ui/icon-button/IconButton'
import { ResizeHandle } from '@/shared/ui/resize-handle/ResizeHandle'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './RightPanel.module.css'

// Lucide "panel-right-close": the panel outline with an arrow pointing out to the right.
const HIDE_ICON = (
  <svg viewBox="0 0 24 24" aria-hidden="true" className={styles.icon}>
    <rect width="18" height="18" x="3" y="3" rx="2" />
    <path d="M15 3v18" />
    <path d="m8 9 3 3-3 3" />
  </svg>
)

interface RightPanelProps {
  children?: ReactNode
}

export function RightPanel({ children }: RightPanelProps) {
  const {
    state: { layout },
    actions: { update },
  } = useShellLayout()
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
        onChange={(rightPanelWidth) => update({ rightPanelWidth })}
      />
      <IconButton label="Hide side panel" className={styles.hide} onClick={() => update({ rightPanelVisible: false })}>
        {HIDE_ICON}
      </IconButton>
      <div className={styles.content}>{children}</div>
    </aside>
  )
}
