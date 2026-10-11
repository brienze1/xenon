import { Tabs } from '@/shared/ui/tabs/Tabs'
import { SIDE_PANEL_VIEWS, type SidePanelView } from '@/features/side-panel/model'
import { CodePanel } from './CodePanel'
import { InfoPanel } from './InfoPanel'
import styles from './SidePanel.module.css'
import { VariablesPanel } from './VariablesPanel'

interface SidePanelProps {
  tabId: string
  view: string
  onViewChange: (view: SidePanelView) => void
}

function isSidePanelView(view: string): view is SidePanelView {
  return (SIDE_PANEL_VIEWS as readonly string[]).includes(view)
}

/** The right panel of a request: Variables, Code and Info views, the chosen one remembered per workspace. */
export function SidePanel({ tabId, view, onViewChange }: SidePanelProps) {
  const selected = isSidePanelView(view) ? view : 'info'
  return (
    <Tabs.Root value={selected} onValueChange={(next) => onViewChange(next as SidePanelView)} className={styles.root}>
      <Tabs.List label="Side panel views">
        <Tabs.Tab value="variables">Variables</Tabs.Tab>
        <Tabs.Tab value="code">Code</Tabs.Tab>
        <Tabs.Tab value="info">Info</Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="variables" className={styles.panel}>
        <VariablesPanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="code" className={styles.panel}>
        <CodePanel tabId={tabId} />
      </Tabs.Panel>
      <Tabs.Panel value="info" className={styles.panel}>
        <InfoPanel tabId={tabId} />
      </Tabs.Panel>
    </Tabs.Root>
  )
}
