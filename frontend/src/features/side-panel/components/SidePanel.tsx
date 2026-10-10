import { Tabs } from '@/shared/ui/tabs/Tabs'
import { InfoPanel } from './InfoPanel'
import styles from './SidePanel.module.css'
import { VariablesPanel } from './VariablesPanel'

export const SIDE_PANEL_VIEWS = ['variables', 'info'] as const

export type SidePanelView = (typeof SIDE_PANEL_VIEWS)[number]

interface SidePanelProps {
  tabId: string
  view: string
  onViewChange: (view: SidePanelView) => void
}

function isSidePanelView(view: string): view is SidePanelView {
  return (SIDE_PANEL_VIEWS as readonly string[]).includes(view)
}

/** The right panel of a request: Variables and Info views, the chosen one remembered per workspace. */
export function SidePanel({ tabId, view, onViewChange }: SidePanelProps) {
  const selected = isSidePanelView(view) ? view : 'info'
  return (
    <Tabs.Root value={selected} onValueChange={(next) => onViewChange(next as SidePanelView)} className={styles.root}>
      <Tabs.List label="Side panel views">
        <Tabs.Tab value="variables">Variables</Tabs.Tab>
        <Tabs.Tab value="info">Info</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="variables" className={styles.panel}>
        <VariablesPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="info" className={styles.panel}>
        <InfoPanel tabId={tabId} />
      </Tabs.Panel>
    </Tabs.Root>
  )
}
