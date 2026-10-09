import type { ReactNode } from 'react'
import { LAYOUT_LIMITS } from '@/features/workspaces/lib/layout'
import { ResizeHandle } from '@/shared/ui/resize-handle/ResizeHandle'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './Sidebar.module.css'

interface SidebarProps {
  children?: ReactNode
}

export function Sidebar({ children }: SidebarProps) {
  const { layout, onLayoutChange } = useShellLayout()
  if (!layout.sidebarVisible) {
    return null
  }
  return (
    <nav aria-label="Sidebar" className={styles.sidebar}>
      {children}
      <ResizeHandle
        label="Resize sidebar"
        orientation="vertical"
        edge="end"
        value={layout.sidebarWidth}
        min={LAYOUT_LIMITS.sidebar.min}
        max={LAYOUT_LIMITS.sidebar.max}
        onChange={(sidebarWidth) => onLayoutChange({ sidebarWidth })}
      />
    </nav>
  )
}
