import type { ReactNode } from 'react'
import type { LayoutState } from '@/platform/mappers/ui-state'
import { cn } from '@/shared/lib/cn'
import { useShellLayout } from '@/app/layout/use-shell-layout'
import styles from './Footer.module.css'

type RegionToggle = 'sidebarVisible' | 'rightPanelVisible' | 'consoleVisible'

const REGION_TOGGLES: Array<{ key: RegionToggle; label: string; icon: string }> = [
  { key: 'sidebarVisible', label: 'Sidebar', icon: 'M2 3h12v10H2zM6 3v10' },
  { key: 'consoleVisible', label: 'Console', icon: 'M2 3h12v10H2zM4 7l2 2-2 2M8 11h3' },
  { key: 'rightPanelVisible', label: 'Side panel', icon: 'M2 3h12v10H2zM10 3v10' },
]

interface ToggleButtonProps {
  label: string
  icon: string
  pressed: boolean
  onToggle: () => void
}

function ToggleButton({ label, icon, pressed, onToggle }: ToggleButtonProps) {
  return (
    <button type="button" aria-pressed={pressed} title={label} className={cn(styles.toggle, pressed && styles.on)} onClick={onToggle}>
      <svg viewBox="0 0 16 16" aria-hidden="true" className={styles.icon}>
        <path d={icon} />
      </svg>
      <span className={styles.toggleLabel}>{label}</span>
    </button>
  )
}

interface FooterProps {
  children?: ReactNode
}

export function Footer({ children }: FooterProps) {
  const { layout, onLayoutChange } = useShellLayout()
  const sideBySide = layout.split === 'side-by-side'

  function toggle(key: RegionToggle) {
    onLayoutChange({ [key]: !layout[key] } as Partial<LayoutState>)
  }

  return (
    <footer className={styles.footer}>
      <div className={styles.status}>{children}</div>
      <div className={styles.toggles} role="group" aria-label="Layout">
        {REGION_TOGGLES.map((region) => (
          <ToggleButton key={region.key} label={region.label} icon={region.icon} pressed={layout[region.key]} onToggle={() => toggle(region.key)} />
        ))}
        <ToggleButton
          label="Side by side"
          icon="M2 3h12v10H2zM8 3v10"
          pressed={sideBySide}
          onToggle={() => onLayoutChange({ split: sideBySide ? 'stacked' : 'side-by-side' })}
        />
      </div>
    </footer>
  )
}
