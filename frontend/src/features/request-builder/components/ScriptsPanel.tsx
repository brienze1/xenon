import type { ReactNode } from 'react'
import { PanelSection } from '@/shared/ui/panel-section/PanelSection'

const NO_SCRIPTS_EDITOR = (
  <PanelSection.Root label="Scripts">
    <PanelSection.Heading>Pre-Request and Post-Response Scripts</PanelSection.Heading>
    <PanelSection.Note>Scripts can be edited in the request tab.</PanelSection.Note>
  </PanelSection.Root>
)

interface ScriptsPanelProps {
  /** The scripts editor, composed by the app so the request builder stays unaware of it. */
  children?: ReactNode
}

export function ScriptsPanel({ children }: ScriptsPanelProps) {
  return children ?? NO_SCRIPTS_EDITOR
}
